// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jose"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
)

// OpenIDConnectProvider is an external identity provider which is an OpenID Connect 1.0 Provider.
type OpenIDConnectProvider struct {
	id           string
	name         string
	issuer       string
	clientID     string
	scopes       []string
	responseMode string
	amr          authenticationMethodsReferencePolicy

	clientSecret            string
	tokenEndpointAuthMethod string

	algIDToken  string
	algUserInfo string

	endpointAuthorization string
	endpointPAR           string
	endpointToken         string
	endpointUserInfo      string

	jwksURI               string
	issParameterSupported bool
	parRequired           bool

	discover bool
	resolved bool
	mutex    sync.Mutex

	keys   KeySet
	client *retryablehttp.Client
}

func newOpenIDConnectProvider(config *schema.AuthenticationBackendExternalIdentityProvider, client *retryablehttp.Client) (provider *OpenIDConnectProvider) {
	provider = &OpenIDConnectProvider{
		id:                      config.ID,
		name:                    config.Name,
		issuer:                  config.Issuer,
		clientID:                config.ClientID,
		scopes:                  config.Scopes,
		responseMode:            config.ResponseMode,
		amr:                     newAuthenticationMethodsReferencePolicy(config.AuthenticationMethodsReference),
		clientSecret:            config.ClientSecret,
		tokenEndpointAuthMethod: config.TokenEndpointAuthMethod,
		algIDToken:              config.IDTokenSignedResponseAlg,
		endpointAuthorization:   config.Endpoints.Authorization,
		endpointToken:           config.Endpoints.Token,
		endpointUserInfo:        config.Endpoints.UserInfo,
		jwksURI:                 config.Endpoints.JSONWebKeys,
		issParameterSupported:   config.AuthorizationResponseIssParameterSupported,
		endpointPAR:             config.Endpoints.PushedAuthorizationRequest,
		parRequired:             config.RequirePushedAuthorizationRequests,
		algUserInfo:             config.UserInfoSignedResponseAlg,
		discover:                !config.Discovery.Disable,
		client:                  client,
	}

	if provider.responseMode == "" {
		provider.responseMode = ResponseModeQuery
	}

	// An unset algorithm is selected during discovery, so it only needs a default here when there is no discovery.
	if provider.algIDToken == "" && !provider.discover {
		provider.algIDToken = defaultIDTokenSigningAlg
	}

	if len(config.JSONWebKeys) != 0 {
		provider.keys = &staticKeySet{jwks: oidc.NewJSONWebKeySetPublic(config.JSONWebKeys)}
	} else {
		provider.keys = oauthelia2.NewDefaultJWKSFetcherStrategy(oauthelia2.JWKSFetcherWithHTTPClient(client))
	}

	return provider
}

// ID implements Provider.
func (p *OpenIDConnectProvider) ID() string {
	return p.id
}

// Name implements Provider.
func (p *OpenIDConnectProvider) Name() string {
	return p.name
}

// Type implements Provider.
func (p *OpenIDConnectProvider) Type() string {
	return ProviderTypeOpenIDConnect
}

// Issuer implements Provider.
func (p *OpenIDConnectProvider) Issuer() string {
	return p.issuer
}

// ResponseMode implements Provider.
func (p *OpenIDConnectProvider) ResponseMode() string {
	return p.responseMode
}

// AuthorizationResponseIssuerRequired implements Provider. A provider known to send the 'iss' parameter must send it,
// as RFC9207 section 2.4 requires the response of such a provider be rejected without it. Discovery is resolved first
// as the callback may be the first use of the provider since the process started.
func (p *OpenIDConnectProvider) AuthorizationResponseIssuerRequired(ctx context.Context) (required bool, err error) {
	if err = p.Resolve(ctx); err != nil {
		return false, err
	}

	return p.issParameterSupported, nil
}

// AuthenticationMethodsReference implements Provider.
func (p *OpenIDConnectProvider) AuthenticationMethodsReference(asserted []string) []string {
	return p.amr.resolve(asserted)
}

// AuthorizationEndpoint returns the resolved authorization endpoint.
func (p *OpenIDConnectProvider) AuthorizationEndpoint() string {
	return p.endpointAuthorization
}

// TokenEndpoint returns the resolved token endpoint.
func (p *OpenIDConnectProvider) TokenEndpoint() string {
	return p.endpointToken
}

