// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	"authelia.com/provider/jose"
	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/clock"
)

// GetEndpointURLs returns the absolute URLs of the endpoints of this Authorization Server for the issuer of the
// request.
func (c *Config) GetEndpointURLs(ctx context.Context) (urls oauthelia2.EndpointURLs) {
	issuer := c.GetIssuer(ctx)

	if issuer == "" {
		return urls
	}

	return oauthelia2.EndpointURLs{
		Authorization:              fmt.Sprintf("%s%s", issuer, EndpointPathAuthorization),
		Token:                      fmt.Sprintf("%s%s", issuer, EndpointPathToken),
		UserInfo:                   fmt.Sprintf("%s%s", issuer, EndpointPathUserinfo),
		JWKS:                       fmt.Sprintf("%s%s", issuer, EndpointPathJWKs),
		Revocation:                 fmt.Sprintf("%s%s", issuer, EndpointPathRevocation),
		Introspection:              fmt.Sprintf("%s%s", issuer, EndpointPathIntrospection),
		PushedAuthorizationRequest: fmt.Sprintf("%s%s", issuer, EndpointPathPushedAuthorizationRequest),
		DeviceAuthorization:        fmt.Sprintf("%s%s", issuer, EndpointPathDeviceAuthorization),
	}
}

// GetScopesSupported returns the 'scopes_supported' values.
func (c *Config) GetScopesSupported(ctx context.Context) (scopes []string) {
	return []string{ScopeOfflineAccess, ScopeOpenID, ScopeProfile, ScopeEmail, ScopeAddress, ScopePhone, ScopeGroups}
}

// GetClaimsSupported returns the 'claims_supported' values.
func (c *Config) GetClaimsSupported(ctx context.Context) (claims []string) {
	return []string{
		ClaimAuthenticationMethodsReference,
		ClaimAudience,
		ClaimAuthorizedParty,
		ClaimClientIdentifier,
		ClaimExpirationTime,
		ClaimIssuedAt,
		ClaimIssuer,
		ClaimJWTID,
		ClaimRequestedAt,
		ClaimAuthenticationTime,
		ClaimNonce,
		ClaimGroups,
		ClaimSubject,
		ClaimFullName,
		ClaimGivenName,
		ClaimFamilyName,
		ClaimMiddleName,
		ClaimNickname,
		ClaimPreferredUsername,
		ClaimProfile,
		ClaimPicture,
		ClaimWebsite,
		ClaimEmail,
		ClaimEmailVerified,
		ClaimEmailAlts,
		ClaimGender,
		ClaimBirthdate,
		ClaimZoneinfo,
		ClaimLocale,
		ClaimPhoneNumber,
		ClaimPhoneNumberVerified,
		ClaimAddress,
		ClaimUpdatedAt,
	}
}

// GetACRValuesSupported returns nil as no 'acr_values_supported' values are advertised.
func (c *Config) GetACRValuesSupported(ctx context.Context) (values []string) {
	return nil
}

// GetDisplayValuesSupported returns nil as no 'display_values_supported' values are advertised.
func (c *Config) GetDisplayValuesSupported(ctx context.Context) (values []string) {
	return nil
}

// GetClaimTypesSupported returns the 'claim_types_supported' values.
func (c *Config) GetClaimTypesSupported(ctx context.Context) (types []string) {
	return []string{ClaimTypeNormal}
}

// GetSubjectTypesSupported returns the 'subject_types_supported' values.
func (c *Config) GetSubjectTypesSupported(ctx context.Context) (types []string) {
	return []string{SubjectTypePublic, SubjectTypePairwise}
}

// GetUILocalesSupported returns nil as no 'ui_locales_supported' values are advertised.
func (c *Config) GetUILocalesSupported(ctx context.Context) (locales []string) {
	return nil
}

// GetClaimsLocalesSupported returns the 'claims_locales_supported' values, which this Authorization Server does not
// advertise.
func (c *Config) GetClaimsLocalesSupported(ctx context.Context) (locales []string) {
	return nil
}

