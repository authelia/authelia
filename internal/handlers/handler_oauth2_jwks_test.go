package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestOAuth2JSONWebKeySetGET(t *testing.T) {
	t.Run("ShouldReturnKeySet", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		OAuth2JSONWebKeySetGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
		assert.Equal(t, "application/json; charset=utf-8", string(mock.Ctx.Response.Header.ContentType()))

		out := struct {
			Keys []map[string]any `json:"keys"`
		}{}

		require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &out))
		require.Len(t, out.Keys, 1)

		assert.Equal(t, testOIDCKeyID, out.Keys[0]["kid"])
		assert.Equal(t, "RS256", out.Keys[0]["alg"])
		assert.Equal(t, "sig", out.Keys[0]["use"])

		assert.NotContains(t, out.Keys[0], "d")
	})

	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		OAuth2JSONWebKeySetGET(mock.Ctx)

		assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpJSONWebKeySetIssuerError, "missing required X-Forwarded-Host header")
	})
}
