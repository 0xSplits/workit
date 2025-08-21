package proxy

import (
	"fmt"

	"github.com/0xSplits/workit/handler"
	"github.com/xh3b4sd/tracer"
)

type Config struct {
	Han handler.Ensure
}

// Proxy is a handler implementation to resolve optional implementations of
// handler.Cooler and handler.Unwrap. A worker engine may wrap its configured
// worker handlers within this proxy implementation in order to cover interface
// requirements for cases where those functions may not be implemented by the
// wrapped handlers.
type Proxy struct {
	han handler.Ensure
}

func New(c Config) *Proxy {
	if c.Han == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Han must not be empty", c)))
	}

	return &Proxy{
		han: c.Han,
	}
}