// GetServiceDocumentation returns an empty value as no 'service_documentation' URL is advertised.
func (c *Config) GetServiceDocumentation(ctx context.Context) (value string) {
	return ""
}

// GetOPPolicyURI returns an empty value as no 'op_policy_uri' URL is advertised.
func (c *Config) GetOPPolicyURI(ctx context.Context) (value string) {
	return ""
}

// GetOPTOSURI returns an empty value as no 'op_tos_uri' URL is advertised.
func (c *Config) GetOPTOSURI(ctx context.Context) (value string) {
	return ""
}

// GetTokenEndpointAuthMethods returns the client authentication methods accepted at the token endpoint.
func (c *Config) GetTokenEndpointAuthMethods(ctx context.Context) (methods []string) {
	return []string{ClientAuthMethodClientSecretBasic, ClientAuthMethodClientSecretPost, ClientAuthMethodClientSecretJWT, ClientAuthMethodPrivateKeyJWT, ClientAuthMethodNone}
}

// GetRevocationEndpointAuthMethods returns the client authentication methods accepted at the revocation endpoint.
func (c *Config) GetRevocationEndpointAuthMethods(ctx context.Context) (methods []string) {
	return []string{ClientAuthMethodClientSecretBasic, ClientAuthMethodClientSecretPost, ClientAuthMethodClientSecretJWT, ClientAuthMethodPrivateKeyJWT, ClientAuthMethodNone}
}

// GetRevocationEndpointAuthSigningAlgValues returns the JWS algorithms accepted for a client assertion at the
// revocation endpoint.
func (c *Config) GetRevocationEndpointAuthSigningAlgValues(ctx context.Context) (algs []string) {
	return SigningAlgValuesSupported()
}

// GetIntrospectionEndpointAuthMethods returns nil so the introspection endpoint accepts an access token as well as the
// client authentication method registered for the client. The advertised value is set by the discovery metadata hook.
func (c *Config) GetIntrospectionEndpointAuthMethods(ctx context.Context) (methods []string) {
	return nil
}

// GetIntrospectionEndpointAuthSigningAlgValues returns the JWS algorithms accepted for a client assertion at the
// introspection endpoint.
func (c *Config) GetIntrospectionEndpointAuthSigningAlgValues(ctx context.Context) (algs []string) {
	return SigningAlgValuesSupported()
}

// GetClientAssertionSigningAlgValues returns the JWS algorithms accepted for a client assertion.
func (c *Config) GetClientAssertionSigningAlgValues(ctx context.Context) (algs []string) {
	return SigningAlgValuesSupported()
}

// GetClientAssertionEncryptionAlgValues returns the JWE 'alg' values accepted for an encrypted client assertion.
func (c *Config) GetClientAssertionEncryptionAlgValues(ctx context.Context) (algs []string) {
	return EncryptionAlgValuesSupported()
}

// GetClientAssertionEncryptionEncValues returns the JWE 'enc' values accepted for an encrypted client assertion.
func (c *Config) GetClientAssertionEncryptionEncValues(ctx context.Context) (encs []string) {
	return EncryptionEncValuesSupported()
}

// GetClientAssertionClientSecretEncryptionDisabled returns true as client assertions encrypted with a key derived
// from the client secret are not accepted, since client secrets are generally stored as a hash.
func (c *Config) GetClientAssertionClientSecretEncryptionDisabled(ctx context.Context) (disabled bool) {
	return true
}

// GetRequestObjectSigningAlgValues returns the JWS algorithms accepted for a request object, which only includes
// 'none' when signed request objects are not required.
func (c *Config) GetRequestObjectSigningAlgValues(ctx context.Context) (algs []string) {
	if c.RequireSignedRequestObject {
		return SigningAlgValuesSupported()
	}

	return SigningAlgValuesSupported(SigningAlgNone)
}

// GetRequestObjectEncryptionAlgValues returns the JWE 'alg' values accepted for an encrypted request object.
func (c *Config) GetRequestObjectEncryptionAlgValues(ctx context.Context) (algs []string) {
	return EncryptionAlgValuesSupported()
}

