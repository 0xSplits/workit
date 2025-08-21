package proxy

import (
	"github.com/0xSplits/workit/handler"
)

func (p *Proxy) Unwrap() handler.Ensure {
	v, i := p.han.(handler.Unwrap)
	if i {
		return v.Unwrap()
	}

	return p.han
}
