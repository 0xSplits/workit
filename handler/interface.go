package handler

import "time"

// Interface describes the internally wrapped worker handlers used for proper
// management inside of the various worker engines. External users do usually
// not have to be concerned with this entire interface.
type Interface interface {
	Cooler
	Ensure
	Unwrap
}

// Active is a scheduler primitive that allows worker handlers to be disabled
// for arbitrary conditions.
type Active interface {
	// Active returns the underlying worker handler implementation is supposed to
	// be executed during reconciliation. This might be useful to disable certain
	// worker handlers for e.g. specific environments.
	Active() bool
}

// Cooler is manadatory to be implemented for worker handlers executed by the
// *parallel.Worker engine, because those worker handlers do all run inside
// their own isolated failure domains, which require individual cooler durations
// to be provided. Cooler is irrelevant for the worker handlers executed by the
// *sequence.Worker engine, because those handlers run inside a single pipeline
// with a cooler duration congiured on the engine level.
type Cooler interface {
	// Cooler is the amount of time that any given handler specifies to wait
	// before being executed again. This is not an interval on a strict schedule.
	// This is simply the time to sleep after execution, before another cycle
	// repeats.
	Cooler() time.Duration
	Ensure
}

// Ensure is the minimal worker handler interface that all users have to
// implement for their own business logic, regardless of the underlying worker
// engine.
type Ensure interface {
	Active
	// Ensure executes the handler specific business logic in order to complete
	// the given task, if possible. Any error returned will be emitted using the
	// underlying logger interface, unless the injected metrics registry is
	// configured to filter the received error.
	Ensure() error
}

// Unwrap is an administrative interface that is most useful for our internal
// wrapper handlers, e.g. metrics and proxy. Most users do not have to worry
// about this.
type Unwrap interface {
	// Unwrap returns the underlying worker handler implementation, if any.
	Unwrap() Ensure
}
