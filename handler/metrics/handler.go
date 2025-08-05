package metrics

import (
	"fmt"

	"github.com/0xSplits/otelgo/registry"
	"github.com/0xSplits/workit/handler"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/tracer"
	"go.opentelemetry.io/otel/metric"
)

const (
	MetricTotal    = "worker_handler_execution_total"
	MetricDuration = "worker_handler_execution_duration_seconds"
)

type Config struct {
	Env string
	Fil func(error) bool
	Han handler.Interface
	Log logger.Interface
	Met metric.Meter
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
	if c.Met == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Met must not be empty", c)))
	}

	var nam string
	{
		nam = handler.Name(c.Han)
	}

	var reg registry.Interface
	{
		reg = newRegistry(c.Env, c.Log, c.Met, nam)
	}

	return &Metrics{
		fil: c.Fil,
		han: c.Han,
		log: c.Log,
		nam: nam,
		reg: reg,
	}
}
