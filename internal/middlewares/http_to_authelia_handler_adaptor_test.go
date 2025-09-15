package middlewares_test

import (
	"io"
	"net/http"
	"testing"

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
