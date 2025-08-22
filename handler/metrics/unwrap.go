package metrics

import "github.com/0xSplits/workit/handler"

// Unwrap only forwards the unwrap of the wrapped handler implementation. That
// means the metrics handler does not have its own unwrap behaviour, but only
// acts as proxy for the underlying handler.
func (m *Metrics) Unwrap() handler.Ensure {
	return m.han.Unwrap()
}
