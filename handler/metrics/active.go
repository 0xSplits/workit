package metrics

// Active only forwards the scheduler primitive of the wrapped handler
// implementation. That means the metrics handler does not have its own
// activation setting, but only acts as proxy for the underlying handler.
func (m *Metrics) Active() bool {
	return m.han.Active()
}
