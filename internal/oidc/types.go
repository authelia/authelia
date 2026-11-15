// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"authelia.com/provider/jose"
	oauthelia2 "authelia.com/provider/oauth2"
	fjwt "authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/storage"
)

// OpenIDConnectProvider for OpenID Connect.
type OpenIDConnectProvider struct {
	*Store
	*Config

	Issuer *Issuer

	oauthelia2.Provider
}

// Store is Authelia's internal representation of the oauthelia2.Storage interface. It maps the following
// interfaces to the storage.Provider interface:
// oauthelia2.Storage, oauthelia2.ClientManager, storage.Transactional, oauth2.AuthorizeCodeStorage, oauth2.AccessTokenStorage,
// oauth2.RefreshTokenStorage, oauth2.TokenRevocationStorage, pkce.PKCERequestStorage,
// openid.OpenIDConnectRequestStorage, and partially implements rfc7523.RFC7523KeyStorage.
type Store struct {
	ClientStore

	provider storage.Provider
}

// ClientStore is an abstraction used for the Store struct which stores clients.
type ClientStore interface {
	// GetRegisteredClient returns a Client matching the provided id.
	GetRegisteredClient(ctx context.Context, id string) (client Client, err error)
}

// MemoryClientStore is an implementation of the ClientStore which just stores the clients in memory.
type MemoryClientStore struct {
	clients map[string]Client
}

// RegisteredClient represents a registered client.
type RegisteredClient struct {
	ID                   string
	Name                 string
	ClientSecret         *ClientSecretDigest
	RotatedClientSecrets []*ClientSecretDigest
	SectorIdentifierURI  *url.URL
	Public               bool
	ApplicationType      string

	RequirePushedAuthorizationRequests bool
	DisableRefreshTokenRotation        bool
	RequireSignedRequestObject         bool

	RequirePKCE                bool
	RequirePKCEChallengeMethod bool
	PKCEChallengeMethod        string

	Audience      []string
	Resource      []string
	Scopes        []string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	ResponseModes []oauthelia2.ResponseModeType

	Lifespans      schema.IdentityProvidersOpenIDConnectLifespan
	ClaimsStrategy ClaimsStrategy

	AuthorizationSignedResponseAlg      string
	AuthorizationSignedResponseKeyID    string
	AuthorizationEncryptedResponseAlg   string
	AuthorizationEncryptedResponseEnc   string
	AuthorizationEncryptedResponseKeyID string

	IDTokenSignedResponseAlg      string
	IDTokenSignedResponseKeyID    string
	IDTokenEncryptedResponseAlg   string
	IDTokenEncryptedResponseEnc   string
	IDTokenEncryptedResponseKeyID string

	AccessTokenSignedResponseAlg      string
	AccessTokenSignedResponseKeyID    string
	AccessTokenEncryptedResponseAlg   string
	AccessTokenEncryptedResponseEnc   string
	AccessTokenEncryptedResponseKeyID string

	UserinfoSignedResponseAlg      string
	UserinfoSignedResponseKeyID    string
	UserinfoEncryptedResponseAlg   string
	UserinfoEncryptedResponseEnc   string
	UserinfoEncryptedResponseKeyID string

	IntrospectionAudience []string

	IntrospectionSignedResponseAlg      string
	IntrospectionSignedResponseKeyID    string
	IntrospectionEncryptedResponseAlg   string
	IntrospectionEncryptedResponseEnc   string
	IntrospectionEncryptedResponseKeyID string

	RequestObjectSigningAlg    string
	RequestObjectEncryptionAlg string
	RequestObjectEncryptionEnc string

	TokenEndpointAuthMethod     string
	TokenEndpointAuthSigningAlg string

	RevocationEndpointAuthMethod     string
	RevocationEndpointAuthSigningAlg string

	IntrospectionEndpointAuthMethod     string
	IntrospectionEndpointAuthSigningAlg string

	PushedAuthorizationRequestEndpointAuthMethod     string
	PushedAuthorizationRequestEndpointAuthSigningAlg string

	RefreshFlowIgnoreOriginalGrantedScopes  bool
	AllowMultipleAuthenticationMethods      bool
	ClientCredentialsFlowAllowImplicitScope bool
	DPoPBoundAccessTokens                   bool

	ClientAssertionJWTValidationHeaderAllowEmptyType bool
	ClientAssertionJWTValidationHeaderAllowTypes     []string

	AuthorizationPolicy ClientAuthorizationPolicy

	ConsentPolicy         ClientConsentPolicy
	RequestedAudienceMode ClientRequestedAudienceMode
	RequestedResourceMode ClientRequestedResourceMode
	AudienceStrategy      oauthelia2.AudienceStrategy
	ResourceStrategy      oauthelia2.ResourceStrategy

	RequestURIs    []string
	JSONWebKeys    *jose.JSONWebKeySet
	JSONWebKeysURI *url.URL
}

