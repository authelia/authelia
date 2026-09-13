// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

const (
	discordIssuer                = "https://discord.com"
	discordAuthorizationEndpoint = "https://discord.com/oauth2/authorize"
	discordTokenEndpoint         = "https://discord.com/api/oauth2/token" //nolint:gosec // This is a URL, not a credential.
	discordUserEndpoint          = "https://discord.com/api/v10/users/@me"
)

// DiscordProvider is an external identity provider which is Discord. Discord is an OAuth 2.0 provider rather than an
// OpenID Connect 1.0 Provider, so there is no ID Token to validate: the identity is the Discord user the access token,
// obtained directly from the token endpoint over TLS, belongs to as returned by the current user endpoint.
type DiscordProvider struct {
	id                      string
	name                    string
	clientID                string
	clientSecret            string
	tokenEndpointAuthMethod string
	scopes                  []string
	amr                     authenticationMethodsReferencePolicy

	authorizationEndpoint string
	tokenEndpoint         string
	userEndpoint          string

	client *retryablehttp.Client
}

func newDiscordProvider(config *schema.AuthenticationBackendExternalIdentityProvider, client *retryablehttp.Client) *DiscordProvider {
	return &DiscordProvider{
		id:                      config.ID,
		name:                    config.Name,
		clientID:                config.ClientID,
		clientSecret:            config.ClientSecret,
		tokenEndpointAuthMethod: config.TokenEndpointAuthMethod,
		scopes:                  config.Scopes,
		amr:                     newAssertionlessAuthenticationMethodsReferencePolicy(config.AuthenticationMethodsReference),
		authorizationEndpoint:   discordAuthorizationEndpoint,
		tokenEndpoint:           discordTokenEndpoint,
		userEndpoint:            discordUserEndpoint,
		client:                  client,
	}
}

// ID implements Provider.
func (p *DiscordProvider) ID() string {
	return p.id
}

// Name implements Provider.
func (p *DiscordProvider) Name() string {
	return p.name
}

// Type implements Provider.
func (p *DiscordProvider) Type() string {
	return ProviderTypeDiscord
}

// Issuer implements Provider.
func (p *DiscordProvider) Issuer() string {
	return discordIssuer
}

// ResponseMode implements Provider.
func (p *DiscordProvider) ResponseMode() string {
	return ResponseModeQuery
}

// AuthorizationResponseIssuerRequired implements Provider. Discord does not send the 'iss' parameter.
func (p *DiscordProvider) AuthorizationResponseIssuerRequired(_ context.Context) (required bool, err error) {
	return false, nil
}

// AuthenticationMethodsReference implements Provider. Discord asserts nothing about how the user authenticated, so
// the configured default values, or the values of a password sign in, are always adopted.
func (p *DiscordProvider) AuthenticationMethodsReference(asserted []string) []string {
	return p.amr.resolve(asserted)
}

// AuthorizationRequest implements Provider.
func (p *DiscordProvider) AuthorizationRequest(_ context.Context, rand random.Provider, options AuthorizationRequestOptions) (request *AuthorizationRequest, err error) {
	return newAuthorizationRequest(rand, p.oauth2Config(options.RedirectURI), false)
}

// Complete implements Provider.
func (p *DiscordProvider) Complete(ctx context.Context, request CompletionRequest) (claims *IdentityClaims, err error) {
	var token *oauth2.Token

	if token, err = exchangeCode(ctx, p.client, p.oauth2Config(request.RedirectURI), request.Code, request.CodeVerifier); err != nil {
		return nil, err
	}

	var user *discordUser

	if user, err = p.user(ctx, token.AccessToken); err != nil {
		return nil, err
	}

	claims = &IdentityClaims{
		Issuer:            discordIssuer,
		Subject:           user.ID,
		PreferredUsername: user.Username,
		Name:              user.Username,
	}

	if user.GlobalName != nil && *user.GlobalName != "" {
		claims.Name = *user.GlobalName
	}

	if user.Email != nil && user.Verified != nil && *user.Verified {
		claims.Email = *user.Email
	}

	return claims, nil
}

type discordUser struct {
	ID         string  `json:"id"`
	Username   string  `json:"username"`
	GlobalName *string `json:"global_name"`
	Email      *string `json:"email"`
	Verified   *bool   `json:"verified"`
}

func (p *DiscordProvider) user(ctx context.Context, accessToken string) (user *discordUser, err error) {
	if accessToken == "" {
		return nil, fmt.Errorf("error requesting the discord user: %w", ErrUserInfoAccessTokenMissing)
	}

	var req *retryablehttp.Request

	if req, err = retryablehttp.NewRequestWithContext(ctx, http.MethodGet, p.userEndpoint, nil); err != nil {
		return nil, fmt.Errorf("error requesting the discord user: %w", err)
	}

	req.Header.Set(headerAccept, mimeApplicationJSON)
	req.Header.Set(headerAuthorization, "Bearer "+accessToken)

	var resp *http.Response

	if resp, err = p.client.Do(req); err != nil {
		return nil, fmt.Errorf("error requesting the discord user: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error requesting the discord user: %w: the user endpoint returned status code %d", ErrDiscordUserInvalid, resp.StatusCode)
	}

	if mediaType, _, _ := mime.ParseMediaType(resp.Header.Get(headerContentType)); mediaType != mimeApplicationJSON {
		return nil, fmt.Errorf("error requesting the discord user: %w: the content type '%s' is not '%s'", ErrDiscordUserInvalid, resp.Header.Get(headerContentType), mimeApplicationJSON)
	}

	var data []byte

	if data, err = io.ReadAll(io.LimitReader(resp.Body, userinfoResponseLimit)); err != nil {
		return nil, fmt.Errorf("error requesting the discord user: %w", err)
	}

	user = &discordUser{}

	if err = json.Unmarshal(data, user); err != nil {
		return nil, fmt.Errorf("error requesting the discord user: %w: %w", ErrDiscordUserInvalid, err)
	}

	if user.ID == "" {
		return nil, fmt.Errorf("error requesting the discord user: %w: the 'id' is required but it is absent", ErrDiscordUserInvalid)
	}

	return user, nil
}

func (p *DiscordProvider) oauth2Config(redirectURI string) (cfg *oauth2.Config) {
	return newOAuth2Config(p.clientID, p.clientSecret, p.tokenEndpointAuthMethod, p.scopes, p.authorizationEndpoint, p.tokenEndpoint, redirectURI)
}
