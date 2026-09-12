// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	oauthelia2 "authelia.com/provider/oauth2"
)

func TestRedirectAuthorizeErrorFieldResponseStrategy(t *testing.T) {
	testCases := []struct {
		name        string
		sendDebug   bool
		rfc         *oauthelia2.RFC6749Error
		contains    []string
		notContains []string
	}{
		{
			"ShouldRedirectWithErrorFields",
			false,
			&oauthelia2.RFC6749Error{
				ErrorField:       "invalid_request",
				DescriptionField: "The request is missing a required parameter.",
				CodeField:        http.StatusBadRequest,
				HintField:        "Parameter 'client_id' is missing.",
			},
			[]string{"error=invalid_request", "error_description=", "error_status_code=400", "error_hint=", FrontendEndpointPathConsentCompletion},
			[]string{"error_debug"},
		},
		{
			"ShouldIncludeDebugWhenEnabled",
			true,
			&oauthelia2.RFC6749Error{
				ErrorField: "server_error",
				DebugField: "internal debug info",
			},
			[]string{"error_debug="},
			nil,
		},
		{
			"ShouldNotIncludeDebugWhenDisabled",
			false,
			&oauthelia2.RFC6749Error{
				ErrorField: "server_error",
				DebugField: "internal debug info",
			},
			nil,
			[]string{"error_debug"},
		},
		{
			"ShouldHandleNilError",
			false,
			nil,
			[]string{"error=server_error"},
			nil,
		},
		{
			"ShouldHandleEmptyFields",
			false,
			&oauthelia2.RFC6749Error{},
			nil,
			[]string{"error=", "error_description="},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issuer := &url.URL{Scheme: "https", Host: "auth.example.com"}

			strategy := &RedirectAuthorizeErrorFieldResponseStrategy{
				Config: &testErrorConfig{issuer: issuer, sendDebug: tc.sendDebug},
			}

			rw := httptest.NewRecorder()

			strategy.WriteErrorFieldResponse(context.Background(), rw, nil, tc.rfc)

			assert.Equal(t, http.StatusSeeOther, rw.Code)
			assert.Equal(t, "no-store", rw.Header().Get("Cache-Control"))
			assert.Equal(t, "no-cache", rw.Header().Get("Pragma"))

			location := rw.Header().Get("Location")

			for _, s := range tc.contains {
				assert.Contains(t, location, s)
			}

			for _, s := range tc.notContains {
				assert.NotContains(t, location, s)
			}
		})
	}
}

func TestNewConsentCompletionErrorURI(t *testing.T) {
	issuer := &url.URL{Scheme: "https", Host: "auth.example.com", Path: "/base"}

	rfc := &oauthelia2.RFC6749Error{
		ErrorField:       "invalid_request",
		DescriptionField: "The request is missing a required parameter.",
		CodeField:        http.StatusBadRequest,
		HintField:        "Parameter 'client_id' is missing.",
		DebugField:       "internal debug info",
	}

	t.Run("ShouldBuildTheCompletionURIWithoutDebug", func(t *testing.T) {
		location := NewConsentCompletionErrorURI(issuer, rfc, false)

		assert.Equal(t, "/base"+FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, url.Values{
			"error":             []string{"invalid_request"},
			"error_description": []string{"The request is missing a required parameter."},
			"error_status_code": []string{"400"},
			"error_hint":        []string{"Parameter 'client_id' is missing."},
		}, location.Query())
	})

	t.Run("ShouldIncludeDebugWhenEnabled", func(t *testing.T) {
		assert.Equal(t, "internal debug info", NewConsentCompletionErrorURI(issuer, rfc, true).Query().Get("error_debug"))
	})

	t.Run("ShouldTreatANilErrorAsAServerError", func(t *testing.T) {
		assert.Equal(t, "server_error", NewConsentCompletionErrorURI(issuer, nil, false).Query().Get("error"))
	})

	t.Run("ShouldNotModifyTheIssuer", func(t *testing.T) {
		NewConsentCompletionErrorURI(issuer, rfc, false)

		assert.Equal(t, "https://auth.example.com/base", issuer.String())
	})
}

type testErrorConfig struct {
	issuer    *url.URL
	sendDebug bool
}

func (c *testErrorConfig) GetSendDebugMessagesToClients(_ context.Context) bool {
	return c.sendDebug
}

func (c *testErrorConfig) GetContext(_ context.Context) Context {
	return &testErrorContext{issuer: c.issuer}
}

type testErrorContext struct {
	Context

	issuer *url.URL
}

func (c *testErrorContext) IssuerURL() (*url.URL, error) {
	if c.issuer == nil {
		return nil, fmt.Errorf("failed")
	}

	return c.issuer, nil
}