// Client represents the internal client definitions.
type Client interface {
	oauthelia2.Client
	oauthelia2.ResponseModeClient
	RefreshFlowScopeClient

	GetName() (name string)
	GetSectorIdentifierURI() (sector string)

	GetClaimsStrategy() (strategy ClaimsStrategy)

	GetAuthorizationSignedResponseKeyID() (kid string)
	GetAuthorizationSignedResponseAlg() (alg string)
	GetAuthorizationEncryptedResponseKeyID() (kid string)
	GetAuthorizationEncryptedResponseAlg() (alg string)
	GetAuthorizationEncryptedResponseEnc() (enc string)

	GetIDTokenSignedResponseKeyID() (kid string)
	GetIDTokenSignedResponseAlg() (alg string)
	GetIDTokenEncryptedResponseKeyID() (kid string)
	GetIDTokenEncryptedResponseAlg() (alg string)
	GetIDTokenEncryptedResponseEnc() (enc string)

	GetAccessTokenSignedResponseKeyID() (kid string)
	GetAccessTokenSignedResponseAlg() (alg string)
	GetAccessTokenEncryptedResponseKeyID() (kid string)
	GetAccessTokenEncryptedResponseAlg() (alg string)
	GetAccessTokenEncryptedResponseEnc() (enc string)
	GetEnableJWTProfileOAuthAccessTokens() bool

	GetUserinfoSignedResponseKeyID() (kid string)
	GetUserinfoSignedResponseAlg() (alg string)
	GetUserinfoEncryptedResponseKeyID() (kid string)
	GetUserinfoEncryptedResponseAlg() (alg string)
	GetUserinfoEncryptedResponseEnc() (enc string)

	GetIntrospectionSignedResponseKeyID() (kid string)
	GetIntrospectionSignedResponseAlg() (alg string)
	GetIntrospectionEncryptedResponseKeyID() (kid string)
	GetIntrospectionEncryptedResponseAlg() (alg string)
	GetIntrospectionEncryptedResponseEnc() (enc string)

	GetRequirePushedAuthorizationRequests() (enforce bool)

	GetJSONWebKeys() (jwks *jose.JSONWebKeySet)
	GetJSONWebKeysURI() (uri string)

	GetEnforcePKCE() (enforce bool)
	GetEnforcePKCEChallengeMethod() (enforce bool)
	GetPKCEChallengeMethod() (method string)

	ValidateResponseModePolicy(r oauthelia2.AuthorizeRequester) (err error)

	GetConsentResponseBody(session RequesterFormSession, form url.Values, authTime time.Time, disablePreConf bool) (body ConsentGetResponseBody)
	GetConsentPolicy() ClientConsentPolicy
	IsAuthenticationLevelSufficient(level authentication.Level, subject authorization.Subject) (sufficient bool)
	GetAuthorizationPolicyRequiredLevel(subject authorization.Subject) (level authorization.Level)
	GetAuthorizationPolicy() (policy ClientAuthorizationPolicy)

	GetEffectiveLifespan(gt oauthelia2.GrantType, tt oauthelia2.TokenType, fallback time.Duration) (lifespan time.Duration)
}

// RefreshFlowScopeClient is a client which can be customized to ignore scopes that were not originally granted.
type RefreshFlowScopeClient interface {
	oauthelia2.Client

	GetRefreshFlowIgnoreOriginalGrantedScopes(ctx context.Context) (ignoreOriginalGrantedScopes bool)
}

// Context represents the context implementation that is used by some OpenID Connect 1.0 implementations.
type Context interface {
	IssuerURL() (issuerURL *url.URL, err error)
	GetClock() (clock clock.Provider)
	GetRandom() (random random.Provider)
	GetConfiguration() (config *schema.Configuration)
	GetProviderStorage() storage.Provider
	GetUserProvider() authentication.UserProvider
	GetProviderUserAttributeResolver() expression.UserAttributeResolver

	context.Context
}

// ClaimsStrategyContext is a context used for the CustomClaimsStrategy implementation.
type ClaimsStrategyContext interface {
	GetProviderUserAttributeResolver() expression.UserAttributeResolver

	context.Context
}

// ClientContext is a context which provides the [http.Client] used for outbound requests.
type ClientContext interface {
	GetHTTPClient() *http.Client

	context.Context
}

// ClientRequesterResponder is a oauthelia2.Requster or fosite.Responder with a GetClient method.
type ClientRequesterResponder interface {
	GetClient() oauthelia2.Client
}

// IDTokenClaimsSession is a session which can return the IDTokenClaims type.
type IDTokenClaimsSession interface {
	GetIDTokenClaims() *fjwt.IDTokenClaims
}

// IDTokenSessionContainer is similar to the oauth2.JWTSessionContainer to facilitate obtaining the headers as appropriate.
type IDTokenSessionContainer interface {
	IDTokenHeaders() *fjwt.Headers
	IDTokenClaims() *fjwt.IDTokenClaims
}

