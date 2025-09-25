package sequence

import (
	"fmt"
	"time"

	"github.com/0xSplits/workit/handler"
	"github.com/0xSplits/workit/registry"
	"github.com/xh3b4sd/choreo/ticker"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/tracer"
)

type Config struct {
	// Coo is the optional amount of time that this sequence worker engine
	// specifies to wait before being executed again. This cooler duration is not
	// an interval on a strict schedule. This is simply the time to sleep after
	// execution, before another cycle repeats. Note that Worker.Daemon is
	// explicitly disabled if Coo is not provided. Regardless, Worker.Ensure may
	// be used on demand even without specified cooler duration.
	Coo time.Duration

	// Han is the list of worker handlers implementing the actual business logic
	// as a directed acyclic graph. The worker handlers configured here may be
	// wrapped in administrative handler implementations to e.g. instrument
	// handler execution latency and handler error rates. All worker handlers
	// provided here will be executed sequentially within the same failure domain.
	Han [][]handler.Ensure

	// Log is a standard logger interface to forward structured log messages to
	// any output interface e.g. stdout.
	Log logger.Interface

	// Reg is the metrics interface used to wrap the internally managed handlers
	// for instrumentation purposes. The metrics handlers created by this registry
	// will record all worker handler execution metrics.
	Reg *registry.Registry
}

type Worker struct {
	han [][]handler.Interface
	log logger.Interface
	reg *registry.Registry
	tic ticker.Interface
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

	// Verify early on that no handler slice is empty and that no handler leaf is
	// ever nil.

	for i, x := range c.Han {
		if len(x) == 0 {
			tracer.Panic(tracer.Mask(fmt.Errorf("%T.Han[%d] must not be empty", c, i)))
		}

		for j, y := range x {
			if y == nil {
				tracer.Panic(tracer.Mask(fmt.Errorf("%T.Han[%d][%d] must not be empty", c, i, j)))
			}
		}
	}

	// Wrap the list of injected worker handlers into their own metrics handler,
	// so that we can instrument the underlying handler interfaces.

	var han [][]handler.Interface
	for _, x := range c.Han {
		var row []handler.Interface

		for _, y := range x {
			row = append(row, c.Reg.New(y))
		}

		{
			han = append(han, row)
		}
	}

	// Allocate a real or fake ticker based on the injected cooler duration, so
	// that Worker.Ensure may be used without the need for Worker.Daemon.

	var tic ticker.Interface
	if c.Coo > 0 {
		tic = ticker.New(ticker.Config{Dur: c.Coo})
	} else {
		tic = ticker.Fake{}
	}

	return &Worker{
		han: han,
		log: c.Log,
		reg: c.Reg,
		tic: tic,
	}
}
