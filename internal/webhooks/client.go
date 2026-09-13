// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"crypto/x509"
	"net/http"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewClient returns the HTTP client for a destination. Each destination has its own client and transport because each
// owns its own TLS configuration.
func NewClient(config *schema.WebhookDestination, caCertPool *x509.CertPool) (client *http.Client) {
	transport := &http.Transport{
		TLSClientConfig:     utils.NewTLSConfig(config.TLS, caCertPool),
		MaxIdleConns:        2,
		MaxIdleConnsPerHost: 2,
	}

	return &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
