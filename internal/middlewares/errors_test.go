package middlewares

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverErr(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		expected any
	}{
		{
			"ShouldHandleNil",
			nil,
			nil,
		},
		{
			"ShouldHandleString",
			"a string",
			fmt.Errorf("recovered panic: a string"),
		},
		{
			"ShouldHandleWrapped",
			fmt.Errorf("a string"),
			fmt.Errorf("recovered panic: %w", fmt.Errorf("a string")),
		},
		{
			"ShouldHandleInt",
			5,
			fmt.Errorf("recovered panic with unknown type: 5"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, recoverErr(tc.have))
		})
	}
}
