package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestShouldDetectIfConsentIsMissing(t *testing.T) {
	var workflow *session.OIDCWorkflowSession

	requestedScopes := []string{"openid", "profile"}
	requestedAudience := []string{"https://authelia.com"}

	assert.True(t, isConsentMissing(workflow, requestedScopes, requestedAudience))

	workflow = &session.OIDCWorkflowSession{
		GrantedScopes:   []string{"openid", "profile"},
		GrantedAudience: []string{"https://authelia.com"},
	}

	assert.False(t, isConsentMissing(workflow, requestedScopes, requestedAudience))

	requestedScopes = []string{"openid", "profile", "group"}

	assert.True(t, isConsentMissing(workflow, requestedScopes, requestedAudience))

	requestedScopes = []string{"openid", "profile"}
	requestedAudience = []string{"https://not.authelia.com"}
	assert.True(t, isConsentMissing(workflow, requestedScopes, requestedAudience))
}

func TestShouldGrantAppropriateClaimsForScopeOpenID(t *testing.T) {
	extraClaims := oidcGrantRequests(nil, []string{oidc.ScopeOpenID}, []string{}, &oidcUserDetailsJohn)

	assert.Len(t, extraClaims, 1)

	require.Contains(t, extraClaims, oidc.ClaimPreferredUsername)
	assert.Equal(t, "john", extraClaims[oidc.ClaimPreferredUsername])
}

func TestShouldGrantAppropriateClaimsForScopeOpenIDAndGroups(t *testing.T) {
	extraClaims := oidcGrantRequests(nil, []string{oidc.ScopeOpenID, oidc.ScopeGroups}, []string{}, &oidcUserDetailsJohn)

	assert.Len(t, extraClaims, 2)

	require.Contains(t, extraClaims, oidc.ClaimPreferredUsername)
	assert.Equal(t, "john", extraClaims[oidc.ClaimPreferredUsername])

	require.Contains(t, extraClaims, oidc.ClaimGroups)
	assert.Len(t, extraClaims[oidc.ClaimGroups], 2)
	assert.Contains(t, extraClaims[oidc.ClaimGroups], "admin")
	assert.Contains(t, extraClaims[oidc.ClaimGroups], "dev")

	extraClaims = oidcGrantRequests(nil, []string{oidc.ScopeOpenID, oidc.ScopeGroups}, []string{}, &oidcUserDetailsFred)

	assert.Len(t, extraClaims, 2)

	require.Contains(t, extraClaims, oidc.ClaimPreferredUsername)
	assert.Equal(t, "fred", extraClaims[oidc.ClaimPreferredUsername])

	require.Contains(t, extraClaims, oidc.ClaimGroups)
	assert.Len(t, extraClaims[oidc.ClaimGroups], 1)
	assert.Contains(t, extraClaims[oidc.ClaimGroups], "dev")
}

func TestShouldGrantAppropriateClaimsForScopeOpenIDAndEmail(t *testing.T) {
	extraClaims := oidcGrantRequests(nil, []string{oidc.ScopeOpenID, oidc.ScopeEmail}, []string{}, &oidcUserDetailsJohn)

	assert.Len(t, extraClaims, 4)

	require.Contains(t, extraClaims, oidc.ClaimPreferredUsername)
	assert.Equal(t, "john", extraClaims[oidc.ClaimPreferredUsername])

	require.Contains(t, extraClaims, oidc.ClaimEmail)
	assert.Equal(t, "j.smith@authelia.com", extraClaims[oidc.ClaimEmail])

	require.Contains(t, extraClaims, oidc.ClaimEmailAlts)
	assert.Len(t, extraClaims[oidc.ClaimEmailAlts], 1)
	assert.Contains(t, extraClaims[oidc.ClaimEmailAlts], "admin@authelia.com")

	require.Contains(t, extraClaims, oidc.ClaimEmailVerified)
	assert.Equal(t, true, extraClaims[oidc.ClaimEmailVerified])

	extraClaims = oidcGrantRequests(nil, []string{oidc.ScopeOpenID, oidc.ScopeEmail}, []string{}, &oidcUserDetailsFred)

	assert.Len(t, extraClaims, 3)

	require.Contains(t, extraClaims, oidc.ClaimPreferredUsername)
	assert.Equal(t, "fred", extraClaims[oidc.ClaimPreferredUsername])

	require.Contains(t, extraClaims, oidc.ClaimEmail)
	assert.Equal(t, "f.smith@authelia.com", extraClaims[oidc.ClaimEmail])

	require.Contains(t, extraClaims, oidc.ClaimEmailVerified)
	assert.Equal(t, true, extraClaims[oidc.ClaimEmailVerified])
}

func TestShouldGrantAppropriateClaimsForScopeOpenIDAndProfile(t *testing.T) {
	extraClaims := oidcGrantRequests(nil, []string{oidc.ScopeOpenID, oidc.ScopeProfile}, []string{}, &oidcUserDetailsJohn)

	assert.Len(t, extraClaims, 2)

	require.Contains(t, extraClaims, oidc.ClaimPreferredUsername)
	assert.Equal(t, "john", extraClaims[oidc.ClaimPreferredUsername])

	require.Contains(t, extraClaims, oidc.ClaimDisplayName)
	assert.Equal(t, "John Smith", extraClaims[oidc.ClaimDisplayName])

	extraClaims = oidcGrantRequests(nil, []string{oidc.ScopeOpenID, oidc.ScopeProfile}, []string{}, &oidcUserDetailsFred)

	assert.Len(t, extraClaims, 2)

	require.Contains(t, extraClaims, oidc.ClaimPreferredUsername)
	assert.Equal(t, "fred", extraClaims[oidc.ClaimPreferredUsername])

	require.Contains(t, extraClaims, oidc.ClaimDisplayName)
	assert.Equal(t, extraClaims[oidc.ClaimDisplayName], "Fred Smith")
}

var (
	oidcUserDetailsJohn = model.UserDetails{
		Username:    "john",
		Groups:      []string{"admin", "dev"},
		DisplayName: "John Smith",
		Emails:      []string{"j.smith@authelia.com", "admin@authelia.com"},
	}

	oidcUserDetailsFred = model.UserDetails{
		Username:    "fred",
		Groups:      []string{"dev"},
		DisplayName: "Fred Smith",
		Emails:      []string{"f.smith@authelia.com"},
	}
)
