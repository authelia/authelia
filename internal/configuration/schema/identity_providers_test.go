// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/jsonschema"
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

	t.Run("ShouldHaveExpectedIdentityAssertionAuthorizationGrant", func(t *testing.T) {
		assert.Equal(t, 5*time.Minute, DefaultOpenIDConnectConfiguration.IdentityAssertionAuthorizationGrant.Lifespan)
		assert.Nil(t, DefaultOpenIDConnectConfiguration.IdentityAssertionAuthorizationGrant.SingleUse)
		assert.Equal(t, time.Duration(0), DefaultOpenIDConnectConfiguration.IdentityAssertionAuthorizationGrant.AccessTokenAllowance)
		assert.False(t, DefaultOpenIDConnectConfiguration.IdentityAssertionAuthorizationGrant.ResourceOptional)
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

func TestIdentityProvidersOpenIDConnectClientTokenExchangePolicyJSONSchema(t *testing.T) {
	field, ok := reflect.TypeOf(IdentityProvidersOpenIDConnectClientTokenExchangePolicy{}).FieldByName("RequestedTokenTypes")

	require.True(t, ok)

	var (
		tagged      []any
		uniqueItems bool
	)

	for _, part := range strings.Split(field.Tag.Get("jsonschema"), ",") {
		switch {
		case strings.HasPrefix(part, "enum="):
			tagged = append(tagged, strings.TrimPrefix(part, "enum="))
		case part == "uniqueItems":
			uniqueItems = true
		}
	}

	require.NotEmpty(t, tagged)
	require.True(t, uniqueItems)

	schema := IdentityProvidersOpenIDConnectClientTokenExchangePolicy{}.JSONSchema()

	require.Len(t, schema.OneOf, 2)

	property, ok := schema.OneOf[1].Properties.Get("requested_token_types")

	require.True(t, ok)

	mapping, ok := property.(*jsonschema.Schema)

	require.True(t, ok)
	require.NotNil(t, mapping.Items)

	assert.Equal(t, tagged, mapping.Items.Enum)
	assert.Equal(t, uniqueItems, mapping.UniqueItems)
	assert.Contains(t, mapping.Items.Enum, "urn:ietf:params:oauth:token-type:id-jag")
}

func TestDefaultOpenIDConnectConfigurationTokenExchange(t *testing.T) {
	te := DefaultOpenIDConnectConfiguration.TokenExchange

	assert.Equal(t, []string{
		"urn:ietf:params:oauth:token-type:access_token",
		"urn:ietf:params:oauth:token-type:refresh_token",
		"urn:ietf:params:oauth:token-type:id_token",
	}, te.TokenTypes)
	assert.Equal(t, "urn:ietf:params:oauth:token-type:access_token", te.DefaultRequestedTokenType)
	assert.Equal(t, 5*time.Minute, te.TransactionTokens.Lifespan)
	assert.Equal(t, "RS256", te.TransactionTokens.SigningAlgorithm)
	assert.Equal(t, 8, te.TransactionTokens.Replacement.MaxChainLength)
	assert.False(t, te.TransactionTokens.Replacement.Enabled)
}

func TestIdentityProvidersOpenIDConnectClientTokenExchangeSchemaEnums(t *testing.T) {
	all := []string{
		"urn:ietf:params:oauth:token-type:access_token",
		"urn:ietf:params:oauth:token-type:refresh_token",
		"urn:ietf:params:oauth:token-type:id_token",
		"urn:ietf:params:oauth:token-type:txn_token",
	}

	testCases := []struct {
		name     string
		expected []string
	}{
		{"SubjectTokenTypesSupported", all},
		{"ActorTokenTypesSupported", all[:3]},
		{"RequestTokenTypesSupported", append(slices.Clone(all), "urn:ietf:params:oauth:token-type:id-jag")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			field, ok := reflect.TypeOf(IdentityProvidersOpenIDConnectClient{}).FieldByName(tc.name)

			require.True(t, ok)

			var actual []string

			for _, part := range strings.Split(field.Tag.Get("jsonschema"), ",") {
				if strings.HasPrefix(part, "enum=") {
					actual = append(actual, strings.TrimPrefix(part, "enum="))
				}
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestIdentityProvidersOpenIDConnectTokenExchangeSchemaEnums(t *testing.T) {
	expected := []string{
		"urn:ietf:params:oauth:token-type:access_token",
		"urn:ietf:params:oauth:token-type:refresh_token",
		"urn:ietf:params:oauth:token-type:id_token",
		"urn:ietf:params:oauth:token-type:txn_token",
		"urn:ietf:params:oauth:token-type:id-jag",
	}

	for _, name := range []string{"TokenTypes", "DefaultRequestedTokenType"} {
		field, ok := reflect.TypeOf(IdentityProvidersOpenIDConnectTokenExchange{}).FieldByName(name)

		require.True(t, ok)

		var actual []string

		for _, part := range strings.Split(field.Tag.Get("jsonschema"), ",") {
			if strings.HasPrefix(part, "enum=") {
				actual = append(actual, strings.TrimPrefix(part, "enum="))
			}
		}

		assert.Equal(t, expected, actual, name)
	}
}

func TestIdentityProvidersOpenIDConnectClientTokenExchangeSchemaEnumsIDJAG(t *testing.T) {
	testCases := []struct {
		name     string
		typ      reflect.Type
		expected bool
	}{
		{"SubjectTokenTypesSupported", reflect.TypeOf(IdentityProvidersOpenIDConnectClient{}), false},
		{"ActorTokenTypesSupported", reflect.TypeOf(IdentityProvidersOpenIDConnectClient{}), false},
		{"RequestTokenTypesSupported", reflect.TypeOf(IdentityProvidersOpenIDConnectClient{}), true},
		{"RequestedTokenTypes", reflect.TypeOf(IdentityProvidersOpenIDConnectClientTokenExchangePolicy{}), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			field, ok := tc.typ.FieldByName(tc.name)

			require.True(t, ok)

			assert.Equal(t, tc.expected, slices.Contains(strings.Split(field.Tag.Get("jsonschema"), ","), "enum=urn:ietf:params:oauth:token-type:id-jag"))
		})
	}
}
