// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestWellKnownOAuthAuthorizationServerGET(t *testing.T) {
	t.Run("ShouldReturnMetadata", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		WellKnownOAuthAuthorizationServerGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

		metadata := map[string]any{}

		require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &metadata))

		assert.Equal(t, "https://login.example.com:8080", metadata["issuer"])
		assert.Equal(t, "https://login.example.com:8080/jwks.json", metadata["jwks_uri"])
		assert.Equal(t, "https://login.example.com:8080/api/oidc/authorization", metadata["authorization_endpoint"])
		assert.Equal(t, "https://login.example.com:8080/api/oidc/token", metadata["token_endpoint"])
		assert.Equal(t, "https://login.example.com:8080/api/oidc/introspection", metadata["introspection_endpoint"])
		assert.Equal(t, "https://login.example.com:8080/api/oidc/revocation", metadata["revocation_endpoint"])

		assert.NotContains(t, metadata, "userinfo_endpoint")
		assert.NotContains(t, metadata, "signed_metadata")
	})

	t.Run("ShouldReturnSignedMetadataWithKeyID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.DiscoverySignedResponseKeyID = testOIDCKeyID
		config.DiscoverySignedResponseAlg = oidc.SigningAlgRSAUsingSHA256

		setupTestOIDCProvider(t, mock, config)

		WellKnownOAuthAuthorizationServerGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

		metadata := map[string]any{}

		require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &metadata))

		signed, ok := metadata["signed_metadata"].(string)

		require.True(t, ok)
		assert.Len(t, strings.Split(signed, "."), 3)
	})

	t.Run("ShouldReturnSignedMetadataWithAlg", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.DiscoverySignedResponseAlg = "RS256"

		setupTestOIDCProvider(t, mock, config)

		WellKnownOAuthAuthorizationServerGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

		metadata := map[string]any{}

		require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &metadata))

		signed, ok := metadata["signed_metadata"].(string)

		require.True(t, ok)
		assert.Len(t, strings.Split(signed, "."), 3)
	})

	t.Run("ShouldReturnUnsignedMetadataWithAlgNone", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.DiscoverySignedResponseAlg = oidc.SigningAlgNone

		setupTestOIDCProvider(t, mock, config)

		WellKnownOAuthAuthorizationServerGET(mock.Ctx)

		require.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

		metadata := map[string]any{}

		require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &metadata))

		assert.NotContains(t, metadata, "signed_metadata")
	})

	t.Run("ShouldReturnSignedMetadataUsingTheConfiguredKey", func(t *testing.T) {
		testCases := []struct {
			name        string
			alg         string
			kid         string
			expectedAlg string
			expectedKID string
		}{
			{"ShouldSignWithRSA", oidc.SigningAlgRSAUsingSHA256, testOIDCKeyID, oidc.SigningAlgRSAUsingSHA256, testOIDCKeyID},
			{"ShouldSignWithECDSA", oidc.SigningAlgECDSAUsingP256AndSHA256, "ecdsa-default", oidc.SigningAlgECDSAUsingP256AndSHA256, "ecdsa-default"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mock := mocks.NewMockAutheliaCtx(t)
				defer mock.Close()

				config := newTestOIDCConfig(t)
				config.JSONWebKeys = append(config.JSONWebKeys, newTestOIDCConfigNoRSAKey(t).JSONWebKeys...)
				config.DiscoverySignedResponseAlg = tc.alg
				config.DiscoverySignedResponseKeyID = tc.kid

				setupTestOIDCProvider(t, mock, config)

				WellKnownOAuthAuthorizationServerGET(mock.Ctx)

				require.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

				metadata := map[string]any{}

				require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &metadata))

				signed, ok := metadata["signed_metadata"].(string)

				require.True(t, ok)

				parts := strings.Split(signed, ".")

				require.Len(t, parts, 3)

				raw, err := base64.RawURLEncoding.DecodeString(parts[0])

				require.NoError(t, err)

				header := map[string]any{}

				require.NoError(t, json.Unmarshal(raw, &header))

				assert.Equal(t, tc.expectedAlg, header["alg"])
				assert.Equal(t, tc.expectedKID, header["kid"])
			})
		}
	})

	t.Run("ShouldHandleSigningError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfigNoRSAKey(t)
		config.DiscoverySignedResponseAlg = "RS256"

		setupTestOIDCProvider(t, mock, config)

		WellKnownOAuthAuthorizationServerGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusInternalServerError, mock.Ctx.Response.StatusCode())

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "OAuth 2.0 Discovery Request could not be processed: error occurred building the metadata", regexpUnableToFindJSONWebKey)
	})

	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		WellKnownOAuthAuthorizationServerGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpOAuth2DiscoveryIssuerError, "missing required X-Forwarded-Host header")
	})
}

