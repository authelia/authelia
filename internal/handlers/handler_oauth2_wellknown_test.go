package handlers

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/mocks"
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

	t.Run("ShouldHandleSigningError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfigNoRSAKey(t)
		config.DiscoverySignedResponseAlg = "RS256"

		setupTestOIDCProvider(t, mock, config)

		WellKnownOAuthAuthorizationServerGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusInternalServerError, mock.Ctx.Response.StatusCode())

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred signing metadata", regexpUnableToFindJSONWebKey)
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

	t.Run("ShouldHandleSigningError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfigNoRSAKey(t)
		config.DiscoverySignedResponseAlg = "RS256"

		setupTestOIDCProvider(t, mock, config)

		WellKnownOpenIDConfigurationGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusInternalServerError, mock.Ctx.Response.StatusCode())

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred signing metadata", regexpUnableToFindJSONWebKey)
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
