package metrics

import "github.com/0xSplits/workit/handler"

func (m *Metrics) Unwrap() handler.Ensure {
	return m.han.Unwrap()
}
