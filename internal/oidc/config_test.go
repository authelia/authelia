// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/compose"
	"authelia.com/provider/oauth2/handler/idjag"
	"authelia.com/provider/oauth2/handler/oauth2"
	"authelia.com/provider/oauth2/handler/oidckb"
	"authelia.com/provider/oauth2/handler/rfc8628"
	"authelia.com/provider/oauth2/handler/rfc8693"
	"authelia.com/provider/oauth2/handler/rfc9449"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestConfig_GetAllowedPrompts(t *testing.T) {
	ctx := context.Background()

	config := &oidc.Config{}

	assert.Equal(t, []string(nil), config.AllowedPrompts)
	assert.Equal(t, []string{oidc.PromptNone, oidc.PromptLogin, oidc.PromptConsent, oidc.PromptSelectAccount}, config.GetAllowedPrompts(ctx))
	assert.Equal(t, []string{oidc.PromptNone, oidc.PromptLogin, oidc.PromptConsent, oidc.PromptSelectAccount}, config.AllowedPrompts)

	config.AllowedPrompts = []string{oidc.PromptNone}
	assert.Equal(t, []string{oidc.PromptNone}, config.AllowedPrompts)
}

func TestConfig_PKCE(t *testing.T) {
	ctx := context.Background()

	config := &oidc.Config{}

	assert.False(t, config.GetEnforcePKCE(ctx))
	assert.True(t, config.GetAllowPublicClientsWithoutPKCE(ctx))

	config.ProofKeyCodeExchange.Enforce = true
	assert.True(t, config.GetEnforcePKCE(ctx))
	assert.False(t, config.GetAllowPublicClientsWithoutPKCE(ctx))

	config.ProofKeyCodeExchange.Enforce = false

	assert.True(t, config.GetAllowPublicClientsWithoutPKCE(ctx))

	config.ProofKeyCodeExchange.EnforcePublicClients = true

	assert.False(t, config.GetAllowPublicClientsWithoutPKCE(ctx))

	assert.False(t, config.GetEnablePKCEPlainChallengeMethod(ctx))
	config.ProofKeyCodeExchange.AllowPlainChallengeMethod = true

	assert.True(t, config.GetEnablePKCEPlainChallengeMethod(ctx))
}

func TestConfig_GrantTypeJWTBearer(t *testing.T) {
	ctx := context.Background()

	config := &oidc.Config{}
	assert.False(t, config.GetGrantTypeJWTBearerIDOptional(ctx))
	assert.False(t, config.GetGrantTypeJWTBearerCanSkipClientAuth(ctx))
	assert.False(t, config.GetGrantTypeJWTBearerIssuedDateOptional(ctx))

	config.GrantTypeJWTBearer.OptionalJTIClaim = true
	assert.True(t, config.GetGrantTypeJWTBearerIDOptional(ctx))
	assert.False(t, config.GetGrantTypeJWTBearerCanSkipClientAuth(ctx))
	assert.False(t, config.GetGrantTypeJWTBearerIssuedDateOptional(ctx))

	config.GrantTypeJWTBearer.OptionalClientAuth = true
	assert.True(t, config.GetGrantTypeJWTBearerIDOptional(ctx))
	assert.True(t, config.GetGrantTypeJWTBearerCanSkipClientAuth(ctx))
	assert.False(t, config.GetGrantTypeJWTBearerIssuedDateOptional(ctx))

	config.GrantTypeJWTBearer.OptionalIssuedDate = true
	assert.True(t, config.GetGrantTypeJWTBearerIDOptional(ctx))
	assert.True(t, config.GetGrantTypeJWTBearerCanSkipClientAuth(ctx))
	assert.True(t, config.GetGrantTypeJWTBearerIssuedDateOptional(ctx))
}

func TestConfig_Durations(t *testing.T) {
	ctx := context.Background()

	config := &oidc.Config{}
	assert.Equal(t, time.Duration(0), config.JWTMaxDuration)
	assert.Equal(t, time.Hour*24, config.GetJWTMaxDuration(ctx))
	assert.Equal(t, time.Hour*24, config.JWTMaxDuration)

	assert.Equal(t, time.Duration(0), config.Lifespans.IDToken)
	assert.Equal(t, time.Hour, config.GetIDTokenLifespan(ctx))
	assert.Equal(t, time.Hour, config.Lifespans.IDToken)

	assert.Equal(t, time.Duration(0), config.Lifespans.AccessToken)
	assert.Equal(t, time.Hour, config.GetAccessTokenLifespan(ctx))
	assert.Equal(t, time.Hour, config.Lifespans.AccessToken)

	assert.Equal(t, time.Duration(0), config.Lifespans.RefreshToken)
	assert.Equal(t, time.Hour*24*30, config.GetRefreshTokenLifespan(ctx))
	assert.Equal(t, time.Hour*24*30, config.Lifespans.RefreshToken)

	assert.Equal(t, time.Duration(0), config.GetRefreshTokenMaximumLifespan(ctx))

	config.Lifespans.RefreshTokenMaximum = time.Hour * 24 * 90
	assert.Equal(t, time.Hour*24*90, config.GetRefreshTokenMaximumLifespan(ctx))

	assert.Equal(t, time.Duration(0), config.GetRefreshTokenGracePeriod(ctx))

	config.Lifespans.RefreshTokenGracePeriod = time.Second * 10
	assert.Equal(t, time.Second*10, config.GetRefreshTokenGracePeriod(ctx))

	assert.Equal(t, time.Duration(0), config.Lifespans.AuthorizeCode)
	assert.Equal(t, time.Minute*15, config.GetAuthorizeCodeLifespan(ctx))
	assert.Equal(t, time.Minute*15, config.Lifespans.AuthorizeCode)

	assert.Equal(t, time.Second*10, config.GetJWTClockSkew(ctx))

	config.JWTClockSkew = time.Second * 30
	assert.Equal(t, time.Second*30, config.GetJWTClockSkew(ctx))

	config.JWTClockSkew = time.Hour
	assert.Equal(t, time.Minute, config.GetJWTClockSkew(ctx))

	config.JWTClockSkew = -1
	assert.Equal(t, time.Duration(0), config.GetJWTClockSkew(ctx))

	assert.Equal(t, time.Hour, config.GetRequestObjectMaximumLifetime(ctx))

	config.RequestObjectMaximumLifetime = time.Minute * 30
	assert.Equal(t, time.Minute*30, config.GetRequestObjectMaximumLifetime(ctx))

	config.RequestObjectMaximumLifetime = -1
	assert.Equal(t, time.Duration(0), config.GetRequestObjectMaximumLifetime(ctx))
}

