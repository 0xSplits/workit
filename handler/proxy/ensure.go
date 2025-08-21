package proxy

func (p *Proxy) Ensure() error {
	return p.han.Ensure()
}
