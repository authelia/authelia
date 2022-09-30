package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
