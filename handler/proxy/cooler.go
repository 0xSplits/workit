package proxy

import (
	"time"

	"github.com/0xSplits/workit/handler"
)

// Cooler returns the wait duration of the underlying worker handler if that
// handler implements the handler.Cooler interface. Otherwise 0 is returned.
func (p *Proxy) Cooler() time.Duration {
	v, i := p.han.(handler.Cooler)
	if i {
		return v.Cooler()
	}

	return 0
}
