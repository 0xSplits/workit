package worker

import (
	"fmt"
	"strconv"
	"time"

	"github.com/0xSplits/workit/handler"
	"github.com/0xSplits/workit/handler/metrics"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/tracer"
	"go.opentelemetry.io/otel/metric"
)

type Config struct {
	// Env is the environment identifier injected to the internally managed
	// registry interface to annotate all metrics with the respective label, e.g.
	// "env=staging".
	Env string

	// Fil is an optional error matcher to ignore certain errors returned by the
	// executed worker handlers, in order to suppress their associated error logs.
	// All errors will be logged by default.
	Fil func(error) bool

	// Han is the list of worker handlers implementing the actual business logic.
	// The worker handlers configured here may be wrapped in administrative
	// handler implementations to e.g. instrument handler execution latency and
	// handler error rates. All worker handlers provided here will be executed
	// concurrently within their own isolated failure domain.
	Han []handler.Interface

	// Log is a standard logger interface to forward structured log messages to
	// any output interface e.g. stdout.
	Log logger.Interface

	// Met is the open telemetry meter interface injected to the internally
	// managed registry interface. This meter will record all worker handler
	// execution metrics.
	Met metric.Meter
}

type Worker struct {
	fil func(error) bool
	han []handler.Interface
	log logger.Interface
	rdy chan struct{}
}

func New(c Config) *Worker {
	if c.Env == "" {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Env must not be empty", c)))
	}
	if c.Fil == nil {
		c.Fil = func(_ error) bool { return false }
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

	// Wrap the list of injected worker handlers into their own metrics handler,
	// so that we can instrument the runtime latency and error rates of every
	// single worker handler provided.

	var han []handler.Interface
	for _, x := range c.Han {
		han = append(han, metrics.New(metrics.Config{
			Env: c.Env,
			Fil: c.Fil,
			Han: x,
			Log: c.Log,
			Met: c.Met,
		}))
	}

	var rdy chan struct{}
	{
		rdy = make(chan struct{})
	}

	return &Worker{
		fil: c.Fil,
		han: han,
		log: c.Log,
		rdy: rdy,
	}
}

func (w *Worker) Daemon() {
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

	// Signal the worker engine's readiness by closing the internal ready channel.
	// This mechanism implies that Worker.Daemon() must never be called twice,
	// because closing a closed channel results in a runtime panic. Time based
	// systems are often a source of race conditions. Providing this mechanism may
	// help facilitate e.g. unit tests concerned with concurrency patterns, so
	// that we do not have to rely on time based systems within event driven
	// problem domains.

	{
		close(w.rdy)
	}

	// Once the static worker pool created all necessary goroutines, we block
	// Worker.Daemon forever as a long running process, so that we do not risk
	// terminating the goroutines that we just bootstrapped.

	{
		select {}
	}
}

func (w *Worker) daemon(han handler.Interface) {
	for {
		// Execute the worker handler and log any runtime error of this handler's
		// business logic if the configured error matcher permits it. Note that any
		// error caught here may never originate from the worker engine's internal
		// metric registry.

		err := han.Ensure()
		if err != nil && !w.fil(err) {
			w.error(tracer.Mask(err, tracer.Context{Key: "handler", Value: handler.Name(han)}))
		}

		// Sleep for the given duration after this worker handler has been executed.
		// This specific cycle repeats again for the given worker handler only,
		// after the sleep below is over.

		{
			time.Sleep(han.Cooler())
		}
	}
}

func (w *Worker) error(err error) {
	w.log.Log(
		"level", "error",
		"message", "worker execution failed",
		"stack", tracer.Json(err),
	)
}
