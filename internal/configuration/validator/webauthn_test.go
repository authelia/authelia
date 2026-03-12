package validator

import (
	"net/url"
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
			"ShouldAcceptStrictCachePolicy",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					Metadata: schema.WebAuthnMetadata{
						Enabled:     true,
						CachePolicy: "strict",
					},
				},
			},
			nil,
			nil,
		},
		{
			"ShouldAcceptRelaxedCachePolicy",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					Metadata: schema.WebAuthnMetadata{
						Enabled:     true,
						CachePolicy: "relaxed",
					},
				},
			},
			nil,
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

func TestValidateWebAuthnRelatedOrigins(t *testing.T) {
	testCases := []struct {
		name   string
		have   *schema.Configuration
		errors []string
	}{
		{
			"ShouldPassWithNoRelatedOrigins",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{},
			},
			nil,
		},
		{
			"ShouldPassValidConfig",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
						"example.com": {
							Origins: []*url.URL{mustParseURL("https://example.com"), mustParseURL("https://auth.example.com")},
						},
					},
				},
				Session: schema.Session{
					Cookies: []schema.SessionCookie{
						{AutheliaURL: mustParseURL("https://auth.example.com")},
						{AutheliaURL: mustParseURL("https://example.com")},
					},
				},
			},
			nil,
		},
		{
			"ShouldErrorOnEmptyRelyingPartyID",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
						"": {
							Origins: []*url.URL{mustParseURL("https://example.com")},
						},
					},
				},
			},
			[]string{
				"webauthn: related_origins: : option 'relying_party_id' is empty but it must have a value",
			},
		},
		{
			"ShouldErrorOnUpperCaseRelyingPartyID",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
						"Example.com": {
							Origins: []*url.URL{mustParseURL("https://Example.com")},
						},
					},
				},
				Session: schema.Session{
					Cookies: []schema.SessionCookie{
						{AutheliaURL: mustParseURL("https://Example.com")},
					},
				},
			},
			[]string{
				"webauthn: related_origins: Example.com: relying party id is not lower case",
			},
		},
		{
			"ShouldErrorOnNilOrigin",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
						"example.com": {
							Origins: []*url.URL{nil},
						},
					},
				},
			},
			[]string{
				"webauthn: related_origins: example.com: option 'origins' item #1 is empty",
				"error rpid example.com does not match any origin",
			},
		},
		{
			"ShouldErrorOnOriginWithPath",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
						"example.com": {
							Origins: []*url.URL{mustParseURL("https://example.com/path")},
						},
					},
				},
				Session: schema.Session{
					Cookies: []schema.SessionCookie{
						{AutheliaURL: mustParseURL("https://example.com")},
					},
				},
			},
			[]string{
				"webauthn: related_origins: example.com: option 'origins' item #1 with value 'https://example.com/path' is invalid as it doesn't have an empty path",
			},
		},
		{
			"ShouldErrorOnOriginNotMatchingSessionCookie",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
						"example.com": {
							Origins: []*url.URL{mustParseURL("https://example.com")},
						},
					},
				},
				Session: schema.Session{
					Cookies: []schema.SessionCookie{
						{AutheliaURL: mustParseURL("https://other.com")},
					},
				},
			},
			[]string{
				"webauthn: related_origins: example.com: option 'origins' item #1 has value 'https://example.com' but this value is not a valid origin for any 'authelia_url' configured in the session cookies",
			},
		},
		{
			"ShouldErrorOnOriginNotMatchingSessionCookieNilAutheliaURL",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
						"example.com": {
							Origins: []*url.URL{mustParseURL("https://example.com")},
						},
					},
				},
				Session: schema.Session{
					Cookies: []schema.SessionCookie{
						{AutheliaURL: nil},
					},
				},
			},
			[]string{
				"webauthn: related_origins: example.com: option 'origins' item #1 has value 'https://example.com' but this value is not a valid origin for any 'authelia_url' configured in the session cookies",
			},
		},
		{
			"ShouldErrorOnDuplicateOrigins",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
						"example.com": {
							Origins: []*url.URL{mustParseURL("https://example.com"), mustParseURL("https://example.com")},
						},
					},
				},
				Session: schema.Session{
					Cookies: []schema.SessionCookie{
						{AutheliaURL: mustParseURL("https://example.com")},
					},
				},
			},
			[]string{
				"webauthn: related_origins: option 'origins' has value 'https://example.com' can only be defined in one relying party but it's defined in ''https://example.com' and 'https://example.com''",
			},
		},
		{
			"ShouldErrorWhenRPIDDoesNotMatchAnyOrigin",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
						"example.com": {
							Origins: []*url.URL{mustParseURL("https://other.com")},
						},
					},
				},
				Session: schema.Session{
					Cookies: []schema.SessionCookie{
						{AutheliaURL: mustParseURL("https://other.com")},
					},
				},
			},
			[]string{
				"error rpid example.com does not match any origin",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			validateWebAuthnRelatedOrigins(tc.have, validator)

			errors := validator.Errors()

			require.Len(t, errors, len(tc.errors))

			for i, err := range errors {
				assert.EqualError(t, err, tc.errors[i])
			}
		})
	}
}

func TestOriginMatchesCookieAutheliaURL(t *testing.T) {
	testCases := []struct {
		name     string
		config   *schema.Configuration
		origin   *url.URL
		expected bool
	}{
		{
			"ShouldMatchWhenHostnameMatches",
			&schema.Configuration{
				Session: schema.Session{
					Cookies: []schema.SessionCookie{
						{AutheliaURL: mustParseURL("https://auth.example.com")},
					},
				},
			},
			mustParseURL("https://auth.example.com"),
			true,
		},
		{
			"ShouldNotMatchWhenHostnameDiffers",
			&schema.Configuration{
				Session: schema.Session{
					Cookies: []schema.SessionCookie{
						{AutheliaURL: mustParseURL("https://auth.example.com")},
					},
				},
			},
			mustParseURL("https://other.example.com"),
			false,
		},
		{
			"ShouldSkipNilAutheliaURL",
			&schema.Configuration{
				Session: schema.Session{
					Cookies: []schema.SessionCookie{
						{AutheliaURL: nil},
						{AutheliaURL: mustParseURL("https://auth.example.com")},
					},
				},
			},
			mustParseURL("https://auth.example.com"),
			true,
		},
		{
			"ShouldReturnFalseWhenNoCookies",
			&schema.Configuration{
				Session: schema.Session{
					Cookies: []schema.SessionCookie{},
				},
			},
			mustParseURL("https://auth.example.com"),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, originMatchesCookieAutheliaURL(tc.config, tc.origin))
		})
	}
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return u
}
