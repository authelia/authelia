package oidc

import (
	"context"
	"crypto/sha512"
	"hash"
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	"authelia.com/provider/oauth2/handler/par"
	"authelia.com/provider/oauth2/handler/pkce"
	"authelia.com/provider/oauth2/handler/rfc8628"
	"authelia.com/provider/oauth2/i18n"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewConfig returns a new *Config given the identity providers configuration, issuer, and templates provider.
func NewConfig(config *schema.IdentityProvidersOpenIDConnect, issuer *Issuer, templates *templates.Provider) (c *Config) {
	c = &Config{
		GlobalSecret:               []byte(utils.HashSHA256FromString(config.HMACSecret)),
		SendDebugMessagesToClients: config.EnableClientDebugMessages,
		MinParameterEntropy:        config.MinimumParameterEntropy,
		Lifespans: LifespansConfig{
			IdentityProvidersOpenIDConnectLifespanToken: config.Lifespans.IdentityProvidersOpenIDConnectLifespanToken,
			RFC8628Code: config.Lifespans.DeviceCode,
		},
		ProofKeyCodeExchange: ProofKeyCodeExchangeConfig{
			Enforce:                   config.EnforcePKCE == "always",
			EnforcePublicClients:      config.EnforcePKCE != "never",
			AllowPlainChallengeMethod: config.EnablePKCEPlainChallenge,
		},
		PAR: PARConfig{
			Require:         config.RequirePushedAuthorizationRequests,
			ContextLifespan: 5 * time.Minute,
			URIPrefix:       RedirectURIPrefixPushedAuthorizationRequestURN,
		},
		JWTAccessToken: JWTAccessTokenConfig{
			Enable:                       config.Discovery.JWTResponseAccessTokens,
			EnableStatelessIntrospection: config.EnableJWTAccessTokenStatelessIntrospection,
		},
		Strategy:                        StrategyConfig{},
		JWTSecuredAuthorizationLifespan: config.Lifespans.JWTSecuredAuthorization,
		RevokeRefreshTokensExplicit:     true,
		EnforceRevokeFlowRevokeRefreshTokensExplicitClient: true,
		EnforceClientAssertionIssuerAudience:               false,
		ClientCredentialsFlowImplicitGrantRequested:        true,
		Templates: templates,
	}

	c.Strategy.JWT = &jwt.DefaultStrategy{
		Config: c,
		Issuer: issuer,
	}

	if config.Discovery.JWTResponseAccessTokens {
		c.Strategy.Core = oauth2.NewCoreStrategy(c, fmtAutheliaOpaqueOAuth2Token, c.Strategy.JWT)
	} else {
		c.Strategy.Core = oauth2.NewCoreStrategy(c, fmtAutheliaOpaqueOAuth2Token, nil)
	}

	c.Strategy.OpenID = &openid.DefaultStrategy{
		Strategy: c.Strategy.JWT,
		Config:   c,
	}

	return c
}

// Config is an implementation of the oauthelia2.Configurator.
type Config struct {
	// GlobalSecret is the global secret used to sign and verify signatures.
	GlobalSecret []byte

	// RotatedGlobalSecrets is a list of global secrets that are used to verify signatures.
	RotatedGlobalSecrets [][]byte

	Issuers IssuersConfig

	SendDebugMessagesToClients    bool
	DisableRefreshTokenValidation bool
	OmitRedirectScopeParameter    bool

	DPoPEnabled       bool
	DPoPEnforce       bool
	DPoPNonceRequired bool

	DPoPAllowedJWSAlgorithms []string
	DPoPClockSkew            time.Duration
	DPoPNonceLifespan        time.Duration
	DPoPProofLifespan        time.Duration
	DPoPStrategy             oauthelia2.DPoPStrategy

	EnforceClientAssertionIssuerAudience bool

	BackChannelLogoutLifespan    time.Duration
	BackChannelLogoutConcurrency int

	JWTScopeField  jwt.JWTScopeFieldEnum
	JWTMaxDuration time.Duration

	JWTSecuredAuthorizationLifespan time.Duration

	JWTAccessToken JWTAccessTokenConfig

	Hash      HashConfig
	Strategy  StrategyConfig
	PAR       PARConfig
	Handlers  HandlersConfig
	Lifespans LifespansConfig
	RFC8693   RFC8693Config

	ProofKeyCodeExchange ProofKeyCodeExchangeConfig
	GrantTypeJWTBearer   GrantTypeJWTBearerConfig

	TokenURL string

	RFC8628UserVerificationURL string

	RevokeRefreshTokensExplicit                        bool
	EnforceRevokeFlowRevokeRefreshTokensExplicitClient bool
	EnforceJWTProfileAccessTokens                      bool
	ClientCredentialsFlowImplicitGrantRequested        bool

	TokenEntropy        int
	MinParameterEntropy int

	SanitationWhiteList []string
	AllowedPrompts      []string
	RefreshTokenScopes  []string

	HTTPClient     *retryablehttp.Client
	MessageCatalog i18n.MessageCatalog

	Templates *templates.Provider
}

// GetJWTStrategy returns the JWT strategy.
func (c *Config) GetJWTStrategy(ctx context.Context) jwt.Strategy {
	return c.Strategy.JWT
}

// RFC8693Config holds specific oauthelia2.Configurator information for RFC8693 Token Exchange.
type RFC8693Config struct {
	TokenTypes                map[string]oauthelia2.RFC8693TokenType
	DefaultRequestedTokenType string
}

// LifespansConfig holds specific oauthelia2.Configurator information for token lifespans.
type LifespansConfig struct {
	schema.IdentityProvidersOpenIDConnectLifespanToken

	VerifiableCredentialsNonce time.Duration

	RFC8628Code    time.Duration
	RFC8628Polling time.Duration
}

// HashConfig holds specific oauthelia2.Configurator information for hashing.
type HashConfig struct {
	HMAC func() (h hash.Hash)
}

// StrategyConfig holds specific oauthelia2.Configurator information for various strategies.
type StrategyConfig struct {
	Core                            oauth2.CoreStrategy
	OpenID                          openid.OpenIDConnectTokenStrategy
	Audience                        oauthelia2.AudienceStrategy
	Resource                        oauthelia2.ResourceStrategy
	Scope                           oauthelia2.ScopeStrategy
	JWT                             jwt.Strategy
	JWKSFetcher                     jwt.JWKSFetcherStrategy
	ClientAuthentication            oauthelia2.ClientAuthenticationStrategy
	AuthorizeErrorFieldResponse     oauthelia2.AuthorizeErrorFieldResponseStrategy
	TokenEndpointClientAuth         oauthelia2.EndpointClientAuthStrategy
	RevocationEndpointClientAuth    oauthelia2.EndpointClientAuthStrategy
	IntrospectionEndpointClientAuth oauthelia2.EndpointClientAuthStrategy
	IDTokenValidation               oauthelia2.TokenValidationStrategy
}

// JWTAccessTokenConfig represents the JWT Access Token config.
type JWTAccessTokenConfig struct {
	Enable                       bool
	EnableStatelessIntrospection bool
}

// PARConfig holds specific oauthelia2.Configurator information for Pushed Authorization Requests.
type PARConfig struct {
	Require         bool
	URIPrefix       string
	ContextLifespan time.Duration
}

// IssuersConfig holds specific oauthelia2.Configurator information for the issuer.
type IssuersConfig struct {
	IDToken       string
	AccessToken   string
	Introspection string

	AuthorizationServerIssuerIdentification string
	JWTSecuredResponseMode                  string
}

// HandlersConfig holds specific oauthelia2.Configurator handlers configuration information.
type HandlersConfig struct {
	// ResponseMode provides an extension handler for custom response modes.
	ResponseMode oauthelia2.ResponseModeHandlers

	// ResponseModeParameter provides an extension handler for custom response mode parameters added later after the
	// response mode is assured.
	ResponseModeParameter oauthelia2.ResponseModeParameterHandlers

	// AuthorizeEndpoint is a list of handlers that are called before the authorization endpoint is served.
	AuthorizeEndpoint oauthelia2.AuthorizeEndpointHandlers

	// TokenEndpoint is a list of handlers that are called before the token endpoint is served.
	TokenEndpoint oauthelia2.TokenEndpointHandlers

	// TokenIntrospection is a list of handlers that are called before the token introspection endpoint is served.
	TokenIntrospection oauthelia2.TokenIntrospectionHandlers

	// Revocation is a list of handlers that are called before the revocation endpoint is served.
	Revocation oauthelia2.RevocationHandlers

	// PushedAuthorizeEndpoint is a list of handlers that are called before the PAR endpoint is served.
	PushedAuthorizeEndpoint oauthelia2.PushedAuthorizeEndpointHandlers

	RFC8628DeviceAuthorizeEndpoint oauthelia2.RFC8628DeviceAuthorizeEndpointHandlers

	RFC8628UserAuthorizeEndpoint oauthelia2.RFC8628UserAuthorizeEndpointHandlers

	// AuthorizeEndpointBinding is a list of handlers which bind a sender-constrained credential to an authorization
	// response.
	AuthorizeEndpointBinding oauthelia2.AuthorizeEndpointBindingHandlers

	// TokenEndpointBinding is a list of handlers which bind a sender-constrained credential to a token response.
	TokenEndpointBinding oauthelia2.TokenEndpointBindingHandlers

	// RFC7591ClientRegistrationEndpoint is a list of handlers that are called before the RFC 7591 Dynamic Client
	// Registration endpoint is served. It is empty as this Authorization Server does not serve that endpoint.
	RFC7591ClientRegistrationEndpoint oauthelia2.RFC7591ClientRegistrationEndpointHandlers

	// RFC7592ClientConfigurationEndpoint is a list of handlers that are called before the RFC 7592 Dynamic Client
	// Registration Management endpoint is served. It is empty as this Authorization Server does not serve that
	// endpoint.
	RFC7592ClientConfigurationEndpoint oauthelia2.RFC7592ClientConfigurationEndpointHandlers
}

// GrantTypeJWTBearerConfig holds specific oauthelia2.Configurator information for the JWT Bearer Grant Type.
type GrantTypeJWTBearerConfig struct {
	OptionalClientAuth bool
	OptionalJTIClaim   bool
	OptionalIssuedDate bool
}

// ProofKeyCodeExchangeConfig holds specific oauthelia2.Configurator information for PKCE.
type ProofKeyCodeExchangeConfig struct {
	Enforce                   bool
	EnforcePublicClients      bool
	AllowPlainChallengeMethod bool
}

// StatelessJWTStrategy is a strategy which handles stateless JWT profile access tokens.
type StatelessJWTStrategy struct {
	jwt.Strategy
	oauth2.CoreStrategy
}

// LoadHandlers reloads the handlers based on the current configuration.
func (c *Config) LoadHandlers(store *Store) {
	validator := openid.NewOpenIDConnectRequestValidator(c.Strategy.JWT, c)

	var statelessJWT any

	if c.JWTAccessToken.Enable && c.JWTAccessToken.EnableStatelessIntrospection {
		statelessJWT = &oauth2.StatelessJWTValidator{
			StatelessJWTStrategy: &StatelessJWTStrategy{
				Strategy:     c.Strategy.JWT,
				CoreStrategy: c.Strategy.Core,
			},
			Config: c,
		}
	}

	handlers := []any{
		&oauth2.AuthorizeExplicitGrantHandler{
			AccessTokenStrategy:    c.Strategy.Core,
			RefreshTokenStrategy:   c.Strategy.Core,
			AuthorizeCodeStrategy:  c.Strategy.Core,
			CoreStorage:            store,
			TokenRevocationStorage: store,
			Config:                 c,
		},
		&oauth2.AuthorizeImplicitGrantTypeHandler{
			AccessTokenStrategy: c.Strategy.Core,
			AccessTokenStorage:  store,
			Config:              c,
		},
		&oauth2.ClientCredentialsGrantHandler{
			HandleHelper: &oauth2.HandleHelper{
				AccessTokenStrategy: c.Strategy.Core,
				AccessTokenStorage:  store,
				Config:              c,
			},
			Config: c,
		},
		&oauth2.RefreshTokenGrantHandler{
			AccessTokenStrategy:    c.Strategy.Core,
			RefreshTokenStrategy:   c.Strategy.Core,
			TokenRevocationStorage: store,
			Config:                 c,
		},
		&rfc8628.DeviceAuthorizeHandler{
			Storage:  store,
			Strategy: c.Strategy.Core,
			Config:   c,
		},
		&rfc8628.UserAuthorizeHandler{
			Storage:  store,
			Strategy: c.Strategy.Core,
			Config:   c,
		},
		&rfc8628.DeviceAuthorizeTokenEndpointHandler{
			GenericCodeTokenEndpointHandler: oauth2.GenericCodeTokenEndpointHandler{
				CodeTokenEndpointHandler: &rfc8628.DeviceCodeTokenHandler{
					Strategy: c.Strategy.Core,
					Storage:  store,
					Config:   c,
				},
				AccessTokenStrategy:    c.Strategy.Core,
				RefreshTokenStrategy:   c.Strategy.Core,
				CoreStorage:            store,
				TokenRevocationStorage: store,
				Config:                 c,
			},
		},

		&openid.OpenIDConnectExplicitHandler{
			IDTokenHandleHelper: &openid.IDTokenHandleHelper{
				IDTokenStrategy: c.Strategy.OpenID,
			},
			OpenIDConnectRequestValidator: validator,
			OpenIDConnectRequestStorage:   store,
			Config:                        c,
		},
		&openid.OpenIDConnectImplicitHandler{
			AuthorizeImplicitGrantTypeHandler: &oauth2.AuthorizeImplicitGrantTypeHandler{
				AccessTokenStrategy: c.Strategy.Core,
				AccessTokenStorage:  store,
				Config:              c,
			},
			IDTokenHandleHelper: &openid.IDTokenHandleHelper{
				IDTokenStrategy: c.Strategy.OpenID,
			},
			OpenIDConnectRequestValidator: validator,
			Config:                        c,
		},
		&openid.OpenIDConnectHybridHandler{
			AuthorizeExplicitGrantHandler: &oauth2.AuthorizeExplicitGrantHandler{
				AccessTokenStrategy:   c.Strategy.Core,
				RefreshTokenStrategy:  c.Strategy.Core,
				AuthorizeCodeStrategy: c.Strategy.Core,
				CoreStorage:           store,
				Config:                c,
			},
			Config: c,
			AuthorizeImplicitGrantTypeHandler: &oauth2.AuthorizeImplicitGrantTypeHandler{
				AccessTokenStrategy: c.Strategy.Core,
				AccessTokenStorage:  store,
				Config:              c,
			},
			IDTokenHandleHelper: &openid.IDTokenHandleHelper{
				IDTokenStrategy: c.Strategy.OpenID,
			},
			OpenIDConnectRequestValidator: validator,
			OpenIDConnectRequestStorage:   store,
		},
		&openid.OpenIDConnectRefreshHandler{
			IDTokenHandleHelper: &openid.IDTokenHandleHelper{
				IDTokenStrategy: c.Strategy.OpenID,
			},
			Config: c,
		},
		&openid.OpenIDConnectDeviceAuthorizeHandler{
			OpenIDConnectRequestStorage:   store,
			OpenIDConnectRequestValidator: validator,
			CodeTokenEndpointHandler: &rfc8628.DeviceCodeTokenHandler{
				Strategy: c.Strategy.Core,
				Storage:  store,
				Config:   c,
			},
			Config: c,
			IDTokenHandleHelper: &openid.IDTokenHandleHelper{
				IDTokenStrategy: c.Strategy.OpenID,
			},
		},

		statelessJWT,
		&oauth2.CoreValidator{
			CoreStrategy: c.Strategy.Core,
			CoreStorage:  store,
			Config:       c,
		},

		&oauth2.TokenRevocationHandler{
			AccessTokenStrategy:    c.Strategy.Core,
			RefreshTokenStrategy:   c.Strategy.Core,
			TokenRevocationStorage: store,
			Config:                 c,
		},

		&pkce.Handler{
			AuthorizeCodeStrategy: c.Strategy.Core,
			Storage:               store,
			Config:                c,
		},

		&par.PushedAuthorizeHandler{
			Storage: store,
			Config:  c,
		},

		// Response Modes Handling.
		&oauthelia2.DefaultResponseModeHandler{
			Config: c,
		},
		&oauthelia2.RFC9207ResponseModeParameterHandler{
			Config: c,
		},
	}

	x := HandlersConfig{
		ResponseMode: []oauthelia2.ResponseModeHandler{&oauthelia2.DefaultResponseModeHandler{Config: c}},
	}

	for _, handler := range handlers {
		if handler == nil {
			continue
		}

		if h, ok := handler.(oauthelia2.AuthorizeEndpointHandler); ok {
			x.AuthorizeEndpoint.Append(h)
		}

		if h, ok := handler.(oauthelia2.RFC8628DeviceAuthorizeEndpointHandler); ok {
			x.RFC8628DeviceAuthorizeEndpoint.Append(h)
		}

		if h, ok := handler.(oauthelia2.RFC8628UserAuthorizeEndpointHandler); ok {
			x.RFC8628UserAuthorizeEndpoint.Append(h)
		}

		if h, ok := handler.(oauthelia2.TokenEndpointHandler); ok {
			x.TokenEndpoint.Append(h)
		}

		if h, ok := handler.(oauthelia2.TokenIntrospector); ok {
			x.TokenIntrospection.Append(h)
		}

		if h, ok := handler.(oauthelia2.RevocationHandler); ok {
			x.Revocation.Append(h)
		}

		if h, ok := handler.(oauthelia2.PushedAuthorizeEndpointHandler); ok {
			x.PushedAuthorizeEndpoint.Append(h)
		}

		if h, ok := handler.(oauthelia2.ResponseModeHandler); ok {
			x.ResponseMode.Append(h)
		}

		if h, ok := handler.(oauthelia2.ResponseModeParameterHandler); ok {
			x.ResponseModeParameter.Append(h)
		}
	}

	c.Handlers = x
}

// GetAllowedPrompts returns the allowed prompts.
func (c *Config) GetAllowedPrompts(ctx context.Context) (prompts []string) {
	if len(c.AllowedPrompts) == 0 {
		c.AllowedPrompts = []string{PromptNone, PromptLogin, PromptConsent, PromptSelectAccount}
	}

	return c.AllowedPrompts
}

// GetEnforcePKCE returns the enforcement of PKCE.
func (c *Config) GetEnforcePKCE(ctx context.Context) (enforce bool) {
	return c.ProofKeyCodeExchange.Enforce
}

// GetEnforcePKCEForPublicClients returns the enforcement of PKCE for public clients.
func (c *Config) GetEnforcePKCEForPublicClients(ctx context.Context) (enforce bool) {
	return c.GetEnforcePKCE(ctx) || c.ProofKeyCodeExchange.EnforcePublicClients
}

// GetEnablePKCEPlainChallengeMethod returns the enable PKCE plain challenge method.
func (c *Config) GetEnablePKCEPlainChallengeMethod(ctx context.Context) (enable bool) {
	return c.ProofKeyCodeExchange.AllowPlainChallengeMethod
}

// GetGrantTypeJWTBearerCanSkipClientAuth returns the grant type JWT bearer can skip client auth.
func (c *Config) GetGrantTypeJWTBearerCanSkipClientAuth(ctx context.Context) (skip bool) {
	return c.GrantTypeJWTBearer.OptionalClientAuth
}

// GetGrantTypeJWTBearerIDOptional returns the grant type JWT bearer ID optional.
func (c *Config) GetGrantTypeJWTBearerIDOptional(ctx context.Context) (optional bool) {
	return c.GrantTypeJWTBearer.OptionalJTIClaim
}

// GetGrantTypeJWTBearerIssuedDateOptional returns the grant type JWT bearer issued date optional.
func (c *Config) GetGrantTypeJWTBearerIssuedDateOptional(ctx context.Context) (optional bool) {
	return c.GrantTypeJWTBearer.OptionalIssuedDate
}

// GetJWTMaxDuration returns the JWT max duration.
func (c *Config) GetJWTMaxDuration(ctx context.Context) (duration time.Duration) {
	if c.JWTMaxDuration == 0 {
		c.JWTMaxDuration = time.Hour * 24
	}

	return c.JWTMaxDuration
}

// GetRedirectSecureChecker returns the redirect URL security validator.
func (c *Config) GetRedirectSecureChecker(ctx context.Context) func(context.Context, *url.URL) (secure bool) {
	return oauthelia2.IsRedirectURISecure
}

// GetOmitRedirectScopeParam must be set to true if the scope query param is to be omitted
// in the authorization's redirect URI.
func (c *Config) GetOmitRedirectScopeParam(ctx context.Context) (omit bool) {
	return c.OmitRedirectScopeParameter
}

// GetSanitationWhiteList is a whitelist of form values that are required by the token endpoint. These values
// are safe for storage in a database (cleartext).
func (c *Config) GetSanitationWhiteList(ctx context.Context) (whitelist []string) {
	return c.SanitationWhiteList
}

// GetJWTScopeField returns the JWT scope field.
func (c *Config) GetJWTScopeField(ctx context.Context) (field jwt.JWTScopeFieldEnum) {
	if c.JWTScopeField == jwt.JWTScopeFieldUnset {
		c.JWTScopeField = jwt.JWTScopeFieldList
	}

	return c.JWTScopeField
}

// GetIssuerFallback returns the issuer from the ctx or returns the fallback value.
func (c *Config) GetIssuerFallback(ctx context.Context, fallback string) (issuer string) {
	if octx := c.GetContext(ctx); octx != nil {
		if iss, err := octx.IssuerURL(); err == nil {
			return iss.String()
		}
	}

	return fallback
}

// GetIDTokenIssuer returns the ID token issuer.
func (c *Config) GetIDTokenIssuer(ctx context.Context) (issuer string) {
	return c.GetIssuerFallback(ctx, c.Issuers.IDToken)
}

// GetAccessTokenIssuer returns the access token issuer.
func (c *Config) GetAccessTokenIssuer(ctx context.Context) (issuer string) {
	return c.GetIssuerFallback(ctx, c.Issuers.AccessToken)
}

// GetAuthorizationServerIdentificationIssuer returns the Authorization Server Identification issuer.
func (c *Config) GetAuthorizationServerIdentificationIssuer(ctx context.Context) (issuer string) {
	return c.GetIssuerFallback(ctx, c.Issuers.AuthorizationServerIssuerIdentification)
}

// GetIntrospectionIssuer returns the Introspection issuer.
func (c *Config) GetIntrospectionIssuer(ctx context.Context) (issuer string) {
	return c.GetIssuerFallback(ctx, c.Issuers.Introspection)
}

// GetIntrospectionJWTResponseStrategy returns jwt.Signer for Introspection JWT Responses.
func (c *Config) GetIntrospectionJWTResponseStrategy(ctx context.Context) jwt.Strategy {
	return c.Strategy.JWT
}

// GetDisableRefreshTokenValidation returns the disable refresh token validation flag.
func (c *Config) GetDisableRefreshTokenValidation(ctx context.Context) (disable bool) {
	return c.DisableRefreshTokenValidation
}

// GetJWTSecuredAuthorizeResponseModeLifespan returns the configured JWT Secured Authorization lifespan.
func (c *Config) GetJWTSecuredAuthorizeResponseModeLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.JWTSecuredAuthorizationLifespan.Seconds() <= 0 {
		c.JWTSecuredAuthorizationLifespan = lifespanJWTSecuredAuthorizationDefault
	}

	return c.JWTSecuredAuthorizationLifespan
}

