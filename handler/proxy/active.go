package proxy

import (
	"github.com/0xSplits/workit/handler"
)

// Active returns the scheduler primitive of the underlying worker handler if
// that handler implements the handler.Active interface. Otherwise true is
// returned.
func (p *Proxy) Active() bool {
	v, i := p.han.(handler.Active)
	if i {
		return v.Active()
	}

	return true
}