func TestWellKnownOpenIDConfigurationGET(t *testing.T) {
	t.Run("ShouldReturnMetadata", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		WellKnownOpenIDConfigurationGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

		metadata := map[string]any{}

		require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &metadata))

		assert.Equal(t, "https://login.example.com:8080", metadata["issuer"])
		assert.Equal(t, "https://login.example.com:8080/jwks.json", metadata["jwks_uri"])
		assert.Equal(t, "https://login.example.com:8080/api/oidc/userinfo", metadata["userinfo_endpoint"])
		assert.Equal(t, "https://login.example.com:8080/api/oidc/token", metadata["token_endpoint"])

		assert.NotContains(t, metadata, "signed_metadata")
	})

	t.Run("ShouldReturnSignedMetadataWithKeyID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.DiscoverySignedResponseKeyID = testOIDCKeyID
		config.DiscoverySignedResponseAlg = oidc.SigningAlgRSAUsingSHA256

		setupTestOIDCProvider(t, mock, config)

		WellKnownOpenIDConfigurationGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

		metadata := map[string]any{}

		require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &metadata))

		signed, ok := metadata["signed_metadata"].(string)

		require.True(t, ok)
		assert.Len(t, strings.Split(signed, "."), 3)
	})

	t.Run("ShouldReturnSignedMetadataWithAlg", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.DiscoverySignedResponseAlg = "RS256"

		setupTestOIDCProvider(t, mock, config)

		WellKnownOpenIDConfigurationGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

		metadata := map[string]any{}

		require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &metadata))

		signed, ok := metadata["signed_metadata"].(string)

		require.True(t, ok)
		assert.Len(t, strings.Split(signed, "."), 3)
	})

	t.Run("ShouldReturnUnsignedMetadataWithAlgNone", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.DiscoverySignedResponseAlg = oidc.SigningAlgNone

		setupTestOIDCProvider(t, mock, config)

		WellKnownOpenIDConfigurationGET(mock.Ctx)

		require.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

		metadata := map[string]any{}

		require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &metadata))

		assert.NotContains(t, metadata, "signed_metadata")
	})

	t.Run("ShouldReturnSignedMetadataUsingTheConfiguredKey", func(t *testing.T) {
		testCases := []struct {
			name        string
			alg         string
			kid         string
			expectedAlg string
			expectedKID string
		}{
			{"ShouldSignWithRSA", oidc.SigningAlgRSAUsingSHA256, testOIDCKeyID, oidc.SigningAlgRSAUsingSHA256, testOIDCKeyID},
			{"ShouldSignWithECDSA", oidc.SigningAlgECDSAUsingP256AndSHA256, "ecdsa-default", oidc.SigningAlgECDSAUsingP256AndSHA256, "ecdsa-default"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mock := mocks.NewMockAutheliaCtx(t)
				defer mock.Close()

				config := newTestOIDCConfig(t)
				config.JSONWebKeys = append(config.JSONWebKeys, newTestOIDCConfigNoRSAKey(t).JSONWebKeys...)
				config.DiscoverySignedResponseAlg = tc.alg
				config.DiscoverySignedResponseKeyID = tc.kid

				setupTestOIDCProvider(t, mock, config)

				WellKnownOpenIDConfigurationGET(mock.Ctx)

				require.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

				metadata := map[string]any{}

				require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &metadata))

				signed, ok := metadata["signed_metadata"].(string)

				require.True(t, ok)

				parts := strings.Split(signed, ".")

				require.Len(t, parts, 3)

				raw, err := base64.RawURLEncoding.DecodeString(parts[0])

				require.NoError(t, err)

				header := map[string]any{}

				require.NoError(t, json.Unmarshal(raw, &header))

				assert.Equal(t, tc.expectedAlg, header["alg"])
				assert.Equal(t, tc.expectedKID, header["kid"])
			})
		}
	})

	t.Run("ShouldHandleSigningError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfigNoRSAKey(t)
		config.DiscoverySignedResponseAlg = "RS256"

		setupTestOIDCProvider(t, mock, config)

		WellKnownOpenIDConfigurationGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusInternalServerError, mock.Ctx.Response.StatusCode())

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "OpenID Connect 1.0 Discovery Request could not be processed: error occurred building the metadata", "the openid provider metadata requires 'id_token_signing_alg_values_supported' to include 'RS256' but the id token issuer has no 'RS256' signing key")
	})

	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		WellKnownOpenIDConfigurationGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpOpenIDConnectDiscoveryIssuerError, "missing required X-Forwarded-Host header")
	})
}
