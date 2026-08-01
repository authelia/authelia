package middlewares_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	. "github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestNewHTTPToAutheliaHandlerAdaptor(t *testing.T) {
	t.Run("ShouldHandle", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			_, _ = rw.Write([]byte("Hello World"))
			rw.WriteHeader(http.StatusOK)
			rw.Header().Set(fasthttp.HeaderContentType, "application/json")
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)
	})

	t.Run("ShouldHandleBadRequestURI", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			_, _ = rw.Write([]byte("Hello World"))
			rw.WriteHeader(http.StatusOK)
			rw.Header().Set(fasthttp.HeaderContentType, "application/json")
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.SetRequestURI("!@&*#(^TY!@&*#!^Y@$")

		handler(mock.Ctx)
	})

	t.Run("ShouldHandleDefaultStatusCode", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			_, _ = rw.Write([]byte("Hello World"))
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)
	})

	t.Run("ShouldHandleRequest", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			data, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			require.NoError(t, err)

			_, _ = rw.Write(append([]byte("Hello World"), data...))
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)
	})

	t.Run("ShouldPreserveRepeatedHeaders", func(t *testing.T) {
		var values []string

		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			values = r.Header.Values("DPoP")
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.Header.Add("DPoP", "proof-one")
		mock.Ctx.Request.Header.Add("DPoP", "proof-two")

		handler(mock.Ctx)

		assert.Equal(t, []string{"proof-one", "proof-two"}, values)
	})

	t.Run("ShouldPreserveFirstStatusCode", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusCreated)
			rw.WriteHeader(http.StatusInternalServerError)
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)

		assert.Equal(t, http.StatusCreated, mock.Ctx.Response.StatusCode())
	})

	t.Run("ShouldSetProtocolFromRequest", func(t *testing.T) {
		testCases := []struct {
			name     string
			protocol string
			expected string
			major    int
			minor    int
		}{
			{"HTTP11", "HTTP/1.1", "HTTP/1.1", 1, 1},
			{"HTTP10", "HTTP/1.0", "HTTP/1.0", 1, 0},
			{"Invalid", "NOTAPROTOCOL", "HTTP/1.1", 1, 1},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var r *http.Request

				handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
					r = req
				})

				mock := mocks.NewMockAutheliaCtx(t)

				defer mock.Close()

				mock.Ctx.Request.Header.SetProtocol(tc.protocol)

				handler(mock.Ctx)

				require.NotNil(t, r)
				assert.Equal(t, tc.expected, r.Proto)
				assert.Equal(t, tc.major, r.ProtoMajor)
				assert.Equal(t, tc.minor, r.ProtoMinor)
			})
		}
	})

	t.Run("ShouldHandleRequestWithData", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			data, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			require.NoError(t, err)

			_, _ = rw.Write(append([]byte("Hello World"), data...))
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.SetBodyString("example")

		handler(mock.Ctx)
	})
}
