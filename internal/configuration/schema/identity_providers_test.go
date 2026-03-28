package schema

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCustomClaimByName(t *testing.T) {
	claims := IdentityProvidersOpenIDConnectCustomClaims{
		"claim_a": {Name: "email", Attribute: "mail"},
		"claim_b": {Name: "groups", Attribute: "memberOf"},
		"claim_c": {Name: "display_name", Attribute: "displayName"},
	}

	t.Run("ShouldReturnMatchingClaim", func(t *testing.T) {
		result := claims.GetCustomClaimByName("email")

		assert.Equal(t, "email", result.Name)
		assert.Equal(t, "mail", result.Attribute)
	})

	t.Run("ShouldReturnMatchingClaimGroups", func(t *testing.T) {
		result := claims.GetCustomClaimByName("groups")

		assert.Equal(t, "groups", result.Name)
		assert.Equal(t, "memberOf", result.Attribute)
	})

	t.Run("ShouldReturnEmptyForNonExistentClaim", func(t *testing.T) {
		result := claims.GetCustomClaimByName("nonexistent")

		assert.Equal(t, "", result.Name)
		assert.Equal(t, "", result.Attribute)
	})

	t.Run("ShouldReturnEmptyForEmptyMap", func(t *testing.T) {
		empty := IdentityProvidersOpenIDConnectCustomClaims{}

		result := empty.GetCustomClaimByName("email")

		assert.Equal(t, "", result.Name)
		assert.Equal(t, "", result.Attribute)
	})

	t.Run("ShouldReturnEmptyForNilMap", func(t *testing.T) {
		var nilClaims IdentityProvidersOpenIDConnectCustomClaims

		result := nilClaims.GetCustomClaimByName("email")

		assert.Equal(t, "", result.Name)
		assert.Equal(t, "", result.Attribute)
	})
}

func TestDefaultOpenIDConnectConfiguration(t *testing.T) {
	t.Run("ShouldHaveExpectedLifespans", func(t *testing.T) {
		assert.Equal(t, time.Hour, DefaultOpenIDConnectConfiguration.Lifespans.AccessToken)
		assert.Equal(t, time.Minute, DefaultOpenIDConnectConfiguration.Lifespans.AuthorizeCode)
		assert.Equal(t, time.Hour, DefaultOpenIDConnectConfiguration.Lifespans.IDToken)
		assert.Equal(t, 90*time.Minute, DefaultOpenIDConnectConfiguration.Lifespans.RefreshToken)
		assert.Equal(t, 10*time.Minute, DefaultOpenIDConnectConfiguration.Lifespans.DeviceCode)
	})

	t.Run("ShouldHaveExpectedPKCEEnforcement", func(t *testing.T) {
		assert.Equal(t, "public_clients_only", DefaultOpenIDConnectConfiguration.EnforcePKCE)
	})
}

func TestDefaultOpenIDConnectPolicyConfiguration(t *testing.T) {
	assert.Equal(t, policyTwoFactor, DefaultOpenIDConnectPolicyConfiguration.DefaultPolicy)
}

func TestDefaultOpenIDConnectClientConfiguration(t *testing.T) {
	t.Run("ShouldHaveExpectedDefaults", func(t *testing.T) {
		assert.Equal(t, policyTwoFactor, DefaultOpenIDConnectClientConfiguration.AuthorizationPolicy)
		assert.Equal(t, []string{"openid", "groups", "profile", "email"}, DefaultOpenIDConnectClientConfiguration.Scopes)
		assert.Equal(t, []string{"code"}, DefaultOpenIDConnectClientConfiguration.ResponseTypes)
		assert.Equal(t, []string{"form_post"}, DefaultOpenIDConnectClientConfiguration.ResponseModes)
	})

	t.Run("ShouldHaveExpectedSigningAlgorithms", func(t *testing.T) {
		assert.Equal(t, "RS256", DefaultOpenIDConnectClientConfiguration.AuthorizationSignedResponseAlg)
		assert.Equal(t, "RS256", DefaultOpenIDConnectClientConfiguration.IDTokenSignedResponseAlg)
		assert.Equal(t, "none", DefaultOpenIDConnectClientConfiguration.AccessTokenSignedResponseAlg)
		assert.Equal(t, "none", DefaultOpenIDConnectClientConfiguration.UserinfoSignedResponseAlg)
		assert.Equal(t, "none", DefaultOpenIDConnectClientConfiguration.IntrospectionSignedResponseAlg)
	})

	t.Run("ShouldHaveExpectedConsentSettings", func(t *testing.T) {
		assert.Equal(t, "explicit", DefaultOpenIDConnectClientConfiguration.RequestedAudienceMode)
		assert.Equal(t, "auto", DefaultOpenIDConnectClientConfiguration.ConsentMode)
		assert.NotNil(t, DefaultOpenIDConnectClientConfiguration.ConsentPreConfiguredDuration)
		assert.Equal(t, 7*24*time.Hour, *DefaultOpenIDConnectClientConfiguration.ConsentPreConfiguredDuration)
	})
}
