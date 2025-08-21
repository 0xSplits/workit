# workit

This package provides some shared code for our Golang microservice background
jobs. The project name `workit` means as much as "worker kit". The main concept
here are **worker handlers**, which are executed by soecialized worker engines.

- [\*parallel.Worker](./worker/parallel/worker.go) implements concurrent execution within isolated failure domains
- [\*sequence.Worker](./worker/sequence/worker.go) implements sequential execution of a directed acyclic graph

```golang
// Interface describes the internally wrapped worker handlers for proper
// management inside of the various worker engines.
type Interface interface {
	// Cooler is the amount of time that any given handler specifies to wait
	// before being executed again. This is not an interval on a strict schedule.
	// This is simply the time to sleep after execution, before another cycle
	// repeats.
	Cooler() time.Duration

	// Ensure executes the handler specific business logic in order to complete
	// the given task, if possible. Any error returned will be emitted using the
	// underlying logger interface, unless the injected metrics registry is
	// configured to filter the received error.
	Ensure() error

	// Unwrap returns the underlying worker handler implementation, if any.
	Unwrap() Ensure
}
```