// GetJWTSecuredAuthorizeResponseModeStrategy returns jwt.Signer for JWT Secured Authorization Responses.
func (c *Config) GetJWTSecuredAuthorizeResponseModeStrategy(ctx context.Context) (strategy jwt.Strategy) {
	return c.Strategy.JWT
}

// GetJWTSecuredAuthorizeResponseModeIssuer returns the issuer for JWT Secured Authorization Responses.
func (c *Config) GetJWTSecuredAuthorizeResponseModeIssuer(ctx context.Context) string {
	return c.GetIssuerFallback(ctx, c.Issuers.JWTSecuredResponseMode)
}

// GetAccessTokenLifespan returns the access token lifespan.
func (c *Config) GetAccessTokenLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.Lifespans.AccessToken.Seconds() <= 0 {
		c.Lifespans.AccessToken = lifespanTokenDefault
	}

	return c.Lifespans.AccessToken
}

// GetRefreshTokenLifespan returns the refresh token lifespan.
func (c *Config) GetRefreshTokenLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.Lifespans.RefreshToken.Seconds() <= 0 {
		c.Lifespans.RefreshToken = lifespanRefreshTokenDefault
	}

	return c.Lifespans.RefreshToken
}

// GetIDTokenLifespan returns the ID token lifespan.
func (c *Config) GetIDTokenLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.Lifespans.IDToken.Seconds() <= 0 {
		c.Lifespans.IDToken = lifespanTokenDefault
	}

	return c.Lifespans.IDToken
}

