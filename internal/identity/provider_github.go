// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

// GitHubProvider is an external identity provider which is GitHub. GitHub is an OAuth 2.0 provider rather than an
// OpenID Connect 1.0 Provider, so there is no ID Token to validate: the identity is the GitHub user the access token,
// obtained directly from the token endpoint over TLS, belongs to as returned by the authenticated user endpoint.
type GitHubProvider struct {
	id                      string
	name                    string
	logoURI                 string
	clientID                string
	clientSecret            string
	tokenEndpointAuthMethod string
	scopes                  []string
	sharedRedirectURI       bool
	amr                     authenticationMethodsReferencePolicy

	authorizationEndpoint string
	tokenEndpoint         string
	userEndpoint          string
	emailsEndpoint        string
	organizationsEndpoint string
	teamsEndpoint         string

	client *retryablehttp.Client
}

func newGitHubProvider(config *schema.AuthenticationBackendExternalIdentityProvider, client *retryablehttp.Client) *GitHubProvider {
	return &GitHubProvider{
		id:                      config.ID,
		name:                    config.Name,
		logoURI:                 config.LogoURI,
		clientID:                config.ClientID,
		clientSecret:            config.ClientSecret,
		tokenEndpointAuthMethod: config.TokenEndpointAuthMethod,
		scopes:                  config.Scopes,
		sharedRedirectURI:       config.SharedRedirectURI,
		amr:                     newAssertionlessAuthenticationMethodsReferencePolicy(config.AuthenticationMethodsReference),
		authorizationEndpoint:   githubAuthorizationEndpoint,
		tokenEndpoint:           githubTokenEndpoint,
		userEndpoint:            githubUserEndpoint,
		emailsEndpoint:          githubEmailsEndpoint,
		organizationsEndpoint:   githubOrganizationsEndpoint,
		teamsEndpoint:           githubTeamsEndpoint,
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

// LogoURI implements Provider.
func (p *GitHubProvider) LogoURI() string {
	return p.logoURI
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

// SharedRedirectURI implements Provider.
func (p *GitHubProvider) SharedRedirectURI() bool {
	return p.sharedRedirectURI
}

// AuthorizationResponseIssuerRequired implements Provider. GitHub sends the 'iss' parameter, which is validated when it
// is present, but it is not required as GitHub does not advertise it in any metadata.
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

	if containsString(p.scopes, githubScopeReadOrg) {
		if err = p.memberships(ctx, token.AccessToken, user.raw); err != nil {
			return nil, err
		}
	}

	var primary string

	// Any failure is returned, rather than the claims being incomplete, as a claims mapping may decide whether a user can
	// sign in based on them. This is decided by the configured scopes and not the granted scopes, as a user can reduce the
	// scopes they grant and must not be able to avoid a claims mapping by doing so.
	if p.emailsScoped() {
		var emails []any

		if emails, primary, err = p.emails(ctx, token.AccessToken); err != nil {
			return nil, fmt.Errorf("error requesting the github emails: %w", err)
		}

		user.raw[githubClaimEmails] = emails
	}

	claims = &IdentityClaims{
		Issuer:            githubIssuer,
		Subject:           strconv.FormatInt(user.ID, 10),
		PreferredUsername: user.Login,
		Name:              user.Login,
		Email:             primary,
		RawProviders:      map[string]map[string]any{ProviderTypeGitHub: user.raw},
	}

	if user.Name != nil && *user.Name != "" {
		claims.Name = *user.Name
	}

	return claims, nil
}

// memberships adds the organizations and teams of the user to the raw claims. Any failure is returned, rather than the
// claims being incomplete, as a claims mapping may decide whether a user can sign in based on them.
func (p *GitHubProvider) memberships(ctx context.Context, accessToken string, raw map[string]any) (err error) {
	if raw[githubClaimOrganizations], err = p.list(ctx, p.organizationsEndpoint, accessToken); err != nil {
		return fmt.Errorf("error requesting the github organizations: %w", err)
	}

	if raw[githubClaimTeams], err = p.list(ctx, p.teamsEndpoint, accessToken); err != nil {
		return fmt.Errorf("error requesting the github teams: %w", err)
	}

	return nil
}

func (p *GitHubProvider) list(ctx context.Context, endpoint, accessToken string) (items []any, err error) {
	var parsed *url.URL

	if parsed, err = url.Parse(endpoint); err != nil {
		return nil, err
	}

	items = []any{}

	for page := 1; page <= githubListPagesLimit; page++ {
		query := parsed.Query()

		query.Set("per_page", strconv.Itoa(githubListPerPage))
		query.Set("page", strconv.Itoa(page))

		parsed.RawQuery = query.Encode()

		var current []any

		if err = p.get(ctx, parsed.String(), accessToken, &current); err != nil {
			return nil, err
		}

		items = append(items, current...)

		if len(current) < githubListPerPage {
			return items, nil
		}
	}

	return nil, fmt.Errorf("%w: the list exceeds %d pages", ErrGitHubResponseInvalid, githubListPagesLimit)
}

// emailsScoped returns true if a scope which grants access to the email addresses of the user is configured.
func (p *GitHubProvider) emailsScoped() bool {
	return containsString(p.scopes, githubScopeUserEmail) || containsString(p.scopes, githubScopeUser)
}

// emails returns all of the email addresses of the user and the verified primary email address among them. The response
// is rejected in full if any of the email addresses lacks the properties which describe it.
func (p *GitHubProvider) emails(ctx context.Context, accessToken string) (emails []any, primary string, err error) {
	if emails, err = p.list(ctx, p.emailsEndpoint, accessToken); err != nil {
		return nil, "", err
	}

	for i, item := range emails {
		entry, ok := item.(map[string]any)
		if !ok {
			return nil, "", fmt.Errorf("%w: the email address at index %d is not an object", ErrGitHubResponseInvalid, i)
		}

		address, okAddress := entry[githubClaimEmail].(string)
		isPrimary, okPrimary := entry[githubClaimPrimary].(bool)
		isVerified, okVerified := entry[githubClaimVerified].(bool)

		if !okAddress || !okPrimary || !okVerified {
			return nil, "", fmt.Errorf("%w: the email address at index %d requires the '%s', '%s', and '%s' properties", ErrGitHubResponseInvalid, i, githubClaimEmail, githubClaimPrimary, githubClaimVerified)
		}

		if primary == "" && isPrimary && isVerified {
			primary = address
		}
	}

	return emails, primary, nil
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

	if mediaType, _, _ := mime.ParseMediaType(resp.Header.Get(headerContentType)); mediaType != mimeApplicationJSON && mediaType != mimeApplicationGitHubJSON {
		return fmt.Errorf("%w: the content type '%s' is not '%s' or '%s'", ErrGitHubResponseInvalid, resp.Header.Get(headerContentType), mimeApplicationJSON, mimeApplicationGitHubJSON)
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
