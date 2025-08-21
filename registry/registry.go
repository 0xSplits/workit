package registry

import (
	"fmt"

	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/tracer"
	"go.opentelemetry.io/otel/metric"
)

type Config struct {
	// Env is the environment identifier injected to the internally managed
	// registry interface to annotate all metrics with the respective label, e.g.
	// "env=staging".
	Env string

	// Fil is an optional error matcher to ignore certain errors returned by the
	// executed worker handlers, in order to suppress their associated error logs.
	// All errors will be logged by default.
	Fil func(error) bool

	// Log is a standard logger interface to forward structured log messages to
	// any output interface e.g. stdout.
	Log logger.Interface

	// Met is the open telemetry meter interface injected to the internally
	// managed registry interface. This meter will record all worker handler
	// execution metrics.
	Met metric.Meter
}

type Registry struct {
	env string
	fil func(error) bool
	log logger.Interface
	met metric.Meter
}

func New(c Config) *Registry {
	if c.Env == "" {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Env must not be empty", c)))
	}
	if c.Fil == nil {
		c.Fil = func(_ error) bool { return false }
	}
	if c.Log == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Log must not be empty", c)))
	}
	if c.Met == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Met must not be empty", c)))
	}

	return &Registry{
		env: c.Env,
		fil: c.Fil,
		log: c.Log,
		met: c.Met,
	}
}
