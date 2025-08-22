package parallel

import (
	"fmt"
	"strconv"

	"github.com/0xSplits/workit/handler"
	"github.com/0xSplits/workit/registry"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/tracer"
)

type Config struct {
	// Han is the list of worker handlers implementing the actual business logic
	// as distinct execution pipelines. The worker handlers configured here may be
	// wrapped in administrative handler implementations to e.g. instrument
	// handler execution latency and handler error rates. All worker handlers
	// provided here will be executed concurrently within their own isolated
	// failure domain.
	Han []handler.Cooler

	// Log is a standard logger interface to forward structured log messages to
	// any output interface e.g. stdout.
	Log logger.Interface

	// Reg is the metrics interface used to wrap the internally managed handlers
	// for instrumentation purposes. The metrics handlers created by this registry
	// will record all worker handler execution metrics.
	Reg *registry.Registry
}

type Worker struct {
	han []handler.Interface
	log logger.Interface
	reg *registry.Registry
	rdy chan struct{}
}

func New(c Config) *Worker {
	if len(c.Han) == 0 {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Han must not be empty", c)))
	}
	if c.Log == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Log must not be empty", c)))
	}
	if c.Reg == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Reg must not be empty", c)))
	}

	// Verify early on that no handler leaf is ever nil.

	for i, x := range c.Han {
		if x == nil {
			tracer.Panic(tracer.Mask(fmt.Errorf("%T.Han[%d] must not be empty", c, i)))
		}
	}

	// Wrap the list of injected worker handlers into their own metrics handler,
	// so that we can instrument the runtime latency and error rates of every
	// single worker handler provided.

	var han []handler.Interface
	for _, x := range c.Han {
		han = append(han, c.Reg.New(x))
	}

	var rdy chan struct{}
	{
		rdy = make(chan struct{})
	}

	return &Worker{
		han: han,
		log: c.Log,
		reg: c.Reg,
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
		go w.ensure(h)
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

func (w *Worker) error(err error) {
	w.log.Log(
		"level", "error",
		"message", "worker execution failed",
		"stack", tracer.Json(err),
	)
}
