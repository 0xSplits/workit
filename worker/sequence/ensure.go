package sequence

import (
	"github.com/0xSplits/workit/handler"
	"github.com/xh3b4sd/tracer"
	"golang.org/x/sync/errgroup"
)

func (g *Worker) ensure() error {
	for _, x := range g.han {
		var err error

		if len(x) == 1 {
			err = g.ensSeq(x)
		} else {
			err = g.ensPar(x)
		}

		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (g *Worker) ensPar(han []handler.Interface) error {
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

func (g *Worker) ensSeq(han []handler.Interface) error {
	var x handler.Interface
	{
		x = han[0]
	}

	// Note that our worker handlers may be wrapped. So we have to call unwrap
	// before resolving the implementation's identifier in the error case.

	err := x.Ensure()
	if err != nil {
		return tracer.Mask(err, tracer.Context{Key: "handler", Value: handler.Name(x.Unwrap())})
	}

	return nil
}
