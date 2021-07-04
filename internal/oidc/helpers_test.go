package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopeNamesToScopes(t *testing.T) {
	scopeNames := []string{"openid"}

	scopes := scopeNamesToScopes(scopeNames)
	assert.Equal(t, "openid", scopes[0].Name)
	assert.Equal(t, "Use OpenID to verify your identity", scopes[0].Description)

	scopeNames = []string{"groups"}

	scopes = scopeNamesToScopes(scopeNames)
	assert.Equal(t, "groups", scopes[0].Name)
	assert.Equal(t, "Access your group membership", scopes[0].Description)

	scopeNames = []string{"profile"}

	scopes = scopeNamesToScopes(scopeNames)
	assert.Equal(t, "profile", scopes[0].Name)
	assert.Equal(t, "Access your display name", scopes[0].Description)

	scopeNames = []string{"email"}

	scopes = scopeNamesToScopes(scopeNames)
	assert.Equal(t, "email", scopes[0].Name)
	assert.Equal(t, "Access your email addresses", scopes[0].Description)

	scopeNames = []string{"another"}

	scopes = scopeNamesToScopes(scopeNames)
	assert.Equal(t, "another", scopes[0].Name)
	assert.Equal(t, "another", scopes[0].Description)
}

func TestAudienceNamesToScopes(t *testing.T) {
	audienceNames := []string{"audience", "another_aud"}

	audiences := audienceNamesToAudience(audienceNames)
	assert.Equal(t, "audience", audiences[0].Name)
	assert.Equal(t, "audience", audiences[0].Description)
	assert.Equal(t, "another_aud", audiences[1].Name)
	assert.Equal(t, "another_aud", audiences[1].Description)
}