// GetAuthorizeCodeLifespan returns the authorization code lifespan.
func (c *Config) GetAuthorizeCodeLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.Lifespans.AuthorizeCode.Seconds() <= 0 {
		c.Lifespans.AuthorizeCode = lifespanAuthorizeCodeDefault
	}

	return c.Lifespans.AuthorizeCode
}

// GetRFC8628CodeLifespan returns the RFC8628 device code lifespan.
func (c *Config) GetRFC8628CodeLifespan(ctx context.Context) time.Duration {
	if c.Lifespans.RFC8628Code.Seconds() <= 0 {
		c.Lifespans.RFC8628Code = lifespanRFC8628CodeDefault
	}

	return c.Lifespans.RFC8628Code
}

// GetPushedAuthorizeContextLifespan is the lifespan of the short-lived PAR context.
func (c *Config) GetPushedAuthorizeContextLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.PAR.ContextLifespan.Seconds() <= 0 {
		c.PAR.ContextLifespan = lifespanPARContextDefault
	}

	return c.PAR.ContextLifespan
}

// GetVerifiableCredentialsNonceLifespan is the lifespan of the verifiable credentials' nonce.
func (c *Config) GetVerifiableCredentialsNonceLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.Lifespans.VerifiableCredentialsNonce.Seconds() == 0 {
		c.Lifespans.VerifiableCredentialsNonce = lifespanVerifiableCredentialsNonceDefault
	}

	return c.Lifespans.VerifiableCredentialsNonce
}

