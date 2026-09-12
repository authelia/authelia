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
	"strconv"
	"strings"
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

			assert.Equal(t, http.StatusFound, rw.Code)
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

func TestConsentCompletionURL(t *testing.T) {
	testCases := []struct {
		name     string
		rfc      *oauthelia2.RFC6749Error
		debug    bool
		expected url.Values
	}{
		{
			"ShouldEncodeAllPopulatedFields",
			&oauthelia2.RFC6749Error{
				ErrorField:       "invalid_request",
				DescriptionField: "The request is otherwise malformed.",
				CodeField:        http.StatusBadRequest,
				HintField:        "Could not perform consent.",
			},
			false,
			url.Values{
				"error":             []string{"invalid_request"},
				"error_description": []string{"The request is otherwise malformed."},
				"error_status_code": []string{"400"},
				"error_hint":        []string{"Could not perform consent."},
			},
		},
		{
			"ShouldOmitEmptyFields",
			&oauthelia2.RFC6749Error{ErrorField: "server_error"},
			false,
			url.Values{"error": []string{"server_error"}},
		},
		{
			"ShouldIncludeDebugWhenEnabled",
			&oauthelia2.RFC6749Error{ErrorField: "server_error", DebugField: "internal debug info"},
			true,
			url.Values{"error": []string{"server_error"}, "error_debug": []string{"internal debug info"}},
		},
		{
			"ShouldExcludeDebugWhenDisabled",
			&oauthelia2.RFC6749Error{ErrorField: "server_error", DebugField: "internal debug info"},
			false,
			url.Values{"error": []string{"server_error"}},
		},
		{
			"ShouldDefaultToServerErrorWhenNil",
			nil,
			false,
			url.Values{
				"error":             []string{oauthelia2.ErrServerError.ErrorField},
				"error_description": []string{oauthelia2.ErrServerError.DescriptionField},
				"error_status_code": []string{strconv.Itoa(oauthelia2.ErrServerError.CodeField)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issuer := &url.URL{Scheme: "https", Host: "auth.example.com"}

			actual := ConsentCompletionURL(issuer, tc.rfc, tc.debug)

			assert.True(t, strings.HasPrefix(actual.String(), "https://auth.example.com"+FrontendEndpointPathConsentCompletion+"?"), "unexpected URL: %s", actual)
			assert.Equal(t, tc.expected, actual.Query())

			assert.Equal(t, "", issuer.Path, "the issuer must not be mutated")
		})
	}

	t.Run("ShouldRespectIssuerBasePath", func(t *testing.T) {
		issuer := &url.URL{Scheme: "https", Host: "auth.example.com", Path: "/auth"}

		actual := ConsentCompletionURL(issuer, oauthelia2.ErrInvalidRequest, false)

		assert.Equal(t, "/auth"+FrontendEndpointPathConsentCompletion, actual.Path)
		assert.Equal(t, "/auth", issuer.Path, "the issuer must not be mutated")
	})
}