// GetRequestObjectEncryptionEncValues returns the JWE 'enc' values accepted for an encrypted request object.
func (c *Config) GetRequestObjectEncryptionEncValues(ctx context.Context) (encs []string) {
	return EncryptionEncValuesSupported()
}

// GetResponseEncryptionAlgValues returns the JWE 'alg' values supported for the given response.
func (c *Config) GetResponseEncryptionAlgValues(ctx context.Context, target oauthelia2.ResponseEncryptionTarget) (algs []string) {
	return EncryptionAlgValuesSupported()
}

// GetResponseEncryptionEncValues returns the JWE 'enc' values supported for the given response.
func (c *Config) GetResponseEncryptionEncValues(ctx context.Context, target oauthelia2.ResponseEncryptionTarget) (encs []string) {
	return EncryptionEncValuesSupported()
}

// GetRFC7523AssertionSigningAlgValues returns the JWS algorithms accepted for an RFC 7523 JWT authorization grant.
func (c *Config) GetRFC7523AssertionSigningAlgValues(ctx context.Context) (algs []string) {
	return SigningAlgValuesSupported()
}

// GetIDJAGRedeemSigningAlgValues returns nil as this Authorization Server does not redeem Identity Assertion JWT
// Authorization Grants.
func (c *Config) GetIDJAGRedeemSigningAlgValues(ctx context.Context) (algs []string) {
	return nil
}

// GetAuthorizationServerIdentificationEnabled returns true when the 'iss' authorization response parameter is
// returned per RFC 9207.
func (c *Config) GetAuthorizationServerIdentificationEnabled(ctx context.Context) (enabled bool) {
	return true
}

// GetAllowInsecureRedirectURIForPublicClients returns false as every client is subject to the redirect URI security
// check required by RFC 9700 Section 2.6.
func (c *Config) GetAllowInsecureRedirectURIForPublicClients(ctx context.Context) (allow bool) {
	return false
}

// GetAuthorizationDetailsTypeHandlers returns nil as RFC 9396 Rich Authorization Requests are not supported.
func (c *Config) GetAuthorizationDetailsTypeHandlers(ctx context.Context) (handlers map[string]oauthelia2.AuthorizationDetailsTypeHandler) {
	return nil
}

// GetClientAuthenticationRealm returns the realm of the 'Basic' client authentication challenge.
func (c *Config) GetClientAuthenticationRealm(ctx context.Context) (realm string) {
	return clientAuthenticationRealm
}

// GetDPoPStrictRefreshTokenBinding returns false so a confidential client's DPoP bound refresh token is re-bound to
// the key of the presented proof as RFC 9449 Section 5 permits.
func (c *Config) GetDPoPStrictRefreshTokenBinding(ctx context.Context) (strict bool) {
	return false
}

// GetMTLSStrictRefreshTokenBinding returns false so a confidential client's certificate bound refresh token is
// re-bound to the presented certificate as RFC 8705 Section 4 permits.
func (c *Config) GetMTLSStrictRefreshTokenBinding(ctx context.Context) (strict bool) {
	return false
}

// GetDisablePushedAuthorizationRequestClientRefetch returns false so the current client registration is used when a
// pushed authorization request is redeemed as RFC 9126 Section 7.4 recommends.
func (c *Config) GetDisablePushedAuthorizationRequestClientRefetch(ctx context.Context) (disable bool) {
	return false
}

// GetDiscoveryMetadataHooks returns the hooks which run after the discovery metadata is built.
func (c *Config) GetDiscoveryMetadataHooks(ctx context.Context) (hooks []oauthelia2.DiscoveryMetadataHook) {
	return []oauthelia2.DiscoveryMetadataHook{discoveryMetadataHook{}}
}

// GetDiscoveryMetadataSigningStrategy returns the strategy which signs the discovery metadata as the
// 'signed_metadata' member, or nil when the discovery metadata is not signed.
func (c *Config) GetDiscoveryMetadataSigningStrategy(ctx context.Context) (strategy jwt.Strategy) {
	if c.DiscoverySignedResponseKeyID == "" && (c.DiscoverySignedResponseAlg == "" || c.DiscoverySignedResponseAlg == SigningAlgNone) {
		return nil
	}

	if c.Strategy.JWT == nil {
		return nil
	}

	return &discoverySigningStrategy{
		Strategy: c.Strategy.JWT,
		kid:      c.DiscoverySignedResponseKeyID,
		alg:      c.DiscoverySignedResponseAlg,
	}
}

