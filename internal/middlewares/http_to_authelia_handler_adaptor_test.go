package middlewares_test

import (
	"crypto/tls"
	"io"
	"net"
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

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
		assert.Equal(t, "application/json", string(mock.Ctx.Response.Header.ContentType()))
		assert.Equal(t, "Hello World", string(mock.Ctx.Response.Body()))
	})

	t.Run("ShouldHandleBadRequestURI", func(t *testing.T) {
		invoked := false

		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			invoked = true

			_, _ = rw.Write([]byte("Hello World"))
			rw.WriteHeader(http.StatusOK)
			rw.Header().Set(fasthttp.HeaderContentType, "application/json")
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.SetRequestURI("!@&*#(^TY!@&*#!^Y@$")

		handler(mock.Ctx)

		assert.False(t, invoked)
		assert.Equal(t, fasthttp.StatusInternalServerError, mock.Ctx.Response.StatusCode())
		assert.Equal(t, "Internal Server Error", string(mock.Ctx.Response.Body()))
		mock.AssertLastLogMessage(t, `Cannot parse requestURI "!@&*#(^TY!@&*#!^Y@$": parse "!@&*#(^TY!@&*#!^Y@$": invalid URI for request`, "")
	})

	t.Run("ShouldHandleDefaultStatusCode", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			_, _ = rw.Write([]byte("Hello World"))
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
		assert.Equal(t, "Hello World", string(mock.Ctx.Response.Body()))
	})

	t.Run("ShouldHandleExplicitStatusCodeWithoutBody", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusNoContent)
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusNoContent, mock.Ctx.Response.StatusCode())
		assert.Empty(t, mock.Ctx.Response.Body())
	})

	t.Run("ShouldHandleRequest", func(t *testing.T) {
		var request *http.Request

		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			request = r

			data, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			require.NoError(t, err)

			_, _ = rw.Write(append([]byte("Hello World"), data...))
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)

		require.NotNil(t, request)

		assert.Equal(t, int64(0), request.ContentLength)
		assert.Equal(t, "Hello World", string(mock.Ctx.Response.Body()))
	})

	t.Run("ShouldMapRequest", func(t *testing.T) {
		var request *http.Request

		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			request = r
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.Header.SetMethod(fasthttp.MethodPost)
		mock.Ctx.Request.Header.SetHost("auth.example.com")
		mock.Ctx.Request.SetRequestURI("/api/oidc/token?example=abc")
		mock.Ctx.Request.SetBodyString("grant_type=authorization_code")
		mock.Ctx.SetRemoteAddr(&net.TCPAddr{IP: net.IPv4(192, 168, 1, 5), Port: 41234})

		handler(mock.Ctx)

		require.NotNil(t, request)

		assert.Equal(t, fasthttp.MethodPost, request.Method)
		assert.Equal(t, "auth.example.com", request.Host)
		assert.Equal(t, "/api/oidc/token?example=abc", request.RequestURI)
		assert.Equal(t, "192.168.1.5:41234", request.RemoteAddr)
		assert.Equal(t, int64(29), request.ContentLength)
		assert.Nil(t, request.TLS)

		require.NotNil(t, request.URL)

		assert.Equal(t, "/api/oidc/token", request.URL.Path)
		assert.Equal(t, "example=abc", request.URL.RawQuery)
		assert.Equal(t, "abc", request.URL.Query().Get("example"))

		assert.Equal(t, mock.Ctx, request.Context())
	})

	t.Run("ShouldMapProtocolHTTP11", func(t *testing.T) {
		var request *http.Request

		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			request = r
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)

		require.NotNil(t, request)

		assert.Equal(t, "HTTP/1.1", request.Proto)
		assert.Equal(t, 1, request.ProtoMajor)
		assert.Equal(t, 1, request.ProtoMinor)
	})

	t.Run("ShouldMapProtocolHTTP2", func(t *testing.T) {
		var request *http.Request

		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			request = r
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.Header.SetProtocol("HTTP/2")

		handler(mock.Ctx)

		require.NotNil(t, request)

		assert.Equal(t, "HTTP/2", request.Proto)
		assert.Equal(t, 2, request.ProtoMajor)
	})

	t.Run("ShouldMapTLSConnectionState", func(t *testing.T) {
		var request *http.Request

		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			request = r
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Init2(&testTLSConn{state: tls.ConnectionState{Version: tls.VersionTLS13, ServerName: "auth.example.com", NegotiatedProtocol: "h2"}}, nil, true)

		handler(mock.Ctx)

		require.NotNil(t, request)
		require.NotNil(t, request.TLS)

		assert.Equal(t, uint16(tls.VersionTLS13), request.TLS.Version)
		assert.Equal(t, "auth.example.com", request.TLS.ServerName)
		assert.Equal(t, "h2", request.TLS.NegotiatedProtocol)
		assert.Equal(t, "10.0.0.1:41234", request.RemoteAddr)
	})

	t.Run("ShouldMapTransferEncoding", func(t *testing.T) {
		var request *http.Request

		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			request = r
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.SetBodyString("example")
		mock.Ctx.Request.Header.SetContentLength(-1)

		handler(mock.Ctx)

		require.NotNil(t, request)

		assert.Equal(t, []string{"chunked"}, request.TransferEncoding)
		assert.Empty(t, request.Header.Values(fasthttp.HeaderTransferEncoding))
		assert.Equal(t, int64(7), request.ContentLength)
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

	t.Run("ShouldPreserveCookieHeaderAfterBufferReuse", func(t *testing.T) {
		var cookie string

		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			cookie = r.Header.Get(fasthttp.HeaderCookie)
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.Header.SetCookie("authelia_session", "aaaaaaaaaaaaaaaaaaaa")

		handler(mock.Ctx)

		assert.Equal(t, "authelia_session=aaaaaaaaaaaaaaaaaaaa", cookie)

		mock.Ctx.Request.Header.SetCookie("authelia_session", "bbbbbbbbbbbbbbbbbbbb")

		for range mock.Ctx.Request.Header.All() {
			continue
		}

		assert.Equal(t, "authelia_session=aaaaaaaaaaaaaaaaaaaa", cookie)
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
		var request *http.Request

		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			request = r

			data, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			require.NoError(t, err)

			_, _ = rw.Write(append([]byte("Hello World"), data...))
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.SetBodyString("example")

		handler(mock.Ctx)

		require.NotNil(t, request)

		assert.Equal(t, int64(7), request.ContentLength)
		assert.Equal(t, "Hello Worldexample", string(mock.Ctx.Response.Body()))
	})

	t.Run("ShouldWriteResponseInMultipleParts", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			n, err := rw.Write([]byte("Hello "))

			require.NoError(t, err)
			assert.Equal(t, 6, n)

			if n, err = rw.Write([]byte("World")); err != nil {
				require.NoError(t, err)
			}

			assert.Equal(t, 5, n)
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)

		assert.Equal(t, "Hello World", string(mock.Ctx.Response.Body()))
	})

	t.Run("ShouldWriteRepeatedResponseHeaders", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
			rw.Header().Add("X-Example", "value-one")
			rw.Header().Add("X-Example", "value-two")
			rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
			rw.WriteHeader(http.StatusCreated)
		})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusCreated, mock.Ctx.Response.StatusCode())
		assert.Equal(t, "no-store", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderCacheControl)))

		values := mock.Ctx.Response.Header.PeekAll("X-Example")

		require.Len(t, values, 2)
		assert.Equal(t, "value-one", string(values[0]))
		assert.Equal(t, "value-two", string(values[1]))
	})

	t.Run("ShouldNotWriteResponseHeadersWhenNoneSet", func(t *testing.T) {
		handler := NewHTTPToAutheliaHandlerAdaptor(func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request) {})

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
		assert.Empty(t, mock.Ctx.Response.Body())
	})
}

type testTLSConn struct {
	net.Conn

	state tls.ConnectionState
}

func (c *testTLSConn) Handshake() (err error) {
	return nil
}

func (c *testTLSConn) ConnectionState() tls.ConnectionState {
	return c.state
}

func (c *testTLSConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(10, 0, 0, 2), Port: 443}
}

func (c *testTLSConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(10, 0, 0, 1), Port: 41234}
}
