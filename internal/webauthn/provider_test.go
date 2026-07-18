package webauthn_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

func TestNewProvider(t *testing.T) {
	testCases := []struct {
		name     string
		config   schema.WebAuthn
		origin   *url.URL
		err      string
		expected func(t *testing.T, provider *webauthn.Provider)
	}{
		{
			"ShouldUseGlobalConfigurationWithoutRelyingParties",
			schema.WebAuthn{
				WebAuthnBase: schema.WebAuthnBase{
					DisplayName:          "Authelia",
					Timeout:              time.Second * 60,
					ConveyancePreference: protocol.PreferIndirectAttestation,
					SelectionCriteria: schema.WebAuthnSelectionCriteria{
						Discoverability:  protocol.ResidentKeyRequirementPreferred,
						UserVerification: protocol.VerificationPreferred,
					},
				},
			},
			mustParseURL("https://auth.example.com"),
			"",
			func(t *testing.T, provider *webauthn.Provider) {
				assert.Equal(t, "auth.example.com", provider.WebAuthn.Config.RPID)
				assert.Equal(t, "Authelia", provider.WebAuthn.Config.RPDisplayName)
				assert.Equal(t, []string{"https://auth.example.com"}, provider.WebAuthn.Config.RPOrigins)
				assert.Equal(t, protocol.PreferIndirectAttestation, provider.WebAuthn.Config.AttestationPreference)
				assert.Equal(t, time.Second*60, provider.WebAuthn.Config.Timeouts.Login.Timeout)
				assert.Equal(t, time.Second*60, provider.WebAuthn.Config.Timeouts.Registration.Timeout)
			},
		},
		{
			"ShouldUseRelyingPartyConfigurationAndOrigins",
			schema.WebAuthn{
				WebAuthnBase: schema.WebAuthnBase{
					DisplayName:          "Global",
					Timeout:              time.Second * 60,
					ConveyancePreference: protocol.PreferIndirectAttestation,
					SelectionCriteria: schema.WebAuthnSelectionCriteria{
						Discoverability:  protocol.ResidentKeyRequirementPreferred,
						UserVerification: protocol.VerificationPreferred,
					},
				},
				RelyingParties: map[string]schema.WebAuthnRelyingParty{
					"example.com": {
						WebAuthnBase: schema.WebAuthnBase{
							DisplayName:          "Example",
							Timeout:              time.Second * 30,
							ConveyancePreference: protocol.PreferDirectAttestation,
							SelectionCriteria: schema.WebAuthnSelectionCriteria{
								Attachment:       protocol.CrossPlatform,
								Discoverability:  protocol.ResidentKeyRequirementRequired,
								UserVerification: protocol.VerificationRequired,
							},
						},
						Origins: []*url.URL{mustParseURL("https://example.com"), mustParseURL("https://auth.example.com")},
					},
				},
			},
			mustParseURL("https://example.com"),
			"",
			func(t *testing.T, provider *webauthn.Provider) {
				assert.Equal(t, "example.com", provider.WebAuthn.Config.RPID)
				assert.Equal(t, "Example", provider.WebAuthn.Config.RPDisplayName)
				assert.Equal(t, []string{"https://example.com", "https://auth.example.com"}, provider.WebAuthn.Config.RPOrigins)
				assert.Equal(t, protocol.PreferDirectAttestation, provider.WebAuthn.Config.AttestationPreference)
				assert.Equal(t, time.Second*30, provider.WebAuthn.Config.Timeouts.Login.Timeout)
				assert.Equal(t, protocol.CrossPlatform, provider.WebAuthn.Config.AuthenticatorSelection.AuthenticatorAttachment)
				assert.Equal(t, protocol.ResidentKeyRequirementRequired, provider.WebAuthn.Config.AuthenticatorSelection.ResidentKey)
				assert.Equal(t, protocol.VerificationRequired, provider.WebAuthn.Config.AuthenticatorSelection.UserVerification)
				require.NotNil(t, provider.WebAuthn.Config.AuthenticatorSelection.RequireResidentKey)
				assert.True(t, *provider.WebAuthn.Config.AuthenticatorSelection.RequireResidentKey)
			},
		},
		{
			"ShouldAppendOpaqueOriginsToOrigins",
			schema.WebAuthn{
				WebAuthnBase: schema.WebAuthnBase{
					DisplayName:          "Authelia",
					Timeout:              time.Second * 60,
					ConveyancePreference: protocol.PreferIndirectAttestation,
					SelectionCriteria: schema.WebAuthnSelectionCriteria{
						Discoverability:  protocol.ResidentKeyRequirementPreferred,
						UserVerification: protocol.VerificationPreferred,
					},
				},
				RelyingParties: map[string]schema.WebAuthnRelyingParty{
					"example.com": {
						Origins: []*url.URL{mustParseURL("https://example.com")},
						OpaqueOrigins: []string{
							"android:apk-key-hash:-IfWtPXXRFX9gAijCaxCw-f8tty8Azji56EQwBGYuj4",
							"ios:bundle-id:com.example.app",
						},
					},
				},
			},
			mustParseURL("https://example.com"),
			"",
			func(t *testing.T, provider *webauthn.Provider) {
				assert.Equal(t, []string{
					"https://example.com",
					"android:apk-key-hash:-IfWtPXXRFX9gAijCaxCw-f8tty8Azji56EQwBGYuj4",
					"ios:bundle-id:com.example.app",
				}, provider.WebAuthn.Config.RPOrigins)
			},
		},
		{
			"ShouldNotAppendEmptyOriginWhenNoOpaqueOriginsConfigured",
			schema.WebAuthn{
				WebAuthnBase: schema.WebAuthnBase{
					DisplayName:          "Authelia",
					Timeout:              time.Second * 60,
					ConveyancePreference: protocol.PreferIndirectAttestation,
					SelectionCriteria: schema.WebAuthnSelectionCriteria{
						Discoverability:  protocol.ResidentKeyRequirementPreferred,
						UserVerification: protocol.VerificationPreferred,
					},
				},
				RelyingParties: map[string]schema.WebAuthnRelyingParty{
					"example.com": {
						Origins: []*url.URL{mustParseURL("https://example.com")},
					},
				},
			},
			mustParseURL("https://example.com"),
			"",
			func(t *testing.T, provider *webauthn.Provider) {
				assert.Equal(t, []string{"https://example.com"}, provider.WebAuthn.Config.RPOrigins)
			},
		},
		{
			"ShouldSetResidentKeyNotRequiredWhenDiscouraged",
			schema.WebAuthn{
				WebAuthnBase: schema.WebAuthnBase{
					DisplayName:          "Authelia",
					Timeout:              time.Second * 60,
					ConveyancePreference: protocol.PreferIndirectAttestation,
					SelectionCriteria: schema.WebAuthnSelectionCriteria{
						Discoverability:  protocol.ResidentKeyRequirementDiscouraged,
						UserVerification: protocol.VerificationPreferred,
					},
				},
			},
			mustParseURL("https://auth.example.com"),
			"",
			func(t *testing.T, provider *webauthn.Provider) {
				require.NotNil(t, provider.WebAuthn.Config.AuthenticatorSelection.RequireResidentKey)
				assert.False(t, *provider.WebAuthn.Config.AuthenticatorSelection.RequireResidentKey)
			},
		},
		{
			"ShouldErrorWhenDisabled",
			schema.WebAuthn{
				Disable: true,
			},
			mustParseURL("https://auth.example.com"),
			"webauthn is disabled",
			nil,
		},
		{
			"ShouldErrorWhenOriginError",
			schema.WebAuthn{
				WebAuthnBase: schema.WebAuthnBase{
					DisplayName: "Authelia",
					Timeout:     time.Second * 60,
				},
			},
			nil,
			"error occurred determining the origin for the request: no origin",
			nil,
		},
		{
			"ShouldErrorWhenNoRelyingPartyMatchesOrigin",
			schema.WebAuthn{
				WebAuthnBase: schema.WebAuthnBase{
					DisplayName: "Authelia",
					Timeout:     time.Second * 60,
				},
				RelyingParties: map[string]schema.WebAuthnRelyingParty{
					"example.com": {
						Origins: []*url.URL{mustParseURL("https://example.com")},
					},
				},
			},
			mustParseURL("https://other.com"),
			"error occurred finding the relying party: no related origin found for origin 'https://other.com'",
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider, err := webauthn.NewProvider(&testWebAuthnContext{config: &schema.Configuration{WebAuthn: tc.config}, origin: tc.origin})

			if tc.err != "" {
				assert.Nil(t, provider)
				assert.EqualError(t, err, tc.err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, provider)

			assert.Equal(t, tc.config.Disable, provider.Config.Disable)
			assert.Equal(t, tc.config.EnablePasskeyLogin, provider.Config.EnablePasskeyLogin)
			assert.Equal(t, tc.config.EnablePasskey2FA, provider.Config.EnablePasskey2FA)
			assert.Equal(t, tc.config.EnablePasskeyUpgrade, provider.Config.EnablePasskeyUpgrade)

			if tc.expected != nil {
				tc.expected(t, provider)
			}
		})
	}
}

