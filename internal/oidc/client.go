package oidc

import (
	"context"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/x/errorsx"
	jose "github.com/go-jose/go-jose/v4"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// NewClient creates a new Client.
func NewClient(config schema.IdentityProvidersOpenIDConnectClient, c *schema.IdentityProvidersOpenIDConnect) (client Client) {
	registered := &RegisteredClient{
		ID:                  config.ID,
		Name:                config.Name,
		SectorIdentifierURI: config.SectorIdentifierURI,
		Public:              config.Public,

		Audience:      config.Audience,
		Scopes:        config.Scopes,
		RedirectURIs:  config.RedirectURIs,
		RequestURIs:   config.RequestURIs,
		GrantTypes:    config.GrantTypes,
		ResponseTypes: config.ResponseTypes,
		ResponseModes: []oauthelia2.ResponseModeType{},

		RequirePKCE:                config.RequirePKCE || config.PKCEChallengeMethod != "",
		RequirePKCEChallengeMethod: config.PKCEChallengeMethod != "",
		PKCEChallengeMethod:        config.PKCEChallengeMethod,

		RequirePushedAuthorizationRequests:      config.RequirePushedAuthorizationRequests,
		ClientCredentialsFlowAllowImplicitScope: false,
		AllowMultipleAuthenticationMethods:      config.AllowMultipleAuthenticationMethods,

		AuthorizationPolicy:   NewClientAuthorizationPolicy(config.AuthorizationPolicy, c),
		ConsentPolicy:         NewClientConsentPolicy(config.ConsentMode, config.ConsentPreConfiguredDuration),
		RequestedAudienceMode: NewClientRequestedAudienceMode(config.RequestedAudienceMode),

		AuthorizationSignedResponseAlg:   config.AuthorizationSignedResponseAlg,
		AuthorizationSignedResponseKeyID: config.AuthorizationSignedResponseKeyID,
		IDTokenSignedResponseAlg:         config.IDTokenSignedResponseAlg,
		IDTokenSignedResponseKeyID:       config.IDTokenSignedResponseKeyID,
		AccessTokenSignedResponseAlg:     config.AccessTokenSignedResponseAlg,
		AccessTokenSignedResponseKeyID:   config.AccessTokenSignedResponseKeyID,
		UserinfoSignedResponseAlg:        config.UserinfoSignedResponseAlg,
		UserinfoSignedResponseKeyID:      config.UserinfoSignedResponseKeyID,
		IntrospectionSignedResponseAlg:   config.IntrospectionSignedResponseAlg,
		IntrospectionSignedResponseKeyID: config.IntrospectionSignedResponseKeyID,
		RequestObjectSigningAlg:          config.RequestObjectSigningAlg,
		TokenEndpointAuthSigningAlg:      config.TokenEndpointAuthSigningAlg,
		TokenEndpointAuthMethod:          config.TokenEndpointAuthMethod,

		JSONWebKeysURI: config.JSONWebKeysURI,
		JSONWebKeys:    NewPublicJSONWebKeySetFromSchemaJWK(config.JSONWebKeys),
	}

	if config.Secret != nil && config.Secret.Digest != nil {
		registered.ClientSecret = &ClientSecretDigest{PasswordDigest: config.Secret}
	}

	if len(config.Lifespan) != 0 {
		if lifespans, ok := c.Lifespans.Custom[config.Lifespan]; ok {
			registered.Lifespans = lifespans
		}
	}

	for _, mode := range config.ResponseModes {
		registered.ResponseModes = append(registered.ResponseModes, oauthelia2.ResponseModeType(mode))
	}

	return registered
}

// GetID returns the ID for the client.
func (c *RegisteredClient) GetID() string {
	return c.ID
}

// GetName returns the Name for the client.
func (c *RegisteredClient) GetName() (name string) {
	if c.Name == "" {
		c.Name = c.GetID()
	}

	return c.Name
}

// GetClientSecret returns the oauth2.ClientSecret.
func (c *RegisteredClient) GetClientSecret() (secret oauthelia2.ClientSecret) {
	return c.ClientSecret
}

// GetRotatedClientSecrets returns the rotated oauth2.ClientSecret values.
func (c *RegisteredClient) GetRotatedClientSecrets() (secrets []oauthelia2.ClientSecret) {
	secrets = make([]oauthelia2.ClientSecret, len(c.RotatedClientSecrets))

	for i, secret := range c.RotatedClientSecrets {
		secrets[i] = secret
	}

	return secrets
}

// GetSectorIdentifierURI returns the SectorIdentifier for this client.
func (c *RegisteredClient) GetSectorIdentifierURI() (sector string) {
	if c.SectorIdentifierURI == nil {
		return ""
	}

	return c.SectorIdentifierURI.String()
}

// GetRedirectURIs returns the RedirectURIs.
func (c *RegisteredClient) GetRedirectURIs() (redirectURIs []string) {
	return c.RedirectURIs
}

// GetGrantTypes returns the GrantTypes.
func (c *RegisteredClient) GetGrantTypes() (types oauthelia2.Arguments) {
	if len(c.GrantTypes) == 0 {
		return oauthelia2.Arguments{"authorization_code"}
	}

	return c.GrantTypes
}

// GetResponseTypes returns the ResponseTypes.
func (c *RegisteredClient) GetResponseTypes() (types oauthelia2.Arguments) {
	if len(c.ResponseTypes) == 0 {
		return oauthelia2.Arguments{"code"}
	}

	return c.ResponseTypes
}

// GetScopes returns the Scopes.
func (c *RegisteredClient) GetScopes() (scopes oauthelia2.Arguments) {
	return c.Scopes
}

// GetAudience returns the Audience.
func (c *RegisteredClient) GetAudience() (audience oauthelia2.Arguments) {
	return c.Audience
}

// GetResponseModes returns the valid response modes for this client.
//
// Implements the oauthelia2.ResponseModeClient.
func (c *RegisteredClient) GetResponseModes() (modes []oauthelia2.ResponseModeType) {
	return c.ResponseModes
}

// GetAuthorizationSignedResponseAlg returns the AuthorizationSignedResponseAlg.
func (c *RegisteredClient) GetAuthorizationSignedResponseAlg() (alg string) {
	return c.AuthorizationSignedResponseAlg
}

// GetAuthorizationSignedResponseKeyID returns the AuthorizationSignedResponseKeyID.
func (c *RegisteredClient) GetAuthorizationSignedResponseKeyID() (kid string) {
	if c.AuthorizationSignedResponseKeyID == "" {
		c.AuthorizationSignedResponseKeyID = SigningAlgNone
	}

	return c.AuthorizationSignedResponseKeyID
}

func (c *RegisteredClient) GetAuthorizationEncryptedResponseAlg() (alg string) {
	return c.AuthorizationEncryptedResponseAlg
}

func (c *RegisteredClient) GetAuthorizationEncryptedResponseEncryptionAlg() (alg string) {
	return ""
}

// GetIDTokenSignedResponseAlg returns the IDTokenSignedResponseAlg.
func (c *RegisteredClient) GetIDTokenSignedResponseAlg() (alg string) {
	if c.IDTokenSignedResponseAlg == "" {
		c.IDTokenSignedResponseAlg = SigningAlgRSAUsingSHA256
	}

	return c.IDTokenSignedResponseAlg
}

// GetIDTokenSignedResponseKeyID returns the IDTokenSignedResponseKeyID.
func (c *RegisteredClient) GetIDTokenSignedResponseKeyID() (kid string) {
	return c.IDTokenSignedResponseKeyID
}

// GetAccessTokenSignedResponseAlg returns the AccessTokenSignedResponseAlg.
func (c *RegisteredClient) GetAccessTokenSignedResponseAlg() (alg string) {
	if c.AccessTokenSignedResponseAlg == "" {
		c.AccessTokenSignedResponseAlg = SigningAlgNone
	}

	return c.AccessTokenSignedResponseAlg
}

// GetAccessTokenSignedResponseKeyID returns the AccessTokenSignedResponseKeyID.
func (c *RegisteredClient) GetAccessTokenSignedResponseKeyID() (kid string) {
	return c.AccessTokenSignedResponseKeyID
}

// GetUserinfoSignedResponseAlg returns the UserinfoSignedResponseAlg.
func (c *RegisteredClient) GetUserinfoSignedResponseAlg() (alg string) {
	if c.UserinfoSignedResponseAlg == "" {
		c.UserinfoSignedResponseAlg = SigningAlgNone
	}

	return c.UserinfoSignedResponseAlg
}

// GetUserinfoSignedResponseKeyID returns the UserinfoSignedResponseKeyID.
func (c *RegisteredClient) GetUserinfoSignedResponseKeyID() (kid string) {
	return c.UserinfoSignedResponseKeyID
}

// GetIntrospectionSignedResponseAlg returns the IntrospectionSignedResponseAlg.
func (c *RegisteredClient) GetIntrospectionSignedResponseAlg() (alg string) {
	if c.IntrospectionSignedResponseAlg == "" {
		c.IntrospectionSignedResponseAlg = SigningAlgNone
	}

	return c.IntrospectionSignedResponseAlg
}

// GetIntrospectionSignedResponseKeyID returns the IntrospectionSignedResponseKeyID.
func (c *RegisteredClient) GetIntrospectionSignedResponseKeyID() (alg string) {
	return c.IntrospectionSignedResponseKeyID
}

// GetTokenEndpointAuthSigningAlg returns the JWS [JWS] alg algorithm [JWA] that MUST be used for signing the JWT
// [JWT] used to authenticate the Client at the Token Endpoint for the private_key_jwt and client_secret_jwt
// authentication methods.
func (c *RegisteredClient) GetTokenEndpointAuthSigningAlg() (alg string) {
	if c.TokenEndpointAuthSigningAlg == "" {
		c.TokenEndpointAuthSigningAlg = SigningAlgRSAUsingSHA256
	}

	return c.TokenEndpointAuthSigningAlg
}

// GetTokenEndpointAuthMethod returns the requested Client Authentication Method for the Token Endpoint. The options are
// client_secret_post, client_secret_basic, client_secret_jwt, private_key_jwt, and none.
func (c *RegisteredClient) GetTokenEndpointAuthMethod() (method string) {
	if c.TokenEndpointAuthMethod == "" {
		if c.Public {
			c.TokenEndpointAuthMethod = ClientAuthMethodNone
		} else {
			c.TokenEndpointAuthMethod = ClientAuthMethodClientSecretBasic
		}
	}

	return c.TokenEndpointAuthMethod
}

// GetIntrospectionEndpointAuthSigningAlg returns the JWS [JWS] alg algorithm [JWA] that MUST be used for signing the
// JWT [JWT] used to authenticate the Client at the Introspection Endpoint for the private_key_jwt and client_secret_jwt
// authentication methods.
func (c *RegisteredClient) GetIntrospectionEndpointAuthSigningAlg() (alg string) {
	return ""
}

// GetIntrospectionEndpointAuthMethod returns the requested Client Authentication Method for the Revocation Endpoint.
// The options are client_secret_post, client_secret_basic, client_secret_jwt, private_key_jwt, and none.
func (c *RegisteredClient) GetIntrospectionEndpointAuthMethod() (method string) {
	return ""
}

// GetRevocationEndpointAuthSigningAlg returns the JWS [JWS] alg algorithm [JWA] that MUST be used for signing the
// JWT [JWT] used to authenticate the Client at the Introspection Endpoint for the private_key_jwt and client_secret_jwt
// authentication methods.
func (c *RegisteredClient) GetRevocationEndpointAuthSigningAlg() (alg string) {
	return ""
}

// GetRevocationEndpointAuthMethod returns the requested Client Authentication Method for the Revocation Endpoint.
// The options are client_secret_post, client_secret_basic, client_secret_jwt, private_key_jwt, and none.
func (c *RegisteredClient) GetRevocationEndpointAuthMethod() (method string) {
	return ""
}

// GetEnableJWTProfileOAuthAccessTokens returns true if this client is configured to return the
// RFC9068 JWT Profile for OAuth 2.0 Access Tokens.
func (c *RegisteredClient) GetEnableJWTProfileOAuthAccessTokens() (enable bool) {
	return c.GetAccessTokenSignedResponseAlg() != SigningAlgNone || len(c.GetAccessTokenSignedResponseKeyID()) > 0
}

// GetRequirePushedAuthorizationRequests should return true if this client MUST use a Pushed Authorization Request.
func (c *RegisteredClient) GetRequirePushedAuthorizationRequests() (require bool) {
	return c.RequirePushedAuthorizationRequests
}

// GetPushedAuthorizeContextLifespan should return a custom lifespan or a duration of 0 seconds to utilize the
// global lifespan.
func (c *RegisteredClient) GetPushedAuthorizeContextLifespan() (lifespan time.Duration) {
	return lifespan
}

// GetEnforcePKCE returns RequirePKCE.
func (c *RegisteredClient) GetEnforcePKCE() (enforce bool) {
	return c.RequirePKCE
}

// GetEnforcePKCEChallengeMethod returns RequirePKCEChallengeMethod.
func (c *RegisteredClient) GetEnforcePKCEChallengeMethod() (enforce bool) {
	return c.RequirePKCEChallengeMethod
}

// GetPKCEChallengeMethod returns PKCEChallengeMethod.
func (c *RegisteredClient) GetPKCEChallengeMethod() (method string) {
	return c.PKCEChallengeMethod
}

// GetConsentResponseBody returns the proper consent response body for this session.OIDCWorkflowSession.
func (c *RegisteredClient) GetConsentResponseBody(consent *model.OAuth2ConsentSession) ConsentGetResponseBody {
	body := ConsentGetResponseBody{
		ClientID:          c.ID,
		ClientDescription: c.Name,
		PreConfiguration:  c.ConsentPolicy.Mode == ClientConsentModePreConfigured,
	}

	if consent != nil {
		body.Scopes = consent.RequestedScopes
		body.Audience = consent.RequestedAudience
	}

	return body
}

// GetConsentPolicy returns Consent.
func (c *RegisteredClient) GetConsentPolicy() (policy ClientConsentPolicy) {
	return c.ConsentPolicy
}

// IsAuthenticationLevelSufficient returns if the provided authentication.Level is sufficient for the client of the AutheliaClient.
func (c *RegisteredClient) IsAuthenticationLevelSufficient(level authentication.Level, subject authorization.Subject) (sufficient bool) {
	if level == authentication.NotAuthenticated {
		return false
	}

	return authorization.IsAuthLevelSufficient(level, c.GetAuthorizationPolicyRequiredLevel(subject))
}

// GetAuthorizationPolicyRequiredLevel returns the required authorization.Level given an authorization.Subject.
func (c *RegisteredClient) GetAuthorizationPolicyRequiredLevel(subject authorization.Subject) (level authorization.Level) {
	return c.AuthorizationPolicy.GetRequiredLevel(subject)
}

// GetAuthorizationPolicy returns the ClientAuthorizationPolicy from the Policy.
func (c *RegisteredClient) GetAuthorizationPolicy() (policy ClientAuthorizationPolicy) {
	return c.AuthorizationPolicy
}

// IsPublic returns the value of the Public property.
func (c *RegisteredClient) IsPublic() (public bool) {
	return c.Public
}

// ValidateResponseModePolicy is an additional check to the response mode parameter to ensure if it's omitted that the
// default response mode for the oauthelia2.AuthorizeRequester is permitted.
func (c *RegisteredClient) ValidateResponseModePolicy(r oauthelia2.AuthorizeRequester) (err error) {
	if r.GetResponseMode() != oauthelia2.ResponseModeDefault {
		return nil
	}

	m := r.GetDefaultResponseMode()

	modes := c.GetResponseModes()

	if len(modes) == 0 {
		return nil
	}

	for _, mode := range modes {
		if m == mode {
			return nil
		}
	}

	return errorsx.WithStack(oauthelia2.ErrUnsupportedResponseMode.WithHintf(`The request omitted the response_mode making the default response_mode "%s" based on the other authorization request parameters but registered OAuth 2.0 client doesn't support this response_mode`, m))
}

// GetRefreshFlowIgnoreOriginalGrantedScopes returns the value which indicates if the client should ignore the
// originally granted scopes when the scope parameter is present. The specification requires that this is always false,
// however some misbehaving clients may need this option.
func (c *RegisteredClient) GetRefreshFlowIgnoreOriginalGrantedScopes(ctx context.Context) (ignore bool) {
	return c.RefreshFlowIgnoreOriginalGrantedScopes
}

func (c *RegisteredClient) GetRevokeRefreshTokensExplicit(ctx context.Context) (explicit bool) {
	return false
}

// GetRequestURIs is an array of request_uri values that are pre-registered by the RP for use at the OP. Servers MAY
// cache the contents of the files referenced by these URIs and not retrieve them at the time they are used in a request.
// OPs can require that request_uri values used be pre-registered with the require_request_uri_registration
// discovery parameter.
func (c *RegisteredClient) GetRequestURIs() (uris []string) {
	return c.RequestURIs
}

// GetJSONWebKeys returns the JSON Web Key Set containing the public key used by the client to authenticate.
func (c *RegisteredClient) GetJSONWebKeys() (keys *jose.JSONWebKeySet) {
	return c.JSONWebKeys
}

// SetJSONWebKeys sets the JSON Web Key Set containing the public key used by the client to authenticate.
func (c *RegisteredClient) SetJSONWebKeys(jwks *jose.JSONWebKeySet) {
	c.JSONWebKeys = jwks
}

// GetJSONWebKeysURI returns the URL for lookup of JSON Web Key Set containing the
// public key used by the client to authenticate.
func (c *RegisteredClient) GetJSONWebKeysURI() (uri string) {
	if c.JSONWebKeysURI == nil {
		return ""
	}

	return c.JSONWebKeysURI.String()
}

// GetRequestObjectSigningAlg returns the JWS [JWS] alg algorithm [JWA] that MUST be used for signing Request
// Objects sent to the OP. All Request Objects from this Client MUST be rejected, if not signed with this algorithm.
func (c *RegisteredClient) GetRequestObjectSigningAlg() (alg string) {
	return c.RequestObjectSigningAlg
}

func (c *RegisteredClient) GetAllowMultipleAuthenticationMethods() (allow bool) {
	return c.AllowMultipleAuthenticationMethods
}

func (c *RegisteredClient) GetClientCredentialsFlowRequestedScopeImplicit() (allow bool) {
	return c.ClientCredentialsFlowAllowImplicitScope
}

func (c *RegisteredClient) GetRequestedAudienceImplicit() (implicit bool) {
	return c.RequestedAudienceMode == ClientRequestedAudienceModeImplicit
}

// GetEffectiveLifespan returns the effective lifespan for a grant type and token type otherwise returns the fallback
// value. This implements the oauthelia2.ClientWithCustomTokenLifespans interface.
func (c *RegisteredClient) GetEffectiveLifespan(gt oauthelia2.GrantType, tt oauthelia2.TokenType, fallback time.Duration) time.Duration {
	gtl := c.getGrantTypeLifespan(gt)

	switch tt {
	case oauthelia2.AccessToken:
		switch {
		case gtl.AccessToken > durationZero:
			return gtl.AccessToken
		case c.Lifespans.AccessToken > durationZero:
			return c.Lifespans.AccessToken
		default:
			return fallback
		}
	case oauthelia2.AuthorizeCode:
		switch {
		case gtl.AuthorizeCode > durationZero:
			return gtl.AuthorizeCode
		case c.Lifespans.AuthorizeCode > durationZero:
			return c.Lifespans.AuthorizeCode
		default:
			return fallback
		}
	case oauthelia2.IDToken:
		switch {
		case gtl.IDToken > durationZero:
			return gtl.IDToken
		case c.Lifespans.IDToken > durationZero:
			return c.Lifespans.IDToken
		default:
			return fallback
		}
	case oauthelia2.RefreshToken:
		switch {
		case gtl.RefreshToken > durationZero:
			return gtl.RefreshToken
		case c.Lifespans.RefreshToken > durationZero:
			return c.Lifespans.RefreshToken
		default:
			return fallback
		}
	default:
		return fallback
	}
}

func (c *RegisteredClient) getGrantTypeLifespan(gt oauthelia2.GrantType) (gtl schema.IdentityProvidersOpenIDConnectLifespanToken) {
	switch gt {
	case oauthelia2.GrantTypeAuthorizationCode:
		return c.Lifespans.Grants.AuthorizeCode
	case oauthelia2.GrantTypeImplicit:
		return c.Lifespans.Grants.Implicit
	case oauthelia2.GrantTypeClientCredentials:
		return c.Lifespans.Grants.ClientCredentials
	case oauthelia2.GrantTypeRefreshToken:
		return c.Lifespans.Grants.RefreshToken
	case oauthelia2.GrantTypeJWTBearer:
		return c.Lifespans.Grants.JWTBearer
	default:
		return gtl
	}
}