// GetTokenEntropy returns the token entropy.
func (c *Config) GetTokenEntropy(ctx context.Context) (entropy int) {
	if c.TokenEntropy == 0 {
		c.TokenEntropy = 32
	}

	return c.TokenEntropy
}

// GetGlobalSecret returns the global secret.
func (c *Config) GetGlobalSecret(ctx context.Context) (secret []byte, err error) {
	return c.GlobalSecret, nil
}

// GetRotatedGlobalSecrets returns the rotated global secrets.
func (c *Config) GetRotatedGlobalSecrets(ctx context.Context) (secrets [][]byte, err error) {
	return c.RotatedGlobalSecrets, nil
}

// GetHTTPClient returns the HTTP client provider.
func (c *Config) GetHTTPClient(ctx context.Context) (client *retryablehttp.Client) {
	if c.HTTPClient == nil {
		c.HTTPClient = retryablehttp.NewClient()
	}

	return c.HTTPClient
}

// GetRefreshTokenScopes returns the refresh token scopes.
func (c *Config) GetRefreshTokenScopes(ctx context.Context) (scopes []string) {
	if c.RefreshTokenScopes == nil {
		c.RefreshTokenScopes = []string{ScopeOffline, ScopeOfflineAccess}
	}

	return c.RefreshTokenScopes
}

