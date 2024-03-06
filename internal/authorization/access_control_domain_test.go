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
			"ShouldMatchDomainSuffixUserWildcard",
			&AccessControlDomainMatcher{
				Name:         "-user.domain.com",
				UserWildcard: true,
			},
			"a-user.domain.com",
			Subject{},
			true,
		},
		{
			"ShouldMatchDomainSuffixGroupWildcard",
			&AccessControlDomainMatcher{
				Name:          "-group.domain.com",
				GroupWildcard: true,
			},
			"a-group.domain.com",
			Subject{},
			true,
		},
		{
			"ShouldNotMatchExactDomainWithUserWildcard",
			&AccessControlDomainMatcher{
				Name:         "-user.domain.com",
				UserWildcard: true,
			},
			"-user.domain.com",
			Subject{},
			false,
		},
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.IsMatch(tc.domain, tc.subject))
		})
	}
}
