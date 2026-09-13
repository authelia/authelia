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
	"strconv"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

const (
	githubIssuer                = "https://github.com"
	githubAuthorizationEndpoint = "https://github.com/login/oauth/authorize"
	githubTokenEndpoint         = "https://github.com/login/oauth/access_token" //nolint:gosec // This is a URL, not a credential.
	githubUserEndpoint          = "https://api.github.com/user"
	githubEmailsEndpoint        = "https://api.github.com/user/emails"

	headerGitHubAPIVersion    = "X-GitHub-Api-Version"
	githubAPIVersion          = "2022-11-28"
	mimeApplicationGitHubJSON = "application/vnd.github+json"
)

// GitHubProvider is an external identity provider which is GitHub. GitHub is an OAuth 2.0 provider rather than an
// OpenID Connect 1.0 Provider, so there is no ID Token to validate: the identity is the GitHub user the access token,
// obtained directly from the token endpoint over TLS, belongs to as returned by the authenticated user endpoint.
type GitHubProvider struct {
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
	emailsEndpoint        string

	client *retryablehttp.Client
}

func newGitHubProvider(config *schema.AuthenticationBackendExternalIdentityProvider, client *retryablehttp.Client) *GitHubProvider {
	return &GitHubProvider{
		id:                      config.ID,
		name:                    config.Name,
		clientID:                config.ClientID,
		clientSecret:            config.ClientSecret,
		tokenEndpointAuthMethod: config.TokenEndpointAuthMethod,
		scopes:                  config.Scopes,
		amr:                     newAssertionlessAuthenticationMethodsReferencePolicy(config.AuthenticationMethodsReference),
		authorizationEndpoint:   githubAuthorizationEndpoint,
		tokenEndpoint:           githubTokenEndpoint,
		userEndpoint:            githubUserEndpoint,
		emailsEndpoint:          githubEmailsEndpoint,
		client:                  client,
	}
}

// ID implements Provider.
func (p *GitHubProvider) ID() string {
	return p.id
}

// Name implements Provider.
func (p *GitHubProvider) Name() string {
	return p.name
}

// Type implements Provider.
func (p *GitHubProvider) Type() string {
	return ProviderTypeGitHub
}

// Issuer implements Provider.
func (p *GitHubProvider) Issuer() string {
	return githubIssuer
}

// ResponseMode implements Provider.
func (p *GitHubProvider) ResponseMode() string {
	return ResponseModeQuery
}

// AuthorizationResponseIssuerRequired implements Provider. GitHub does not send the 'iss' parameter.
func (p *GitHubProvider) AuthorizationResponseIssuerRequired(_ context.Context) (required bool, err error) {
	return false, nil
}

// AuthenticationMethodsReference implements Provider. GitHub asserts nothing about how the user authenticated, so the
// configured default values, or the values of a password sign in, are always adopted.
func (p *GitHubProvider) AuthenticationMethodsReference(asserted []string) []string {
	return p.amr.resolve(asserted)
}

// AuthorizationRequest implements Provider.
func (p *GitHubProvider) AuthorizationRequest(_ context.Context, rand random.Provider, options AuthorizationRequestOptions) (request *AuthorizationRequest, err error) {
	return newAuthorizationRequest(rand, p.oauth2Config(options.RedirectURI), false)
}

// Complete implements Provider.
func (p *GitHubProvider) Complete(ctx context.Context, request CompletionRequest) (claims *IdentityClaims, err error) {
	var token *oauth2.Token

	if token, err = exchangeCode(ctx, p.client, p.oauth2Config(request.RedirectURI), request.Code, request.CodeVerifier); err != nil {
		return nil, err
	}

	if token.AccessToken == "" {
		return nil, fmt.Errorf("error requesting the github user: %w", ErrUserInfoAccessTokenMissing)
	}

	user := &githubUser{}

	if err = p.get(ctx, p.userEndpoint, token.AccessToken, user); err != nil {
		return nil, fmt.Errorf("error requesting the github user: %w", err)
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("error requesting the github user: %w: the 'id' is required but it is absent", ErrGitHubResponseInvalid)
	}

	claims = &IdentityClaims{
		Issuer:            githubIssuer,
		Subject:           strconv.FormatInt(user.ID, 10),
		PreferredUsername: user.Login,
		Name:              user.Login,
		Email:             p.email(ctx, token.AccessToken),
	}

	if user.Name != nil && *user.Name != "" {
		claims.Name = *user.Name
	}

	return claims, nil
}

type githubUser struct {
	ID    int64   `json:"id"`
	Login string  `json:"login"`
	Name  *string `json:"name"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (p *GitHubProvider) email(ctx context.Context, accessToken string) string {
	var emails []githubEmail

	if err := p.get(ctx, p.emailsEndpoint, accessToken, &emails); err != nil {
		return ""
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email
		}
	}

	return ""
}

func (p *GitHubProvider) get(ctx context.Context, endpoint, accessToken string, v any) (err error) {
	var req *retryablehttp.Request

	if req, err = retryablehttp.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil); err != nil {
		return err
	}

	req.Header.Set(headerAccept, mimeApplicationGitHubJSON)
	req.Header.Set(headerAuthorization, "Bearer "+accessToken)
	req.Header.Set(headerGitHubAPIVersion, githubAPIVersion)

	var resp *http.Response

	if resp, err = p.client.Do(req); err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: the endpoint returned status code %d", ErrGitHubResponseInvalid, resp.StatusCode)
	}

	if mediaType, _, _ := mime.ParseMediaType(resp.Header.Get(headerContentType)); mediaType != mimeApplicationJSON {
		return fmt.Errorf("%w: the content type '%s' is not '%s'", ErrGitHubResponseInvalid, resp.Header.Get(headerContentType), mimeApplicationJSON)
	}

	var data []byte

	if data, err = io.ReadAll(io.LimitReader(resp.Body, userinfoResponseLimit)); err != nil {
		return err
	}

	if err = json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("%w: %w", ErrGitHubResponseInvalid, err)
	}

	return nil
}

func (p *GitHubProvider) oauth2Config(redirectURI string) (cfg *oauth2.Config) {
	return newOAuth2Config(p.clientID, p.clientSecret, p.tokenEndpointAuthMethod, p.scopes, p.authorizationEndpoint, p.tokenEndpoint, redirectURI)
}
