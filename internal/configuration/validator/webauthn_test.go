package validator

import (
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestWebAuthnShouldSetDefaultValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		WebAuthn: schema.WebAuthn{},
	}

	ValidateWebAuthn(config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultWebAuthnConfiguration.DisplayName, config.WebAuthn.DisplayName)
	assert.Equal(t, schema.DefaultWebAuthnConfiguration.Timeout, config.WebAuthn.Timeout)
	assert.Equal(t, schema.DefaultWebAuthnConfiguration.ConveyancePreference, config.WebAuthn.ConveyancePreference)
	assert.Equal(t, schema.DefaultWebAuthnConfiguration.SelectionCriteria.UserVerification, config.WebAuthn.SelectionCriteria.UserVerification)
	assert.Equal(t, schema.DefaultWebAuthnConfiguration.SelectionCriteria.Discoverability, config.WebAuthn.SelectionCriteria.Discoverability)
	assert.Equal(t, schema.DefaultWebAuthnConfiguration.SelectionCriteria.Attachment, config.WebAuthn.SelectionCriteria.Attachment)
}

func TestWebAuthnShouldSetDefaultTimeoutWhenNegative(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		WebAuthn: schema.WebAuthn{
			Timeout: -1,
		},
	}

	ValidateWebAuthn(config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultWebAuthnConfiguration.Timeout, config.WebAuthn.Timeout)
}

func TestWebAuthnShouldNotSetDefaultValuesWhenConfigured(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		WebAuthn: schema.WebAuthn{
			DisplayName:          "Test",
			Timeout:              time.Second * 50,
			ConveyancePreference: protocol.PreferNoAttestation,
			SelectionCriteria: schema.WebAuthnSelectionCriteria{
				UserVerification: protocol.VerificationDiscouraged,
			},
		},
	}

	ValidateWebAuthn(config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "Test", config.WebAuthn.DisplayName)
	assert.Equal(t, time.Second*50, config.WebAuthn.Timeout)
	assert.Equal(t, protocol.PreferNoAttestation, config.WebAuthn.ConveyancePreference)
	assert.Equal(t, protocol.VerificationDiscouraged, config.WebAuthn.SelectionCriteria.UserVerification)

	config.WebAuthn.ConveyancePreference = protocol.PreferIndirectAttestation
	config.WebAuthn.SelectionCriteria.UserVerification = protocol.VerificationPreferred

	ValidateWebAuthn(config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, protocol.PreferIndirectAttestation, config.WebAuthn.ConveyancePreference)
	assert.Equal(t, protocol.VerificationPreferred, config.WebAuthn.SelectionCriteria.UserVerification)

	config.WebAuthn.ConveyancePreference = protocol.PreferDirectAttestation
	config.WebAuthn.SelectionCriteria.UserVerification = protocol.VerificationRequired

	ValidateWebAuthn(config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, protocol.PreferDirectAttestation, config.WebAuthn.ConveyancePreference)
	assert.Equal(t, protocol.VerificationRequired, config.WebAuthn.SelectionCriteria.UserVerification)
}

func TestWebAuthnShouldRaiseErrorsOnInvalidOptions(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		WebAuthn: schema.WebAuthn{
			DisplayName:          "Test",
			Timeout:              time.Second * 50,
			ConveyancePreference: "no",
			SelectionCriteria: schema.WebAuthnSelectionCriteria{
				UserVerification: "yes",
			},
		},
	}

	ValidateWebAuthn(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "webauthn: option 'attestation_conveyance_preference' must be one of 'none', 'indirect', or 'direct' but it's configured as 'no'")
	assert.EqualError(t, validator.Errors()[1], "webauthn: selection_criteria: option 'user_verification' must be one of 'discouraged', 'preferred', or 'required' but it's configured as 'yes'")
}

func TestValidateWebAuthn(t *testing.T) {
	testCases := []struct {
		name     string
		have     *schema.Configuration
		warnings []string
		errors   []string
	}{
		{
			"ShouldHandleIncorrectValues",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					SelectionCriteria: schema.WebAuthnSelectionCriteria{
						Attachment:      "bad-attachment",
						Discoverability: "bad-discoverability",
					},
					Filtering: schema.WebAuthnFiltering{
						PermittedAAGUIDs:  []uuid.UUID{uuid.Must(uuid.Parse("cb69481e-8ff7-4039-93ec-0a2729a154a8"))},
						ProhibitedAAGUIDs: []uuid.UUID{uuid.Must(uuid.Parse("cb69481e-8ff7-4039-93ec-0a2729a154a8"))},
					},
				},
			},
			nil,
			[]string{
				"webauthn: selection_criteria: option 'attachment' must be one of 'platform' or 'cross-platform' but it's configured as 'bad-attachment'",
				"webauthn: selection_criteria: option 'discoverability' must be one of 'discouraged', 'preferred', or 'required' but it's configured as 'bad-discoverability'",
				"webauthn: filtering: option 'permitted_aaguids' and 'prohibited_aaguids' are mutually exclusive however both have values",
			},
		},
		{
			"ShouldHandlePasskeyWarning",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					EnablePasskeyLogin: true,
					SelectionCriteria: schema.WebAuthnSelectionCriteria{
						Discoverability: "discouraged",
					},
				},
			},
			[]string{
				"webauthn: selection_criteria: option 'discoverability' should generally be configured as 'preferred' or 'required' when passkey logins are enabled",
			},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			ValidateWebAuthn(tc.have, validator)

			warnings, errors := validator.Warnings(), validator.Errors()

			require.Len(t, warnings, len(tc.warnings))
			require.Len(t, errors, len(tc.errors))

			for i, warning := range warnings {
				assert.EqualError(t, warning, tc.warnings[i])
			}

			for i, err := range errors {
				assert.EqualError(t, err, tc.errors[i])
			}
		})
	}
}
