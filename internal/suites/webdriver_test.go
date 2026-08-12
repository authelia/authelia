package suites

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDevToolsEnabled(t *testing.T) {
	testCases := []struct {
		name     string
		ci       string
		disable  bool
		headless bool
		expected bool
	}{
		{
			"ShouldEnableWhenNotCINotHeadlessAndNotDisabled",
			"",
			false,
			false,
			true,
		},
		{
			"ShouldEnableWhenCIIsNotTrue",
			"false",
			false,
			false,
			true,
		},
		{
			"ShouldDisableWhenCI",
			"true",
			false,
			false,
			false,
		},
		{
			"ShouldDisableWhenHeadless",
			"",
			false,
			true,
			false,
		},
		{
			"ShouldDisableWhenDisabled",
			"",
			true,
			false,
			false,
		},
		{
			"ShouldDisableWhenCIAndHeadless",
			"true",
			false,
			true,
			false,
		},
		{
			"ShouldDisableWhenCIAndDisabled",
			"true",
			true,
			false,
			false,
		},
		{
			"ShouldDisableWhenHeadlessAndDisabled",
			"",
			true,
			true,
			false,
		},
		{
			"ShouldDisableWhenCIAndHeadlessAndDisabled",
			"true",
			true,
			true,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CI", tc.ci)

			assert.Equal(t, tc.expected, isDevToolsEnabled(tc.disable, tc.headless))
		})
	}
}
