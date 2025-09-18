package proxy

type testEnsure struct{}

func (t *testEnsure) Active() bool {
	return true
}

func (t *testEnsure) Ensure() error {
	return nil
}
