package operator

import "time"

type Operator struct{}

func (o *Operator) Cooler() time.Duration {
	return 0
}

func (o *Operator) Ensure() error {
	return nil
}
