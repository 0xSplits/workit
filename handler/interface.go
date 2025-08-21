package handler

import "time"

// Interface describes the internally wrapped worker handlers for proper
// management inside of the various worker engines.
type Interface interface {
	Cooler
	Ensure
	Unwrap
}

type Cooler interface {
	// Cooler is the amount of time that any given handler specifies to wait
	// before being executed again. This is not an interval on a strict schedule.
	// This is simply the time to sleep after execution, before another cycle
	// repeats.
	Cooler() time.Duration
	Ensure
}

type Ensure interface {
	// Ensure executes the handler specific business logic in order to complete
	// the given task, if possible. Any error returned will be emitted using the
	// underlying logger interface, unless the injected metrics registry is
	// configured to filter the received error.
	Ensure() error
}

type Unwrap interface {
	// Unwrap returns the underlying worker handler implementation, if any.
	Unwrap() Ensure
}
