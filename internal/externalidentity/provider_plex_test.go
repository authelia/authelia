// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestPlexProviderAuthorizationRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v2/pins", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("strong"))
		assert.Equal(t, mimeApplicationJSON, r.Header.Get(headerAccept))
		assertPlexDeviceHeaders(t, r, "fr-CA")
		assert.Empty(t, r.Header.Get(headerPlexToken))

		rw.Header().Set(headerContentType, mimeApplicationJSON)
		rw.WriteHeader(http.StatusCreated)

		_ = json.NewEncoder(rw).Encode(map[string]any{"id": 1234, "code": "the-pin-code", "authToken": nil})
	}))

	defer server.Close()

	provider := newTestPlexProvider(server.URL)

	request, err := provider.AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/api/firstfactor/external-identity/plex/callback", Language: "fr-CA"})
	require.NoError(t, err)

	assert.Len(t, request.State, 43)
	assert.Empty(t, request.Nonce)
	assert.Empty(t, request.CodeVerifier)
	assert.Equal(t, "1234:the-pin-code", request.Handle)

	uri, err := url.Parse(request.URL)
	require.NoError(t, err)

	assert.Equal(t, "https", uri.Scheme)
	assert.Equal(t, "app.plex.tv", uri.Host)
	assert.Equal(t, "/auth", uri.Path)
	assert.Empty(t, uri.RawQuery)

	fragment, err := url.ParseQuery(strings.TrimPrefix(uri.EscapedFragment(), "?"))
	require.NoError(t, err)

	assert.Equal(t, "authelia-client", fragment.Get("clientID"))
	assert.Equal(t, "the-pin-code", fragment.Get("code"))
	assert.Equal(t, "Authelia", fragment.Get("context[device][product]"))
	assert.Equal(t, utils.Version(), fragment.Get("context[device][version]"))
	assert.Equal(t, "Authelia (https://auth.example.com)", fragment.Get("context[device][deviceName]"))

	forward, err := url.Parse(fragment.Get("forwardUrl"))
	require.NoError(t, err)

	assert.Equal(t, "https://auth.example.com/api/firstfactor/external-identity/plex/callback", (&url.URL{Scheme: forward.Scheme, Host: forward.Host, Path: forward.Path}).String())
	assert.Equal(t, request.State, forward.Query().Get("state"))
	assert.Equal(t, "1234", forward.Query().Get("code"))
}

func TestPlexProviderAuthorizationRequestShouldOmitAnAbsentLanguage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, ok := r.Header[headerPlexLanguage]

		assert.False(t, ok)

		rw.Header().Set(headerContentType, mimeApplicationJSON)
		rw.WriteHeader(http.StatusCreated)

		_ = json.NewEncoder(rw).Encode(map[string]any{"id": 1234, "code": "the-pin-code"})
	}))

	defer server.Close()

	_, err := newTestPlexProvider(server.URL).AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/cb"})
	require.NoError(t, err)
}

func TestPlexProviderAuthorizationRequestShouldRaiseErrors(t *testing.T) {
	testCases := []struct {
		Name   string
		Status int
		Body   any
		Error  string
	}{
		{
			Name:   "ShouldRaiseErrorOnErrorStatus",
			Status: http.StatusBadRequest,
			Body:   map[string]any{},
			Error:  "error creating the plex pin: the plex response is invalid: the endpoint returned status code 400",
		},
		{
			Name:   "ShouldRaiseErrorOnAbsentCode",
			Status: http.StatusCreated,
			Body:   map[string]any{"id": 1234},
			Error:  "error creating the plex pin: the plex response is invalid: the 'id' and 'code' are required but one or both are absent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set(headerContentType, mimeApplicationJSON)
				rw.WriteHeader(tc.Status)

				_ = json.NewEncoder(rw).Encode(tc.Body)
			}))

			defer server.Close()

			request, err := newTestPlexProvider(server.URL).AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/cb"})

			assert.Nil(t, request)
			require.EqualError(t, err, tc.Error)
		})
	}
}