func TestNewProviderConfig(t *testing.T) {
	testCases := []struct {
		name     string
		config   schema.WebAuthn
		base     schema.WebAuthnBase
		expected webauthn.ProviderConfig
	}{
		{
			"ShouldMapGlobalOptionsAndBase",
			schema.WebAuthn{
				Disable:              true,
				EnablePasskeyLogin:   true,
				EnablePasskey2FA:     true,
				EnablePasskeyUpgrade: true,
			},
			schema.WebAuthnBase{
				DisplayName: "Example",
				Timeout:     time.Second * 30,
			},
			webauthn.ProviderConfig{
				Disable:              true,
				EnablePasskeyLogin:   true,
				EnablePasskey2FA:     true,
				EnablePasskeyUpgrade: true,
				WebAuthnBase: schema.WebAuthnBase{
					DisplayName: "Example",
					Timeout:     time.Second * 30,
				},
			},
		},
		{
			"ShouldMapZeroValues",
			schema.WebAuthn{},
			schema.WebAuthnBase{},
			webauthn.ProviderConfig{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, webauthn.NewProviderConfig(tc.config, tc.base))
		})
	}
}

type testWebAuthnContext struct {
	context.Context

	config *schema.Configuration
	origin *url.URL
}

func (c *testWebAuthnContext) GetOrigin() (origin *url.URL, err error) {
	if c.origin == nil {
		return nil, fmt.Errorf("no origin")
	}

	return c.origin, nil
}

func (c *testWebAuthnContext) GetConfiguration() (config *schema.Configuration) {
	return c.config
}

func (c *testWebAuthnContext) GetWebAuthnMetaDataProvider() webauthn.MetaDataProvider {
	return nil
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return u
}
