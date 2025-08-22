package registry

// Log returns the result of the configured filter function transparently.
// Exposing the injected filter here allows any worker handler to understand the
// same behaviour implemented inside of the metrics handler. E.g. an error
// intended to cancel the reconciliation loop should neither be considered a
// failure inside the metrics handler, as well as inside the worker engine.
func (r *Registry) Log(err error) bool {
	return r.fil(err)
}
