package handlers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestWebAuthnWellKnownGET(t *testing.T) {
	testCases := []struct {
		name                string
		setup               func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expectedStatus      int
		expectedBody        string
		expectedContentType string
	}{
		{
			"ShouldReturnBadRequestWhenOriginCannotBeDetermined",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				t.Helper()

				mock.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)
			},
			fasthttp.StatusBadRequest,
			"",
			"",
		},
		{
			"ShouldReturnNotFoundWhenNoRelatedOriginConfigured",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				t.Helper()

				mock.Ctx.Configuration.WebAuthn.RelatedOrigins = nil
			},
			fasthttp.StatusNotFound,
			"",
			"",
		},
		{
			"ShouldReturnNotFoundWhenRelatedOriginsEmptyMap",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				t.Helper()

				mock.Ctx.Configuration.WebAuthn.RelatedOrigins = map[string]schema.WebAuthnRelatedOrigin{}
			},
			fasthttp.StatusNotFound,
			"",
			"",
		},
		{
			"ShouldReturnNotFoundWhenOriginNotInConfig",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				t.Helper()

				mock.Ctx.Configuration.WebAuthn.RelatedOrigins = map[string]schema.WebAuthnRelatedOrigin{
					"other.com": {
						Origins: []*url.URL{
							{Scheme: "http", Host: "other.com"},
						},
					},
				}
			},
			fasthttp.StatusNotFound,
			"",
			"",
		},
		{
			"ShouldReturnNotFoundWhenOriginSchemeDiffers",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				t.Helper()

				mock.Ctx.Configuration.WebAuthn.RelatedOrigins = map[string]schema.WebAuthnRelatedOrigin{
					"example.com": {
						Origins: []*url.URL{
							{Scheme: "https", Host: "example.com"},
						},
					},
				}
			},
			fasthttp.StatusNotFound,
			"",
			"",
		},
		{
			"ShouldReturnOriginsWhenMatched",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				t.Helper()

				mock.Ctx.Configuration.WebAuthn.RelatedOrigins = map[string]schema.WebAuthnRelatedOrigin{
					"login.example.com": {
						Origins: []*url.URL{
							{Scheme: "https", Host: "login.example.com"},
							{Scheme: "https", Host: "auth.example.com"},
						},
					},
				}
			},
			fasthttp.StatusOK,
			"{\"origins\":[\"https://login.example.com\",\"https://auth.example.com\"]}\n",
			"application/json; charset=utf-8",
		},
		{
			"ShouldReturnSingleOriginWhenMatched",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				t.Helper()

				mock.Ctx.Configuration.WebAuthn.RelatedOrigins = map[string]schema.WebAuthnRelatedOrigin{
					"example.com": {
						Origins: []*url.URL{
							{Scheme: "https", Host: "login.example.com"},
						},
					},
				}
			},
			fasthttp.StatusOK,
			"{\"origins\":[\"https://login.example.com\"]}\n",
			"application/json; charset=utf-8",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			WebAuthnWellKnownGET(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())

			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, string(mock.Ctx.Response.Body()))
			}

			if tc.expectedContentType != "" {
				assert.Equal(t, tc.expectedContentType, string(mock.Ctx.Response.Header.ContentType()))
			}
		})
	}
}
