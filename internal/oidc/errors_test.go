package oidc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/stretchr/testify/assert"
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
			rootURL := &url.URL{Scheme: "https", Host: "auth.example.com"}

			strategy := &RedirectAuthorizeErrorFieldResponseStrategy{
				Config: &testErrorConfig{rootURL: rootURL, sendDebug: tc.sendDebug},
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
	rootURL   *url.URL
	sendDebug bool
}

func (c *testErrorConfig) GetSendDebugMessagesToClients(_ context.Context) bool {
	return c.sendDebug
}

func (c *testErrorConfig) GetContext(_ context.Context) Context {
	return &testErrorContext{rootURL: c.rootURL}
}

type testErrorContext struct {
	Context
	rootURL *url.URL
}

func (c *testErrorContext) RootURL() *url.URL {
	return c.rootURL
}
