package parallel

import (
	"fmt"

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
