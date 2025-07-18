package workit

import (
	"fmt"
	"strconv"
	"time"

	"github.com/0xSplits/otelgo/recorder"
	"github.com/0xSplits/otelgo/registry"
	"github.com/0xSplits/workit/handler"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/tracer"
	"go.opentelemetry.io/otel/metric"
)

const (
	MetricTotal    = "worker_handler_execution_total"
	MetricDuration = "worker_handler_execution_duration_seconds"
)

type Config struct {
	Env string
	// Han are the worker specific handlers implementing the actual business
	// logic.
	Han []handler.Interface
	Log logger.Interface
	Met metric.Meter
}

type Engine struct {
	han []handler.Interface
	log logger.Interface
	rdy chan struct{}
	reg registry.Interface
}

func New(c Config) *Engine {
	if c.Env == "" {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Env must not be empty", c)))
	}
	if len(c.Han) == 0 {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Han must not be empty", c)))
	}
	if c.Log == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Log must not be empty", c)))
	}
	if c.Met == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Met must not be empty", c)))
	}

	cou := map[string]recorder.Interface{}

	{
		cou[MetricTotal] = recorder.NewCounter(recorder.CounterConfig{
			Des: "the total amount of worker handler executions",
			Lab: map[string][]string{
				"success": {"true", "false"},
			},
			Met: c.Met,
			Nam: MetricTotal,
		})
	}

	gau := map[string]recorder.Interface{}

	his := map[string]recorder.Interface{}

	{
		his[MetricDuration] = recorder.NewHistogram(recorder.HistogramConfig{
			Des: "the time it takes for worker handler executions to complete",
			Lab: map[string][]string{
				"handler": handler.Names(c.Han),
				"success": {"true", "false"},
			},
			Buc: []float64{
				0.10, //  100 ms
				0.15, //  150 ms
				0.20, //  200 ms
				0.25, //  250 ms
				0.50, //  500 ms

				1.00, // 1000 ms
				1.50, // 1500 ms
				2.00, // 2000 ms
				2.50, // 2500 ms
				5.00, // 5000 ms
			},
			Met: c.Met,
			Nam: MetricDuration,
		})
	}

	var rdy chan struct{}
	{
		rdy = make(chan struct{})
	}

	var reg registry.Interface
	{
		reg = registry.New(registry.Config{
			Env: c.Env,
			Log: c.Log,

			Cou: cou,
			Gau: gau,
			His: his,
		})
	}

	return &Engine{
		han: c.Han,
		log: c.Log,
		rdy: rdy,
		reg: reg,
	}
}

func (w *Engine) Daemon() {
	w.log.Log(
		"level", "info",
		"message", "worker is executing tasks",
		"pipelines", strconv.Itoa(len(w.han)),
	)

	// Bootstrap a static worker pool of N goroutines, where N is the number of
	// injected worker handlers. This parallel execution isolates worker specific
	// failure domains. Each handler is executed along its own pipeline so that
	// any handler specific runtime errors and execution delays cannot affect the
	// execution of the other worker handlers.

	for _, h := range w.han {
		go w.daemon(h)
	}

	// Signal the worker daemon's readiness by closing the internal ready channel.
	// This mechanism implies that Engine.Daemon() must never be called twice,
	// because closing a closed channel results in a runtime panic. Time based
	// systems are often a source of race conditions. Providing this mechanism may
	// help facilitate e.g. unit tests concerned with concurrency patterns, so
	// that we do not have to rely on time based systems within event driven
	// problem domains.

	{
		close(w.rdy)
	}

	// Once the static worker pool created all necessary goroutines, we block
	// Engine.Daemon forever as a long running process, so that we do not risk
	// terminating the goroutines that we just bootstrapped.

	{
		select {}
	}
}

func (w *Engine) daemon(han handler.Interface) {
	for {
		err := w.ensure(han)
		if err != nil {
			w.error(err)
		}

		// Sleep for the given duration after this worker handler has been executed.
		// This specific cycle repeats again for the given worker handler only,
		// after the sleep below is over.

		{
			time.Sleep(han.Cooler())
		}
	}
}

func (w *Engine) ensure(han handler.Interface) error {
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
		err = han.Ensure()
		if err != nil {
			w.error(err)
		}
	}

	// Record the handler latency immediately after the handler execution. Here as
	// well, time.Since() does not rely on a specific timezone, so we can simply
	// use the time.Now() instance of this cycle's start time.

	var lat time.Duration
	var suc string
	{
		lat = time.Since(sta)
		suc = strconv.FormatBool(err == nil)
	}

	w.log.Log(
		"level", "debug",
		"message", "executed worker handler",
		"handler", handler.Name(han),
		"latency", lat.String(),
		"success", suc,
	)

	{
		lab := map[string]string{
			"success": suc,
		}

		err := w.reg.Counter(MetricTotal, 1, lab)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	{
		lab := map[string]string{
			"handler": handler.Name(han),
			"success": suc,
		}

		err := w.reg.Histogram(MetricDuration, lat.Seconds(), lab)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (w *Engine) error(err error) {
	w.log.Log(
		"level", "error",
		"message", err.Error(),
		"stack", tracer.Json(err),
	)
}
