package oidcrp

import (
	"context"
	"encoding/base64"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/authelia/authelia/v4/internal/random"
)

// AuthorizationRequest represents a constructed authorization request and the values which must be retained to
// validate the eventual callback.
type AuthorizationRequest struct {
	URL          string
	State        string
	Nonce        string
	CodeVerifier string
}

// AuthorizationRequest constructs an authorization request for this provider.
func (p *Provider) AuthorizationRequest(ctx context.Context, rand random.Provider, redirectURI string) (request *AuthorizationRequest, err error) {
	if err = p.Resolve(ctx); err != nil {
		return nil, err
	}

	var state, nonce []byte

	if state, err = rand.BytesCustomErr(32, nil); err != nil {
		return nil, fmt.Errorf("error generating the authorization request: %w", err)
	}

	if nonce, err = rand.BytesCustomErr(32, nil); err != nil {
		return nil, fmt.Errorf("error generating the authorization request: %w", err)
	}

	verifier := oauth2.GenerateVerifier()

	request = &AuthorizationRequest{
		State:        base64.RawURLEncoding.EncodeToString(state),
		Nonce:        base64.RawURLEncoding.EncodeToString(nonce),
		CodeVerifier: verifier,
	}

	cfg := p.oauth2Config(redirectURI)

	request.URL = cfg.AuthCodeURL(request.State, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("nonce", request.Nonce))

	return request, nil
}

// Exchange exchanges an authorization code for an ID Token. The access token in the response is never stored,
// logged, or returned.
func (p *Provider) Exchange(ctx context.Context, code, verifier, redirectURI string) (rawIDToken string, err error) {
	if err = p.Resolve(ctx); err != nil {
		return "", err
	}

	cfg := p.oauth2Config(redirectURI)

	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.client.StandardClient())

	var token *oauth2.Token

	if token, err = cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier)); err != nil {
		return "", fmt.Errorf("error exchanging the authorization code: %w", err)
	}

	var ok bool

	if rawIDToken, ok = token.Extra("id_token").(string); !ok || rawIDToken == "" {
		return "", fmt.Errorf("error exchanging the authorization code: the token response did not contain an id token")
	}

	return rawIDToken, nil
}

func (p *Provider) oauth2Config(redirectURI string) (cfg *oauth2.Config) {
	cfg = &oauth2.Config{
		ClientID:    p.ClientID,
		RedirectURL: redirectURI,
		Scopes:      p.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.authorizationEndpoint,
			TokenURL: p.tokenEndpoint,
		},
	}

	switch p.tokenEndpointAuthMethod {
	case "client_secret_basic":
		cfg.ClientSecret = p.clientSecret
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	case "client_secret_post":
		cfg.ClientSecret = p.clientSecret
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	default:
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	}

	return cfg
}