// UserInfoEndpoint returns the resolved UserInfo endpoint.
func (p *OpenIDConnectProvider) UserInfoEndpoint() string {
	return p.endpointUserInfo
}

// AuthorizationRequest implements Provider.
func (p *OpenIDConnectProvider) AuthorizationRequest(ctx context.Context, rand random.Provider, options AuthorizationRequestOptions) (request *AuthorizationRequest, err error) {
	if err = p.Resolve(ctx); err != nil {
		return nil, err
	}

	var opts []oauth2.AuthCodeOption

	if p.responseMode == ResponseModeFormPost {
		opts = append(opts, oauth2.SetAuthURLParam("response_mode", ResponseModeFormPost))
	}

	if options.Language != "" {
		opts = append(opts, oauth2.SetAuthURLParam("ui_locales", options.Language))
	}

	if request, err = newAuthorizationRequest(rand, p.oauth2Config(options.RedirectURI), true, opts...); err != nil {
		return nil, err
	}

	if p.parRequired {
		if request.URL, err = p.pushAuthorizationRequest(ctx, request.URL); err != nil {
			return nil, err
		}
	}

	return request, nil
}

// Complete implements Provider. The code is exchanged for an ID Token which is fully validated, and the UserInfo
// Endpoint is only requested after that, as its response is only trusted to the extent it describes the subject of
// that ID Token.
func (p *OpenIDConnectProvider) Complete(ctx context.Context, request CompletionRequest) (claims *IdentityClaims, err error) {
	var token *Token

	if token, err = p.Exchange(ctx, request.Code, request.CodeVerifier, request.RedirectURI); err != nil {
		return nil, err
	}

	if claims, err = p.validateIDToken(ctx, token.IDToken, request.Nonce, token.AccessToken, request.Now); err != nil {
		return nil, err
	}

	if err = p.UserInfo(ctx, token.AccessToken, claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// Exchange exchanges an authorization code for an ID Token and an access token. The access token is only returned so
// the UserInfo Endpoint can be requested with it; it is never stored or logged.
func (p *OpenIDConnectProvider) Exchange(ctx context.Context, code, verifier, redirectURI string) (token *Token, err error) {
	if err = p.Resolve(ctx); err != nil {
		return nil, err
	}

	var response *oauth2.Token

	if response, err = exchangeCode(ctx, p.client, p.oauth2Config(redirectURI), code, verifier); err != nil {
		return nil, err
	}

	token = &Token{AccessToken: response.AccessToken}

	var ok bool

	if token.IDToken, ok = response.Extra("id_token").(string); !ok || token.IDToken == "" {
		return nil, fmt.Errorf("error exchanging the authorization code: the token response did not contain an id token")
	}

	return token, nil
}

// ValidateIDToken v alidates a raw ID Token against this provider.
func (p *OpenIDConnectProvider) ValidateIDToken(ctx context.Context, raw, nonce string, now time.Time) (claims *IdentityClaims, err error) {
	return p.validateIDToken(ctx, raw, nonce, "", now)
}

func (p *OpenIDConnectProvider) validateIDToken(ctx context.Context, raw, nonce, accessToken string, now time.Time) (claims *IdentityClaims, err error) {
	if err = p.Resolve(ctx); err != nil {
		return nil, err
	}

	return ValidateIDToken(ctx, p.keys, raw, ValidateOptions{
		Issuer:      p.issuer,
		ClientID:    p.clientID,
		Nonce:       nonce,
		Alg:         p.algIDToken,
		JWKSURI:     p.jwksURI,
		Now:         now,
		Leeway:      time.Minute,
		AccessToken: accessToken,
	})
}

// Resolve performs OpenID Connect Discovery 1.0 for this provider if it has not already been performed successfully.
// Discovery is deliberately not performed while the providers are constructed: the availability of a third party must
// never determine whether Authelia is able to start, as that would take down every other authentication method with
// it. A failure here is retried the next time the provider is used.
func (p *OpenIDConnectProvider) Resolve(ctx context.Context) (err error) {
	if !p.discover {
		return nil
	}

	p.mutex.Lock()

	defer p.mutex.Unlock()

	if p.resolved {
		return nil
	}

	var discovery *Discovery

	if discovery, err = Discover(ctx, p.client, p.issuer); err != nil {
		return fmt.Errorf("error resolving provider '%s': %w", p.id, err)
	}

	if err = p.validateDiscovery(discovery); err != nil {
		return fmt.Errorf("error resolving provider '%s': %w", p.id, err)
	}

	var alg string

	if alg, err = p.selectIDTokenSigningAlg(discovery); err != nil {
		return fmt.Errorf("error resolving provider '%s': %w", p.id, err)
	}

	parEndpoint, parRequired := p.endpointPAR, p.parRequired || discovery.RequirePushedAuthorizationRequests

	if parEndpoint == "" {
		parEndpoint = discovery.PushedAuthorizationRequestEndpoint
	}

	if parRequired && parEndpoint == "" {
		return fmt.Errorf("error resolving provider '%s': %w: pushed authorization requests are required but the discovery document does not include the 'pushed_authorization_request_endpoint'", p.id, ErrDiscoveryEndpointMissing)
	}

	if p.endpointAuthorization == "" {
		p.endpointAuthorization = discovery.AuthorizationEndpoint
	}

	if p.endpointToken == "" {
		p.endpointToken = discovery.TokenEndpoint
	}

	if p.endpointUserInfo == "" {
		p.endpointUserInfo = discovery.UserInfoEndpoint
	}

	if p.jwksURI == "" {
		p.jwksURI = discovery.JWKSURI
	}

	if discovery.AuthorizationResponseIssParameterSupported {
		p.issParameterSupported = true
	}

	p.algIDToken = alg
	p.endpointPAR, p.parRequired = parEndpoint, parRequired

	p.resolved = true

	return nil
}

func (p *OpenIDConnectProvider) validateDiscovery(discovery *Discovery) (err error) {
	checks := []discoveryCheck{
		{"id_token_signing_alg_values_supported", discovery.IDTokenSigningAlgs, p.algIDToken},
		{"code_challenge_methods_supported", discovery.CodeChallengeMethods, "S256"},
		{"token_endpoint_auth_methods_supported", discovery.TokenEndpointAuthMethods, p.tokenEndpointAuthMethod},
		{"userinfo_signing_alg_values_supported", discovery.UserInfoSigningAlgs, p.algUserInfo},
	}

	if p.responseMode == ResponseModeFormPost {
		checks = append(checks, discoveryCheck{"response_modes_supported", discovery.ResponseModes, ResponseModeFormPost})
	}

	for _, check := range checks {
		if len(check.values) == 0 || check.value == "" || containsString(check.values, check.value) {
			continue
		}

		return fmt.Errorf("error validating the discovery document: %w: the value '%s' is not included in the '%s' values '%s'", ErrDiscoveryUnsupported, check.value, check.name, strings.Join(check.values, "', '"))
	}

	return nil
}

func (p *OpenIDConnectProvider) oauth2Config(redirectURI string) (cfg *oauth2.Config) {
	return newOAuth2Config(p.clientID, p.clientSecret, p.tokenEndpointAuthMethod, p.scopes, p.endpointAuthorization, p.endpointToken, redirectURI)
}

func (p *OpenIDConnectProvider) selectIDTokenSigningAlg(discovery *Discovery) (alg string, err error) {
	switch {
	case p.algIDToken != "":
		return p.algIDToken, nil
	case len(discovery.IDTokenSigningAlgs) == 0, containsString(discovery.IDTokenSigningAlgs, defaultIDTokenSigningAlg):
		return defaultIDTokenSigningAlg, nil
	}

	for _, advertised := range discovery.IDTokenSigningAlgs {
		if containsString(supportedIDTokenSigningAlgs, advertised) {
			return advertised, nil
		}
	}

	return "", fmt.Errorf("error validating the discovery document: %w: the 'id_token_signing_alg_values_supported' values are '%s'", ErrDiscoveryNoSupportedAlg, strings.Join(discovery.IDTokenSigningAlgs, "', '"))
}

type discoveryCheck struct {
	name   string
	values []string
	value  string
}

type staticKeySet struct {
	jwks *jose.JSONWebKeySet
}

func (s *staticKeySet) Resolve(_ context.Context, _ string, _ bool) (jwks *jose.JSONWebKeySet, err error) {
	return s.jwks, nil
}