func TestConfig_GetTokenEntropy(t *testing.T) {
	ctx := context.Background()

	config := &oidc.Config{}

	assert.Equal(t, 0, config.TokenEntropy)
	assert.Equal(t, 32, config.GetTokenEntropy(ctx))
	assert.Equal(t, 32, config.TokenEntropy)
}

func TestConfig_Misc(t *testing.T) {
	ctx := context.Background()

	config := &oidc.Config{}

	assert.False(t, config.DisableRefreshTokenValidation)
	assert.False(t, config.GetDisableRefreshTokenValidation(ctx))

	assert.False(t, config.GetDisableRefreshTokenRotation(ctx))

	config.DisableRefreshTokenRotation = true
	assert.True(t, config.GetDisableRefreshTokenRotation(ctx))

	assert.Equal(t, time.Duration(0), config.GetRefreshTokenMaximumLifespan(ctx))
	assert.False(t, config.GetDisableRefreshTokenExpirationParameters(ctx))

	assert.False(t, config.GetRequireRequestObjectAudienceAndLifetime(ctx))

	config.RequireRequestObjectAudienceAndLifetime = true
	assert.True(t, config.GetRequireRequestObjectAudienceAndLifetime(ctx))

	assert.Equal(t, "", config.Issuer)
	assert.Equal(t, "", config.GetIssuer(ctx))

	assert.Equal(t, jwt.JWTScopeFieldUnset, config.JWTScopeField)
	assert.Equal(t, jwt.JWTScopeFieldList, config.GetJWTScopeField(ctx))
	assert.Equal(t, jwt.JWTScopeFieldList, config.JWTScopeField)

	assert.Equal(t, []string(nil), config.SanitationWhiteList)
	assert.Equal(t, []string(nil), config.GetSanitationWhiteList(ctx))
	assert.Equal(t, []string(nil), config.SanitationWhiteList)

	assert.False(t, config.OmitRedirectScopeParameter)
	assert.False(t, config.GetOmitRedirectScopeParam(ctx))

	assert.Nil(t, config.GetRedirectSecureChecker(ctx))
	assert.NotNil(t, config.GetHTTPClient(ctx))

	assert.Nil(t, config.Strategy.Scope)
	assert.NotNil(t, config.GetScopeStrategy(ctx))
	assert.NotNil(t, config.Strategy.Scope)

	assert.Nil(t, config.Strategy.Audience)
	assert.NotNil(t, config.GetAudienceStrategy(ctx))
	assert.NotNil(t, config.Strategy.Audience)

	assert.Equal(t, []string(nil), config.RefreshTokenScopes)
	assert.Equal(t, []string{oidc.ScopeOffline, oidc.ScopeOfflineAccess}, config.GetRefreshTokenScopes(ctx))
	assert.Equal(t, []string{oidc.ScopeOffline, oidc.ScopeOfflineAccess}, config.RefreshTokenScopes)

	assert.Equal(t, 0, config.MinParameterEntropy)
	assert.Equal(t, 8, config.GetMinParameterEntropy(ctx))
	assert.Equal(t, 8, config.MinParameterEntropy)

	assert.False(t, config.SendDebugMessagesToClients)
	assert.False(t, config.GetSendDebugMessagesToClients(ctx))

	config.SendDebugMessagesToClients = true

	assert.True(t, config.GetSendDebugMessagesToClients(ctx))

	assert.Nil(t, config.Strategy.JWKSFetcher)
	assert.NotNil(t, config.GetJWKSFetcherStrategy(ctx))
	assert.NotNil(t, config.Strategy.JWKSFetcher)

	assert.Nil(t, config.Strategy.ClientAuthentication)
	assert.Nil(t, config.GetClientAuthenticationStrategy(ctx))

	assert.Nil(t, config.MessageCatalog)
	assert.Nil(t, config.GetMessageCatalog(ctx))

	assert.Nil(t, config.Templates)
	assert.Nil(t, config.GetFormPostHTMLTemplate(ctx))

	var err error

	config.Templates, err = templates.New(templates.Config{})
	require.NoError(t, err)

	assert.NotNil(t, config.GetFormPostHTMLTemplate(ctx))
	assert.NotNil(t, config.Templates)

	assert.False(t, config.GetUseLegacyErrorFormat(ctx))

	assert.Nil(t, config.GetAuthorizeEndpointHandlers(ctx))
	assert.Nil(t, config.GetTokenEndpointHandlers(ctx))
	assert.Nil(t, config.GetTokenIntrospectionHandlers(ctx))
	assert.Nil(t, config.GetRevocationHandlers(ctx))
	assert.Nil(t, config.GetPushedAuthorizeEndpointHandlers(ctx))

	assert.Equal(t, []string(nil), config.GetAllowedJWTAssertionAudiences(ctx))

	var octx context.Context

	octx = &TestContext{
		Context: ctx,
		IssuerURLFunc: func() (issuerURL *url.URL, err error) {
			return nil, fmt.Errorf("test error")
		},
	}

	octx = context.WithValue(octx, model.CtxKeyAutheliaCtx, octx)

	assert.Equal(t, []string(nil), config.GetAllowedJWTAssertionAudiences(octx))
}

