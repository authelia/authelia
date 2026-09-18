// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

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

// Discover fetches and validates the OpenID Connect Discovery 1.0 document for the given issuer.
func Discover(ctx context.Context, client *retryablehttp.Client, issuer string) (discovery *Discovery, err error) {
	endpoint := strings.TrimSuffix(issuer, "/") + pathWellKnownOpenIDConfiguration

	var req *retryablehttp.Request

	if req, err = retryablehttp.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil); err != nil {
		return nil, fmt.Errorf("error discovering the provider: %w", err)
	}

	req.Header.Set(headerAccept, mimeApplicationJSON)

	var resp *http.Response

	if resp, err = client.Do(req); err != nil {
		return nil, fmt.Errorf("error discovering the provider: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error discovering the provider: the discovery endpoint returned status code %d", resp.StatusCode)
	}

	var data []byte

	if data, err = io.ReadAll(io.LimitReader(resp.Body, 1024*512)); err != nil {
		return nil, fmt.Errorf("error discovering the provider: %w", err)
	}

	discovery = &Discovery{}

	if err = json.Unmarshal(data, discovery); err != nil {
		return nil, fmt.Errorf("error discovering the provider: %w", err)
	}

	if discovery.Issuer != issuer {
		return nil, fmt.Errorf("error discovering the provider: %w", ErrDiscoveryIssuerMismatch)
	}

	if discovery.AuthorizationEndpoint == "" || discovery.TokenEndpoint == "" || discovery.JWKSURI == "" {
		return nil, fmt.Errorf("error discovering the provider: %w", ErrDiscoveryEndpointMissing)
	}

	if err = validateDiscoveryURLs(discovery); err != nil {
		return nil, fmt.Errorf("error discovering the provider: %w", err)
	}

	return discovery, nil
}

// validateDiscoveryURLs ensures every URL in the discovery document uses the https scheme. These URLs are used to send
// the client secret, the authorization code, and the access token, and to retrieve the keys ID Tokens are verified
// with, so a plaintext scheme must never be accepted from the discovery document any more than from the configuration.
func validateDiscoveryURLs(discovery *Discovery) (err error) {
	urls := []struct {
		name  string
		value string
	}{
		{"issuer", discovery.Issuer},
		{"authorization_endpoint", discovery.AuthorizationEndpoint},
		{"token_endpoint", discovery.TokenEndpoint},
		{"userinfo_endpoint", discovery.UserInfoEndpoint},
		{"jwks_uri", discovery.JWKSURI},
		{"pushed_authorization_request_endpoint", discovery.PushedAuthorizationRequestEndpoint},
	}

	for _, u := range urls {
		if u.value == "" {
			continue
		}

		var parsed *url.URL

		if parsed, err = url.Parse(u.value); err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("%w: the '%s' is '%s'", ErrDiscoveryURLInsecure, u.name, u.value)
		}
	}

	return nil
}
