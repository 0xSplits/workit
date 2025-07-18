package handler

import (
	"fmt"
	"testing"

	"github.com/0xSplits/workit/testdata/artefact"
	"github.com/0xSplits/workit/testdata/metadata"
	"github.com/0xSplits/workit/testdata/operator"
	"github.com/google/go-cmp/cmp"
)

func Test_Handler_Names(t *testing.T) {
	testCases := []struct {
		han []Interface
		nam []string
	}{
		// Case 000
		{
			han: []Interface{
				&artefact.Handler{},
				&metadata.Foo{},
				&operator.Operator{},
			},
			nam: []string{
				"artefact",
				"metadata",
				"operator",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {
			nam := Names(tc.han)
			if dif := cmp.Diff(tc.nam, nam); dif != "" {
				t.Fatalf("-expected +actual:\n%s", dif)
			}
		})
	}
}
