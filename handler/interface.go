package handler

import "time"

// Interface describes asynchronous worker handlers that are executed
// iteratively by a custom task engine. New worker handlers can be created by
// implementing this handler interface and registering the handler instance with
// the respective worker engine.
type Interface interface {
	// Cooler is the amount of time that any given handler specifies to wait
	// before being executed again. This is not an interval on a strict schedule.
	// This is simply the time to sleep after execution, before another cycle
	// repeats.
	Cooler() time.Duration

	// Ensure executes the handler specific business logic in order to complete
	// the given task, if possible. Any error returned will be emitted using the
	// underlying logger interface. Calling this method will not interfere with
	// the execution of other handlers, because every handler is managed within
	// its own pipeline to guarantee isolated failure domains.
	Ensure() error
}
