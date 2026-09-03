// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	parameterClientID     = "client_id"
	parameterClientSecret = "client_secret"
	parameterRequestURI   = "request_uri"

	mimeApplicationXWWWFormURLEncoded = "application/x-www-form-urlencoded"

	pushedAuthorizationResponseLimit = 1024 * 64
)

// PushedAuthorizationRequestError is an error response from a Pushed Authorization Request endpoint. It carries the
// parts of the response which describe the error, each sanitized with SanitizeProviderErrorValue, and never the
// response body itself.
type PushedAuthorizationRequestError struct {
	StatusCode  int
	ErrorCode   string
	Description string
	Hint        string
}

// Error implements error.
func (e *PushedAuthorizationRequestError) Error() string {
	b := strings.Builder{}

	fmt.Fprintf(&b, "the pushed authorization request endpoint returned an error response with status code %d", e.StatusCode)

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

func (p *OpenIDConnectProvider) pushAuthorizationRequest(ctx context.Context, authorizationURL string) (location string, err error) {
	if p.endpointPAR == "" {
		return "", fmt.Errorf("error pushing the authorization request: %w", ErrPushedAuthorizationRequestEndpointMissing)
	}

	var authorization *url.URL

	if authorization, err = url.Parse(authorizationURL); err != nil {
		return "", fmt.Errorf("error pushing the authorization request: %w", err)
	}

	form := authorization.Query()

	if p.tokenEndpointAuthMethod == "client_secret_post" {
		form.Set(parameterClientSecret, p.clientSecret)
	}

	var req *retryablehttp.Request

	if req, err = retryablehttp.NewRequestWithContext(ctx, http.MethodPost, p.endpointPAR, []byte(form.Encode())); err != nil {
		return "", fmt.Errorf("error pushing the authorization request: %w", err)
	}

	req.Header.Set(headerContentType, mimeApplicationXWWWFormURLEncoded)
	req.Header.Set(headerAccept, mimeApplicationJSON)

	// The client identifier and secret are form encoded before they are used as the credentials as RFC6749 Section
	// 2.3.1 requires, which is also what the token endpoint request does.
	if p.tokenEndpointAuthMethod == "client_secret_basic" {
		req.SetBasicAuth(url.QueryEscape(p.clientID), url.QueryEscape(p.clientSecret))
	}

	var resp *http.Response

	if resp, err = p.client.Do(req); err != nil {
		return "", fmt.Errorf("error pushing the authorization request: %w", err)
	}

	defer resp.Body.Close()

	var data []byte

	if data, err = io.ReadAll(io.LimitReader(resp.Body, pushedAuthorizationResponseLimit)); err != nil {
		return "", fmt.Errorf("error pushing the authorization request: %w", err)
	}

	body := struct {
		RequestURI  string `json:"request_uri"`
		ErrorCode   string `json:"error"`
		Description string `json:"error_description"`
		Hint        string `json:"error_hint"`
	}{}

	// A body which is not JSON leaves every value empty, which is reported below as an error response without detail
	// or as an invalid response, so the error is not otherwise needed.
	_ = json.Unmarshal(data, &body)

	// RFC9126 Section 2.2 specifies 201 Created, however some providers respond with 200 OK and there is nothing to
	// gain by rejecting a response which is otherwise valid.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error pushing the authorization request: %w", &PushedAuthorizationRequestError{
			StatusCode:  resp.StatusCode,
			ErrorCode:   SanitizeProviderErrorValue(body.ErrorCode),
			Description: SanitizeProviderErrorValue(body.Description),
			Hint:        SanitizeProviderErrorValue(body.Hint),
		})
	}

	if body.RequestURI == "" {
		return "", fmt.Errorf("error pushing the authorization request: %w: the 'request_uri' is absent", ErrPushedAuthorizationResponseInvalid)
	}

	var endpoint *url.URL

	if endpoint, err = url.Parse(p.endpointAuthorization); err != nil {
		return "", fmt.Errorf("error pushing the authorization request: %w", err)
	}

	query := endpoint.Query()

	query.Set(parameterClientID, p.clientID)
	query.Set(parameterRequestURI, body.RequestURI)

	endpoint.RawQuery = query.Encode()

	return endpoint.String(), nil
}
