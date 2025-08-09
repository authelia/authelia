package oidc

import (
	"context"
	"net/url"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jwt"
	"authelia.com/provider/oauth2/x/errorsx"
	"github.com/go-jose/go-jose/v4"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewClient creates a new Client.
func NewClient(config schema.IdentityProvidersOpenIDConnectClient, c *schema.IdentityProvidersOpenIDConnect, policies map[string]ClientAuthorizationPolicy) (client Client) {
	registered := &RegisteredClient{
		ID:                  config.ID,
		Name:                config.Name,
		ClientSecret:        &ClientSecretDigest{PasswordDigest: config.Secret},
		SectorIdentifierURI: config.SectorIdentifierURI,
		Public:              config.Public,

		Audience:      config.Audience,
		Scopes:        config.Scopes,
		RedirectURIs:  config.RedirectURIs,
		RequestURIs:   config.RequestURIs,
		GrantTypes:    config.GrantTypes,
		ResponseTypes: config.ResponseTypes,
		ResponseModes: []oauthelia2.ResponseModeType{},

		ClaimsStrategy: NewCustomClaimsStrategyFromClient(config, c.Scopes, c.ClaimsPolicies),

		RequirePKCE:                config.RequirePKCE || config.PKCEChallengeMethod != "",
		RequirePKCEChallengeMethod: config.PKCEChallengeMethod != "",
		PKCEChallengeMethod:        config.PKCEChallengeMethod,

		RequirePushedAuthorizationRequests:      config.RequirePushedAuthorizationRequests,
		ClientCredentialsFlowAllowImplicitScope: false,
		AllowMultipleAuthenticationMethods:      config.AllowMultipleAuthenticationMethods,

		ConsentPolicy:         NewClientConsentPolicy(config.ConsentMode, config.ConsentPreConfiguredDuration),
		RequestedAudienceMode: NewClientRequestedAudienceMode(config.RequestedAudienceMode),

		AuthorizationSignedResponseAlg:                   config.AuthorizationSignedResponseAlg,
		AuthorizationSignedResponseKeyID:                 config.AuthorizationSignedResponseKeyID,
		AuthorizationEncryptedResponseAlg:                config.AuthorizationEncryptedResponseAlg,
		AuthorizationEncryptedResponseEnc:                config.AuthorizationEncryptedResponseEnc,
		AuthorizationEncryptedResponseKeyID:              config.AuthorizationEncryptedResponseKeyID,
		IDTokenSignedResponseAlg:                         config.IDTokenSignedResponseAlg,
		IDTokenSignedResponseKeyID:                       config.IDTokenSignedResponseKeyID,
		IDTokenEncryptedResponseAlg:                      config.IDTokenEncryptedResponseAlg,
		IDTokenEncryptedResponseEnc:                      config.IDTokenEncryptedResponseEnc,
		IDTokenEncryptedResponseKeyID:                    config.IDTokenEncryptedResponseKeyID,
		AccessTokenSignedResponseAlg:                     config.AccessTokenSignedResponseAlg,
		AccessTokenSignedResponseKeyID:                   config.AccessTokenSignedResponseKeyID,
		AccessTokenEncryptedResponseAlg:                  config.AccessTokenEncryptedResponseAlg,
		AccessTokenEncryptedResponseEnc:                  config.AccessTokenEncryptedResponseEnc,
		AccessTokenEncryptedResponseKeyID:                config.AccessTokenEncryptedResponseKeyID,
		UserinfoSignedResponseAlg:                        config.UserinfoSignedResponseAlg,
		UserinfoSignedResponseKeyID:                      config.UserinfoSignedResponseKeyID,
		UserinfoEncryptedResponseAlg:                     config.UserinfoEncryptedResponseAlg,
		UserinfoEncryptedResponseEnc:                     config.UserinfoEncryptedResponseEnc,
		UserinfoEncryptedResponseKeyID:                   config.UserinfoEncryptedResponseKeyID,
		IntrospectionSignedResponseAlg:                   config.IntrospectionSignedResponseAlg,
		IntrospectionSignedResponseKeyID:                 config.IntrospectionSignedResponseKeyID,
		IntrospectionEncryptedResponseAlg:                config.IntrospectionEncryptedResponseAlg,
		IntrospectionEncryptedResponseEnc:                config.IntrospectionEncryptedResponseEnc,
		IntrospectionEncryptedResponseKeyID:              config.IntrospectionEncryptedResponseKeyID,
		RequestObjectSigningAlg:                          config.RequestObjectSigningAlg,
		RequestObjectEncryptionAlg:                       config.RequestObjectEncryptionAlg,
		RequestObjectEncryptionEnc:                       config.RequestObjectEncryptionEnc,
		TokenEndpointAuthMethod:                          config.TokenEndpointAuthMethod,
		TokenEndpointAuthSigningAlg:                      config.TokenEndpointAuthSigningAlg,
		RevocationEndpointAuthMethod:                     config.RevocationEndpointAuthMethod,
		RevocationEndpointAuthSigningAlg:                 config.RevocationEndpointAuthSigningAlg,
		IntrospectionEndpointAuthMethod:                  config.IntrospectionEndpointAuthMethod,
		IntrospectionEndpointAuthSigningAlg:              config.IntrospectionEndpointAuthSigningAlg,
		PushedAuthorizationRequestEndpointAuthMethod:     config.PushedAuthorizationRequestEndpointAuthMethod,
		PushedAuthorizationRequestEndpointAuthSigningAlg: config.PushedAuthorizationRequestAuthSigningAlg,

		JSONWebKeysURI: config.JSONWebKeysURI,
		JSONWebKeys:    NewJSONWebKeySet(config.JSONWebKeys),
	}

	if policies == nil {
		registered.AuthorizationPolicy = ClientAuthorizationPolicy{DefaultPolicy: authorization.TwoFactor}
	} else if policy, ok := policies[config.AuthorizationPolicy]; ok {
		registered.AuthorizationPolicy = policy
	} else {
		registered.AuthorizationPolicy = ClientAuthorizationPolicy{DefaultPolicy: authorization.TwoFactor}
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

// GetClientSecretPlainText returns the ClientSecret as plaintext if available. The semantics of this function
// return values are important.
// If the client is not configured with a secret the return should be:
//   - secret with value nil, ok with value false, and err with value of nil
//
// If the client is configured with a secret but is hashed or otherwise not a plaintext value:
//   - secret with value nil, ok with value true, and err with value of nil
//
// If an error occurs retrieving the secret other than this:
//   - secret with value nil, ok with value true, and err with value of the error
//
// If the plaintext secret is successful:
//   - secret with value of the bytes of the plaintext secret, ok with value true, and err with value of nil
func (c *RegisteredClient) GetClientSecretPlainText() (secret []byte, ok bool, err error) {
	if c.ClientSecret == nil || !c.ClientSecret.Valid() {
		return nil, false, nil
	}

	if !c.ClientSecret.IsPlainText() {
		return nil, true, nil
	}

	if secret, err = c.ClientSecret.GetPlainTextValue(); err != nil {
		return nil, true, err
	}

	return secret, true, nil
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

func (c *RegisteredClient) GetClaimsStrategy() (strategy ClaimsStrategy) {
	return c.ClaimsStrategy
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

// GetAuthorizationSignedResponseKeyID returns the AuthorizationSignedResponseKeyID.
func (c *RegisteredClient) GetAuthorizationSignedResponseKeyID() (kid string) {
	return c.AuthorizationSignedResponseKeyID
}

// GetAuthorizationSignedResponseAlg returns the AuthorizationSignedResponseAlg.
func (c *RegisteredClient) GetAuthorizationSignedResponseAlg() (alg string) {
	if c.AuthorizationSignedResponseAlg == "" {
		c.AuthorizationSignedResponseAlg = SigningAlgRSAUsingSHA256
	}

	return c.AuthorizationSignedResponseAlg
}

// GetAuthorizationEncryptedResponseKeyID returns the specific key identifier used to satisfy JWE requirements of
// the JWT-secured Authorization Response Method (JARM) specifications. If unspecified the other available parameters will be
// utilized to select an appropriate key.
func (c *RegisteredClient) GetAuthorizationEncryptedResponseKeyID() (kid string) {
	return c.AuthorizationEncryptedResponseKeyID
}

// GetAuthorizationEncryptedResponseAlg is equivalent to the 'authorization_encrypted_response_alg' client metadata
// value which determines the JWE [RFC7516] alg algorithm JWA [RFC7518] REQUIRED for encrypting authorization
// responses. If both signing and encryption are requested, the response will be signed then encrypted, with the
// result being a Nested JWT, as defined in JWT [RFC7519]. The default, if omitted, is that no encryption is
// performed.
func (c *RegisteredClient) GetAuthorizationEncryptedResponseAlg() (alg string) {
	return c.AuthorizationEncryptedResponseAlg
}

// GetAuthorizationEncryptedResponseEnc is equivalent to the 'authorization_encrypted_response_enc' client
// metadata value which determines the JWE [RFC7516] enc algorithm JWA [RFC7518] REQUIRED for encrypting
// authorization responses. If authorization_encrypted_response_alg is specified, the default for this value is
// A128CBC-HS256. When authorization_encrypted_response_enc is included, authorization_encrypted_response_alg MUST
// also be provided.
func (c *RegisteredClient) GetAuthorizationEncryptedResponseEnc() (enc string) {
	return c.AuthorizationEncryptedResponseEnc
}

// GetIDTokenSignedResponseKeyID returns the specific key identifier used to satisfy JWS requirements of the ID
// Token specifications. If unspecified the other available parameters will be utilized to select an appropriate
// key.
func (c *RegisteredClient) GetIDTokenSignedResponseKeyID() (kid string) {
	return c.IDTokenSignedResponseKeyID
}

// GetIDTokenSignedResponseAlg is equivalent to the 'id_token_signed_response_alg' client metadata value which
// determines the JWS alg algorithm [JWA] REQUIRED for signing the ID Token issued to this Client. The value none
// MUST NOT be used as the ID Token alg value unless the Client uses only Response Types that return no ID Token
// from the Authorization Endpoint (such as when only using the Authorization Code Flow). The default, if omitted,
// is RS256. The public key for validating the signature is provided by retrieving the JWK Set referenced by the
// jwks_uri element from OpenID Connect Discovery 1.0 [OpenID.Discovery].
func (c *RegisteredClient) GetIDTokenSignedResponseAlg() (alg string) {
	if c.IDTokenSignedResponseAlg == "" {
		c.IDTokenSignedResponseAlg = SigningAlgRSAUsingSHA256
	}

	return c.IDTokenSignedResponseAlg
}

// GetIDTokenEncryptedResponseKeyID returns the specific key identifier used to satisfy JWE requirements of the ID
// Token specifications. If unspecified the other available parameters will be utilized to select an appropriate
// key.
func (c *RegisteredClient) GetIDTokenEncryptedResponseKeyID() (kid string) {
	return c.IDTokenEncryptedResponseKeyID
}

// GetIDTokenEncryptedResponseAlg is equivalent to the 'id_token_encrypted_response_alg' client metadata value which
// determines the JWE alg algorithm [JWA] REQUIRED for encrypting the ID Token issued to this Client. If this is
// requested, the response will be signed then encrypted, with the result being a Nested JWT, as defined in [JWT].
// The default, if omitted, is that no encryption is performed.
func (c *RegisteredClient) GetIDTokenEncryptedResponseAlg() (alg string) {
	return c.IDTokenEncryptedResponseAlg
}

// GetIDTokenEncryptedResponseEnc is equivalent to the 'id_token_encrypted_response_enc' client metadata value which
// determines the JWE enc algorithm [JWA] REQUIRED for encrypting the ID Token issued to this Client. If
// id_token_encrypted_response_alg is specified, the default id_token_encrypted_response_enc value is A128CBC-HS256.
// When id_token_encrypted_response_enc is included, id_token_encrypted_response_alg MUST also be provided.
func (c *RegisteredClient) GetIDTokenEncryptedResponseEnc() (enc string) {
	return c.IDTokenEncryptedResponseEnc
}

// GetAccessTokenSignedResponseKeyID returns the specific key identifier used to satisfy JWS requirements for
// JWT Profile for OAuth 2.0 Access Tokens specifications. If unspecified the other available parameters will be
// utilized to select an appropriate key.
func (c *RegisteredClient) GetAccessTokenSignedResponseKeyID() (kid string) {
	return c.AccessTokenSignedResponseKeyID
}

// GetAccessTokenSignedResponseAlg determines the JWS [RFC7515] algorithm (alg value) as defined in JWA [RFC7518]
// for signing JWT Profile Access Token responses. If this is specified, the response will be signed using JWS and
// the configured algorithm. The default, if omitted, is none; i.e. unsigned responses unless the
// GetEnableJWTProfileOAuthAccessTokens receiver returns true in which case the default is RS256.
func (c *RegisteredClient) GetAccessTokenSignedResponseAlg() (alg string) {
	if c.AccessTokenSignedResponseAlg == "" {
		c.AccessTokenSignedResponseAlg = SigningAlgNone
	}

	return c.AccessTokenSignedResponseAlg
}

// GetAccessTokenEncryptedResponseKeyID returns the specific key identifier used to satisfy JWE requirements for
// JWT Profile for OAuth 2.0 Access Tokens specifications. If unspecified the other available parameters will be
// utilized to select an appropriate key.
func (c *RegisteredClient) GetAccessTokenEncryptedResponseKeyID() (kid string) {
	return c.AccessTokenEncryptedResponseKeyID
}

// GetAccessTokenEncryptedResponseAlg determines the JWE [RFC7516] algorithm (alg value) as defined in JWA [RFC7518]
// for content key encryption. If this is specified, the response will be encrypted using JWE and the configured
// content encryption algorithm (access_token_encrypted_response_enc). The default, if omitted, is that no
// encryption is performed. If both signing and encryption are requested, the response will be signed then
// encrypted, with the result being a Nested JWT, as defined in JWT [RFC7519].
func (c *RegisteredClient) GetAccessTokenEncryptedResponseAlg() (alg string) {
	return c.AccessTokenEncryptedResponseAlg
}

// GetAccessTokenEncryptedResponseEnc determines the JWE [RFC7516] algorithm (enc value) as defined in JWA [RFC7518]
// for content encryption of access token responses. The default, if omitted, is A128CBC-HS256. Note: This parameter
// MUST NOT be specified without setting access_token_encrypted_response_alg.
func (c *RegisteredClient) GetAccessTokenEncryptedResponseEnc() (enc string) {
	return c.AccessTokenEncryptedResponseEnc
}

// GetUserinfoSignedResponseKeyID returns the specific key identifier used to satisfy JWS requirements of the User
// Info specifications. If unspecified the other available parameters will be utilized to select an appropriate
// key.
func (c *RegisteredClient) GetUserinfoSignedResponseKeyID() (kid string) {
	return c.UserinfoSignedResponseKeyID
}

// GetUserinfoSignedResponseAlg is equivalent to the 'userinfo_signed_response_alg' client metadata value which
// determines the JWS alg algorithm [JWA] REQUIRED for signing UserInfo Responses. If this is specified, the
// response will be JWT [JWT] serialized, and signed using JWS. The default, if omitted, is for the UserInfo
// Response to return the Claims as a UTF-8 [RFC3629] encoded JSON object using the application/json content-type.
func (c *RegisteredClient) GetUserinfoSignedResponseAlg() (alg string) {
	if c.UserinfoSignedResponseAlg == "" {
		c.UserinfoSignedResponseAlg = SigningAlgNone
	}

	return c.UserinfoSignedResponseAlg
}

// GetUserinfoEncryptedResponseKeyID returns the specific key identifier used to satisfy JWE requirements of the
// User Info specifications. If unspecified the other available parameters will be utilized to select an appropriate
// key.
func (c *RegisteredClient) GetUserinfoEncryptedResponseKeyID() (kid string) {
	return c.UserinfoEncryptedResponseKeyID
}

// GetUserinfoEncryptedResponseAlg is equivalent to the 'userinfo_encrypted_response_alg' client metadata value
// which determines the JWE alg algorithm [JWA] REQUIRED for encrypting the ID Token issued to this Client. If
// this is requested, the response will be signed then encrypted, with the result being a Nested JWT, as defined in
// [JWT]. The default, if omitted, is that no encryption is performed.
func (c *RegisteredClient) GetUserinfoEncryptedResponseAlg() (alg string) {
	return c.UserinfoEncryptedResponseAlg
}

// GetUserinfoEncryptedResponseEnc is equivalent to the 'userinfo_encrypted_response_enc' client metadata value
// which determines the JWE enc algorithm [JWA] REQUIRED for encrypting UserInfo Responses. If
// userinfo_encrypted_response_alg is specified, the default userinfo_encrypted_response_enc value is A128CBC-HS256.
// When userinfo_encrypted_response_enc is included, userinfo_encrypted_response_alg MUST also be provided.
func (c *RegisteredClient) GetUserinfoEncryptedResponseEnc() (enc string) {
	return c.UserinfoEncryptedResponseEnc
}

// GetIntrospectionSignedResponseKeyID returns the IntrospectionSignedResponseKeyID.
func (c *RegisteredClient) GetIntrospectionSignedResponseKeyID() (alg string) {
	return c.IntrospectionSignedResponseKeyID
}

// GetIntrospectionSignedResponseAlg returns the IntrospectionSignedResponseAlg.
func (c *RegisteredClient) GetIntrospectionSignedResponseAlg() (alg string) {
	if c.IntrospectionSignedResponseAlg == "" {
		c.IntrospectionSignedResponseAlg = SigningAlgNone
	}

	return c.IntrospectionSignedResponseAlg
}

// GetIntrospectionEncryptedResponseKeyID returns the specific key identifier used to satisfy JWE requirements for
// OAuth 2.0 JWT introspection response specifications. If unspecified the other available parameters will be
//
//	// utilized to select an appropriate key.
func (c *RegisteredClient) GetIntrospectionEncryptedResponseKeyID() (kid string) {
	return c.IntrospectionEncryptedResponseKeyID
}

// GetIntrospectionEncryptedResponseAlg is equivalent to the 'introspection_encrypted_response_alg' client metadata
// value which determines the JWE [RFC7516] algorithm (alg value) as defined in JWA [RFC7518] for content key
// encryption. If this is specified, the response will be encrypted using JWE and the configured content encryption
// algorithm (introspection_encrypted_response_enc). The default, if omitted, is that no encryption is performed.
// If both signing and encryption are requested, the response will be signed then encrypted, with the result being
// a Nested JWT, as defined in JWT [RFC7519].
func (c *RegisteredClient) GetIntrospectionEncryptedResponseAlg() (alg string) {
	return c.IntrospectionEncryptedResponseAlg
}

// GetIntrospectionEncryptedResponseEnc is equivalent to the 'introspection_encrypted_response_enc' client metadata
// value which determines the  JWE [RFC7516] algorithm (enc value) as defined in JWA [RFC7518] for content
// encryption of introspection responses. The default, if omitted, is A128CBC-HS256. Note: This parameter MUST NOT
// be specified without setting introspection_encrypted_response_alg.
func (c *RegisteredClient) GetIntrospectionEncryptedResponseEnc() (enc string) {
	return c.IntrospectionEncryptedResponseEnc
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

// GetTokenEndpointAuthSigningAlg returns the JWS [JWS] alg algorithm [JWA] that MUST be used for signing the JWT
// [JWT] used to authenticate the Client at the Token Endpoint for the private_key_jwt and client_secret_jwt
// authentication methods.
func (c *RegisteredClient) GetTokenEndpointAuthSigningAlg() (alg string) {
	if c.TokenEndpointAuthSigningAlg == "" {
		c.TokenEndpointAuthSigningAlg = SigningAlgRSAUsingSHA256
	}

	return c.TokenEndpointAuthSigningAlg
}

// GetRevocationEndpointAuthMethod returns the requested Client Authentication Method for the Revocation Endpoint.
// The options are client_secret_post, client_secret_basic, client_secret_jwt, private_key_jwt, and none.
func (c *RegisteredClient) GetRevocationEndpointAuthMethod() (method string) {
	if c.Public && c.RevocationEndpointAuthMethod == "" {
		c.RevocationEndpointAuthMethod = ClientAuthMethodNone
	}

	return c.RevocationEndpointAuthMethod
}

// GetRevocationEndpointAuthSigningAlg returns the JWS [JWS] alg algorithm [JWA] that MUST be used for signing the
// JWT [JWT] used to authenticate the Client at the Revocation Endpoint for the private_key_jwt and client_secret_jwt
// authentication methods.
func (c *RegisteredClient) GetRevocationEndpointAuthSigningAlg() (alg string) {
	return c.RevocationEndpointAuthSigningAlg
}

// GetIntrospectionEndpointAuthMethod returns the requested Client Authentication Method for the Introspection Endpoint.
// The options are client_secret_post, client_secret_basic, client_secret_jwt, private_key_jwt, and none.
func (c *RegisteredClient) GetIntrospectionEndpointAuthMethod() (method string) {
	if c.Public && c.IntrospectionEndpointAuthMethod == "" {
		c.IntrospectionEndpointAuthMethod = ClientAuthMethodNone
	}

	return c.IntrospectionEndpointAuthMethod
}

// GetIntrospectionEndpointAuthSigningAlg returns the JWS [JWS] alg algorithm [JWA] that MUST be used for signing the
// JWT [JWT] used to authenticate the Client at the Introspection Endpoint for the private_key_jwt and client_secret_jwt
// authentication methods.
func (c *RegisteredClient) GetIntrospectionEndpointAuthSigningAlg() (alg string) {
	return c.IntrospectionEndpointAuthSigningAlg
}

// GetPushedAuthorizationRequestEndpointAuthMethod returns the requested Client Authentication Method for the
// Pushed Authorization Request Endpoint. The options are client_secret_post, client_secret_basic, client_secret_jwt,
// private_key_jwt, and none.
func (c *RegisteredClient) GetPushedAuthorizationRequestEndpointAuthMethod() (method string) {
	if c.Public && c.PushedAuthorizationRequestEndpointAuthMethod == "" {
		c.PushedAuthorizationRequestEndpointAuthMethod = ClientAuthMethodNone
	}

	return c.PushedAuthorizationRequestEndpointAuthMethod
}

// GetPushedAuthorizationRequestEndpointAuthSigningAlg returns the JWS [JWS] alg algorithm [JWA] that MUST be used for
// signing the JWT [JWT] used to authenticate the Client at the Pushed Authorization Request Endpoint for the
// private_key_jwt and client_secret_jwt authentication methods.
func (c *RegisteredClient) GetPushedAuthorizationRequestEndpointAuthSigningAlg() (alg string) {
	return c.PushedAuthorizationRequestEndpointAuthSigningAlg
}

// GetEnableJWTProfileOAuthAccessTokens returns true if this client is configured to return the
// RFC9068 JWT Profile for OAuth 2.0 Access Tokens.
func (c *RegisteredClient) GetEnableJWTProfileOAuthAccessTokens() (enable bool) {
	return c.GetAccessTokenSignedResponseAlg() != SigningAlgNone && len(c.GetAccessTokenSignedResponseKeyID()) > 0
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
func (c *RegisteredClient) GetConsentResponseBody(session RequesterFormSession, form url.Values, authTime time.Time, disablePreConf bool) ConsentGetResponseBody {
	body := ConsentGetResponseBody{
		ClientID:          c.ID,
		ClientDescription: c.Name,
		PreConfiguration:  c.ConsentPolicy.Mode == ClientConsentModePreConfigured && !disablePreConf,
	}

	if session != nil {
		body.Scopes = session.GetRequestedScopes()
		body.Audience = session.GetRequestedAudience()

		var (
			claims *ClaimsRequests
			err    error
		)

		if form == nil {
			if form, err = session.GetForm(); err != nil {
				return body
			}
		}

		if form != nil {
			if claims, err = NewClaimRequests(form); err == nil {
				body.Claims, body.EssentialClaims = claims.ToSlices()
			}

			body.RequireLogin = RequestFormRequiresLogin(form, session.GetRequestedAt(), authTime)
		}
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

// GetRequestObjectSigningKeyID returns the specific key identifier used to satisfy JWS requirements of the request
// object specifications. If unspecified the other available parameters will be utilized to select an appropriate
// key.
func (c *RegisteredClient) GetRequestObjectSigningKeyID() (kid string) {
	return ""
}

// GetRequestObjectSigningAlg returns the JWS [JWS] alg algorithm [JWA] that MUST be used for signing Request
// Objects sent to the OP. All Request Objects from this Client MUST be rejected, if not signed with this algorithm.
func (c *RegisteredClient) GetRequestObjectSigningAlg() (alg string) {
	return c.RequestObjectSigningAlg
}

// GetRequestObjectEncryptionKeyID returns the specific key identifier used to satisfy JWE requirements of the
// request object specifications. If unspecified the other available parameters will be utilized to select an
// appropriate key.
func (c *RegisteredClient) GetRequestObjectEncryptionKeyID() (kid string) {
	return ""
}

// GetRequestObjectEncryptionAlg is equivalent to the 'request_object_encryption_alg' client metadata value which
// determines the JWE alg algorithm [JWA] the RP is declaring that it may use for encrypting Request Objects sent to
// the OP. This parameter SHOULD be included when symmetric encryption will be used, since this signals to the OP
// that a client_secret value needs to be returned from which the symmetric key will be derived, that might not
// otherwise be returned. The RP MAY still use other supported encryption algorithms or send unencrypted Request
// Objects, even when this parameter is present. If both signing and encryption are requested, the Request Object
// will be signed then encrypted, with the result being a Nested JWT, as defined in [JWT]. The default, if omitted,
// is that the RP is not declaring whether it might encrypt any Request Objects.
func (c *RegisteredClient) GetRequestObjectEncryptionAlg() (alg string) {
	return c.RequestObjectEncryptionAlg
}

// GetRequestObjectEncryptionEnc is equivalent to the 'request_object_encryption_enc' client metadata value which
// determines the JWE enc algorithm [JWA] the RP is declaring that it may use for encrypting Request Objects sent to
// the OP. If request_object_encryption_alg is specified, the default request_object_encryption_enc value is
// A128CBC-HS256. When request_object_encryption_enc is included, request_object_encryption_alg MUST also be
// provided.
func (c *RegisteredClient) GetRequestObjectEncryptionEnc() (enc string) {
	return c.RequestObjectEncryptionEnc
}

// GetAllowMultipleAuthenticationMethods should return true if the client policy allows multiple authentication
// methods due to the client implementation breaching RFC6749 Section 2.3.
//
// See: https://datatracker.ietf.org/doc/html/rfc6749#section-2.3.
func (c *RegisteredClient) GetAllowMultipleAuthenticationMethods() (allow bool) {
	return c.AllowMultipleAuthenticationMethods
}

// GetClientCredentialsFlowRequestedScopeImplicit is indicative of if a client will implicitly request all scopes it
// is allowed to request in the absence of requested scopes during the Client Credentials Flow.
func (c *RegisteredClient) GetClientCredentialsFlowRequestedScopeImplicit() (allow bool) {
	return c.ClientCredentialsFlowAllowImplicitScope
}

// GetRequestedAudienceImplicit is indicative of if a client will implicitly request all audiences it is allowed to
// request in the absence of requested audience during an Authorization Endpoint Flow or Client Credentials Flow.
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
	case oauthelia2.RefreshToken:
		switch {
		case gtl.RefreshToken > durationZero:
			return gtl.RefreshToken
		case c.Lifespans.RefreshToken > durationZero:
			return c.Lifespans.RefreshToken
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
	case oauthelia2.AuthorizeCode:
		switch {
		case gtl.AuthorizeCode > durationZero:
			return gtl.AuthorizeCode
		case c.Lifespans.AuthorizeCode > durationZero:
			return c.Lifespans.AuthorizeCode
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
	case oauthelia2.GrantTypeDeviceCode:
		return c.Lifespans.Grants.DeviceCode
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

func NewUserinfoClient(client Client) jwt.Client {
	return &decoratedUserinfoClient{client: client}
}

type decoratedUserinfoClient struct {
	client Client
}

func (d decoratedUserinfoClient) GetSigningKeyID() (kid string) {
	return d.client.GetUserinfoSignedResponseKeyID()
}

func (d decoratedUserinfoClient) GetSigningAlg() (alg string) {
	return d.client.GetUserinfoSignedResponseAlg()
}

func (d decoratedUserinfoClient) GetEncryptionKeyID() (kid string) {
	return d.client.GetUserinfoEncryptedResponseKeyID()
}

func (d decoratedUserinfoClient) GetEncryptionAlg() (alg string) {
	return d.client.GetUserinfoEncryptedResponseAlg()
}

func (d decoratedUserinfoClient) GetEncryptionEnc() (enc string) {
	return d.client.GetUserinfoEncryptedResponseEnc()
}

func (d decoratedUserinfoClient) IsClientSigned() (is bool) {
	return false
}

func (d decoratedUserinfoClient) GetID() string {
	return d.client.GetID()
}

func (d decoratedUserinfoClient) GetClientSecretPlainText() (secret []byte, ok bool, err error) {
	return d.client.GetClientSecretPlainText()
}

func (d decoratedUserinfoClient) GetJSONWebKeys() (jwks *jose.JSONWebKeySet) {
	return d.client.GetJSONWebKeys()
}

func (d decoratedUserinfoClient) GetJSONWebKeysURI() (uri string) {
	return d.client.GetJSONWebKeysURI()
}