func TestConfig_PAR(t *testing.T) {
	ctx := context.Background()

	config := &oidc.Config{}

	assert.Equal(t, "", config.PAR.URIPrefix)
	assert.Equal(t, "urn:ietf:params:oauth:request_uri:", config.GetPushedAuthorizeRequestURIPrefix(ctx))
	assert.Equal(t, "urn:ietf:params:oauth:request_uri:", config.PAR.URIPrefix)

	assert.False(t, config.PAR.Require)
	assert.False(t, config.GetRequirePushedAuthorizationRequests(ctx))
	assert.False(t, config.PAR.Require)

	config.PAR.Require = true

	assert.True(t, config.GetRequirePushedAuthorizationRequests(ctx))

	assert.False(t, config.GetRequireRedirectURIPushedAuthorizationRequests(ctx))

	config.PAR.RequireRedirectURI = true

	assert.True(t, config.GetRequireRedirectURIPushedAuthorizationRequests(ctx))

	assert.Equal(t, time.Duration(0), config.PAR.ContextLifespan)
	assert.Equal(t, time.Minute*5, config.GetPushedAuthorizeContextLifespan(ctx))
	assert.Equal(t, time.Minute*5, config.PAR.ContextLifespan)
}

func TestConfig_RequireSignedRequestObject(t *testing.T) {
	ctx := context.Background()

	config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{}, nil, nil)

	assert.False(t, config.GetRequireSignedRequestObject(ctx))
	assert.False(t, config.GetRequireSignedRequestObjectSkipPushedAuthorizationRequests(ctx))

	config = oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{RequireSignedRequestObject: true}, nil, nil)

	assert.True(t, config.GetRequireSignedRequestObject(ctx))
	assert.False(t, config.GetRequireSignedRequestObjectSkipPushedAuthorizationRequests(ctx))
}

func TestNewConfig(t *testing.T) {
	c := &schema.IdentityProvidersOpenIDConnect{
		Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
			JWTResponseAccessTokens: true,
		},
	}

	tmpl, err := templates.New(templates.Config{})

	require.NoError(t, err)

	strategy := oidc.NewIssuer(c.JSONWebKeys)

	config := oidc.NewConfig(c, strategy, tmpl)

	assert.IsType(t, &oauth2.JWTProfileCoreStrategy{}, config.Strategy.Core)

	config.LoadHandlers(nil)

	assert.Len(t, config.Handlers.TokenIntrospection, 1)

	config.JWTAccessToken.EnableStatelessIntrospection = true

	config.LoadHandlers(nil)

	assert.Len(t, config.Handlers.TokenIntrospection, 2)
}

func TestConfig_LoadHandlersBinding(t *testing.T) {
	testCases := []struct {
		name          string
		dpop          bool
		keyBinding    bool
		expected      int
		expectedToken int
	}{
		{
			"ShouldRegisterBindingHandlersWhenDPoPEnabled",
			true,
			false,
			1,
			2,
		},
		{
			"ShouldRegisterKeyBindingHandlersWhenBothEnabled",
			true,
			true,
			2,
			2,
		},
		{
			"ShouldOnlyRegisterRejectingTokenBindingHandlerWhenDPoPDisabled",
			false,
			false,
			0,
			1,
		},
		{
			"ShouldNotRegisterKeyBindingHandlersWhenDPoPDisabled",
			false,
			true,
			0,
			1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			config := &oidc.Config{}
			config.DPoP.Enabled = tc.dpop
			config.DPoP.KeyBinding = tc.keyBinding

			config.LoadHandlers(&oidc.Store{})

			assert.Len(t, config.GetAuthorizeEndpointBindingHandlers(ctx), tc.expected)
			assert.Len(t, config.GetRFC8628DeviceAuthorizeEndpointBindingHandlers(ctx), tc.expected)
			assert.Len(t, config.GetTokenEndpointBindingHandlers(ctx), tc.expectedToken)

			// The device flow records the 'bound_key' grant at the user authorization endpoint rather than a binding
			// list, as that is where consent has decided it. It is registered from the same condition, so a key
			// binding deployment which never reaches it issues device flow ID Tokens with no confirmation.
			var userAuthorize bool

			for _, handler := range config.GetRFC8628UserAuthorizeEndpointHandlers(ctx) {
				if _, ok := handler.(*oidckb.UserAuthorizeHandler); ok {
					userAuthorize = true
				}
			}

			assert.Equal(t, tc.dpop && tc.keyBinding, userAuthorize)
		})
	}
}

