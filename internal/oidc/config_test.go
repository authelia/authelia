package oidc_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"authelia.com/provider/oauth2/handler/oauth2"
	"authelia.com/provider/oauth2/token/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
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
	assert.False(t, config.GetEnforcePKCEForPublicClients(ctx))

	config.ProofKeyCodeExchange.Enforce = true
	assert.True(t, config.GetEnforcePKCE(ctx))
	assert.True(t, config.GetEnforcePKCEForPublicClients(ctx))

	config.ProofKeyCodeExchange.Enforce = false

	assert.False(t, config.GetEnforcePKCEForPublicClients(ctx))

	config.ProofKeyCodeExchange.EnforcePublicClients = true

	assert.True(t, config.GetEnforcePKCEForPublicClients(ctx))

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

	assert.Equal(t, time.Duration(0), config.Lifespans.AuthorizeCode)
	assert.Equal(t, time.Minute*15, config.GetAuthorizeCodeLifespan(ctx))
	assert.Equal(t, time.Minute*15, config.Lifespans.AuthorizeCode)
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

	assert.Equal(t, "", config.Issuers.AccessToken)
	assert.Equal(t, "", config.GetAccessTokenIssuer(ctx))

	assert.Equal(t, "", config.Issuers.IDToken)
	assert.Equal(t, "", config.GetIDTokenIssuer(ctx))

	assert.Equal(t, jwt.JWTScopeFieldUnset, config.JWTScopeField)
	assert.Equal(t, jwt.JWTScopeFieldList, config.GetJWTScopeField(ctx))
	assert.Equal(t, jwt.JWTScopeFieldList, config.JWTScopeField)

	assert.Equal(t, []string(nil), config.SanitationWhiteList)
	assert.Equal(t, []string(nil), config.GetSanitationWhiteList(ctx))
	assert.Equal(t, []string(nil), config.SanitationWhiteList)

	assert.False(t, config.OmitRedirectScopeParameter)
	assert.False(t, config.GetOmitRedirectScopeParam(ctx))

	assert.NotNil(t, config.GetRedirectSecureChecker(ctx))
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

	assert.Equal(t, time.Duration(0), config.PAR.ContextLifespan)
	assert.Equal(t, time.Minute*5, config.GetPushedAuthorizeContextLifespan(ctx))
	assert.Equal(t, time.Minute*5, config.PAR.ContextLifespan)
}

func TestNewConfig(t *testing.T) {
	c := &schema.IdentityProvidersOpenIDConnect{
		Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
			JWTResponseAccessTokens: true,
		},
	}

	tmpl, err := templates.New(templates.Config{})

	require.NoError(t, err)

	signer := oidc.NewKeyManager(c)

	config := oidc.NewConfig(c, signer, tmpl)

	assert.IsType(t, &oauth2.JWTProfileCoreStrategy{}, config.Strategy.Core)

	config.LoadHandlers(nil)

	assert.Len(t, config.Handlers.TokenIntrospection, 1)

	config.JWTAccessToken.EnableStatelessIntrospection = true

	config.LoadHandlers(nil)

	assert.Len(t, config.Handlers.TokenIntrospection, 2)
}

func TestConfig_GetIssuerFuncs(t *testing.T) {
	testCases := []struct {
		name                                                                        string
		have                                                                        oidc.IssuersConfig
		ctx                                                                         context.Context
		expectIntrospection, expectIDToken, expectAccessToken, expectAS, expectJARM string
	}{
		{
			"ShouldReturnCtxValues",
			oidc.IssuersConfig{},
			&TestContext{
				Context: context.Background(),
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return &url.URL{Scheme: "https", Host: "example.com", Path: "/issuer"}, nil
				},
			},
			"https://example.com/issuer",
			"https://example.com/issuer",
			"https://example.com/issuer",
			"https://example.com/issuer",
			"https://example.com/issuer",
		},
		{
			"ShouldNotReturnDefaultValues",
			oidc.IssuersConfig{
				IDToken: "https://example.com/id-issuer",
			},
			&TestContext{
				Context: context.Background(),
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return &url.URL{Scheme: "https", Host: "example.com", Path: "/issuer"}, nil
				},
			},
			"https://example.com/issuer",
			"https://example.com/issuer",
			"https://example.com/issuer",
			"https://example.com/issuer",
			"https://example.com/issuer",
		},
		{
			"ShouldReturnDefaultValues",
			oidc.IssuersConfig{
				IDToken:                                 "https://example.com/id-issuer",
				AccessToken:                             "https://example.com/at-issuer",
				Introspection:                           "https://example.com/i-issuer",
				JWTSecuredResponseMode:                  "https://example.com/jarm-issuer",
				AuthorizationServerIssuerIdentification: "https://example.com/as-issuer",
			},
			context.Background(),
			"https://example.com/i-issuer",
			"https://example.com/id-issuer",
			"https://example.com/at-issuer",
			"https://example.com/as-issuer",
			"https://example.com/jarm-issuer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{
				Issuers: tc.have,
			}

			ctx := context.WithValue(tc.ctx, 0, 1) //nolint:staticcheck // This value is used to demonstrate a functionality not used for an actual value.

			assert.Equal(t, tc.expectIntrospection, config.GetIntrospectionIssuer(ctx))
			assert.Equal(t, tc.expectIDToken, config.GetIDTokenIssuer(ctx))
			assert.Equal(t, tc.expectAccessToken, config.GetAccessTokenIssuer(ctx))
			assert.Equal(t, tc.expectAS, config.GetAuthorizationServerIdentificationIssuer(ctx))
			assert.Equal(t, tc.expectJARM, config.GetJWTSecuredAuthorizeResponseModeIssuer(ctx))
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
	assert.Nil(t, config.GetIntrospectionJWTResponseSigner(context.Background()))
	assert.Nil(t, config.GetJWTSecuredAuthorizeResponseModeSigner(context.Background()))

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
