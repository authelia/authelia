package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldRaiseErrorWhenInvalidOIDCServerConfiguration(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.OAuthConfiguration{
		OIDCServer: &schema.OpenIDConnectServerConfiguration{
			HMACSecret:           "abc",
			IssuerPrivateKeyPath: "",
		},
	}

	ValidateOAuth(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "OIDC Server issuer private key path must be provided")
	assert.EqualError(t, validator.Errors()[1], "OIDC Server HMAC secret must be exactly 32 chars long")
}

func TestShouldRaiseErrorWhenOIDCServerIssuerPrivateKeyPathInvalid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.OAuthConfiguration{
		OIDCServer: &schema.OpenIDConnectServerConfiguration{
			HMACSecret:           "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKeyPath: "../abc",
		},
	}

	ValidateOAuth(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "OIDC Server issuer private key path doesn't exist")
}
