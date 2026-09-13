// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/utils"
)

const (
	plexIssuer                = "https://plex.tv"
	plexProduct               = "Authelia"
	plexAuthorizationEndpoint = "https://app.plex.tv/auth"
	plexPINsEndpoint          = "https://plex.tv/api/v2/pins"
	plexUserEndpoint          = "https://plex.tv/api/v2/user"

	headerPlexProduct          = "X-Plex-Product"
	headerPlexVersion          = "X-Plex-Version"
	headerPlexClientIdentifier = "X-Plex-Client-Identifier"
	headerPlexDeviceName       = "X-Plex-Device-Name"
	headerPlexLanguage         = "X-Plex-Language"
	headerPlexToken            = "X-Plex-Token" //nolint:gosec // This is a header name, not a credential.
)

// PlexProvider is an external identity provider which is Plex. Plex offers neither OAuth 2.0 nor OpenID Connect 1.0 to
// third parties, so the user is signed in with the PIN flow Plex apps use: a PIN is created, the user approves it at
// Plex, and the token Plex issues for the approved PIN is retrieved directly from Plex over TLS. The identity is the
// Plex account that token belongs to.
type PlexProvider struct {
	id       string
	name     string
	clientID string
	amr      authenticationMethodsReferencePolicy

	authorizationEndpoint string
	pinsEndpoint          string
	userEndpoint          string

	client *retryablehttp.Client
}

func newPlexProvider(config *schema.AuthenticationBackendExternalIdentityProvider, client *retryablehttp.Client) *PlexProvider {
	return &PlexProvider{
		id:                    config.ID,
		name:                  config.Name,
		clientID:              config.ClientID,
		amr:                   newAssertionlessAuthenticationMethodsReferencePolicy(config.AuthenticationMethodsReference),
		authorizationEndpoint: plexAuthorizationEndpoint,
		pinsEndpoint:          plexPINsEndpoint,
		userEndpoint:          plexUserEndpoint,
		client:                client,
	}
}

// ID implements Provider.
func (p *PlexProvider) ID() string {
	return p.id
}

// Name implements Provider.
func (p *PlexProvider) Name() string {
	return p.name
}

// Type implements Provider.
func (p *PlexProvider) Type() string {
	return ProviderTypePlex
}

// Issuer implements Provider.
func (p *PlexProvider) Issuer() string {
	return plexIssuer
}

// ResponseMode implements Provider.
func (p *PlexProvider) ResponseMode() string {
	return ResponseModeQuery
}

// AuthorizationResponseIssuerRequired implements Provider. Plex does not send the 'iss' parameter.
func (p *PlexProvider) AuthorizationResponseIssuerRequired(_ context.Context) (required bool, err error) {
	return false, nil
}

// AuthenticationMethodsReference implements Provider. Plex asserts nothing about how the user authenticated, so the
// configured default values, or the values of a password sign in, are always adopted.
func (p *PlexProvider) AuthenticationMethodsReference(asserted []string) []string {
	return p.amr.resolve(asserted)
}

// AuthorizationRequest implements Provider. Plex returns the user to the forward URL exactly as it was given, so the
// state and the PIN id are carried in its query. The PIN id and code are also retained in the handle so the PIN which is
// completed is always the one this flow created, whatever the browser returns.
func (p *PlexProvider) AuthorizationRequest(ctx context.Context, rand random.Provider, options AuthorizationRequestOptions) (request *AuthorizationRequest, err error) {
	var state []byte

	if state, err = rand.BytesCustomErr(32, nil); err != nil {
		return nil, fmt.Errorf("error generating the authorization request: %w", err)
	}

	device := newPlexDevice(options.RedirectURI, options.Language)

	pin := &plexPIN{}

	if err = p.do(ctx, http.MethodPost, p.pinsEndpoint+"?strong=true", "", device, pin); err != nil {
		return nil, fmt.Errorf("error creating the plex pin: %w", err)
	}

	if pin.ID == 0 || pin.Code == "" {
		return nil, fmt.Errorf("error creating the plex pin: %w: the 'id' and 'code' are required but one or both are absent", ErrPlexResponseInvalid)
	}

	id := strconv.FormatInt(pin.ID, 10)

	request = &AuthorizationRequest{
		State:  base64.RawURLEncoding.EncodeToString(state),
		Handle: id + ":" + pin.Code,
	}

	forward := options.RedirectURI + "?" + url.Values{"state": []string{request.State}, "code": []string{id}}.Encode()

	request.URL = p.authorizationEndpoint + "#?" + url.Values{
		"clientID":                    []string{p.clientID},
		"code":                        []string{pin.Code},
		"forwardUrl":                  []string{forward},
		"context[device][product]":    []string{plexProduct},
		"context[device][version]":    []string{device.version},
		"context[device][deviceName]": []string{device.name},
	}.Encode()

	return request, nil
}

