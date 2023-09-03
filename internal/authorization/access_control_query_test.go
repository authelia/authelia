package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewAccessControlQuery(t *testing.T) {
	testCases := []struct {
		name     string
		have     [][]schema.AccessControlRuleQuery
		expected []AccessControlQuery
		matches  [][]Object
	}{
		{
			"ShouldSkipInvalidTypeEqual",
			[][]schema.AccessControlRuleQuery{
				{
					{Operator: operatorEqual, Key: "example", Value: 1},
				},
			},
			[]AccessControlQuery{{Rules: []ObjectMatcher(nil)}},
			[][]Object{{{}}},
		},
		{
			"ShouldSkipInvalidTypePattern",
			[][]schema.AccessControlRuleQuery{
				{
					{Operator: operatorPattern, Key: "example", Value: 1},
				},
			},
			[]AccessControlQuery{{Rules: []ObjectMatcher(nil)}},
			[][]Object{{{}}},
		},
		{
			"ShouldSkipInvalidOperator",
			[][]schema.AccessControlRuleQuery{
				{
					{Operator: "nop", Key: "example", Value: 1},
				},
			},
			[]AccessControlQuery{{Rules: []ObjectMatcher(nil)}},
			[][]Object{{{}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewAccessControlQuery(tc.have)
			assert.Equal(t, tc.expected, actual)

			for i, rule := range actual {
				for _, object := range tc.matches[i] {
					assert.True(t, rule.IsMatch(object))
				}
			}
		})
	}
}