func TestConfig_GetDPoPProofLifespan(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name     string
		have     time.Duration
		expected time.Duration
	}{
		{"ShouldUseTheConfiguredValue", time.Second * 45, time.Second * 45},
		{"ShouldFallBackWhenUnset", 0, time.Second * 10},
		{"ShouldFallBackWhenNegative", -time.Second, time.Second * 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{}
			config.DPoP.ProofLifespan = tc.have

			assert.Equal(t, tc.expected, config.GetDPoPProofLifespan(ctx))
		})
	}
}

func TestDPoPSigningAlgValuesSupported(t *testing.T) {
	ctx := context.Background()

	algs := oidc.DPoPSigningAlgValuesSupported()

	assert.NotEmpty(t, algs)

	// A proof is signed by a key the client holds, which a shared secret cannot demonstrate possession of.
	for _, alg := range algs {
		assert.NotContains(t, []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512}, alg)
	}

	// The advertised value and the value proofs are validated against must not be able to disagree.
	config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{
		DPoP: schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true},
	}, nil, nil)

	assert.Equal(t, algs, config.GetDPoPAllowedJWSAlgorithms(ctx))
}

func TestConfig_GetDPoPAllowedJWSAlgorithmsConfigured(t *testing.T) {
	ctx := context.Background()

	config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{
		DPoP: schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, AllowedAlgorithms: []string{oidc.SigningAlgRSAPSSUsingSHA256, oidc.SigningAlgECDSAUsingP256AndSHA256}},
	}, nil, nil)

	assert.Equal(t, []string{oidc.SigningAlgRSAPSSUsingSHA256, oidc.SigningAlgECDSAUsingP256AndSHA256}, config.GetDPoPAllowedJWSAlgorithms(ctx))
}

func TestConfig_GetDPoPStrictRefreshTokenBinding(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name     string
		have     bool
		expected bool
	}{
		{"ShouldReturnFalseByDefault", false, false},
		{"ShouldReturnTrueWhenConfigured", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{
				DPoP: schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, StrictRefreshTokenBinding: tc.have},
			}, nil, nil)

			assert.Equal(t, tc.expected, config.GetDPoPStrictRefreshTokenBinding(ctx))
		})
	}
}

func TestConfig_GetDPoPNonceStateless(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name     string
		have     bool
		expected bool
	}{
		{"ShouldReturnFalseByDefault", false, false},
		{"ShouldReturnTrueWhenConfigured", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{
				DPoP: schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, NonceStateless: tc.have},
			}, nil, nil)

			assert.Equal(t, tc.expected, config.GetDPoPNonceStateless(ctx))
		})
	}
}

func TestNewClientDPoPStrictRefreshTokenBinding(t *testing.T) {
	testCases := []struct {
		name     string
		have     bool
		expected bool
	}{
		{"ShouldReturnFalseByDefault", false, false},
		{"ShouldReturnTrueWhenConfigured", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{ID: "test", DPoPStrictRefreshTokenBinding: tc.have}, &schema.IdentityProvidersOpenIDConnect{}, nil)

			registered, ok := client.(*oidc.RegisteredClient)
			require.True(t, ok)

			assert.Equal(t, tc.expected, registered.GetDPoPStrictRefreshTokenBinding())
		})
	}
}

func TestConfig_LoadHandlersKeyBindingOrder(t *testing.T) {
	ctx := context.Background()

	config := &oidc.Config{}
	config.DPoP.Enabled = true
	config.DPoP.KeyBinding = true

	config.LoadHandlers(&oidc.Store{})

	dpop, binding := -1, -1

	for i, handler := range config.GetTokenEndpointBindingHandlers(ctx) {
		switch handler.(type) {
		case *rfc9449.Handler:
			dpop = i
		case *oidckb.Handler:
			binding = i
		}
	}

	require.NotEqual(t, -1, dpop)
	require.NotEqual(t, -1, binding)
	assert.Less(t, dpop, binding)

	device, granted := -1, -1

	for i, handler := range config.Handlers.RFC8628UserAuthorizeEndpoint {
		switch handler.(type) {
		case *rfc8628.UserAuthorizeHandler:
			device = i
		case *oidckb.UserAuthorizeHandler:
			granted = i
		}
	}

	require.NotEqual(t, -1, device)
	require.NotEqual(t, -1, granted)
	assert.Less(t, granted, device)
}

func TestConfig_GetOIDCKeyBindingEnabled(t *testing.T) {
	testCases := []struct {
		name     string
		have     bool
		expected bool
	}{
		{
			"ShouldReturnFalseByDefault",
			false,
			false,
		},
		{
			"ShouldReturnTrueWhenEnabled",
			true,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{}
			config.DPoP.KeyBinding = tc.have

			assert.Equal(t, tc.expected, config.GetOIDCKeyBindingEnabled(context.Background()))
		})
	}
}

