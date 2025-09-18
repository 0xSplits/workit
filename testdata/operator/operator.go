package operator

type Operator struct{}

func (p *Operator) Active() bool {
	return true
}

func (o *Operator) Ensure() error {
	return nil
}
