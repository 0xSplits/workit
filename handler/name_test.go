package handler

import (
	"fmt"
	"testing"

	"github.com/0xSplits/workit/testdata/artefact"
	"github.com/0xSplits/workit/testdata/metadata"
	"github.com/0xSplits/workit/testdata/operator"
	"github.com/google/go-cmp/cmp"
)

func Test_Handler_Name(t *testing.T) {
	testCases := []struct {
		han Interface
		nam string
	}{
		// Case 000
		{
			han: &artefact.Handler{},
			nam: "artefact",
		},
		// Case 001
		{
			han: &metadata.Handler{},
			nam: "metadata",
		},
		// Case 002
		{
			han: &operator.Handler{},
			nam: "operator",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {
			nam := Name(tc.han)
			if dif := cmp.Diff(tc.nam, nam); dif != "" {
				t.Fatalf("-expected +actual:\n%s", dif)
			}
		})
	}
}
