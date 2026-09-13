// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"authelia.com/provider/oauth2/token/jose"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestProviderUserInfoSigned(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	other, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	sign := func(t *testing.T, signer *rsa.PrivateKey, alg jose.SignatureAlgorithm, kid string, claims map[string]any) string {
		t.Helper()

		s, err := jose.NewSigner(jose.SigningKey{Algorithm: alg, Key: jose.JSONWebKey{Key: signer, KeyID: kid, Algorithm: string(alg)}}, nil)
		require.NoError(t, err)

		payload, err := json.Marshal(claims)
		require.NoError(t, err)

		object, err := s.Sign(payload)
		require.NoError(t, err)

		raw, err := object.CompactSerialize()
		require.NoError(t, err)

		return raw
	}

	testCases := []struct {
		Name        string
		ContentType string
		Body        func(t *testing.T) string
		Expected    IdentityClaims
		Error       string
	}{
		{
			Name:        "ShouldAdoptClaimsOfValidSignedResponse",
			ContentType: "application/jwt; charset=utf-8",
			Body: func(t *testing.T) string {
				return sign(t, key, jose.RS256, "kid1", map[string]any{"iss": "https://op.example.com", "aud": "client", "sub": "abc123", "preferred_username": "john", "name": "John Smith", "email": "john@example.com"})
			},
			Expected: IdentityClaims{Subject: "abc123", PreferredUsername: "john", Name: "John Smith", Email: "john@example.com"},
		},
		{
			Name:        "ShouldAcceptSignedResponseWithoutIssuerOrAudience",
			ContentType: mimeApplicationJWT,
			Body: func(t *testing.T) string {
				return sign(t, key, jose.RS256, "kid1", map[string]any{"sub": "abc123", "name": "John Smith"})
			},
			Expected: IdentityClaims{Subject: "abc123", Name: "John Smith"},
		},
		{
			Name:        "ShouldAcceptSignedResponseWithAudienceArray",
			ContentType: mimeApplicationJWT,
			Body: func(t *testing.T) string {
				return sign(t, key, jose.RS256, "kid1", map[string]any{"aud": []string{"other", "client"}, "sub": "abc123"})
			},
			Expected: IdentityClaims{Subject: "abc123"},
		},
		{
			Name:        "ShouldRaiseErrorOnUnsignedResponse",
			ContentType: mimeApplicationJSON,
			Body: func(t *testing.T) string {
				return `{"sub":"abc123","name":"Mallory"}`
			},
			Expected: IdentityClaims{Subject: "abc123"},
			Error:    "error requesting the userinfo: the userinfo response is invalid: the content type 'application/json' is not 'application/jwt'",
		},
		{
			Name:        "ShouldRaiseErrorOnUnsecuredResponse",
			ContentType: mimeApplicationJWT,
			Body: func(t *testing.T) string {
				return `eyJhbGciOiJub25lIn0.eyJzdWIiOiJhYmMxMjMiLCJuYW1lIjoiTWFsbG9yeSJ9.`
			},
			Expected: IdentityClaims{Subject: "abc123"},
			Error:    "error requesting the userinfo: the userinfo response is invalid: the signature could not be verified",
		},
		{
			Name:        "ShouldRaiseErrorOnSignatureByAnotherKey",
			ContentType: mimeApplicationJWT,
			Body: func(t *testing.T) string {
				return sign(t, other, jose.RS256, "kid1", map[string]any{"sub": "abc123", "name": "Mallory"})
			},
			Expected: IdentityClaims{Subject: "abc123"},
			Error:    "error requesting the userinfo: the userinfo response is invalid: the signature could not be verified",
		},
		{
			Name:        "ShouldRaiseErrorOnAnotherAlgorithm",
			ContentType: mimeApplicationJWT,
			Body: func(t *testing.T) string {
				return sign(t, key, jose.RS512, "kid1", map[string]any{"sub": "abc123", "name": "Mallory"})
			},
			Expected: IdentityClaims{Subject: "abc123"},
			Error:    "error requesting the userinfo: the userinfo response is invalid: the signature could not be verified",
		},
		{
			Name:        "ShouldRaiseErrorOnUnknownKeyID",
			ContentType: mimeApplicationJWT,
			Body: func(t *testing.T) string {
				return sign(t, key, jose.RS256, "kid-absent", map[string]any{"sub": "abc123", "name": "Mallory"})
			},
			Expected: IdentityClaims{Subject: "abc123"},
			Error:    "error requesting the userinfo: the userinfo response is invalid: no key in the json web key set matched the signature",
		},
		{
			Name:        "ShouldRaiseErrorOnIssuerMismatch",
			ContentType: mimeApplicationJWT,
			Body: func(t *testing.T) string {
				return sign(t, key, jose.RS256, "kid1", map[string]any{"iss": "https://evil.example.org", "sub": "abc123", "name": "Mallory"})
			},
			Expected: IdentityClaims{Subject: "abc123"},
			Error:    "error requesting the userinfo: the userinfo response is invalid: the 'iss' claim does not match the configured issuer",
		},
		{
			Name:        "ShouldRaiseErrorOnAudienceForAnotherClient",
			ContentType: mimeApplicationJWT,
			Body: func(t *testing.T) string {
				return sign(t, key, jose.RS256, "kid1", map[string]any{"aud": "other", "sub": "abc123", "name": "Mallory"})
			},
			Expected: IdentityClaims{Subject: "abc123"},
			Error:    "error requesting the userinfo: the userinfo response is invalid: the 'aud' claim does not contain the client id 'client'",
		},
		{
			Name:        "ShouldRaiseErrorOnSubjectMismatch",
			ContentType: mimeApplicationJWT,
			Body: func(t *testing.T) string {
				return sign(t, key, jose.RS256, "kid1", map[string]any{"sub": "abc123invalid", "name": "Mallory"})
			},
			Expected: IdentityClaims{Subject: "abc123"},
			Error:    "error requesting the userinfo: the userinfo response 'sub' claim does not match the id token 'sub' claim",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			body := tc.Body(t)

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, mimeApplicationJWT, r.Header.Get(headerAccept))

				rw.Header().Set(headerContentType, tc.ContentType)
				rw.WriteHeader(http.StatusOK)

				_, _ = rw.Write([]byte(body))
			}))

			defer server.Close()

			provider := newTestProvider("", "client_secret_basic")
			provider.endpointUserInfo = server.URL + "/userinfo"
			provider.algUserInfo = "RS256"
			provider.jwksURI = "https://op.example.com/jwks.json"
			provider.keys = &stubKeySet{jwks: &jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{KeyID: "kid1", Key: key.Public(), Algorithm: "RS256", Use: "sig"}}}}

			claims := IdentityClaims{Subject: "abc123"}

			err := provider.UserInfo(context.Background(), "at", &claims)

			if tc.Error != "" {
				require.EqualError(t, err, tc.Error)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.Expected, claims)
		})
	}
}