func TestConfig_LoadHandlersOrder(t *testing.T) {
	testCases := []struct {
		name                         string
		enableStatelessIntrospection bool
	}{
		{
			"ShouldRegisterHandlersInValidOrder",
			false,
		},
		{
			"ShouldRegisterHandlersInValidOrderWithStatelessIntrospection",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := templates.New(templates.Config{})

			require.NoError(t, err)

			config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{}, oidc.NewIssuer(nil), tmpl)

			config.JWTAccessToken.EnableStatelessIntrospection = tc.enableStatelessIntrospection

			config.LoadHandlers(nil)

			assert.NoError(t, compose.ValidateHandlerOrder(&oauthelia2.Config{
				AuthorizeEndpointHandlers:              config.Handlers.AuthorizeEndpoint,
				TokenEndpointHandlers:                  config.Handlers.TokenEndpoint,
				TokenIntrospectionHandlers:             config.Handlers.TokenIntrospection,
				RevocationHandlers:                     config.Handlers.Revocation,
				PushedAuthorizeEndpointHandlers:        config.Handlers.PushedAuthorizeEndpoint,
				RFC8628DeviceAuthorizeEndpointHandlers: config.Handlers.RFC8628DeviceAuthorizeEndpoint,
				RFC8628UserAuthorizeEndpointHandlers:   config.Handlers.RFC8628UserAuthorizeEndpoint,
				AuthorizeEndpointBindingHandlers:       config.Handlers.AuthorizeEndpointBinding,
				TokenEndpointBindingHandlers:           config.Handlers.TokenEndpointBinding,
			}))
		})
	}
}

func TestNewConfigSecurityOptions(t *testing.T) {
	_, loopback, err := net.ParseCIDR("127.0.0.0/8")

	require.NoError(t, err)

	testCases := []struct {
		name     string
		have     *schema.IdentityProvidersOpenIDConnect
		expected func(t *testing.T, ctx context.Context, config *oidc.Config)
	}{
		{
			"ShouldDefaultToSecureValues",
			&schema.IdentityProvidersOpenIDConnect{},
			func(t *testing.T, ctx context.Context, config *oidc.Config) {
				assert.False(t, config.GetEnableRSA15KeyManagement(ctx))
				assert.False(t, config.GetAccessTokenURIQueryParameterEnabled(ctx))
				assert.False(t, config.GetRFC8628VerificationURIComplete(ctx))
				assert.False(t, config.GetDisableSpecialUseAddressGuard(ctx))
				assert.Nil(t, config.GetSpecialUseAddressGuardAllowedNetworks(ctx))
				assert.False(t, config.GetRFC8628UserCodeSubjectBinding(ctx))
				assert.Equal(t, time.Duration(0), config.GetRefreshTokenGracePeriod(ctx))
			},
		},
		{
			"ShouldMapEnabledValues",
			&schema.IdentityProvidersOpenIDConnect{
				EnableRSA15KeyManagement:                true,
				EnableAccessTokenURIQueryParameter:      true,
				EnableDeviceCodeVerificationURIComplete: true,
				SpecialUseAddressGuard: schema.IdentityProvidersOpenIDConnectSpecialUseAddressGuard{
					Disable:         true,
					AllowedNetworks: []*net.IPNet{loopback},
				},
			},
			func(t *testing.T, ctx context.Context, config *oidc.Config) {
				assert.True(t, config.GetEnableRSA15KeyManagement(ctx))
				assert.True(t, config.GetAccessTokenURIQueryParameterEnabled(ctx))
				assert.True(t, config.GetRFC8628VerificationURIComplete(ctx))
				assert.True(t, config.GetDisableSpecialUseAddressGuard(ctx))
				assert.Equal(t, []*net.IPNet{loopback}, config.GetSpecialUseAddressGuardAllowedNetworks(ctx))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := templates.New(templates.Config{})

			require.NoError(t, err)

			tc.expected(t, context.Background(), oidc.NewConfig(tc.have, oidc.NewIssuer(nil), tmpl))
		})
	}
}

