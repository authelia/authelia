// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"time"

	"authelia.com/provider/oauth2/token/jose"

	"github.com/authelia/authelia/v4/internal/random"
)

// Discovery represents the subset of the OpenID Connect Discovery 1.0 document Authelia consumes.
type Discovery struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserInfoEndpoint      string   `json:"userinfo_endpoint"`
	JWKSURI               string   `json:"jwks_uri"`
	IDTokenSigningAlgs    []string `json:"id_token_signing_alg_values_supported"`

	CodeChallengeMethods     []string `json:"code_challenge_methods_supported"`
	ResponseModes            []string `json:"response_modes_supported"`
	TokenEndpointAuthMethods []string `json:"token_endpoint_auth_methods_supported"`
	UserInfoSigningAlgs      []string `json:"userinfo_signing_alg_values_supported"`

	AuthorizationResponseIssParameterSupported bool `json:"authorization_response_iss_parameter_supported"`

	PushedAuthorizationRequestEndpoint string `json:"pushed_authorization_request_endpoint"`
	RequirePushedAuthorizationRequests bool   `json:"require_pushed_authorization_requests"`
}

// Token represents the subset of a token response Authelia consumes. The access token is only used to request the
// UserInfo Endpoint during the callback which obtained it; it is never stored or logged.
type Token struct {
	IDToken     string
	AccessToken string
}

// IdentityClaims represents the validated claims Authelia consumes from an ID Token.
type IdentityClaims struct {
	Issuer                         string
	Subject                        string
	PreferredUsername              string
	Name                           string
	Email                          string
	AuthenticationMethodsReference []string
}

// KeySet resolves a JSON Web Key Set for a location, optionally bypassing any cache.
type KeySet interface {
	Resolve(ctx context.Context, location string, ignoreCache bool) (jwks *jose.JSONWebKeySet, err error)
}

// ValidateOptions represents the parameters an ID Token is validated against.
type ValidateOptions struct {
	Issuer   string
	ClientID string
	Nonce    string
	Alg      string
	JWKSURI  string
	Now      time.Time
	Leeway   time.Duration

	// AccessToken is the access token issued alongside the ID Token, which the 'at_hash' claim is validated against
	// when the ID Token carries one.
	AccessToken string
}

// Provider is an external identity provider users may sign in to Authelia with.
type Provider interface {
	// ID returns the identifier of the provider, which appears in its URLs and in the account links made with it.
	ID() string

	// Name returns the display name of the provider.
	Name() string

	// Type returns the type of the provider.
	Type() string

	// Issuer returns the issuer which, together with the subject of an identity, identifies the identity in account
	// links.
	Issuer() string

	// ResponseMode returns the response mode the provider delivers the authorization response with.
	ResponseMode() string

	// AuthorizationResponseIssuerRequired returns true when the authorization response must carry the 'iss' parameter
	// described by RFC9207, which is the case for a provider known to send it.
	AuthorizationResponseIssuerRequired(ctx context.Context) (required bool, err error)

	// SharedRedirectURI returns true when the provider uses the redirect URI shared by every provider configured to use
	// it rather than a redirect URI of its own. An authorization response delivered to the shared redirect URI must
	// carry the 'iss' parameter described by RFC9207.
	SharedRedirectURI() bool

	// AuthenticationMethodsReference returns the Authentication Method Reference values a session adopts for an identity
	// the provider asserted the given values for, which may be none.
	AuthenticationMethodsReference(asserted []string) []string

	// AuthorizationRequest constructs an authorization request for the provider.
	AuthorizationRequest(ctx context.Context, rand random.Provider, options AuthorizationRequestOptions) (request *AuthorizationRequest, err error)

	// Complete completes the authorization with the authorization code the provider returned, and returns the identity
	// it established. The identity is only returned once everything the provider type requires of it is validated.
	Complete(ctx context.Context, request CompletionRequest) (claims *IdentityClaims, err error)
}

// AuthorizationRequestOptions represents the values an authorization request is constructed with.
type AuthorizationRequestOptions struct {
	RedirectURI string

	// Language is the language tag of the language the user chose in the portal, which may be empty.
	Language string
}

// AuthorizationRequest represents a constructed authorization request and the values which must be retained to
// complete the authorization.
type AuthorizationRequest struct {
	URL          string
	State        string
	Nonce        string
	CodeVerifier string
	Handle       string
}

// CompletionRequest represents the values an authorization is completed with.
type CompletionRequest struct {
	Code         string
	Nonce        string
	CodeVerifier string
	Handle       string
	RedirectURI  string
	Language     string
	Now          time.Time
}

type discoveryCheck struct {
	name   string
	values []string
	value  string
}