func TestProviderResolveShouldValidateTheUserInfoSigningAlg(t *testing.T) {
	testCases := []struct {
		Name     string
		Metadata string
		Alg      string
		Error    string
	}{
		{Name: "ShouldAcceptAdvertisedAlg", Metadata: `,"userinfo_signing_alg_values_supported":["RS256","ES256"]`, Alg: "ES256"},
		{Name: "ShouldAcceptOmittedAlgs", Alg: "ES256"},
		{Name: "ShouldNotCheckWhenUnsigned", Metadata: `,"userinfo_signing_alg_values_supported":["RS256"]`},
		{
			Name:     "ShouldRaiseErrorOnUnadvertisedAlg",
			Metadata: `,"userinfo_signing_alg_values_supported":["RS256"]`,
			Alg:      "ES256",
			Error:    "error resolving provider 'example': error validating the discovery document: the discovery document does not advertise support for the configured value: the value 'ES256' is not included in the 'userinfo_signing_alg_values_supported' values 'RS256'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var server *httptest.Server

			server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set(headerContentType, mimeApplicationJSON)

				_, _ = rw.Write([]byte(`{"issuer":"` + server.URL + `","authorization_endpoint":"` + server.URL + `/authorize","token_endpoint":"` + server.URL + `/token","jwks_uri":"` + server.URL + `/jwks.json"` + tc.Metadata + `}`))
			}))

			defer server.Close()

			provider, ok := getOpenIDConnectProvider(NewProviders(&schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "example", Name: "Example", Issuer: server.URL,
						ClientID: "client", ClientSecret: "secret",
						PKCE:                      schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
						UserInfoSignedResponseAlg: tc.Alg,
					},
				},
			}, nil))

			require.True(t, ok)

			err := provider.Resolve(context.Background())

			if tc.Error != "" {
				require.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestProviderUserInfo(t *testing.T) {
	testCases := []struct {
		Name        string
		Status      int
		ContentType string
		Body        string
		AccessToken string
		Have        IdentityClaims
		Expected    IdentityClaims
		Error       string
	}{
		{
			Name:        "ShouldAdoptAbsentClaimsWhenTheSubjectMatches",
			Status:      http.StatusOK,
			ContentType: "application/json; charset=utf-8",
			Body:        `{"sub":"abc123","preferred_username":"john","name":"John Smith","email":"john@example.com"}`,
			AccessToken: "at",
			Have:        IdentityClaims{Issuer: "https://op.example.com", Subject: "abc123"},
			Expected:    IdentityClaims{Issuer: "https://op.example.com", Subject: "abc123", PreferredUsername: "john", Name: "John Smith", Email: "john@example.com"},
		},
		{
			Name:        "ShouldPreferUserInfoClaimsOverTheIDToken",
			Status:      http.StatusOK,
			ContentType: mimeApplicationJSON,
			Body:        `{"sub":"abc123","preferred_username":"other","name":"Other","email":"other@example.com"}`,
			AccessToken: "at",
			Have:        IdentityClaims{Subject: "abc123", PreferredUsername: "john", Name: "John Smith", Email: "john@example.com"},
			Expected:    IdentityClaims{Subject: "abc123", PreferredUsername: "other", Name: "Other", Email: "other@example.com"},
		},
		{
			Name:        "ShouldKeepIDTokenClaimsTheUserInfoResponseOmitsOrLeavesEmpty",
			Status:      http.StatusOK,
			ContentType: mimeApplicationJSON,
			Body:        `{"sub":"abc123","name":"Other","email":""}`,
			AccessToken: "at",
			Have:        IdentityClaims{Subject: "abc123", PreferredUsername: "john", Name: "John Smith", Email: "john@example.com"},
			Expected:    IdentityClaims{Subject: "abc123", PreferredUsername: "john", Name: "Other", Email: "john@example.com"},
		},
		{
			Name:        "ShouldIgnoreNonStringClaims",
			Status:      http.StatusOK,
			ContentType: mimeApplicationJSON,
			Body:        `{"sub":"abc123","name":12345}`,
			AccessToken: "at",
			Have:        IdentityClaims{Subject: "abc123"},
			Expected:    IdentityClaims{Subject: "abc123"},
		},
		{
			Name:        "ShouldRaiseErrorOnSubjectMismatch",
			Status:      http.StatusOK,
			ContentType: mimeApplicationJSON,
			Body:        `{"sub":"abc123invalid","name":"Mallory"}`,
			AccessToken: "at",
			Have:        IdentityClaims{Subject: "abc123"},
			Expected:    IdentityClaims{Subject: "abc123"},
			Error:       "error requesting the userinfo: the userinfo response 'sub' claim does not match the id token 'sub' claim",
		},
		{
			Name:        "ShouldRaiseErrorOnAbsentSubject",
			Status:      http.StatusOK,
			ContentType: mimeApplicationJSON,
			Body:        `{"name":"John Smith"}`,
			AccessToken: "at",
			Have:        IdentityClaims{Subject: "abc123"},
			Expected:    IdentityClaims{Subject: "abc123"},
			Error:       "error requesting the userinfo: the userinfo response is invalid: the 'sub' claim is required but it is absent",
		},
		{
			Name:        "ShouldRaiseErrorOnSignedResponse",
			Status:      http.StatusOK,
			ContentType: "application/jwt",
			Body:        `eyJhbGciOiJub25lIn0.eyJzdWIiOiJhYmMxMjMifQ.`,
			AccessToken: "at",
			Have:        IdentityClaims{Subject: "abc123"},
			Expected:    IdentityClaims{Subject: "abc123"},
			Error:       "error requesting the userinfo: the userinfo response is invalid: the content type 'application/jwt' is not 'application/json'",
		},
		{
			Name:        "ShouldRaiseErrorOnMalformedResponse",
			Status:      http.StatusOK,
			ContentType: mimeApplicationJSON,
			Body:        `{"sub":`,
			AccessToken: "at",
			Have:        IdentityClaims{Subject: "abc123"},
			Expected:    IdentityClaims{Subject: "abc123"},
			Error:       "error requesting the userinfo: the userinfo response is invalid: unexpected end of JSON input",
		},
		{
			Name:        "ShouldRaiseErrorOnErrorStatus",
			Status:      http.StatusUnauthorized,
			ContentType: mimeApplicationJSON,
			Body:        `{"error":"invalid_token"}`,
			AccessToken: "at",
			Have:        IdentityClaims{Subject: "abc123"},
			Expected:    IdentityClaims{Subject: "abc123"},
			Error:       "error requesting the userinfo: the userinfo response is invalid: the userinfo endpoint returned status code 401",
		},
		{
			Name:     "ShouldRaiseErrorOnAbsentAccessToken",
			Have:     IdentityClaims{Subject: "abc123"},
			Expected: IdentityClaims{Subject: "abc123"},
			Error:    "error requesting the userinfo: the token response did not contain an access token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "Bearer "+tc.AccessToken, r.Header.Get(headerAuthorization))
				assert.Empty(t, r.URL.RawQuery, "the access token must never be sent in the query")

				rw.Header().Set(headerContentType, tc.ContentType)
				rw.WriteHeader(tc.Status)

				_, _ = rw.Write([]byte(tc.Body))
			}))

			defer server.Close()

			provider := newTestProvider("", "client_secret_basic")
			provider.endpointUserInfo = server.URL + "/userinfo"

			claims := tc.Have

			err := provider.UserInfo(context.Background(), tc.AccessToken, &claims)

			if tc.Error != "" {
				require.EqualError(t, err, tc.Error)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.Expected, claims)
		})
	}
}

func TestProviderUserInfoShouldDoNothingWithoutEndpoint(t *testing.T) {
	provider := newTestProvider("", "client_secret_basic")

	claims := IdentityClaims{Subject: "abc123"}

	require.NoError(t, provider.UserInfo(context.Background(), "", &claims))

	assert.Equal(t, IdentityClaims{Subject: "abc123"}, claims)
}
