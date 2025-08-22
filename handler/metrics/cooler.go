package metrics

import "time"

// Cooler only forwards the cooler of the wrapped handler implementation. That
// means the metrics handler does not have its own cooler setting, but only acts
// as proxy for the underlying handler.
func (m *Metrics) Cooler() time.Duration {
	return m.han.Cooler()
}
