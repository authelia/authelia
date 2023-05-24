package authorization

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexpGroupStringSubjectMatcher_IsMatch(t *testing.T) {
	testCases := []struct {
		name     string
		have     *RegexpGroupStringSubjectMatcher
		input    string
		subject  Subject
		expected bool
	}{
		{
			"Abc",
			&RegexpGroupStringSubjectMatcher{
				MustCompileRegexNoPtr(`^(?P<User>[a-zA-Z0-9]+)\.regex.com$`),
				1,
				0,
			},
			"example.com",
			Subject{Username: "a-user"},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.IsMatch(tc.input, tc.subject))
		})
	}
}

func MustCompileRegexNoPtr(input string) regexp.Regexp {
	out := regexp.MustCompile(input)

	return *out
}
