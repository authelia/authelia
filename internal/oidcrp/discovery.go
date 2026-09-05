package oidcrp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	return discovery, nil
}
