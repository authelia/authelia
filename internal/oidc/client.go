package oidc

import (
	"context"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-jose/go-jose/v4"
	"github.com/ory/x/errorsx"

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

		RequirePKCE:                config.RequirePKCE || config.PKCEChallengeMethod != "",
		RequirePKCEChallengeMethod: config.PKCEChallengeMethod != "",
		PKCEChallengeMethod:        config.PKCEChallengeMethod,

		Audience:      config.Audience,
		Scopes:        config.Scopes,
		RedirectURIs:  config.RedirectURIs,
		GrantTypes:    config.GrantTypes,
		ResponseTypes: config.ResponseTypes,
		ResponseModes: []oauthelia2.ResponseModeType{},

		RequirePushedAuthorizationRequests: config.RequirePushedAuthorizationRequests,

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

		AuthorizationPolicy:         NewClientAuthorizationPolicy(config.AuthorizationPolicy, c),
		ConsentPolicy:               NewClientConsentPolicy(config.ConsentMode, config.ConsentPreConfiguredDuration),
		RequestedAudienceMode:       NewClientRequestedAudienceMode(config.RequestedAudienceMode),
		TokenEndpointAuthMethod:     config.TokenEndpointAuthMethod,
		TokenEndpointAuthSigningAlg: config.TokenEndpointAuthSigningAlg,
		RequestObjectSigningAlg:     config.RequestObjectSigningAlg,

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

// GetSecret returns the Secret.
func (c *RegisteredClient) GetSecret() algorithm.Digest {
	return c.Secret
}

// GetClientSecret returns the oauth2.ClientSecret.
func (c *RegisteredClient) GetClientSecret() oauthelia2.ClientSecret {
	return c.ClientSecret
}

// GetRotatedClientSecrets returns the rotated oauth2.ClientSecret values.
func (c *RegisteredClient) GetRotatedClientSecrets() []oauthelia2.ClientSecret {
	return nil
}

// GetSectorIdentifier returns the SectorIdentifier for this client.
func (c *RegisteredClient) GetSectorIdentifier() string {
	if c.SectorIdentifierURI == nil {
		return ""
	}

	return c.SectorIdentifierURI.String()
}

// GetHashedSecret returns the Secret.
func (c *RegisteredClient) GetHashedSecret() (secret []byte) {
	if c.Secret == nil {
		return []byte(nil)
	}

	return []byte(c.Secret.Encode())
}

// GetRedirectURIs returns the RedirectURIs.
func (c *RegisteredClient) GetRedirectURIs() (redirectURIs []string) {
	return c.RedirectURIs
}

// GetGrantTypes returns the GrantTypes.
func (c *RegisteredClient) GetGrantTypes() oauthelia2.Arguments {
	if len(c.GrantTypes) == 0 {
		return oauthelia2.Arguments{"authorization_code"}
	}

	return c.GrantTypes
}

// GetResponseTypes returns the ResponseTypes.
func (c *RegisteredClient) GetResponseTypes() oauthelia2.Arguments {
	if len(c.ResponseTypes) == 0 {
		return oauthelia2.Arguments{"code"}
	}

	return c.ResponseTypes
}

// GetScopes returns the Scopes.
func (c *RegisteredClient) GetScopes() oauthelia2.Arguments {
	return c.Scopes
}

// GetAudience returns the Audience.
func (c *RegisteredClient) GetAudience() oauthelia2.Arguments {
	return c.Audience
}

// GetResponseModes returns the valid response modes for this client.
//
// Implements the oauthelia2.ResponseModeClient.
func (c *RegisteredClient) GetResponseModes() []oauthelia2.ResponseModeType {
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
	return ""
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
func (c *RegisteredClient) GetIDTokenSignedResponseKeyID() (alg string) {
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
func (c *RegisteredClient) GetAccessTokenSignedResponseKeyID() (alg string) {
	return c.AccessTokenSignedResponseKeyID
}

// GetEnableJWTProfileOAuthAccessTokens returns true if this client is configured to return the
// RFC9068 JWT Profile for OAuth 2.0 Access Tokens.
func (c *RegisteredClient) GetEnableJWTProfileOAuthAccessTokens() bool {
	return c.GetAccessTokenSignedResponseAlg() != SigningAlgNone || len(c.GetAccessTokenSignedResponseKeyID()) > 0
}

// GetUserinfoSignedResponseAlg returns the UserinfoSignedResponseAlg.
func (c *RegisteredClient) GetUserinfoSignedResponseAlg() string {
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

// GetRequirePushedAuthorizationRequests returns RequirePushedAuthorizationRequests.
func (c *RegisteredClient) GetRequirePushedAuthorizationRequests() bool {
	return c.RequirePushedAuthorizationRequests
}

// GetPKCEEnforcement returns RequirePKCE.
func (c *RegisteredClient) GetPKCEEnforcement() bool {
	return c.RequirePKCE
}

// GetPKCEChallengeMethodEnforcement returns RequirePKCEChallengeMethod.
func (c *RegisteredClient) GetPKCEChallengeMethodEnforcement() bool {
	return c.RequirePKCEChallengeMethod
}

// GetPKCEChallengeMethod returns PKCEChallengeMethod.
func (c *RegisteredClient) GetPKCEChallengeMethod() string {
	return c.PKCEChallengeMethod
}

// ApplyRequestedAudiencePolicy applies the requested audience policy to a oauthelia2.Requester.
func (c *RegisteredClient) ApplyRequestedAudiencePolicy(requester oauthelia2.Requester) {
	switch c.RequestedAudienceMode {
	case ClientRequestedAudienceModeExplicit:
		return
	case ClientRequestedAudienceModeImplicit:
		if requester.GetRequestForm().Has(FormParameterAudience) || len(requester.GetRequestedAudience()) != 0 {
			return
		}

		requester.SetRequestedAudience(c.Audience)
	}
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
func (c *RegisteredClient) GetConsentPolicy() ClientConsentPolicy {
	return c.ConsentPolicy
}

// IsAuthenticationLevelSufficient returns if the provided authentication.Level is sufficient for the client of the AutheliaClient.
func (c *RegisteredClient) IsAuthenticationLevelSufficient(level authentication.Level, subject authorization.Subject) bool {
	if level == authentication.NotAuthenticated {
		return false
	}

	return authorization.IsAuthLevelSufficient(level, c.GetAuthorizationPolicyRequiredLevel(subject))
}

// GetAuthorizationPolicyRequiredLevel returns the required authorization.Level given an authorization.Subject.
func (c *RegisteredClient) GetAuthorizationPolicyRequiredLevel(subject authorization.Subject) authorization.Level {
	return c.AuthorizationPolicy.GetRequiredLevel(subject)
}

// GetAuthorizationPolicy returns the ClientAuthorizationPolicy from the Policy.
func (c *RegisteredClient) GetAuthorizationPolicy() ClientAuthorizationPolicy {
	return c.AuthorizationPolicy
}

// IsPublic returns the value of the Public property.
func (c *RegisteredClient) IsPublic() bool {
	return c.Public
}

// ValidatePKCEPolicy is a helper function to validate PKCE policy constraints on a per-client basis.
func (c *RegisteredClient) ValidatePKCEPolicy(r oauthelia2.Requester) (err error) {
	form := r.GetRequestForm()

	if c.RequirePKCE {
		if form.Get(FormParameterCodeChallenge) == "" {
			return errorsx.WithStack(oauthelia2.ErrInvalidRequest.
				WithHint("Clients must include a code_challenge when performing the authorize code flow, but it is missing.").
				WithDebug("The server is configured in a way that enforces PKCE for this client."))
		}

		if c.RequirePKCEChallengeMethod {
			if method := form.Get(FormParameterCodeChallengeMethod); method != c.PKCEChallengeMethod {
				return errorsx.WithStack(oauthelia2.ErrInvalidRequest.
					WithHintf("Client must use code_challenge_method=%s, %s is not allowed.", c.PKCEChallengeMethod, method).
					WithDebugf("The server is configured in a way that enforces PKCE %s as challenge method for this client.", c.PKCEChallengeMethod))
			}
		}
	}

	return nil
}

// ValidatePARPolicy is a helper function to validate additional policy constraints on a per-client basis.
func (c *RegisteredClient) ValidatePARPolicy(r oauthelia2.Requester, prefix string) (err error) {
	if c.RequirePushedAuthorizationRequests {
		if !IsPushedAuthorizedRequest(r, prefix) {
			switch requestURI := r.GetRequestForm().Get(FormParameterRequestURI); requestURI {
			case "":
				return errorsx.WithStack(ErrPAREnforcedClientMissingPAR.WithDebug("The request_uri parameter was empty."))
			default:
				return errorsx.WithStack(ErrPAREnforcedClientMissingPAR.WithDebugf("The request_uri parameter '%s' is malformed.", requestURI))
			}
		}
	}

	return nil
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
func (c *RegisteredClient) GetRefreshFlowIgnoreOriginalGrantedScopes(ctx context.Context) (ignoreOriginalGrantedScopes bool) {
	return c.RefreshFlowIgnoreOriginalGrantedScopes
}

func (c *RegisteredClient) GetRevokeRefreshTokensExplicit(ctx context.Context) (explicit bool) {
	return false
}

// GetRequestURIs is an array of request_uri values that are pre-registered by the RP for use at the OP. Servers MAY
// cache the contents of the files referenced by these URIs and not retrieve them at the time they are used in a request.
// OPs can require that request_uri values used be pre-registered with the require_request_uri_registration
// discovery parameter.
func (c *RegisteredClient) GetRequestURIs() []string {
	return c.RequestURIs
}

// GetJSONWebKeys returns the JSON Web Key Set containing the public key used by the client to authenticate.
func (c *RegisteredClient) GetJSONWebKeys() *jose.JSONWebKeySet {
	return c.JSONWebKeys
}

// SetJSONWebKeys sets the JSON Web Key Set containing the public key used by the client to authenticate.
func (c *RegisteredClient) SetJSONWebKeys(jwks *jose.JSONWebKeySet) {
	c.JSONWebKeys = jwks
}

// GetJSONWebKeysURI returns the URL for lookup of JSON Web Key Set containing the
// public key used by the client to authenticate.
func (c *RegisteredClient) GetJSONWebKeysURI() string {
	if c.JSONWebKeysURI == nil {
		return ""
	}

	return c.JSONWebKeysURI.String()
}

// GetRequestObjectSigningAlgorithm returns the JWS [JWS] alg algorithm [JWA] that MUST be used for signing Request
// Objects sent to the OP. All Request Objects from this Client MUST be rejected, if not signed with this algorithm.
func (c *RegisteredClient) GetRequestObjectSigningAlgorithm() string {
	return c.RequestObjectSigningAlg
}

// GetTokenEndpointAuthMethod returns the requested Client Authentication Method for the Token Endpoint. The options are
// client_secret_post, client_secret_basic, client_secret_jwt, private_key_jwt, and none.
func (c *RegisteredClient) GetTokenEndpointAuthMethod() string {
	if c.TokenEndpointAuthMethod == "" {
		if c.Public {
			c.TokenEndpointAuthMethod = ClientAuthMethodNone
		} else {
			c.TokenEndpointAuthMethod = ClientAuthMethodClientSecretBasic
		}
	}

	return c.TokenEndpointAuthMethod
}

// GetTokenEndpointAuthSigningAlgorithm returns the JWS [JWS] alg algorithm [JWA] that MUST be used for signing the JWT
// [JWT] used to authenticate the Client at the Token Endpoint for the private_key_jwt and client_secret_jwt
// authentication methods.
func (c *RegisteredClient) GetTokenEndpointAuthSigningAlgorithm() string {
	if c.TokenEndpointAuthSigningAlg == "" {
		c.TokenEndpointAuthSigningAlg = SigningAlgRSAUsingSHA256
	}

	return c.TokenEndpointAuthSigningAlg
}

func (c *RegisteredClient) GetAllowMultipleAuthenticationMethods(ctx context.Context) (allow bool) {
	return false
}

func (c *RegisteredClient) GetClientCredentialsFlowAllowImplicitScope() (allow bool) {
	return false
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
