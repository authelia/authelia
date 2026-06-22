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
		{
			"ShouldMatchUserWildcardForAuthenticatedUser",
			&AccessControlDomainMatcher{
				Name:         ".domain.com",
				UserWildcard: true,
			},
			"john.domain.com",
			Subject{Username: "john"},
			true,
		},
		{
			"ShouldNotMatchUserWildcardForDifferentUser",
			&AccessControlDomainMatcher{
				Name:         ".domain.com",
				UserWildcard: true,
			},
			"john.domain.com",
			Subject{Username: "bob"},
			false,
		},
		{
			"ShouldNotMatchUserWildcardWithExtraLabel",
			&AccessControlDomainMatcher{
				Name:         ".domain.com",
				UserWildcard: true,
			},
			"john.sub.domain.com",
			Subject{Username: "john"},
			false,
		},
		{
			"ShouldMatchGroupWildcardForGroupMember",
			&AccessControlDomainMatcher{
				Name:          ".domain.com",
				GroupWildcard: true,
			},
			"admins.domain.com",
			Subject{Groups: []string{"admins"}},
			true,
		},
		{
			"ShouldNotMatchGroupWildcardForNonMember",
			&AccessControlDomainMatcher{
				Name:          ".domain.com",
				GroupWildcard: true,
			},
			"users.domain.com",
			Subject{Groups: []string{"admins"}},
			false,
		},
		{
			"ShouldNotMatchGroupWildcardWithExtraLabel",
			&AccessControlDomainMatcher{
				Name:          ".domain.com",
				GroupWildcard: true,
			},
			"admins.sub.domain.com",
			Subject{Groups: []string{"admins"}},
			false,
		},
		{
			"ShouldNotMatchUserWildcardWhenDomainHasNoSeparator",
			&AccessControlDomainMatcher{
				Name:         "abc",
				UserWildcard: true,
			},
			"johnabc",
			Subject{Username: "john"},
			false,
		},
		{
			"ShouldNotMatchGroupWildcardWhenDomainHasNoSeparator",
			&AccessControlDomainMatcher{
				Name:          "abc",
				GroupWildcard: true,
			},
			"adminsabc",
			Subject{Groups: []string{"admins"}},
			false,
		},
		{
			"ShouldNotMatchUserWildcardWithEmptyName",
			&AccessControlDomainMatcher{
				Name:         "",
				UserWildcard: true,
			},
			"john.domain.com",
			Subject{Username: "john"},
			false,
		},
		{
			"ShouldNotMatchGroupWildcardWithEmptyName",
			&AccessControlDomainMatcher{
				Name:          "",
				GroupWildcard: true,
			},
			"admins.domain.com",
			Subject{Groups: []string{"admins"}},
			false,
		},
		{
			"ShouldNotMatchUserWildcardWithoutDotSeparator",
			&AccessControlDomainMatcher{
				Name:         "-example.domain.com",
				UserWildcard: true,
			},
			"john-example.domain.com",
			Subject{Username: "john"},
			false,
		},
		{
			"ShouldNotMatchGroupWildcardWithoutDotSeparator",
			&AccessControlDomainMatcher{
				Name:          "-example.domain.com",
				GroupWildcard: true,
			},
			"admins-example.domain.com",
			Subject{Groups: []string{"admins"}},
			false,
		},
		{
			"ShouldNotMatchUserWildcardWithDottedUsername",
			&AccessControlDomainMatcher{
				Name:         ".domain.com",
				UserWildcard: true,
			},
			"john.doe.domain.com",
			Subject{Username: "john.doe"},
			false,
		},
		{
			"ShouldNotMatchGroupWildcardWithDottedGroup",
			&AccessControlDomainMatcher{
				Name:          ".domain.com",
				GroupWildcard: true,
			},
			"dev.team.domain.com",
			Subject{Groups: []string{"dev.team"}},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.IsMatch(tc.domain, tc.subject))
		})
	}
}