func TestPlexProviderComplete(t *testing.T) {
	testCases := []struct {
		Name       string
		Code       string
		Handle     string
		PIN        any
		UserStatus int
		User       any
		Expected   *IdentityClaims
		Error      string
	}{
		{
			Name:       "ShouldReturnTheUser",
			Code:       "1234",
			Handle:     "1234:the-pin-code",
			PIN:        map[string]any{"id": 1234, "code": "the-pin-code", "authToken": "the-token"},
			UserStatus: http.StatusOK,
			User:       map[string]any{"id": 1, "uuid": "a1b2c3d4e5f6", "username": "nelly", "title": "Nelly", "email": "nelly@plex.tv"},
			Expected:   &IdentityClaims{Issuer: "https://plex.tv", Subject: "a1b2c3d4e5f6", PreferredUsername: "nelly", Name: "Nelly", Email: "nelly@plex.tv"},
		},
		{
			Name:       "ShouldUseTheUsernameWithoutATitle",
			Code:       "1234",
			Handle:     "1234:the-pin-code",
			PIN:        map[string]any{"id": 1234, "code": "the-pin-code", "authToken": "the-token"},
			UserStatus: http.StatusOK,
			User:       map[string]any{"uuid": "a1b2c3d4e5f6", "username": "nelly"},
			Expected:   &IdentityClaims{Issuer: "https://plex.tv", Subject: "a1b2c3d4e5f6", PreferredUsername: "nelly", Name: "nelly"},
		},
		{
			Name:   "ShouldRaiseErrorOnMismatchedPIN",
			Code:   "9999",
			Handle: "1234:the-pin-code",
			Error:  "error retrieving the plex pin: the plex pin returned does not match the plex pin requested",
		},
		{
			Name:   "ShouldRaiseErrorOnAbsentHandle",
			Code:   "1234",
			Handle: "",
			Error:  "error retrieving the plex pin: the plex pin returned does not match the plex pin requested",
		},
		{
			Name:   "ShouldRaiseErrorOnUnauthorizedPIN",
			Code:   "1234",
			Handle: "1234:the-pin-code",
			PIN:    map[string]any{"id": 1234, "code": "the-pin-code", "authToken": nil},
			Error:  "error retrieving the plex pin: the plex pin has not been authorized",
		},
		{
			Name:       "ShouldRaiseErrorOnAbsentUUID",
			Code:       "1234",
			Handle:     "1234:the-pin-code",
			PIN:        map[string]any{"id": 1234, "code": "the-pin-code", "authToken": "the-token"},
			UserStatus: http.StatusOK,
			User:       map[string]any{"username": "nelly"},
			Error:      "error requesting the plex user: the plex response is invalid: the 'uuid' is required but it is absent",
		},
		{
			Name:       "ShouldRaiseErrorOnUserErrorStatus",
			Code:       "1234",
			Handle:     "1234:the-pin-code",
			PIN:        map[string]any{"id": 1234, "code": "the-pin-code", "authToken": "the-token"},
			UserStatus: http.StatusUnauthorized,
			User:       map[string]any{"error": "unauthorized"},
			Error:      "error requesting the plex user: the plex response is invalid: the endpoint returned status code 401",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mux := http.NewServeMux()

			mux.HandleFunc("/api/v2/pins/1234", func(rw http.ResponseWriter, r *http.Request) {
				if tc.PIN == nil {
					t.Errorf("the pin must not be requested")
				}

				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "the-pin-code", r.URL.Query().Get("code"))
				assertPlexDeviceHeaders(t, r, "de")
				assert.Empty(t, r.Header.Get(headerPlexToken))

				rw.Header().Set(headerContentType, mimeApplicationJSON)

				_ = json.NewEncoder(rw).Encode(tc.PIN)
			})

			mux.HandleFunc("/api/v2/user", func(rw http.ResponseWriter, r *http.Request) {
				if tc.User == nil {
					t.Errorf("the user must not be requested")
				}

				assert.Equal(t, "the-token", r.Header.Get(headerPlexToken))
				assertPlexDeviceHeaders(t, r, "de")

				rw.Header().Set(headerContentType, mimeApplicationJSON)
				rw.WriteHeader(tc.UserStatus)

				_ = json.NewEncoder(rw).Encode(tc.User)
			})

			server := httptest.NewServer(mux)

			defer server.Close()

			claims, err := newTestPlexProvider(server.URL).Complete(context.Background(), CompletionRequest{
				Code: tc.Code, Handle: tc.Handle, RedirectURI: "https://auth.example.com/api/firstfactor/external-identity/plex/callback", Language: "de", Now: time.Now(),
			})

			if tc.Error != "" {
				assert.Nil(t, claims)
				require.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, claims)
		})
	}
}

