// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewClientShouldApplyTheTimeout(t *testing.T) {
	client := NewClient(&schema.WebhookDestination{Timeout: time.Second * 3}, nil)

	assert.Equal(t, time.Second*3, client.Timeout)
}

func TestNewClientShouldNotFollowRedirects(t *testing.T) {
	var reached bool

	target := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		reached = true

		rw.WriteHeader(http.StatusOK)
	}))

	defer target.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, target.URL, http.StatusFound)
	}))

	defer redirector.Close()

	client := NewClient(&schema.WebhookDestination{Timeout: time.Second * 5}, nil)

	req, err := http.NewRequest(http.MethodPost, redirector.URL, nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer secret")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.False(t, reached, "the client must not follow the redirect, which would resend credentials to another host")
}

func TestNewClientShouldUseTheConfiguredTLSMinimumVersion(t *testing.T) {
	client := NewClient(&schema.WebhookDestination{
		Timeout: time.Second,
		TLS:     &schema.TLS{MinimumVersion: schema.TLSVersion{Value: 0x0304}},
		Address: &url.URL{Scheme: "https", Host: "example.com"},
	}, nil)

	transport, ok := client.Transport.(*http.Transport)

	require.True(t, ok)
	assert.Equal(t, uint16(0x0304), transport.TLSClientConfig.MinVersion)
}
