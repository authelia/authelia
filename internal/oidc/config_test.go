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
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestConfig_GetAllowedPrompts(t *testing.T) {
	ctx := context.Background()

	config := &oidc.Config{}

	assert.Equal(t, []string(nil), config.AllowedPrompts)
	assert.Equal(t, []string{oidc.PromptNone, oidc.PromptLogin, oidc.PromptConsent}, config.GetAllowedPrompts(ctx))
	assert.Equal(t, []string{oidc.PromptNone, oidc.PromptLogin, oidc.PromptConsent}, config.AllowedPrompts)

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

	assert.Equal(t, []string{""}, config.GetTokenURLs(ctx))

	octx := &TestContext{
		Context: ctx,
		IssuerURLFunc: func() (issuerURL *url.URL, err error) {
			return nil, fmt.Errorf("test error")
		},
	}

	assert.Equal(t, []string{""}, config.GetTokenURLs(octx))
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
