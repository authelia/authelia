// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

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
	client.HTTPClient = &http.Client{Timeout: time.Second * 10, Transport: newProviderTransport(caCertPool), CheckRedirect: checkProviderRedirect}
	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if errors.Is(err, ErrRedirectInsecure) {
			return false, err
		}

		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	return client
}

func checkProviderRedirect(req *http.Request, via []*http.Request) error {
	if req.URL.Scheme != "https" {
		return fmt.Errorf("%w: the location is '%s'", ErrRedirectInsecure, req.URL.Redacted())
	}

	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}

	return nil
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
