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

func TestWebAuthnPasskeyBooleans(t *testing.T) {
	testCases := []struct {
		name     string
		disable  bool
		passkey  bool
		mfa      bool
		upgrade  bool
		expected []string
	}{
		{
			"ShouldNotErrorOnDefaultValues",
			false,
			false,
			false,
			false,
			nil,
		},
		{
			"ShouldNotErrorDisabled",
			true,
			false,
			false,
			false,
			nil,
		},
		{
			"ShouldNotErrorPasskeys",
			false,
			true,
			false,
			false,
			nil,
		},
		{
			"ShouldNotErrorPasskeysAllEnabled",
			false,
			true,
			true,
			true,
			nil,
		},
		{
			"ShouldErrorPasskeysWebAuthnDisabled",
			true,
			true,
			false,
			false,
			[]string{
				"webauthn: option 'enable_passkey_login' is true but it must be false when 'disable' is true",
			},
		},
		{
			"ShouldErrorNoPasskeysWithPasskeyOptionUpgrade",
			false,
			false,
			false,
			true,
			[]string{
				"webauthn: option 'experimental_enable_passkey_upgrade' is true but it must be false when 'enable_passkey_login' is false",
			},
		},
		{
			"ShouldErrorNoPasskeysWithPasskeyOptionMFA",
			false,
			false,
			true,
			false,
			[]string{
				"webauthn: option 'experimental_enable_passkey_uv_two_factors' is true but it must be false when 'enable_passkey_login' is false",
			},
		},
		{
			"ShouldErrorNoPasskeysWithPasskeyOptionMultiple",
			false,
			false,
			true,
			true,
			[]string{
				"webauthn: option 'experimental_enable_passkey_uv_two_factors' is true but it must be false when 'enable_passkey_login' is false",
				"webauthn: option 'experimental_enable_passkey_upgrade' is true but it must be false when 'enable_passkey_login' is false",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.Configuration{
				WebAuthn: schema.WebAuthn{
					Disable:              tc.disable,
					EnablePasskeyLogin:   tc.passkey,
					EnablePasskey2FA:     tc.mfa,
					EnablePasskeyUpgrade: tc.upgrade,
				},
			}

			val := schema.NewStructValidator()

			ValidateWebAuthn(config, val)

			assert.False(t, val.HasWarnings())

			errs := val.Errors()

			require.Len(t, errs, len(tc.expected))

			for i, err := range errs {
				assert.EqualError(t, err, tc.expected[i])
			}
		})
	}
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
		{
			"ShouldHandleBadCachePolicy",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					Metadata: schema.WebAuthnMetadata{
						Enabled:     true,
						CachePolicy: "x",
					},
				},
			},
			nil,
			[]string{
				"webauthn: metadata: option 'cache_policy' is 'x' but it must be 'strict' or 'relaxed'",
			},
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
