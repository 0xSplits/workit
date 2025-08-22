package proxy

// Ensure executes the business logic of the wrapped worker handler
// transparently without any additional behaviour change.
func (p *Proxy) Ensure() error {
	return p.han.Ensure()
}
