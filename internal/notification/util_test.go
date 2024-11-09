package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSanitizedPassword(t *testing.T) {
	tests := []struct {
		name     string
		have     string
		expected string
	}{
		{
			"ShouldHandleShortString",
			"abc",
			"***",
		},
		{
			"ShouldHandleMediumString",
			"abc123",
			"ab****",
		},
		{
			"ShouldHandleLongString",
			"abc123abc123abc123abc123",
			"abc12*******************",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, len(tc.expected), len(tc.have))
			assert.Equal(t, tc.expected, getSanitizedPassword(tc.have))
		})
	}
}
