package sequence

import (
	"github.com/0xSplits/workit/handler"
	"github.com/xh3b4sd/tracer"
	"golang.org/x/sync/errgroup"
)

// Ensure executes a single reconciliation loop of the directed acyclic graph.
// This method is exposed publicly so that not only the worker daemon can run
// this sequence of worker handlers continuously, but also to enable other
// commands to run this sequence once in a controlled fashion.
func (w *Worker) Ensure() error {
	for _, x := range w.han {
		var err error

		if len(x) == 1 {
			err = w.ensSeq(x) // execute a single worker handler
		} else {
			err = w.ensPar(x) // execute all worker handlers concurrently
		}

		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (w *Worker) ensPar(han []handler.Interface) error {
	var grp errgroup.Group
	{
		grp = errgroup.Group{}
	}

	// Bootstrap a static worker pool of N goroutines, where N is the number of
	// injected worker handlers for this iteration. This parallel execution
	// isolates handler specific failure domains. Each handler is executed along
	// its own pipeline so that any handler specific runtime errors and execution
	// delays cannot affect the execution of the other worker handlers during this
	// iteration.

	for _, x := range han {
		// Continue with the next worker handler without doing any work for this
		// specific worker handler if this worker handler declares itself as not
		// active for this reconciliation loop.

		if !x.Active() {
			continue
		}

		grp.Go(func() error {
			// Note that our worker handlers may be wrapped. So we have to call unwrap
			// before resolving the implementation's identifier in the error case.

			err := x.Ensure()
			if err != nil {
				return tracer.Mask(err, tracer.Context{Key: "handler", Value: handler.Name(x.Unwrap())})
			}

			return nil
		})
	}

	{
		err := grp.Wait()
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (w *Worker) ensSeq(han []handler.Interface) error {
	var x handler.Interface
	{
		x = han[0] // the factory at sequence.New must validate against empty steps
	}

	// Return early without doing any work if this worker handler declares itself
	// as not active for this reconciliation loop.

	if !x.Active() {
		return nil
	}

	// Note that our worker handlers may be wrapped. So we have to call unwrap
	// before resolving the implementation's identifier in the error case.

	err := x.Ensure()
	if err != nil {
		return tracer.Mask(err, tracer.Context{Key: "handler", Value: handler.Name(x.Unwrap())})
	}

	return nil
}
