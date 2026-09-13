// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"

	"github.com/authelia/authelia/v4/internal/random"
)

const (
	// ProviderTypeOpenIDConnect is the type of a provider which is an OpenID Connect 1.0 Provider.
	ProviderTypeOpenIDConnect = "openid_connect"

	// ProviderTypeDiscord is the type of a provider which is Discord.
	ProviderTypeDiscord = "discord"

	// ProviderTypePlex is the type of a provider which is Plex.
	ProviderTypePlex = "plex"

	// ProviderTypeGitHub is the type of a provider which is GitHub.
	ProviderTypeGitHub = "github"

	// ResponseModeQuery is the response mode which delivers the authorization response in the query of a GET request
	// to the redirect URI.
	ResponseModeQuery = "query"

	// ResponseModeFormPost is the response mode which delivers the authorization response in the form body of a POST
	// request to the redirect URI.
	ResponseModeFormPost = "form_post"
)

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

func newAuthorizationRequest(rand random.Provider, cfg *oauth2.Config, nonce bool, opts ...oauth2.AuthCodeOption) (request *AuthorizationRequest, err error) {
	var state []byte

	if state, err = rand.BytesCustomErr(32, nil); err != nil {
		return nil, fmt.Errorf("error generating the authorization request: %w", err)
	}

	verifier := oauth2.GenerateVerifier()

	request = &AuthorizationRequest{
		State:        base64.RawURLEncoding.EncodeToString(state),
		CodeVerifier: verifier,
	}

	options := []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(verifier)}

	if nonce {
		var value []byte

		if value, err = rand.BytesCustomErr(32, nil); err != nil {
			return nil, fmt.Errorf("error generating the authorization request: %w", err)
		}

		request.Nonce = base64.RawURLEncoding.EncodeToString(value)

		options = append(options, oauth2.SetAuthURLParam("nonce", request.Nonce))
	}

	request.URL = cfg.AuthCodeURL(request.State, append(options, opts...)...)

	return request, nil
}

func newOAuth2Config(clientID, clientSecret, authMethod string, scopes []string, authorizationEndpoint, tokenEndpoint, redirectURI string) (cfg *oauth2.Config) {
	cfg = &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: redirectURI,
		Scopes:      scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authorizationEndpoint,
			TokenURL: tokenEndpoint,
		},
	}

	switch authMethod {
	case "client_secret_basic":
		cfg.ClientSecret = clientSecret
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	case "client_secret_post":
		cfg.ClientSecret = clientSecret
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	default:
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	}

	return cfg
}

func exchangeCode(ctx context.Context, client *retryablehttp.Client, cfg *oauth2.Config, code, verifier string) (token *oauth2.Token, err error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, client.StandardClient())

	if token, err = cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier)); err != nil {
		if retrieve, ok := errors.AsType[*oauth2.RetrieveError](err); ok {
			return nil, fmt.Errorf("error exchanging the authorization code: %w", newTokenEndpointError(retrieve))
		}

		return nil, fmt.Errorf("error exchanging the authorization code: %w", err)
	}

	return token, nil
}

// TokenEndpointError is an error response from a token endpoint. It carries the parts of the response which describe
// the error, each sanitized with SanitizeProviderErrorValue, and never the response body itself.
type TokenEndpointError struct {
	StatusCode  int
	ErrorCode   string
	Description string
	Hint        string

	err *oauth2.RetrieveError
}

func newTokenEndpointError(retrieve *oauth2.RetrieveError) (err *TokenEndpointError) {
	err = &TokenEndpointError{
		ErrorCode:   SanitizeProviderErrorValue(retrieve.ErrorCode),
		Description: SanitizeProviderErrorValue(retrieve.ErrorDescription),
		err:         retrieve,
	}

	if retrieve.Response != nil {
		err.StatusCode = retrieve.Response.StatusCode
	}

	body := struct {
		Hint string `json:"error_hint"`
	}{}

	if json.Unmarshal(retrieve.Body, &body) == nil {
		err.Hint = SanitizeProviderErrorValue(body.Hint)
	}

	return err
}

// Error implements error.
func (e *TokenEndpointError) Error() string {
	b := strings.Builder{}

	b.WriteString("the token endpoint returned an error response")

	if e.StatusCode != 0 {
		fmt.Fprintf(&b, " with status code %d", e.StatusCode)
	}

	if e.ErrorCode != "" {
		fmt.Fprintf(&b, ": error '%s'", e.ErrorCode)
	}

	if e.Description != "" {
		fmt.Fprintf(&b, ", description '%s'", e.Description)
	}

	if e.Hint != "" {
		fmt.Fprintf(&b, ", hint '%s'", e.Hint)
	}

	return b.String()
}

// Unwrap returns the underlying *oauth2.RetrieveError.
func (e *TokenEndpointError) Unwrap() error {
	return e.err
}

const providerErrorValueLimit = 256

// SanitizeProviderErrorValue prepares a value an external provider supplied to describe an error for logging. Control
// characters are removed so the value cannot forge or split log entries, and it is truncated as the provider, or anyone
// able to direct the browser to the callback, decides its length.
func SanitizeProviderErrorValue(value string) string {
	value = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}

		return r
	}, value)

	if utf8.RuneCountInString(value) > providerErrorValueLimit {
		value = string([]rune(value)[:providerErrorValueLimit]) + "..."
	}

	return value
}

func newProviderClient(caCertPool *x509.CertPool) (client *retryablehttp.Client) {
	client = retryablehttp.NewClient()
	client.Logger = nil
	client.RetryMax = 2
	client.HTTPClient = &http.Client{Timeout: time.Second * 10, Transport: newProviderTransport(caCertPool)}

	return client
}

func newProviderTransport(caCertPool *x509.CertPool) (transport *http.Transport) {
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = t.Clone()
	} else {
		transport = &http.Transport{}
	}

	transport.TLSClientConfig = &tls.Config{RootCAs: caCertPool, MinVersion: tls.VersionTLS12}

	return transport
}
