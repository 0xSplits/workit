package proxy

type testEnsure struct{}

func (t *testEnsure) Ensure() error {
	return nil
}