func TestConfig_GetClientJWKSFetcherStrategy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"keys":[]}`))
	}))

	t.Cleanup(server.Close)

	_, loopback, err := net.ParseCIDR("127.0.0.0/8")

	require.NoError(t, err)

	_, other, err := net.ParseCIDR("10.0.0.0/8")

	require.NoError(t, err)

	testCases := []struct {
		name  string
		guard schema.IdentityProvidersOpenIDConnectSpecialUseAddressGuard
		err   bool
	}{
		{
			"ShouldRefuseLoopbackByDefault",
			schema.IdentityProvidersOpenIDConnectSpecialUseAddressGuard{},
			true,
		},
		{
			"ShouldRefuseLoopbackWhenNotAllowed",
			schema.IdentityProvidersOpenIDConnectSpecialUseAddressGuard{AllowedNetworks: []*net.IPNet{other}},
			true,
		},
		{
			"ShouldAllowLoopbackWhenAllowed",
			schema.IdentityProvidersOpenIDConnectSpecialUseAddressGuard{AllowedNetworks: []*net.IPNet{loopback}},
			false,
		},
		{
			"ShouldAllowLoopbackWhenDisabled",
			schema.IdentityProvidersOpenIDConnectSpecialUseAddressGuard{Disable: true},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := templates.New(templates.Config{})

			require.NoError(t, err)

			config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{SpecialUseAddressGuard: tc.guard}, oidc.NewIssuer(nil), tmpl)

			ctx := context.Background()

			jwks, err := config.GetClientJWKSFetcherStrategy(ctx, nil).Resolve(ctx, server.URL, true)

			if tc.err {
				assert.Error(t, err)
				assert.Nil(t, jwks)
			} else {
				require.NoError(t, err)
				require.NotNil(t, jwks)
				assert.Len(t, jwks.Keys, 0)
			}
		})
	}
}

func TestConfig_ValidateRedirectURISecurity(t *testing.T) {
	testCases := []struct {
		name            string
		applicationType string
		public          bool
		uri             string
		err             string
	}{
		{
			"ShouldAllowHTTPSWeb",
			oidc.ApplicationTypeWeb,
			false,
			"https://app.example.com/callback",
			"",
		},
		{
			"ShouldAllowLoopbackNative",
			oidc.ApplicationTypeNative,
			true,
			"http://127.0.0.1:8080/callback",
			"",
		},
		{
			"ShouldRejectLoopbackWebConfidential",
			oidc.ApplicationTypeWeb,
			false,
			"http://127.0.0.1:8080/callback",
			"Redirect URL is using an insecure protocol, http is only allowed for native clients using a loopback redirect URI, for example: http://127.0.0.1:8080/callback.",
		},
		{
			"ShouldRejectLocalhostWebPublic",
			oidc.ApplicationTypeWeb,
			true,
			"http://localhost:8080/callback",
			"Redirect URL is using an insecure protocol, http is only allowed for native clients using a loopback redirect URI, for example: http://127.0.0.1:8080/callback.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := templates.New(templates.Config{})

			require.NoError(t, err)

			config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{}, oidc.NewIssuer(nil), tmpl)

			uri, err := url.Parse(tc.uri)

			require.NoError(t, err)

			client := &oidc.RegisteredClient{ID: "test", ApplicationType: tc.applicationType, Public: tc.public}

			err = oauthelia2.ValidateRedirectURISecurity(context.Background(), config, client, uri)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, "invalid_request")
				assert.Equal(t, tc.err, oauthelia2.ErrorToRFC6749Error(err).HintField)
			}
		})
	}
}

func TestConfig_GetIssuer(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		ctx      context.Context
		expected string
	}{
		{
			"ShouldReturnCtxValue",
			"",
			&TestContext{
				Context: context.Background(),
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return &url.URL{Scheme: "https", Host: "example.com", Path: "/issuer"}, nil
				},
			},
			"https://example.com/issuer",
		},
		{
			"ShouldNotReturnFallbackValue",
			"https://example.com/fallback-issuer",
			&TestContext{
				Context: context.Background(),
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return &url.URL{Scheme: "https", Host: "example.com", Path: "/issuer"}, nil
				},
			},
			"https://example.com/issuer",
		},
		{
			"ShouldReturnFallbackValue",
			"https://example.com/fallback-issuer",
			context.Background(),
			"https://example.com/fallback-issuer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{
				Issuer: tc.have,
			}

			ctx := context.WithValue(tc.ctx, 0, 1) //nolint:staticcheck // This value is used to demonstrate a functionality not used for an actual value.

			assert.Equal(t, tc.expected, config.GetIssuer(ctx))
		})
	}
}

func TestMisc(t *testing.T) {
	tctx := &TestContext{
		Context: context.Background(),
		IssuerURLFunc: func() (issuerURL *url.URL, err error) {
			return &url.URL{Scheme: "https", Host: "example.com", Path: "/issuer"}, nil
		},
	}

	config := &oidc.Config{}
	assert.Nil(t, config.GetIntrospectionJWTResponseStrategy(context.Background()))
	assert.Nil(t, config.GetJWTSecuredAuthorizeResponseModeStrategy(context.Background()))

	secret, err := config.GetGlobalSecret(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, secret)

	secrets, err := config.GetRotatedGlobalSecrets(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, secrets)

	assert.Equal(t, time.Minute*5, config.GetJWTSecuredAuthorizeResponseModeLifespan(context.Background()))

	assert.False(t, config.GetRevokeRefreshTokensExplicit(context.Background()))
	assert.False(t, config.GetEnforceRevokeFlowRevokeRefreshTokensExplicitClient(context.Background()))
	assert.False(t, config.GetClientCredentialsFlowImplicitGrantRequested(context.Background()))
	assert.False(t, config.GetEnforceJWTProfileAccessTokens(context.Background()))
	config.ClientCredentialsFlowImplicitGrantRequested = true

	assert.True(t, config.GetClientCredentialsFlowImplicitGrantRequested(context.Background()))

	assert.NotNil(t, config.GetHMACHasher(context.Background()))
	assert.NotNil(t, config.GetFormPostResponseWriter(context.Background()))

	assert.Equal(t, time.Hour, config.GetVerifiableCredentialsNonceLifespan(context.Background()))
	assert.Nil(t, config.GetResponseModeHandlers(context.Background()))
	assert.Nil(t, config.GetResponseModeParameterHandlers(context.Background()))
	assert.Nil(t, config.GetRFC8628DeviceAuthorizeEndpointHandlers(context.Background()))
	assert.Nil(t, config.GetRFC8628UserAuthorizeEndpointHandlers(context.Background()))
	assert.Nil(t, config.GetRFC8693TokenTypes(context.Background()))

	assert.Equal(t, "", config.GetDefaultRFC8693RequestedTokenType(context.Background()))
	assert.Equal(t, time.Minute*10, config.GetRFC8628CodeLifespan(context.Background()))
	assert.Equal(t, time.Second*10, config.GetRFC8628TokenPollingInterval(context.Background()))

	assert.Equal(t, []string{"https://example.com/issuer", "https://example.com/issuer/api/oidc/token", "https://example.com/issuer/api/oidc/pushed-authorization-request"}, config.GetAllowedJWTAssertionAudiences(tctx))

	assert.Equal(t, "https://example.com/issuer/consent/openid/device-authorization", config.GetRFC8628UserVerificationURL(tctx))
}

func TestConfig_GetAuthorizeErrorFieldResponseStrategy(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func() *oidc.Config
		callTwice bool
	}{
		{
			"ShouldLazilyInitializeStrategy",
			func() *oidc.Config {
				return &oidc.Config{}
			},
			false,
		},
		{
			"ShouldReturnSameStrategyOnSecondCall",
			func() *oidc.Config {
				return &oidc.Config{}
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.setup()

			ctx := context.Background()

			strategy := config.GetAuthorizeErrorFieldResponseStrategy(ctx)

			assert.NotNil(t, strategy)

			if tc.callTwice {
				strategy2 := config.GetAuthorizeErrorFieldResponseStrategy(ctx)

				assert.Equal(t, strategy, strategy2)
			}
		})
	}
}

func TestConfig_GetJWTStrategy(t *testing.T) {
	strategy := &jwt.DefaultStrategy{}

	config := &oidc.Config{
		Strategy: oidc.StrategyConfig{
			JWT: strategy,
		},
	}

	assert.Equal(t, strategy, config.GetJWTStrategy(context.Background()))
}

func TestConfig_GetContext(t *testing.T) {
	testCases := []struct {
		name     string
		ctx      context.Context
		expected bool
	}{
		{
			"ShouldReturnNilForBackgroundContext",
			context.Background(),
			false,
		},
		{
			"ShouldReturnContextFromValue",
			context.WithValue(context.Background(), model.CtxKeyAutheliaCtx, &TestContext{Context: context.Background()}),
			true,
		},
		{
			"ShouldReturnContextFromDirectTypeAssertion",
			&testConfigContext{Context: context.Background()},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{}

			result := config.GetContext(tc.ctx)

			if tc.expected {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

type testConfigContext struct {
	context.Context
}

func (t *testConfigContext) IssuerURL() (issuerURL *url.URL, err error) { return nil, nil }

func (t *testConfigContext) GetClock() clock.Provider { return nil }

func (t *testConfigContext) GetRandom() random.Provider { return nil }

func (t *testConfigContext) GetConfiguration() *schema.Configuration { return nil }

func (t *testConfigContext) GetProviderStorage() storage.Provider { return nil }

func (t *testConfigContext) GetUserProvider() authentication.UserProvider { return nil }

func (t *testConfigContext) GetProviderUserAttributeResolver() expression.UserAttributeResolver {
	return nil
}

func TestNewConfigRegistersRFC8693TokenTypes(t *testing.T) {
	c := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{}, nil, nil)

	types := c.GetRFC8693TokenTypes(context.TODO())

	require.NotNil(t, types)
	assert.Len(t, types, 3)

	for _, name := range []string{oidc.TokenTypeAccessToken, oidc.TokenTypeRefreshToken, oidc.TokenTypeIDToken} {
		tokenType, ok := types[name]

		require.True(t, ok, "token type %s must be registered", name)
		require.NotNil(t, tokenType)
		assert.Equal(t, name, tokenType.GetName(context.TODO()))
	}

	assert.NotContains(t, types, oidc.TokenTypeJWT)
	assert.Equal(t, oidc.TokenTypeAccessToken, c.GetDefaultRFC8693RequestedTokenType(context.TODO()))
}

func TestConfig_GetRFC8693MaxActorChainDepth(t *testing.T) {
	tmpl, err := templates.New(templates.Config{})

	require.NoError(t, err)

	c := &schema.IdentityProvidersOpenIDConnect{TokenExchange: schema.IdentityProvidersOpenIDConnectTokenExchange{MaxActorChainDepth: 3}}

	assert.Equal(t, 3, oidc.NewConfig(c, oidc.NewIssuer(c.JSONWebKeys), tmpl).GetRFC8693MaxActorChainDepth(context.Background()))

	c = &schema.IdentityProvidersOpenIDConnect{}

	assert.Equal(t, 8, oidc.NewConfig(c, oidc.NewIssuer(c.JSONWebKeys), tmpl).GetRFC8693MaxActorChainDepth(context.Background()))
}

func TestNewConfigTransactionTokens(t *testing.T) {
	ctx := context.Background()

	c := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{
		TokenExchange: schema.IdentityProvidersOpenIDConnectTokenExchange{
			TokenTypes:                []string{oidc.TokenTypeAccessToken, oidc.TokenTypeIDToken},
			DefaultRequestedTokenType: oidc.TokenTypeIDToken,
		},
	}, oidc.NewIssuer(nil), nil)

	assert.Equal(t, oidc.TokenTypeIDToken, c.GetDefaultRFC8693RequestedTokenType(ctx))
	assert.Len(t, c.GetRFC8693TokenTypes(ctx), 2)

	c.LoadHandlers(&oidc.Store{})

	for _, h := range c.Handlers.TokenEndpoint {
		_, isTxn := h.(*rfc8693.TxnTokenTypeHandler)

		assert.False(t, isTxn)
	}

	c = oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{
		TokenExchange: schema.IdentityProvidersOpenIDConnectTokenExchange{
			TokenTypes:                []string{oidc.TokenTypeAccessToken, oidc.TokenTypeTxnToken},
			DefaultRequestedTokenType: oidc.TokenTypeAccessToken,
			TransactionTokens:         schema.IdentityProvidersOpenIDConnectTransactionTokens{TrustDomain: "example.com"},
		},
	}, oidc.NewIssuer(nil), nil)

	c.LoadHandlers(&oidc.Store{})

	var found bool

	for _, h := range c.Handlers.TokenEndpoint {
		if _, ok := h.(*rfc8693.TxnTokenTypeHandler); ok {
			found = true
		}
	}

	assert.True(t, found)
}

func TestNewConfigIDJAGHandlerOrder(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name    string
		listed  bool
		ras     bool
		trusted bool
		dpop    bool
		txn     bool
	}{
		{"ShouldNotRegisterWhenUnconfigured", false, false, false, false, false},
		{"ShouldNotRegisterIssueWhenNotListed", false, true, false, false, false},
		{"ShouldRegisterIssueOnly", true, true, false, false, false},
		{"ShouldRegisterIssueWithTxnToken", true, true, false, false, true},
		{"ShouldRegisterRedeemOnly", false, false, true, false, false},
		{"ShouldRegisterBothWithDPoP", true, true, true, true, false},
		{"ShouldRegisterBothWithoutDPoP", true, true, true, false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{}
			config.DPoP.Enabled = tc.dpop
			config.IDJAG = &oidc.IDJAGConfig{
				ResourceAuthorizationServers: map[string]*oidc.IDJAGResourceAuthorizationServer{},
				TrustedIssuers:               map[string]*oidc.IDJAGTrustedIssuer{},
			}

			if tc.ras {
				config.IDJAG.ResourceAuthorizationServers["https://ras.example.com"] = &oidc.IDJAGResourceAuthorizationServer{}
			}

			if tc.trusted {
				config.IDJAG.TrustedIssuers["https://idp.example.com"] = &oidc.IDJAGTrustedIssuer{}
			}

			te := &schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{oidc.TokenTypeAccessToken, oidc.TokenTypeRefreshToken, oidc.TokenTypeIDToken}}

			if tc.listed {
				te.TokenTypes = append(te.TokenTypes, oidc.TokenTypeIDJAG)
			}

			if tc.txn {
				te.TokenTypes = append(te.TokenTypes, oidc.TokenTypeTxnToken)
				te.TransactionTokens.TrustDomain = "example.com"
			}

			config.RFC8693.TokenTypes = oidc.NewRFC8693TokenTypes(te, nil)

			config.LoadHandlers(&oidc.Store{})

			handlers, binding := config.GetTokenEndpointHandlers(ctx), config.GetTokenEndpointBindingHandlers(ctx)

			grant := getTestHandlerIndex[*rfc8693.TokenExchangeGrantHandler](handlers)
			issue := getTestHandlerIndex[*idjag.IssueHandler](handlers)
			txn := getTestHandlerIndex[*rfc8693.TxnTokenTypeHandler](handlers)
			validator := getTestHandlerIndex[*rfc8693.ActorTokenValidationHandler](handlers)
			redeem := getTestHandlerIndex[*idjag.RedeemHandler](handlers)
			dpop := getTestHandlerIndex[*rfc9449.Handler](binding)
			bound := getTestHandlerIndex[*idjag.RedeemHandler](binding)

			_, typed := config.GetRFC8693TokenTypes(ctx)[oidc.TokenTypeIDJAG]

			assert.Equal(t, tc.listed, typed)

			if tc.txn {
				require.NotEqual(t, -1, txn)
				assert.Less(t, issue, txn)
				assert.Less(t, txn, validator)
			}

			if tc.listed {
				require.NotEqual(t, -1, grant)
				require.NotEqual(t, -1, issue)
				require.NotEqual(t, -1, validator)
				assert.Less(t, grant, issue)
				assert.Less(t, issue, validator)
			} else {
				assert.Equal(t, -1, issue)
			}

			if tc.trusted {
				require.NotEqual(t, -1, redeem)
				require.NotEqual(t, -1, bound)

				if tc.dpop {
					require.NotEqual(t, -1, dpop)
					assert.Less(t, dpop, bound)
				}
			} else {
				assert.Equal(t, -1, redeem)
				assert.Equal(t, -1, bound)
			}

			assert.NoError(t, compose.ValidateHandlerOrder(&oauthelia2.Config{
				TokenEndpointHandlers:        config.GetTokenEndpointHandlers(ctx),
				TokenEndpointBindingHandlers: config.GetTokenEndpointBindingHandlers(ctx),
				AuthorizeEndpointHandlers:    config.GetAuthorizeEndpointHandlers(ctx),
			}))
		})
	}
}

func getTestHandlerIndex[T any, H any](handlers []H) int {
	for i, handler := range handlers {
		if _, ok := any(handler).(T); ok {
			return i
		}
	}

	return -1
}