// GetIntrospectionTokenTypeEnabled returns true so the introspection response includes the 'token_type' of an active
// access token per RFC 7662 Section 2.2.
func (c *Config) GetIntrospectionTokenTypeEnabled(ctx context.Context) (enabled bool) {
	return true
}

// GetRFC8693MaxActorChainDepth returns the maximum actor chain depth of a token issued by Token Exchange.
func (c *Config) GetRFC8693MaxActorChainDepth(ctx context.Context) (depth int) {
	return rfc8693MaxActorChainDepth
}

// GetSSFConfig returns nil as the Shared Signals Framework is not supported.
func (c *Config) GetSSFConfig(ctx context.Context) (config *oauthelia2.SSFConfig) {
	return nil
}

// GetRFC7591ClientRegistrationPolicyClaim returns an empty value as RFC 7591 Dynamic Client Registration is not
// supported.
func (c *Config) GetRFC7591ClientRegistrationPolicyClaim(ctx context.Context) (claim string) {
	return ""
}

// GetRFC7591ClientRegistrationStatelessAccessTokens returns false as RFC 7591 Dynamic Client Registration is not
// supported.
func (c *Config) GetRFC7591ClientRegistrationStatelessAccessTokens(ctx context.Context) (enabled bool) {
	return false
}

// GetRFC7591ClientRegistrationTrustedIssuers returns nil as RFC 7591 Dynamic Client Registration is not supported.
func (c *Config) GetRFC7591ClientRegistrationTrustedIssuers(ctx context.Context) (issuers []oauthelia2.RFC7591TrustedIssuer) {
	return nil
}

// GetRFC7591ClientRegistrationPatternIssuers returns nil as RFC 7591 Dynamic Client Registration is not supported.
func (c *Config) GetRFC7591ClientRegistrationPatternIssuers(ctx context.Context) (issuers []string) {
	return nil
}

// GetRFC7591SoftwareStatementIssuers returns nil as RFC 7591 Dynamic Client Registration is not supported.
func (c *Config) GetRFC7591SoftwareStatementIssuers(ctx context.Context) (issuers []oauthelia2.RFC7591TrustedIssuer) {
	return nil
}

// GetRFC7591SoftwareStatementJWTValidationHeaderAllowEmptyType returns false as RFC 7591 Dynamic Client Registration
// is not supported.
func (c *Config) GetRFC7591SoftwareStatementJWTValidationHeaderAllowEmptyType(ctx context.Context) (allow bool) {
	return false
}

// GetRFC7591SoftwareStatementJWTValidationHeaderAllowTypes returns nil as RFC 7591 Dynamic Client Registration is not
// supported.
func (c *Config) GetRFC7591SoftwareStatementJWTValidationHeaderAllowTypes(ctx context.Context) (types []string) {
	return nil
}

// GetRFC7591SoftwareStatementRequireSoftwareID returns false as RFC 7591 Dynamic Client Registration is not supported.
func (c *Config) GetRFC7591SoftwareStatementRequireSoftwareID(ctx context.Context) (require bool) {
	return false
}

// GetRFC7591TrustedIssuerSigningAlgValues returns nil as RFC 7591 Dynamic Client Registration is not supported.
func (c *Config) GetRFC7591TrustedIssuerSigningAlgValues(ctx context.Context) (algs []string) {
	return nil
}

// GetEnforceClaimsParameter returns true so the requirements the 'claims' request parameter places on the
// Authorization Server are enforced, as required by OpenID Connect Core 1.0 Section 5.5.
func (c *Config) GetEnforceClaimsParameter(ctx context.Context) (enforce bool) {
	return true
}

