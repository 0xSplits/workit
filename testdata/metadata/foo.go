package metadata

type Foo struct{}

func (f *Foo) Active() bool {
	return false
}

func (f *Foo) Ensure() error {
	return nil
}
