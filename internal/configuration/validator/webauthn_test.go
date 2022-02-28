package validator

import (
	"testing"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestWebauthnShouldSetDefaultValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Webauthn: schema.WebauthnConfiguration{},
	}

	ValidateWebauthn(config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultWebauthnConfiguration.DisplayName, config.Webauthn.DisplayName)
	assert.Equal(t, schema.DefaultWebauthnConfiguration.Timeout, config.Webauthn.Timeout)
	assert.Equal(t, schema.DefaultWebauthnConfiguration.ConveyancePreference, config.Webauthn.ConveyancePreference)
	assert.Equal(t, schema.DefaultWebauthnConfiguration.UserVerification, config.Webauthn.UserVerification)
}

func TestWebauthnShouldNotSetDefaultValuesWhenConfigured(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Webauthn: schema.WebauthnConfiguration{
			DisplayName:          "Test",
			Timeout:              "50s",
			ConveyancePreference: protocol.PreferNoAttestation,
			UserVerification:     protocol.VerificationDiscouraged,
		},
	}

	ValidateWebauthn(config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "Test", config.Webauthn.DisplayName)
	assert.Equal(t, "50s", config.Webauthn.Timeout)
	assert.Equal(t, protocol.PreferNoAttestation, config.Webauthn.ConveyancePreference)
	assert.Equal(t, protocol.VerificationDiscouraged, config.Webauthn.UserVerification)

	config.Webauthn.ConveyancePreference = protocol.PreferIndirectAttestation
	config.Webauthn.UserVerification = protocol.VerificationPreferred

	ValidateWebauthn(config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, protocol.PreferIndirectAttestation, config.Webauthn.ConveyancePreference)
	assert.Equal(t, protocol.VerificationPreferred, config.Webauthn.UserVerification)

	config.Webauthn.ConveyancePreference = protocol.PreferDirectAttestation
	config.Webauthn.UserVerification = protocol.VerificationRequired

	ValidateWebauthn(config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, protocol.PreferDirectAttestation, config.Webauthn.ConveyancePreference)
	assert.Equal(t, protocol.VerificationRequired, config.Webauthn.UserVerification)
}

func TestWebauthnShouldRaiseErrorsOnInvalidOptions(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Webauthn: schema.WebauthnConfiguration{
			DisplayName:          "Test",
			Timeout:              "50s",
			ConveyancePreference: "no",
			UserVerification:     "yes",
		},
	}

	ValidateWebauthn(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "webauthn: option 'attestation_conveyance_preference' must be one of 'none', 'indirect', 'direct' but it is configured as 'no'")
	assert.EqualError(t, validator.Errors()[1], "webauthn: option 'user_verification' must be one of 'discouraged', 'preferred', 'required' but it is configured as 'yes'")
}

func TestWebauthnShouldRaiseErrorsOnInvalidTimeout(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Webauthn: schema.WebauthnConfiguration{
			Timeout: "abc",
		},
	}

	ValidateWebauthn(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "webauthn: option 'timeout' could not be parsed: could not parse 'abc' as a duration")
}
