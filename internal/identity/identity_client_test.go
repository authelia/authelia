// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProviderClientShouldRefuseInsecureRedirect(t *testing.T) {
	var targetRequests, sourceRequests int

	target := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		targetRequests++

		rw.WriteHeader(http.StatusOK)
	}))

	defer target.Close()

	source := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		sourceRequests++

		http.Redirect(rw, r, target.URL+"/token", http.StatusFound)
	}))

	defer source.Close()

	pool := x509.NewCertPool()
	pool.AddCert(source.Certificate())

	client := newProviderClient(pool)

	req, err := retryablehttp.NewRequest(http.MethodGet, source.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	if resp != nil {
		_ = resp.Body.Close()
	}

	require.ErrorIs(t, err, ErrRedirectInsecure)
	assert.Equal(t, 1, sourceRequests)
	assert.Equal(t, 0, targetRequests)
}

func TestNewProviderClientShouldFollowSecureRedirect(t *testing.T) {
	var server *httptest.Server

	server = httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/final" {
			rw.WriteHeader(http.StatusOK)

			return
		}

		http.Redirect(rw, r, server.URL+"/final", http.StatusFound)
	}))

	defer server.Close()

	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())

	client := newProviderClient(pool)

	req, err := retryablehttp.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "/final", resp.Request.URL.Path)
}