// UserDetailer is an interface which provides the user details used to hydrate claims.
type UserDetailer interface {
	GetUsername() (username string)
	GetGroups() (groups []string)
	GetDisplayName() (name string)
	GetEmails() (emails []string)
	GetGivenName() (given string)
	GetFamilyName() (family string)
	GetMiddleName() (middle string)
	GetNickname() (nickname string)
	GetProfile() (profile string)
	GetPicture() (picture string)
	GetWebsite() (website string)
	GetGender() (gender string)
	GetBirthdate() (birthdate string)
	GetZoneInfo() (info string)
	GetLocale() (locale string)
	GetPhoneNumber() (number string)
	GetPhoneExtension() (extension string)
	GetPhoneNumberRFC3966() (number string)
	GetStreetAddress() (address string)
	GetLocality() (locality string)
	GetRegion() (region string)
	GetPostalCode() (postcode string)
	GetCountry() (country string)
	GetExtra() (extra map[string]any)
}

// ConsentGetResponseBody schema of the response body of the consent GET endpoint.
type ConsentGetResponseBody struct {
	ClientID          string   `json:"client_id"`
	ClientDescription string   `json:"client_description"`
	Scopes            []string `json:"scopes"`
	Audience          []string `json:"audience"`
	Resource          []string `json:"resource"`
	PreConfiguration  bool     `json:"pre_configuration"`
	Claims            []string `json:"claims"`
	EssentialClaims   []string `json:"essential_claims"`
	RequireLogin      bool     `json:"require_login"`
}

// ConsentPostRequestBody schema of the request body of the consent POST endpoint.
type ConsentPostRequestBody struct {
	FlowID       *string  `json:"flow_id"`
	ClientID     string   `json:"client_id"`
	Consent      bool     `json:"consent"`
	PreConfigure bool     `json:"pre_configure"`
	Claims       []string `json:"claims"`
	SubFlow      *string  `json:"subflow"`
	UserCode     *string  `json:"user_code"`
}

// ConsentPostResponseBody schema of the response body of the consent POST endpoint.
type ConsentPostResponseBody struct {
	RedirectURI string `json:"redirect_uri,omitempty"`
	FlowID      string `json:"flow_id,omitempty"`
}

// FormSession is a session which retains the original request form.
type FormSession interface {
	GetForm() (form url.Values, err error)
}

// RequesterFormSession is a FormSession which also retains the requested time and client id.
type RequesterFormSession interface {
	FormSession

	GetRequestedAt() time.Time

	GetRequestedScopes() []string
	GetRequestedAudience() []string
	GetRequestedResource() []string

	GetGrantedScopes() []string
	GetGrantedAudience() []string
	GetGrantedResource() []string
}

// Number is a constraint which permits any integer or floating point type.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

var (
	_ Client                                                              = (*RegisteredClient)(nil)
	_ oauthelia2.Client                                                   = (*RegisteredClient)(nil)
	_ oauthelia2.UserInfoClient                                           = (*RegisteredClient)(nil)
	_ oauthelia2.RotatedClientSecretsClient                               = (*RegisteredClient)(nil)
	_ oauthelia2.ProofKeyCodeExchangeClient                               = (*RegisteredClient)(nil)
	_ oauthelia2.ClientAuthenticationPolicyClient                         = (*RegisteredClient)(nil)
	_ oauthelia2.JARClient                                                = (*RegisteredClient)(nil)
	_ oauthelia2.JWTSecuredAuthorizationRequestJWTValidationOptionsClient = (*RegisteredClient)(nil)
	_ oauthelia2.ClientAssertionJWTValidationOptionsClient                = (*RegisteredClient)(nil)
	_ oauthelia2.AuthenticationMethodClient                               = (*RegisteredClient)(nil)
	_ oauthelia2.RefreshFlowScopeClient                                   = (*RegisteredClient)(nil)
	_ oauthelia2.RevokeFlowRevokeRefreshTokensExplicitClient              = (*RegisteredClient)(nil)
	_ oauthelia2.JARMClient                                               = (*RegisteredClient)(nil)
	_ oauthelia2.PushedAuthorizationRequestClient                         = (*RegisteredClient)(nil)
	_ oauthelia2.ResponseModeClient                                       = (*RegisteredClient)(nil)
	_ oauthelia2.ClientCredentialsFlowRequestedScopeImplicitClient        = (*RegisteredClient)(nil)
	_ oauthelia2.RequestedAudienceResourceImplicitClient                  = (*RegisteredClient)(nil)
	_ oauthelia2.AudienceStrategyProvider                                 = (*RegisteredClient)(nil)
	_ oauthelia2.ResourceStrategyProvider                                 = (*RegisteredClient)(nil)
	_ oauthelia2.JWTProfileClient                                         = (*RegisteredClient)(nil)
	_ oauthelia2.IntrospectionJWTResponseClient                           = (*RegisteredClient)(nil)
	_ oauthelia2.IntrospectionClient                                      = (*RegisteredClient)(nil)
	_ oauthelia2.ApplicationTypeClient                                    = (*RegisteredClient)(nil)
	_ oauthelia2.RefreshTokenMaximumLifespanClient                        = (*RegisteredClient)(nil)
	_ oauthelia2.RefreshTokenRotationClient                               = (*RegisteredClient)(nil)

	_ RequesterFormSession = (*model.OAuth2ConsentSession)(nil)
	_ RequesterFormSession = (*model.OAuth2DeviceCodeSession)(nil)
)
