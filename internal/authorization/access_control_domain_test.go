package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccessControlDomain_IsMatch(t *testing.T) {
	testCases := []struct {
		name     string
		have     *AccessControlDomainMatcher
		domain   string
		subject  Subject
		expected bool
	}{
		{
			"ShouldMatchWildcard",
			&AccessControlDomainMatcher{
				Name:     "user.domain.com",
				Wildcard: true,
			},
			"abc.user.domain.com",
			Subject{},
			true,
		},
		{
			"ShouldMatchWildcardWithMixedCaseDomain",
			&AccessControlDomainMatcher{
				Name:     "user.domain.com",
				Wildcard: true,
			},
			"ABC.User.Domain.COM",
			Subject{},
			true,
		},
		{
			"ShouldNotMatchWildcardWithDifferentSuffix",
			&AccessControlDomainMatcher{
				Name:     "user.domain.com",
				Wildcard: true,
			},
			"abc.user.example.com",
			Subject{},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.IsMatch(tc.domain, tc.subject))
		})
	}
}