// GetScopeStrategy returns the scope strategy.
func (c *Config) GetScopeStrategy(ctx context.Context) (strategy oauthelia2.ScopeStrategy) {
	if c.Strategy.Scope == nil {
		c.Strategy.Scope = oauthelia2.ExactScopeStrategy
	}

	return c.Strategy.Scope
}

// GetAudienceStrategy returns the audience strategy.
func (c *Config) GetAudienceStrategy(ctx context.Context) (strategy oauthelia2.AudienceStrategy) {
	if c.Strategy.Audience == nil {
		c.Strategy.Audience = oauthelia2.DefaultAudienceStrategy
	}

	return c.Strategy.Audience
}

// GetResourceStrategy returns the audience strategy.
func (c *Config) GetResourceStrategy(ctx context.Context) (strategy oauthelia2.ResourceStrategy) {
	if c.Strategy.Resource == nil {
		c.Strategy.Resource = oauthelia2.DefaultResourceStrategy
	}

	return c.Strategy.Resource
}

// GetClientCredentialsFlowImplicitGrantRequested returns the client credentials flow implicit grant requested flag.
func (c *Config) GetClientCredentialsFlowImplicitGrantRequested(ctx context.Context) (implicit bool) {
	return c.ClientCredentialsFlowImplicitGrantRequested
}

// GetMinParameterEntropy returns the minimum parameter entropy.
func (c *Config) GetMinParameterEntropy(_ context.Context) (entropy int) {
	if c.MinParameterEntropy == 0 {
		c.MinParameterEntropy = oauthelia2.MinParameterEntropy
	}

	return c.MinParameterEntropy
}

// GetHMACHasher returns the hash function.
func (c *Config) GetHMACHasher(ctx context.Context) func() (h hash.Hash) {
	if c.Hash.HMAC == nil {
		c.Hash.HMAC = sha512.New512_256
	}

	return c.Hash.HMAC
}

// GetSendDebugMessagesToClients returns the send debug messages to clients.
func (c *Config) GetSendDebugMessagesToClients(ctx context.Context) (send bool) {
	return c.SendDebugMessagesToClients
}

// GetJWKSFetcherStrategy returns the JWKS fetcher strategy.
func (c *Config) GetJWKSFetcherStrategy(ctx context.Context) (strategy jwt.JWKSFetcherStrategy) {
	if c.Strategy.JWKSFetcher == nil {
		c.Strategy.JWKSFetcher = oauthelia2.NewDefaultJWKSFetcherStrategy()
	}

	return c.Strategy.JWKSFetcher
}

// GetClientAuthenticationStrategy returns the client authentication strategy.
func (c *Config) GetClientAuthenticationStrategy(ctx context.Context) (strategy oauthelia2.ClientAuthenticationStrategy) {
	return c.Strategy.ClientAuthentication
}

// GetMessageCatalog returns the message catalog.
func (c *Config) GetMessageCatalog(ctx context.Context) (catalog i18n.MessageCatalog) {
	return c.MessageCatalog
}

// GetFormPostHTMLTemplate returns the form post HTML template.
func (c *Config) GetFormPostHTMLTemplate(ctx context.Context) (tmpl *template.Template) {
	if c.Templates == nil {
		return nil
	}

	return c.Templates.GetOpenIDConnectAuthorizeResponseFormPostTemplate()
}

// GetFormPostResponseWriter returns a FormPostResponseWriter which should be utilized for writing the
// form post response type.
func (c *Config) GetFormPostResponseWriter(ctx context.Context) oauthelia2.FormPostResponseWriter {
	return oauthelia2.DefaultFormPostResponseWriter
}

func (c *Config) getEndpointURL(ctx context.Context, path, fallback string) (endpointURL string) {
	var octx Context

	if octx = c.GetContext(ctx); octx == nil {
		return fallback
	}

	switch issuerURL, err := octx.IssuerURL(); err {
	case nil:
		return strings.ToLower(issuerURL.JoinPath(path).String())
	default:
		return fallback
	}
}