func TestNewPlexDevice(t *testing.T) {
	testCases := []struct {
		Name        string
		RedirectURI string
		Expected    string
	}{
		{"ShouldNameTheDeviceAfterTheOrigin", "https://auth.example.com/api/firstfactor/external-identity/plex/callback", "Authelia (https://auth.example.com)"},
		{"ShouldIncludeThePort", "https://auth.example.com:8443/authelia/api/firstfactor/external-identity/plex/callback", "Authelia (https://auth.example.com:8443)"},
		{"ShouldFallBackWithoutAnOrigin", "/api/firstfactor/external-identity/plex/callback", "Authelia"},
		{"ShouldFallBackOnAnInvalidURI", "https://auth.example.com/%zz", "Authelia"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			device := newPlexDevice(tc.RedirectURI, "en")

			assert.Equal(t, tc.Expected, device.name)
			assert.Equal(t, utils.Version(), device.version)
			assert.Equal(t, "en", device.language)
		})
	}
}

func TestPlexProviderMetadata(t *testing.T) {
	provider := newPlexProvider(&schema.AuthenticationBackendExternalIdentityProvider{
		ID: "plex", Name: "Plex", ClientID: "authelia-client",
		AuthenticationMethodsReference: schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd"}},
	}, newProviderClient(nil))

	assert.Equal(t, "plex", provider.ID())
	assert.Equal(t, "Plex", provider.Name())
	assert.Equal(t, ProviderTypePlex, provider.Type())
	assert.Equal(t, "https://plex.tv", provider.Issuer())
	assert.Equal(t, ResponseModeQuery, provider.ResponseMode())
	assert.Equal(t, []string{"pwd"}, provider.AuthenticationMethodsReference(nil))
	assert.Equal(t, []string{"pwd", "kba"}, newTestPlexProvider("https://plex.tv").AuthenticationMethodsReference(nil))
}

func assertPlexDeviceHeaders(t *testing.T, r *http.Request, language string) {
	t.Helper()

	assert.Equal(t, "Authelia", r.Header.Get(headerPlexProduct))
	assert.Equal(t, utils.Version(), r.Header.Get(headerPlexVersion))
	assert.Equal(t, "authelia-client", r.Header.Get(headerPlexClientIdentifier))
	assert.Equal(t, "Authelia (https://auth.example.com)", r.Header.Get(headerPlexDeviceName))
	assert.Equal(t, language, r.Header.Get(headerPlexLanguage))
}

func newTestPlexProvider(base string) *PlexProvider {
	provider := newPlexProvider(&schema.AuthenticationBackendExternalIdentityProvider{
		ID: "plex", Name: "Plex", ClientID: "authelia-client",
	}, newProviderClient(nil))

	provider.client.RetryMax = 0
	provider.pinsEndpoint = base + "/api/v2/pins"
	provider.userEndpoint = base + "/api/v2/user"

	return provider
}
