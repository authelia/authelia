package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsIntegerInSlice(t *testing.T) {
	testCases := []struct {
		name     string
		needle   int
		haystack []int
		expected bool
	}{
		{
			"ShouldFindInt",
			5,
			[]int{1, 5},
			true,
		},
		{
			"ShouldNotFindIntNil",
			5,
			nil,
			false,
		},
		{
			"ShouldNotFindIntEmpty",
			5,
			[]int{},
			false,
		},
		{
			"ShouldNotFindIntMismatched",
			5,
			[]int{1},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsIntegerInSlice(tc.needle, tc.haystack))
		})
	}
}
