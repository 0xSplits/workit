package sequence

import (
	"time"

	"github.com/xh3b4sd/tracer"
)

func (w *Worker) Daemon() {
	w.log.Log(
		"level", "info",
		"message", "worker is executing tasks",
		"pipelines", "1",
	)

	for {
		// Execute the configured worker handler and log any runtime error of this
		// handler's business logic. Any error returned to us may be annotated with
		// the underlying handler name.

		err := w.Ensure()
		if err != nil && !w.reg.Log(err) {
			w.error(tracer.Mask(err))
		}

		// Sleep for the given cooler duration after the configured worker handler
		// has been executed. Not all worker handlers implement the handler.Cooler
		// interface, so in case we receive no cooler duration at all, we do not
		// invoke the runtime sleep functions. If we were to call time.Sleep with 0,
		// then we would invoke the Golang scheduler, which is simply not necessary.

		if w.coo > 0 {
			time.Sleep(w.coo)
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
