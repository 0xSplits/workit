package handler

import (
	"fmt"
	"time"

	"github.com/xh3b4sd/tracer"
)

type Config struct {
	// Coo is the cooler duration for this handler instance, transparently called
	// with Handler.Cooler without any other modification.
	Coo time.Duration

	// Ens is the ensure function for this handler instance, transparently called
	// with Handler.Ensure without any other modification.
	Ens func() error
}

type Handler struct {
	coo time.Duration
	ens func() error
}

func New(c Config) *Handler {
	if c.Coo == 0 {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Coo must not be empty", c)))
	}
	if c.Ens == nil {
		tracer.Panic(tracer.Mask(fmt.Errorf("%T.Ens must not be empty", c)))
	}

	return &Handler{
		coo: c.Coo,
		ens: c.Ens,
	}
}
