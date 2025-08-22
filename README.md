# workit

This package provides some shared code for our Golang microservice background
jobs. The project name `workit` means as much as "worker kit". The main concept
here are **worker handlers**, which are executed by soecialized worker engines.

- [\*parallel.Worker](./worker/parallel/worker.go) implements concurrent execution within isolated failure domains
- [\*sequence.Worker](./worker/sequence/worker.go) implements sequential execution of a directed acyclic graph

```golang
// Interface describes the internally wrapped worker handlers used for proper
// management inside of the various worker engines. External users do not have
// to be concerned with this interface.
type Interface interface {
	// Cooler is manadatory to be implemented for worker handlers executed by the
	// *parallel.Worker engine, because those worker handlers do all run inside
	// their own isolated failure domains, which require individual cooler durations
	// to be provided. Cooler is irrelevant for the worker handlers executed by the
	// *sequence.Worker engine, because those handlers run inside a single pipeline
	// with a cooler duration congiured on the engine level.
	//
	// Cooler is the amount of time that any given handler specifies to wait
	// before being executed again. This is not an interval on a strict schedule.
	// This is simply the time to sleep after execution, before another cycle
	// repeats.
	Cooler() time.Duration

	// Ensure is the minimal worker handler interface that all users have to
	// implement for their own business logic, regardless of the underlying worker
	// engine.
	//
	// Ensure executes the handler specific business logic in order to complete
	// the given task, if possible. Any error returned will be emitted using the
	// underlying logger interface, unless the injected metrics registry is
	// configured to filter the received error.
	Ensure() error

	// Unwrap is an administrative interface that is most useful for our internal
	// wrapper handlers, e.g. metrics and proxy. Most users do not have to worry
	// about this.
	//
	// Unwrap returns the underlying worker handler implementation, if any.
	Unwrap() Ensure
}
```
