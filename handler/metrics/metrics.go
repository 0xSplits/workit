package metrics

import (
	"fmt"

	"github.com/0xSplits/otelgo/registry"
	"github.com/0xSplits/workit/handler"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/tracer"
)

const (
	MetricTotal    = "worker_handler_execution_total"
	MetricDuration = "worker_handler_execution_duration_seconds"
)

type Config struct {
	Fil func(error) bool
	Han handler.Interface
	Log logger.Interface
	Nam string
	Reg registry.Interface
}

type Metrics struct {
	fil func(error) bool
	han handler.Interface
	log logger.Interface
	nam string
	reg registry.Interface
}

func New(c Config) *Metrics {
	if c.Fil == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Fil must not be empty", c)))
	}
	if c.Han == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Han must not be empty", c)))
	}
	if c.Log == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Log must not be empty", c)))
	}
	if c.Nam == "" {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Nam must not be empty", c)))
	}
	if c.Reg == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%TReg must not be empty", c)))
	}

	return &Metrics{
		fil: c.Fil,
		han: c.Han,
		log: c.Log,
		nam: c.Nam,
		reg: c.Reg,
	}
}
