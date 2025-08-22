package proxy

import (
	"fmt"
	"testing"

	"github.com/0xSplits/workit/handler"
	"github.com/google/go-cmp/cmp"
)

func Test_Handler_Proxy_Unwrap(t *testing.T) {
	var ens handler.Ensure
	{
		ens = &testEnsure{}
	}

	testCases := []struct {
		han handler.Ensure
	}{
		// Case 000, handler.Unwrap not implemented
		{
			han: ens,
		},
		// Case 001, handler.Unwrap implemented
		{
			han: &testUnwrap{wrp: ens},
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

			var wrp handler.Ensure
			{
				wrp = pro.Unwrap()
			}

			if dif := cmp.Diff(ens, wrp); dif != "" {
				t.Fatalf("-expected +actual:\n%s", dif)
			}
		})
	}
}

type testUnwrap struct {
	wrp handler.Ensure
}

func (t *testUnwrap) Ensure() error {
	return nil
}

func (t *testUnwrap) Unwrap() handler.Ensure {
	return t.wrp
}
