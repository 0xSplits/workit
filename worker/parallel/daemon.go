package parallel

import (
	"strconv"

	"github.com/xh3b4sd/tracer"
)

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
