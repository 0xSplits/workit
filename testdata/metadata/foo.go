package metadata

import "time"

type Foo struct{}

func (f *Foo) Cooler() time.Duration {
	return 0
}

func (f *Foo) Ensure() error {
	return nil
}
