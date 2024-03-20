package handlers

import (
	"fmt"
	"testing"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestShouldGrantAppropriateClaimsForScopeProfile(t *testing.T) {
	consent := &model.OAuth2ConsentSession{
		GrantedScopes: []string{oidc.ScopeProfile},
	}

	extraClaims := oidcGrantRequests(nil, consent, &oidcUserSessionJohn)

	assert.Len(t, extraClaims, 2)

	require.Contains(t, extraClaims, oidc.ClaimPreferredUsername)
	assert.Equal(t, "john", extraClaims[oidc.ClaimPreferredUsername])

	require.Contains(t, extraClaims, oidc.ClaimFullName)
	assert.Equal(t, "John Smith", extraClaims[oidc.ClaimFullName])
}

func TestShouldGrantAppropriateClaimsForScopeGroups(t *testing.T) {
	consent := &model.OAuth2ConsentSession{
		GrantedScopes: []string{oidc.ScopeGroups},
	}

	extraClaims := oidcGrantRequests(nil, consent, &oidcUserSessionJohn)

	assert.Len(t, extraClaims, 1)

	require.Contains(t, extraClaims, oidc.ClaimGroups)
	assert.Len(t, extraClaims[oidc.ClaimGroups], 2)
	assert.Contains(t, extraClaims[oidc.ClaimGroups], "admin")
	assert.Contains(t, extraClaims[oidc.ClaimGroups], "dev")

	extraClaims = oidcGrantRequests(nil, consent, &oidcUserSessionFred)

	assert.Len(t, extraClaims, 1)

	require.Contains(t, extraClaims, oidc.ClaimGroups)
	assert.Len(t, extraClaims[oidc.ClaimGroups], 1)
	assert.Contains(t, extraClaims[oidc.ClaimGroups], "dev")
}

func TestShouldGrantAppropriateClaimsForScopeEmail(t *testing.T) {
	consent := &model.OAuth2ConsentSession{
		GrantedScopes: []string{oidc.ScopeEmail},
	}

	extraClaims := oidcGrantRequests(nil, consent, &oidcUserSessionJohn)

	assert.Len(t, extraClaims, 3)

	require.Contains(t, extraClaims, oidc.ClaimPreferredEmail)
	assert.Equal(t, "j.smith@authelia.com", extraClaims[oidc.ClaimPreferredEmail])

	require.Contains(t, extraClaims, oidc.ClaimEmailAlts)
	assert.Len(t, extraClaims[oidc.ClaimEmailAlts], 1)
	assert.Contains(t, extraClaims[oidc.ClaimEmailAlts], "admin@authelia.com")

	require.Contains(t, extraClaims, oidc.ClaimEmailVerified)
	assert.Equal(t, true, extraClaims[oidc.ClaimEmailVerified])

	extraClaims = oidcGrantRequests(nil, consent, &oidcUserSessionFred)

	assert.Len(t, extraClaims, 2)

	require.Contains(t, extraClaims, oidc.ClaimPreferredEmail)
	assert.Equal(t, "f.smith@authelia.com", extraClaims[oidc.ClaimPreferredEmail])

	require.Contains(t, extraClaims, oidc.ClaimEmailVerified)
	assert.Equal(t, true, extraClaims[oidc.ClaimEmailVerified])
}

func TestShouldGrantAppropriateClaimsForScopeOpenIDAndProfile(t *testing.T) {
	consent := &model.OAuth2ConsentSession{
		GrantedScopes: []string{oidc.ScopeOpenID, oidc.ScopeProfile},
	}

	extraClaims := oidcGrantRequests(nil, consent, &oidcUserSessionJohn)

	assert.Len(t, extraClaims, 2)

	require.Contains(t, extraClaims, oidc.ClaimPreferredUsername)
	assert.Equal(t, "john", extraClaims[oidc.ClaimPreferredUsername])

	require.Contains(t, extraClaims, oidc.ClaimFullName)
	assert.Equal(t, "John Smith", extraClaims[oidc.ClaimFullName])

	extraClaims = oidcGrantRequests(nil, consent, &oidcUserSessionFred)

	assert.Len(t, extraClaims, 2)

	require.Contains(t, extraClaims, oidc.ClaimPreferredUsername)
	assert.Equal(t, "fred", extraClaims[oidc.ClaimPreferredUsername])

	require.Contains(t, extraClaims, oidc.ClaimFullName)
	assert.Equal(t, extraClaims[oidc.ClaimFullName], "Fred Smith")
}

func TestOIDCApplyUserInfoClaims(t *testing.T) {
	testCases := []struct {
		name               string
		clientID           string
		scopes             oauthelia2.Arguments
		resolver           oidcDetailResolver
		details            *authentication.UserDetails
		original, expected map[string]any
	}{
		{
			name:     "ShouldNotMapClaimsWhenSubjectAbsent",
			clientID: "test",
			scopes:   []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail},
			details: &authentication.UserDetails{
				Username:    "john",
				DisplayName: "John Smith",
				Groups:      []string{"abc", "123"},
				Emails:      []string{"john@example.com", "john.smith@example.com"},
			},
			original: map[string]any{},
			expected: map[string]any{oidc.ClaimAudience: []string{"test"}},
		},
		{
			name:     "ShouldNotMapClaimsWhenSubjectNotUUID",
			clientID: "test",
			scopes:   []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail},
			details: &authentication.UserDetails{
				Username:    "john",
				DisplayName: "John Smith",
				Groups:      []string{"abc", "123"},
				Emails:      []string{"john@example.com", "john.smith@example.com"},
			},
			original: map[string]any{oidc.ClaimSubject: "abc"},
			expected: map[string]any{oidc.ClaimAudience: []string{"test"}, oidc.ClaimSubject: "abc"},
		},
		{
			name:     "ShouldNotMapClaimsWhenSubjectNotString",
			clientID: "test",
			scopes:   []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail},
			details: &authentication.UserDetails{
				Username:    "john",
				DisplayName: "John Smith",
				Groups:      []string{"abc", "123"},
				Emails:      []string{"john@example.com", "john.smith@example.com"},
			},
			original: map[string]any{oidc.ClaimSubject: 1},
			expected: map[string]any{oidc.ClaimAudience: []string{"test"}, oidc.ClaimSubject: 1},
		},
		{
			name:     "ShouldNotMapClaimsWhenScopesAbsent",
			clientID: "test",
			scopes:   []string{oidc.ScopeOpenID},
			details: &authentication.UserDetails{
				Username:    "john",
				DisplayName: "John Smith",
				Groups:      []string{"abc", "123"},
				Emails:      []string{"john@example.com", "john.smith@example.com"},
			},
			original: map[string]any{oidc.ClaimSubject: "6f05a84f-de27-47e7-8b95-351966532c42"},
			expected: map[string]any{oidc.ClaimAudience: []string{"test"}, oidc.ClaimSubject: "6f05a84f-de27-47e7-8b95-351966532c42"},
		},
		{
			name:     "ShouldNotMapClaimsWhenResolverError",
			clientID: "test",
			scopes:   []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail},
			resolver: func(subject uuid.UUID) (detailer oidc.UserDetailer, err error) {
				return nil, fmt.Errorf("an error")
			},
			original: map[string]any{oidc.ClaimSubject: "6f05a84f-de27-47e7-8b95-351966532c42"},
			expected: map[string]any{oidc.ClaimAudience: []string{"test"}, oidc.ClaimSubject: "6f05a84f-de27-47e7-8b95-351966532c42"},
		},
		{
			name:     "ShouldMapAllClaims",
			clientID: "test",
			scopes:   []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail},
			details: &authentication.UserDetails{
				Username:    "john",
				DisplayName: "John Smith",
				Groups:      []string{"abc", "123"},
				Emails:      []string{"john@example.com", "john.smith@example.com"},
			},
			original: map[string]any{oidc.ClaimSubject: "6f05a84f-de27-47e7-8b95-351966532c42"},
			expected: map[string]any{
				oidc.ClaimAudience:          []string{"test"},
				oidc.ClaimSubject:           "6f05a84f-de27-47e7-8b95-351966532c42",
				oidc.ClaimFullName:          "John Smith",
				oidc.ClaimPreferredUsername: "john",
				oidc.ClaimGroups:            []string{"abc", "123"},
				oidc.ClaimPreferredEmail:    "john@example.com",
				oidc.ClaimEmailVerified:     true,
				oidc.ClaimEmailAlts:         []string{"john.smith@example.com"},
			},
		},
		{
			name:     "ShouldMapAllClaimsWithExtras",
			clientID: "test",
			scopes:   []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail},
			details: &authentication.UserDetails{
				Username:    "john",
				DisplayName: "John Smith",
				Groups:      []string{"abc", "123"},
				Emails:      []string{"john@example.com", "john.smith@example.com"},
			},
			original: map[string]any{
				oidc.ClaimSubject:           "6f05a84f-de27-47e7-8b95-351966532c42",
				oidc.ClaimAudience:          []string{"example"},
				oidc.ClaimAccessTokenHash:   "abc",
				oidc.ClaimPreferredUsername: "not-john",
				oidc.ClaimGroups:            []string{"old", "999"},
				oidc.ClaimEmailVerified:     false,
				oidc.ClaimPreferredEmail:    "not-john@example.com",
				oidc.ClaimFullName:          "John Smithy",
				oidc.ClaimEmailAlts:         []string{"john.smithy@example.com"},
			},
			expected: map[string]any{
				oidc.ClaimAudience:          []string{"example", "test"},
				oidc.ClaimSubject:           "6f05a84f-de27-47e7-8b95-351966532c42",
				oidc.ClaimFullName:          "John Smith",
				oidc.ClaimPreferredUsername: "john",
				oidc.ClaimGroups:            []string{"abc", "123"},
				oidc.ClaimPreferredEmail:    "john@example.com",
				oidc.ClaimEmailVerified:     true,
				oidc.ClaimEmailAlts:         []string{"john.smith@example.com"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims := map[string]any{}

			resolver := tc.resolver

			if resolver == nil {
				resolver = oidcTestDetailerFromSubject(tc.details)
			}

			oidcApplyUserInfoClaims(tc.clientID, tc.scopes, tc.original, claims, resolver)

			assert.Equal(t, tc.expected, claims)
		})
	}
}

func oidcTestDetailerFromSubject(details *authentication.UserDetails) oidcDetailResolver {
	return func(subject uuid.UUID) (detailer oidc.UserDetailer, err error) {
		return details, nil
	}
}

var (
	oidcUserSessionJohn = session.UserSession{
		Username:    "john",
		Groups:      []string{"admin", "dev"},
		DisplayName: "John Smith",
		Emails:      []string{"j.smith@authelia.com", "admin@authelia.com"},
	}

	oidcUserSessionFred = session.UserSession{
		Username:    "fred",
		Groups:      []string{"dev"},
		DisplayName: "Fred Smith",
		Emails:      []string{"f.smith@authelia.com"},
	}
)
