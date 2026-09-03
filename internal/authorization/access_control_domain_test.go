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

func TestNewAccessControlDomainUserGroup(t *testing.T) {
	john := Subject{Username: "john", Groups: []string{"admins"}}

	testCases := []struct {
		name     string
		domain   string
		object   string
		subject  Subject
		expected bool
	}{
		{"ShouldMatchUser", "{user}.example.com", "john.example.com", john, true},
		{"ShouldMatchUserMixedCase", "{user}.example.com", "John.Example.COM", john, true},
		{"ShouldNotMatchDifferentUser", "{user}.example.com", "jane.example.com", john, false},
		{"ShouldNotMatchMultipleLabels", "{user}.example.com", "john.dev.example.com", john, false},
		{"ShouldMatchAnonymousWhenLabelPresent", "{user}.example.com", "john.example.com", Subject{}, true},
		{"ShouldNotMatchAnonymousWhenNoLabel", "{user}.example.com", "example.com", Subject{}, false},
		{"ShouldMatchGroup", "{group}.example.com", "admins.example.com", john, true},
		{"ShouldNotMatchWrongGroup", "{group}.example.com", "devs.example.com", john, false},
		{"ShouldTreatRegexMetacharactersLiterally", "{user}.ex+ample.com", "john.ex+ample.com", john, true},
		{"ShouldNotInterpretMetacharacterAsQuantifier", "{user}.ex+ample.com", "john.example.com", john, false},
		{"ShouldNotPanicOnUnbalancedBracket", "{user}.ex[ample.com", "john.ex[ample.com", john, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subjects, rule := NewAccessControlDomain(tc.domain)

			assert.True(t, subjects)
			assert.Equal(t, tc.expected, rule.IsMatch(tc.subject, Object{Domain: tc.object}))
		})
	}
}
