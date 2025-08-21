package proxy

import (
	"fmt"
	"testing"
	"time"

	"github.com/0xSplits/workit/handler"
	"github.com/google/go-cmp/cmp"
)

func Test_Handler_Proxy_Cooler(t *testing.T) {
	testCases := []struct {
		han handler.Ensure
		coo time.Duration
	}{
		// Case 000
		{
			han: &testEnsure{},
			coo: 0,
		},
		// Case 001
		{
			han: &testCooler{coo: 3},
			coo: 3,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {
			var pro handler.Interface
			{
				pro = New(Config{
					Han: tc.han,
				})
			}

			var coo time.Duration
			{
				coo = pro.Cooler()
			}

			if dif := cmp.Diff(tc.coo, coo); dif != "" {
				t.Fatalf("-expected +actual:\n%s", dif)
			}
		})
	}
}

type testCooler struct {
	coo time.Duration
}

func (t *testCooler) Ensure() error {
	return nil
}

func (t *testCooler) Cooler() time.Duration {
	return t.coo
}
