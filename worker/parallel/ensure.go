package parallel

import (
	"time"

	"github.com/0xSplits/workit/handler"
	"github.com/xh3b4sd/tracer"
)

func (w *Worker) ensure(han handler.Interface) {
	for {
		// Execute the worker handler and log any runtime error of this handler's
		// business logic if the configured error matcher permits it. Note that any
		// error caught here may never originate from the worker engine's internal
		// metric registry.

		err := han.Ensure()
		if err != nil && !w.reg.Log(err) {
			w.error(tracer.Mask(err, tracer.Context{Key: "handler", Value: handler.Name(han.Unwrap())}))
		}

		// Sleep for the given duration after this worker handler has been executed.
		// This specific cycle repeats again for the given worker handler only,
		// after the sleep below is over.

		{
			time.Sleep(han.Cooler())
		}
	}
}
