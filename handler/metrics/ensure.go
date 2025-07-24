package metrics

import (
	"strconv"
	"time"

	"github.com/xh3b4sd/tracer"
)

// Ensure tracks the start time of its own execution and runs the business logic
// of the wrapped worker handler. The wrapped business logic is instrumented for
// runtime latency and error rates. Note that Ensure emits debug logs about the
// internal worker handler execution. Any error returned originates from the
// underlying handler implementation.
func (m *Metrics) Ensure() error {
	// Record the start time for our handler latency. The timezone of the duration
	// measurement is irrelavant here, so we are not using time.Now().UTC() as a
	// best practice like we would in other places.

	var sta time.Time
	{
		sta = time.Now()
	}

	// Note that we cannot return the error from the handler execution, because we
	// want to monitor the failure latency as well, if possible. So instead of
	// returning the error early during the error case, we simply log the error
	// and continue below.

	var err error
	{
		err = m.han.Ensure()
	}

	// Record the handler latency immediately after the handler execution. The
	// function call below must instrument the given handler latency or panic,
	// which may only happen in case of registry whitelist failures. If this
	// instrumentation succeeded once, it may never fail again during runtime.

	{
		m.musIns(sta, err)
	}

	return tracer.Mask(err)
}

func (m *Metrics) musIns(sta time.Time, err error) {
	var lat time.Duration
	var suc string
	{
		lat = time.Since(sta)
		suc = strconv.FormatBool(err == nil)
	}

	m.log.Log(
		"level", "debug",
		"message", "executed worker handler",
		"handler", m.nam,
		"latency", lat.String(),
		"success", suc,
	)

	lab := map[string]string{
		"handler": m.nam,
		"success": suc,
	}

	err = m.reg.Counter(MetricTotal, 1, lab)
	if err != nil {
		tracer.Panic(tracer.Mask(err))
	}

	err = m.reg.Histogram(MetricDuration, lat.Seconds(), lab)
	if err != nil {
		tracer.Panic(tracer.Mask(err))
	}
}
