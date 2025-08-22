package proxy

import (
	"github.com/0xSplits/workit/handler"
)

// Unwrap returns the wrapped handler implementation of the underlying worker
// handler if that handler implements the handler.Ensure interface. Otherwise
// the wrapped handler is returned as is. E.g. we might be dealing with a
// wrapper chain like this.
//
//	metrics -> proxy -> artifact
func (p *Proxy) Unwrap() handler.Ensure {
	v, i := p.han.(handler.Unwrap)
	if i {
		return v.Unwrap()
	}

	return p.han
}