// GetUseLegacyErrorFormat returns whether to use the legacy error format.
//
// Deprecated: Do not use this flag anymore.
func (c *Config) GetUseLegacyErrorFormat(ctx context.Context) (use bool) {
	return false
}

// GetAuthorizeEndpointHandlers returns the authorize endpoint handlers.
func (c *Config) GetAuthorizeEndpointHandlers(ctx context.Context) (handlers oauthelia2.AuthorizeEndpointHandlers) {
	return c.Handlers.AuthorizeEndpoint
}

// GetTokenEndpointHandlers returns the token endpoint handlers.
func (c *Config) GetTokenEndpointHandlers(ctx context.Context) (handlers oauthelia2.TokenEndpointHandlers) {
	return c.Handlers.TokenEndpoint
}

// GetTokenIntrospectionHandlers returns the token introspection handlers.
func (c *Config) GetTokenIntrospectionHandlers(ctx context.Context) (handlers oauthelia2.TokenIntrospectionHandlers) {
	return c.Handlers.TokenIntrospection
}

// GetRevocationHandlers returns the revocation handlers.
func (c *Config) GetRevocationHandlers(ctx context.Context) (handlers oauthelia2.RevocationHandlers) {
	return c.Handlers.Revocation
}

// GetPushedAuthorizeEndpointHandlers returns the handlers.
func (c *Config) GetPushedAuthorizeEndpointHandlers(ctx context.Context) oauthelia2.PushedAuthorizeEndpointHandlers {
	return c.Handlers.PushedAuthorizeEndpoint
}

// GetPushedAuthorizeRequestURIPrefix is the request URI prefix. This is
// usually 'urn:ietf:params:oauth:request_uri:'.
func (c *Config) GetPushedAuthorizeRequestURIPrefix(ctx context.Context) string {
	if c.PAR.URIPrefix == "" {
		c.PAR.URIPrefix = RedirectURIPrefixPushedAuthorizationRequestURN
	}

	return c.PAR.URIPrefix
}

// GetRequirePushedAuthorizationRequests indicates if the use of Pushed Authorization Requests is globally required.
// In this mode, a client cannot pass authorize parameters at the 'authorize' endpoint. The 'authorize' endpoint
// must contain the PAR request_uri.
func (c *Config) GetRequirePushedAuthorizationRequests(ctx context.Context) (enforce bool) {
	return c.PAR.Require
}

// GetResponseModeHandlers returns the response mode handlers.
func (c *Config) GetResponseModeHandlers(ctx context.Context) oauthelia2.ResponseModeHandlers {
	return c.Handlers.ResponseMode
}

// GetResponseModeParameterHandlers returns the response mode parameter handlers.
func (c *Config) GetResponseModeParameterHandlers(ctx context.Context) oauthelia2.ResponseModeParameterHandlers {
	return c.Handlers.ResponseModeParameter
}

// GetRevokeRefreshTokensExplicit returns the revoke refresh tokens explicit flag.
func (c *Config) GetRevokeRefreshTokensExplicit(ctx context.Context) (explicit bool) {
	return c.RevokeRefreshTokensExplicit
}

// GetEnforceRevokeFlowRevokeRefreshTokensExplicitClient returns the enforce revoke flow revoke refresh tokens explicit client flag.
func (c *Config) GetEnforceRevokeFlowRevokeRefreshTokensExplicitClient(ctx context.Context) (enforce bool) {
	return c.EnforceRevokeFlowRevokeRefreshTokensExplicitClient
}

// GetAllowedJWTAssertionAudiences returns the audiences which are allowed for JWT assertions.
func (c *Config) GetAllowedJWTAssertionAudiences(ctx context.Context) (audiences []string) {
	var octx Context

	if octx = c.GetContext(ctx); octx == nil {
		return nil
	}

	var (
		issuer *url.URL
		err    error
	)
	if issuer, err = octx.IssuerURL(); err != nil {
		logging.Logger().WithError(err).Error("Error retrieving issuer")
		return nil
	}

	return []string{
		issuer.String(),
		issuer.JoinPath(EndpointPathToken).String(),
		issuer.JoinPath(EndpointPathPushedAuthorizationRequest).String(),
	}
}

// GetRFC8628UserVerificationURL returns the RFC8628 user verification URL.
func (c *Config) GetRFC8628UserVerificationURL(ctx context.Context) string {
	return c.getEndpointURL(ctx, FrontendEndpointPathConsentDeviceAuthorization, c.RFC8628UserVerificationURL)
}

// GetRFC8628TokenPollingInterval returns the RFC8628 token polling interval.
func (c *Config) GetRFC8628TokenPollingInterval(ctx context.Context) (interval time.Duration) {
	if c.Lifespans.RFC8628Polling.Seconds() == 0 {
		c.Lifespans.RFC8628Polling = lifespanRFC8628PollingIntervalDefault
	}

	return c.Lifespans.RFC8628Polling
}

// GetRFC8628DeviceAuthorizeEndpointHandlers returns the RFC8628 Device Authorize Endpoint handlers.
func (c *Config) GetRFC8628DeviceAuthorizeEndpointHandlers(ctx context.Context) oauthelia2.RFC8628DeviceAuthorizeEndpointHandlers {
	return c.Handlers.RFC8628DeviceAuthorizeEndpoint
}

// GetRFC8628UserAuthorizeEndpointHandlers returns the RFC8628 User Authorize Endpoint handlers.
func (c *Config) GetRFC8628UserAuthorizeEndpointHandlers(ctx context.Context) oauthelia2.RFC8628UserAuthorizeEndpointHandlers {
	return c.Handlers.RFC8628UserAuthorizeEndpoint
}

// GetRFC8693TokenTypes returns the RFC8693 token types.
func (c *Config) GetRFC8693TokenTypes(ctx context.Context) map[string]oauthelia2.RFC8693TokenType {
	return c.RFC8693.TokenTypes
}

// GetDefaultRFC8693RequestedTokenType returns the default RFC8693 requested token type.
func (c *Config) GetDefaultRFC8693RequestedTokenType(ctx context.Context) string {
	return c.RFC8693.DefaultRequestedTokenType
}

