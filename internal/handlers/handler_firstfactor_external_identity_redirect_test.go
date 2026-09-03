// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestOIDCRelyingPartyRedirect(t *testing.T) {
	testCases := []struct {
		Name     string
		BasePath string
		Location string
		Expected string
	}{
		{"ShouldResolveTheRootAgainstTheForwardedOrigin", "", "/", "https://login.example.com:8080/"},
		{"ShouldResolveTheQueryAgainstTheForwardedOrigin", "", "/external-identity/link?link_provider=example", "https://login.example.com:8080/external-identity/link?link_provider=example"},
		{"ShouldResolveThePathAgainstTheForwardedOrigin", "", "/settings/external-identity", "https://login.example.com:8080/settings/external-identity"},
		{"ShouldIncludeTheBasePathForTheRoot", "/auth", "/", "https://login.example.com:8080/auth/"},
		{"ShouldIncludeTheBasePathForTheQuery", "/auth", "/external-identity/link?link_provider=example", "https://login.example.com:8080/auth/external-identity/link?link_provider=example"},
		{"ShouldIncludeTheBasePathForThePath", "/auth", "/settings/external-identity", "https://login.example.com:8080/auth/settings/external-identity"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			// The proxy reaches Authelia at an address the browser cannot, which is what a relative location would be
			// resolved against.
			mock.Ctx.Request.SetHost("authelia-backend:9091")

			if tc.BasePath != "" {
				mock.Ctx.SetUserValue(middlewares.UserValueKeyBaseURL, tc.BasePath)
			}

			externalIdentityRedirect(mock.Ctx, tc.Location)

			assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.Expected, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
		})
	}
}

func TestOIDCRelyingPartyRedirectURI(t *testing.T) {
	testCases := []struct {
		Name     string
		BasePath string
		Expected string
	}{
		{"ShouldUseTheForwardedOrigin", "", "https://login.example.com:8080/api/firstfactor/external-identity/example/callback"},
		{"ShouldIncludeTheBasePath", "/auth", "https://login.example.com:8080/auth/api/firstfactor/external-identity/example/callback"},
		{"ShouldIncludeTheBasePathWithoutDoublingTheSlash", "/auth/", "https://login.example.com:8080/auth/api/firstfactor/external-identity/example/callback"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Request.SetHost("authelia-backend:9091")

			if tc.BasePath != "" {
				mock.Ctx.SetUserValue(middlewares.UserValueKeyBaseURL, tc.BasePath)
			}

			uri, err := externalIdentityRedirectURI(mock.Ctx, "example")

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, uri)
		})
	}
}
