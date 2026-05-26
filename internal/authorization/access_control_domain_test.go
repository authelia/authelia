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
			"ShouldMatchDomainSuffixUserWildcardWithMixedCaseDomain",
			&AccessControlDomainMatcher{
				Name:         "-user.domain.com",
				UserWildcard: true,
			},
			"A-User.Domain.COM",
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
			"ShouldMatchDomainSuffixGroupWildcardWithMixedCaseDomain",
			&AccessControlDomainMatcher{
				Name:          "-group.domain.com",
				GroupWildcard: true,
			},
			"A-Group.Domain.COM",
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
			"ShouldNotMatchExactDomainWithUserWildcardWithMixedCaseDomain",
			&AccessControlDomainMatcher{
				Name:         "-user.domain.com",
				UserWildcard: true,
			},
			"-User.Domain.COM",
			Subject{},
			false,
		},
		{
			"ShouldNotMatchExactDomainWithGroupWildcardWithMixedCaseDomain",
			&AccessControlDomainMatcher{
				Name:          "-group.domain.com",
				GroupWildcard: true,
			},
			"-Group.Domain.COM",
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
