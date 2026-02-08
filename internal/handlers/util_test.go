package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedactEmail(t *testing.T) {
	testCases := []struct {
		testName string
		input    string
		expected string
	}{
		{"ShouldRedactEmail", "james.dean@authelia.com", "j********n@authelia.com"},
		{"ShouldRedactShortEmail", "me@authelia.com", "**@authelia.com"},
		{"ShouldRedactInvalidEmail", "invalidEmail.com", ""},
		{"ShouldRedactUnicode", "søren@example.com", "s***n@example.com"},
		{"ShouldReturnUnicodeInRedactedEmail", "øpenme@example.com", "ø****e@example.com"},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			require.Equal(t, tc.expected, redactEmail(tc.input))
		})
	}
}
