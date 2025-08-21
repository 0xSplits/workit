package proxy

import (
	"time"

	"github.com/0xSplits/workit/handler"
)

func (p *Proxy) Cooler() time.Duration {
	v, i := p.han.(handler.Cooler)
	if i {
		return v.Cooler()
	}

	return 0
}