// GetEnforceJWTProfileAccessTokens returns the enforce JWT profile access tokens flag.
func (c *Config) GetEnforceJWTProfileAccessTokens(ctx context.Context) (enforce bool) {
	return c.EnforceJWTProfileAccessTokens
}

// GetAuthorizeErrorFieldResponseStrategy returns the authorize error field response strategy.
func (c *Config) GetAuthorizeErrorFieldResponseStrategy(ctx context.Context) (strategy oauthelia2.AuthorizeErrorFieldResponseStrategy) {
	if c.Strategy.AuthorizeErrorFieldResponse == nil {
		c.Strategy.AuthorizeErrorFieldResponse = &RedirectAuthorizeErrorFieldResponseStrategy{Config: c}
	}

	return c.Strategy.AuthorizeErrorFieldResponse
}

// GetTokenEndpointClientAuthStrategy returns the Token Endpoint client authentication strategy.
func (c *Config) GetTokenEndpointClientAuthStrategy(ctx context.Context) (strategy oauthelia2.EndpointClientAuthStrategy) {
	if c.Strategy.TokenEndpointClientAuth == nil {
		c.Strategy.TokenEndpointClientAuth = &oauthelia2.TokenEndpointClientAuthStrategy{}
	}

	return c.Strategy.TokenEndpointClientAuth
}

// GetRevocationEndpointClientAuthStrategy returns the Revocation Endpoint client authentication strategy.
func (c *Config) GetRevocationEndpointClientAuthStrategy(ctx context.Context) (strategy oauthelia2.EndpointClientAuthStrategy) {
	if c.Strategy.RevocationEndpointClientAuth == nil {
		c.Strategy.RevocationEndpointClientAuth = &oauthelia2.RevocationEndpointClientAuthStrategy{}
	}

	return c.Strategy.RevocationEndpointClientAuth
}

// GetIntrospectionEndpointClientAuthStrategy returns the Introspection Endpoint client authentication strategy.
func (c *Config) GetIntrospectionEndpointClientAuthStrategy(ctx context.Context) (strategy oauthelia2.EndpointClientAuthStrategy) {
	if c.Strategy.IntrospectionEndpointClientAuth == nil {
		c.Strategy.IntrospectionEndpointClientAuth = &oauthelia2.IntrospectionEndpointClientAuthStrategy{}
	}

	return c.Strategy.IntrospectionEndpointClientAuth
}

// GetDPoPEnabled returns the DPoP enabled flag.
func (c *Config) GetDPoPEnabled(ctx context.Context) (enabled bool) {
	return c.DPoPEnabled
}

// GetDPoPEnforce returns the DPoP enforcement flag.
func (c *Config) GetDPoPEnforce(ctx context.Context) (enforce bool) {
	return c.DPoPEnforce
}

// GetDPoPAllowedJWSAlgorithms returns the allowed DPoP JWS algorithms.
func (c *Config) GetDPoPAllowedJWSAlgorithms(ctx context.Context) (algs []string) {
	return c.DPoPAllowedJWSAlgorithms
}

// GetDPoPClockSkew returns the DPoP clock skew.
func (c *Config) GetDPoPClockSkew(ctx context.Context) (skew time.Duration) {
	return c.DPoPClockSkew
}

// GetDPoPNonceRequired returns the DPoP nonce required flag.
func (c *Config) GetDPoPNonceRequired(ctx context.Context) (required bool) {
	return c.DPoPNonceRequired
}

// GetDPoPNonceLifespan returns the DPoP nonce lifespan.
func (c *Config) GetDPoPNonceLifespan(ctx context.Context) (lifespan time.Duration) {
	return c.DPoPNonceLifespan
}

// GetDPoPStrategy returns the DPoP strategy.
func (c *Config) GetDPoPStrategy(ctx context.Context) (strategy oauthelia2.DPoPStrategy) {
	return c.DPoPStrategy
}

// GetMTLSEnabled returns false as RFC 8705 Mutual-TLS Client Authentication and Certificate-Bound Access Tokens is
// not implemented by this Authorization Server. Returning true here would have the provider offer confirmation
// methods and client authentication paths which nothing in this implementation can satisfy.
func (c *Config) GetMTLSEnabled(ctx context.Context) (enabled bool) {
	return false
}

// GetAllowedIntrospectionAudiences returns the audiences an Access Token used to authenticate a request to the
// introspection endpoint may carry. An empty list is not a disabled check: it makes the provider expect the URL the
// request was made to instead, which is the behavior this implementation has always had as it exposes no
// configuration for this value.
func (c *Config) GetAllowedIntrospectionAudiences(ctx context.Context) (audiences []string) {
	return nil
}

// GetDPoPProofLifespan returns the DPoP proof lifespan, which together with the clock skew fixes the window a proof
// is accepted in.
func (c *Config) GetDPoPProofLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.DPoPProofLifespan <= 0 {
		return lifespanDPoPProofDefault
	}

	return c.DPoPProofLifespan
}

// GetMTLSEnforce returns false as RFC 8705 is not implemented by this Authorization Server.
func (c *Config) GetMTLSEnforce(ctx context.Context) (enforce bool) {
	return false
}

// GetMTLSClientCertificateHeader returns an empty string as RFC 8705 is not implemented by this Authorization Server,
// which per the provider contract means a forwarded client certificate is never read.
func (c *Config) GetMTLSClientCertificateHeader(ctx context.Context) (header string) {
	return ""
}

// GetAllowedIntrospectionScopes returns the scopes an Access Token used to authenticate a request to the
// introspection endpoint may carry. No configuration exposes this value so the provider default of no additional
// scope requirement applies.
func (c *Config) GetAllowedIntrospectionScopes(ctx context.Context) (scopes []string) {
	return nil
}

// GetIntrospectionEndpointClientAuthDisabled returns false as the introspection endpoint always requires client
// authentication in this implementation.
func (c *Config) GetIntrospectionEndpointClientAuthDisabled(ctx context.Context) (disabled bool) {
	return false
}

// GetEnforceClientAssertionIssuerAudience returns whether a JWT client authentication assertion must carry this
// server's issuer identifier as the sole value of its 'aud' claim. This is off by default as the tightening lives in
// draft-ietf-oauth-rfc7523bis while the published RFC 7523 permits the token endpoint URL.
func (c *Config) GetEnforceClientAssertionIssuerAudience(ctx context.Context) (enforce bool) {
	return c.EnforceClientAssertionIssuerAudience
}

