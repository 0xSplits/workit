package metrics

import "time"

func (m *Metrics) Cooler() time.Duration {
	return m.han.Cooler()
}