// Complete implements Provider.
func (p *PlexProvider) Complete(ctx context.Context, request CompletionRequest) (claims *IdentityClaims, err error) {
	id, code, ok := strings.Cut(request.Handle, ":")

	if !ok || id == "" || code == "" || subtle.ConstantTimeCompare([]byte(request.Code), []byte(id)) != 1 {
		return nil, fmt.Errorf("error retrieving the plex pin: %w", ErrPlexPINMismatch)
	}

	device := newPlexDevice(request.RedirectURI, request.Language)

	pin := &plexPIN{}

	if err = p.do(ctx, http.MethodGet, p.pinsEndpoint+"/"+url.PathEscape(id)+"?"+url.Values{"code": []string{code}}.Encode(), "", device, pin); err != nil {
		return nil, fmt.Errorf("error retrieving the plex pin: %w", err)
	}

	if pin.AuthToken == nil || *pin.AuthToken == "" {
		return nil, fmt.Errorf("error retrieving the plex pin: %w", ErrPlexPINUnauthorized)
	}

	user := &plexUser{}

	if err = p.do(ctx, http.MethodGet, p.userEndpoint, *pin.AuthToken, device, user); err != nil {
		return nil, fmt.Errorf("error requesting the plex user: %w", err)
	}

	if user.UUID == "" {
		return nil, fmt.Errorf("error requesting the plex user: %w: the 'uuid' is required but it is absent", ErrPlexResponseInvalid)
	}

	claims = &IdentityClaims{
		Issuer:            plexIssuer,
		Subject:           user.UUID,
		PreferredUsername: user.Username,
		Name:              user.Username,
		Email:             user.Email,
	}

	if user.Title != "" {
		claims.Name = user.Title
	}

	return claims, nil
}

type plexPIN struct {
	ID        int64   `json:"id"`
	Code      string  `json:"code"`
	AuthToken *string `json:"authToken"`
}

type plexUser struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Title    string `json:"title"`
	Email    string `json:"email"`
}

type plexDevice struct {
	name     string
	version  string
	language string
}

func newPlexDevice(redirectURI, language string) plexDevice {
	device := plexDevice{name: plexProduct, version: utils.Version(), language: language}

	if uri, err := url.Parse(redirectURI); err == nil && uri.Scheme != "" && uri.Host != "" {
		device.name = fmt.Sprintf("%s (%s://%s)", plexProduct, uri.Scheme, uri.Host)
	}

	return device
}

func (p *PlexProvider) do(ctx context.Context, method, endpoint, token string, device plexDevice, v any) (err error) {
	var req *retryablehttp.Request

	if req, err = retryablehttp.NewRequestWithContext(ctx, method, endpoint, nil); err != nil {
		return err
	}

	req.Header.Set(headerAccept, mimeApplicationJSON)
	req.Header.Set(headerPlexProduct, plexProduct)
	req.Header.Set(headerPlexVersion, device.version)
	req.Header.Set(headerPlexClientIdentifier, p.clientID)
	req.Header.Set(headerPlexDeviceName, device.name)

	if device.language != "" {
		req.Header.Set(headerPlexLanguage, device.language)
	}

	if token != "" {
		req.Header.Set(headerPlexToken, token)
	}

	var resp *http.Response

	if resp, err = p.client.Do(req); err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("%w: the endpoint returned status code %d", ErrPlexResponseInvalid, resp.StatusCode)
	}

	if mediaType, _, _ := mime.ParseMediaType(resp.Header.Get(headerContentType)); mediaType != mimeApplicationJSON {
		return fmt.Errorf("%w: the content type '%s' is not '%s'", ErrPlexResponseInvalid, resp.Header.Get(headerContentType), mimeApplicationJSON)
	}

	var data []byte

	if data, err = io.ReadAll(io.LimitReader(resp.Body, userinfoResponseLimit)); err != nil {
		return err
	}

	if err = json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("%w: %w", ErrPlexResponseInvalid, err)
	}

	return nil
}