// GetRequireSignedRequestObject returns false as this Authorization Server does not globally require an
// authorization request to be provided as a Request Object.
func (c *Config) GetRequireSignedRequestObject(ctx context.Context) (require bool) {
	return false
}

// GetRequireSignedRequestObjectSkipPushedAuthorizationRequests returns false as no global Request Object requirement
// is configured for it to exempt.
func (c *Config) GetRequireSignedRequestObjectSkipPushedAuthorizationRequests(ctx context.Context) (skip bool) {
	return false
}

// GetIDTokenValidationStrategy returns the ID Token validation strategy used by RP-Initiated Logout. It has no
// default and may be nil.
func (c *Config) GetIDTokenValidationStrategy(ctx context.Context) (strategy oauthelia2.TokenValidationStrategy) {
	return c.Strategy.IDTokenValidation
}

// GetBackChannelLogoutTokenStrategy returns the Back-Channel Logout token strategy. It is nil as this Authorization
// Server does not implement OpenID Connect Back-Channel Logout 1.0.
func (c *Config) GetBackChannelLogoutTokenStrategy(ctx context.Context) (strategy oauthelia2.BackChannelLogoutTokenStrategy) {
	return nil
}

// GetBackChannelLogoutLifespan returns the lifespan of a Back-Channel Logout Token.
func (c *Config) GetBackChannelLogoutLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.BackChannelLogoutLifespan <= 0 {
		return lifespanBackChannelLogoutDefault
	}

	return c.BackChannelLogoutLifespan
}

// GetBackChannelLogoutConcurrency returns the number of Back-Channel Logout requests performed concurrently.
func (c *Config) GetBackChannelLogoutConcurrency(ctx context.Context) (n int) {
	if c.BackChannelLogoutConcurrency <= 0 {
		return backChannelLogoutConcurrencyDefault
	}

	return c.BackChannelLogoutConcurrency
}

// GetAuthorizeEndpointBindingHandlers returns the authorization endpoint binding handlers. None are registered as
// this Authorization Server implements no sender-constraining binding at that endpoint.
func (c *Config) GetAuthorizeEndpointBindingHandlers(ctx context.Context) (handlers oauthelia2.AuthorizeEndpointBindingHandlers) {
	return c.Handlers.AuthorizeEndpointBinding
}

// GetTokenEndpointBindingHandlers returns the token endpoint binding handlers. None are registered as this
// Authorization Server implements no sender-constraining binding at that endpoint beyond DPoP, which the provider
// wires itself.
func (c *Config) GetTokenEndpointBindingHandlers(ctx context.Context) (handlers oauthelia2.TokenEndpointBindingHandlers) {
	return c.Handlers.TokenEndpointBinding
}

// The RFC 7591 Dynamic Client Registration and RFC 7592 Dynamic Client Registration Management surfaces below are
// deliberately inert. This Authorization Server registers clients from its configuration, so no registration endpoint
// is served: the handler lists are empty, which is what stops the provider routing to them, and the remaining values
// describe an endpoint which does not exist. The Store reports the same by refusing to create, update, or delete a
// client.

// GetRFC7591ClientRegistrationEndpointURL returns an empty string as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientRegistrationEndpointURL(ctx context.Context) (endpoint string) {
	return ""
}

// GetRFC7591ClientRegistrationGlobalSecret returns no secret as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientRegistrationGlobalSecret(ctx context.Context) (secret []byte, err error) {
	return nil, nil
}

// GetRFC7591ClientRegistrationRotatedGlobalSecrets returns no secrets as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientRegistrationRotatedGlobalSecrets(ctx context.Context) (secrets [][]byte, err error) {
	return nil, nil
}

// GetRFC7591ClientSecretLifespan returns zero as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientSecretLifespan(ctx context.Context) (lifespan time.Duration) {
	return 0
}

// GetRFC7591ClientRegistrationStrategy returns nil as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientRegistrationStrategy(ctx context.Context) (strategy oauthelia2.ClientRegistrationStrategy) {
	return nil
}

// GetRFC7591ClientRegistrationMetadataStrategy returns nil as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientRegistrationMetadataStrategy(ctx context.Context) (strategy oauthelia2.ClientRegistrationMetadataStrategy) {
	return nil
}

// GetRFC7591ClientRegistrationEndpointAuthStrategy returns nil as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientRegistrationEndpointAuthStrategy(ctx context.Context) (strategy oauthelia2.ClientRegistrationEndpointAuthStrategy) {
	return nil
}

// GetRFC7591ClientRegistrationValidators returns no validators as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientRegistrationValidators(ctx context.Context) (validators []oauthelia2.ClientRegistrationValidator) {
	return nil
}

// GetRFC7591ClientRegistrationEndpointAudiences returns no audiences as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientRegistrationEndpointAudiences(ctx context.Context) (audiences []string) {
	return nil
}

// GetRFC7591ClientRegistrationScopes returns no scopes as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientRegistrationScopes(ctx context.Context) (scopes []string) {
	return nil
}

// GetRFC7591ClientRegistrationGrantTypes returns no grant types as no RFC 7591 registration endpoint is served.
func (c *Config) GetRFC7591ClientRegistrationGrantTypes(ctx context.Context) (grantTypes []string) {
	return nil
}

// GetRFC7591ClientRegistrationEndpointHandlers returns no handlers, which is what prevents the provider serving an
// RFC 7591 registration endpoint.
func (c *Config) GetRFC7591ClientRegistrationEndpointHandlers(ctx context.Context) (handlers oauthelia2.RFC7591ClientRegistrationEndpointHandlers) {
	return c.Handlers.RFC7591ClientRegistrationEndpoint
}

// GetRFC7592ClientConfigurationEndpointHandlers returns no handlers, which is what prevents the provider serving an
// RFC 7592 client configuration endpoint.
func (c *Config) GetRFC7592ClientConfigurationEndpointHandlers(ctx context.Context) (handlers oauthelia2.RFC7592ClientConfigurationEndpointHandlers) {
	return c.Handlers.RFC7592ClientConfigurationEndpoint
}

// GetContext returns the OpenID Connect 1.0 Context from the given context.
func (c *Config) GetContext(ctx context.Context) (octx Context) {
	var ok bool

	if octx, ok = ctx.Value(model.CtxKeyAutheliaCtx).(Context); ok {
		return octx
	}

	if octx, ok = ctx.(Context); ok {
		return octx
	}

	return nil
}

var (
	_ oauthelia2.Configurator = (*Config)(nil)
)
