// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"time"

	"authelia.com/provider/oauth2/token/jose"
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
