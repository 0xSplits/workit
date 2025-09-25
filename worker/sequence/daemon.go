package sequence

import (
	"github.com/xh3b4sd/choreo/ticker"
)

func (w *Worker) Daemon() {
	// Explicitly disable Worker.Daemon if no cooler duration was provided. This
	// turns Worker.Daemon into a noop without blocking and side effects.

	if _, typ := w.tic.(ticker.Fake); typ {
		return
	}

	w.log.Log(
		"level", "info",
		"message", "worker is executing tasks",
		"pipelines", "1",
	)

	// Run Worker.Ensure once initially and rely on the underlying ticker
	// implementation to further trigger scheduled execution.

	{
		w.ensure()
	}

	// Execute Worker.Ensure based on the internally managed ticker
	// implementation. Note that the delivered ticks are synchronized with the
	// actual execution of Worker.Ensure, so that external calls reset the
	// effective wait duration.

	for range w.tic.Ticks() {
		w.ensure()
	}
}
