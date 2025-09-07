package combined

import (
	"fmt"

	"github.com/0xSplits/workit/worker/parallel"
	"github.com/0xSplits/workit/worker/sequence"
	"github.com/xh3b4sd/tracer"
)

type Config struct {
	Par *parallel.Worker
	Seq *sequence.Worker
}

// Worker is a simple wrapper that combines the different worker engines in a
// single daemon interface.
type Worker struct {
	par *parallel.Worker
	seq *sequence.Worker
}

func New(c Config) *Worker {
	if c.Par == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Par must not be empty", c)))
	}
	if c.Seq == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Seq must not be empty", c)))
	}

	return &Worker{
		par: c.Par,
		seq: c.Seq,
	}
}