// GetJWTProfileAccessTokenDefaultAudience returns the audience of a JWT profile access token whose request granted no
// audience or resource, as RFC 9068 Section 2.2 requires the claim and Section 3 requires a default resource
// indicator. This is the effective issuer, as this Authorization Server is the only resource such a token is
// implicitly valid for.
func (c *Config) GetJWTProfileAccessTokenDefaultAudience(ctx context.Context) (audience []string) {
	if issuer := c.GetIssuer(ctx); issuer != "" {
		return []string{issuer}
	}

	return nil
}

var (
	_ oauthelia2.JWTProfileAccessTokenDefaultAudienceProvider = (*Config)(nil)
	_ oauthelia2.ClaimsParameterEnforcementProvider           = (*Config)(nil)
)

// NewDiscoverySigningClient returns a jwt.Client which selects the issuer key used to sign the discovery metadata by
// the given key id and algorithm.
func NewDiscoverySigningClient(kid, alg string) jwt.Client {
	return &discoverySigningClient{kid: kid, alg: alg}
}

type discoverySigningClient struct {
	kid string
	alg string
}

func (c *discoverySigningClient) GetID() string {
	return ""
}

func (c *discoverySigningClient) GetClientSecretPlainText() (secret []byte, ok bool, err error) {
	return nil, false, nil
}

func (c *discoverySigningClient) GetSigningKeyID() (kid string) {
	return c.kid
}

func (c *discoverySigningClient) GetSigningAlg() (alg string) {
	return c.alg
}

func (c *discoverySigningClient) GetEncryptionKeyID() (kid string) {
	return ""
}

func (c *discoverySigningClient) GetEncryptionAlg() (alg string) {
	return ""
}

func (c *discoverySigningClient) GetEncryptionEnc() (enc string) {
	return ""
}

func (c *discoverySigningClient) IsClientSigned() (is bool) {
	return false
}

func (c *discoverySigningClient) IsPublic() bool {
	return false
}

func (c *discoverySigningClient) GetJSONWebKeys() (jwks *jose.JSONWebKeySet) {
	return nil
}

func (c *discoverySigningClient) GetJSONWebKeysURI() (uri string) {
	return ""
}

type discoveryMetadataHook struct{}

func (discoveryMetadataHook) HandleDiscoveryMetadata(ctx context.Context, metadata *oauthelia2.OpenIDConnectProviderMetadata) (err error) {
	metadata.ClaimsSupported = slices.DeleteFunc(metadata.ClaimsSupported, func(claim string) bool {
		return claim == ClaimAuthenticationContextClassReference || claim == ClaimSessionID
	})

	if metadata.IntrospectionEndpoint != "" {
		metadata.IntrospectionEndpointAuthMethodsSupported = []string{
			ClientAuthMethodClientSecretBasic,
			ClientAuthMethodClientSecretPost,
			ClientAuthMethodClientSecretJWT,
			ClientAuthMethodPrivateKeyJWT,
		}
	}

	return nil
}

type discoverySigningStrategy struct {
	jwt.Strategy

	kid string
	alg string
}

func (s *discoverySigningStrategy) SigningAlgValues(ctx context.Context) (algs []string) {
	if s.alg != "" && s.alg != SigningAlgNone {
		return []string{s.alg}
	}

	return s.Strategy.SigningAlgValues(ctx)
}

func (s *discoverySigningStrategy) Encode(ctx context.Context, claims jwt.Claims, _ ...jwt.StrategyOpt) (tokenString string, signature string, err error) {
	if mapClaims, ok := claims.(jwt.MapClaims); ok {
		var now time.Time

		if c, ok := ctx.(clockContext); ok {
			now = c.GetClock().Now().UTC()
		} else {
			now = time.Now().UTC()
		}

		mapClaims[ClaimJWTID] = uuid.New().String()
		mapClaims[ClaimIssuedAt] = now.Unix()
		mapClaims[ClaimExpirationTime] = now.Add(time.Hour).Unix()
	}

	return s.Strategy.Encode(ctx, claims, jwt.WithClient(NewDiscoverySigningClient(s.kid, s.alg)))
}

type clockContext interface {
	GetClock() (clock clock.Provider)
}
