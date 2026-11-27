// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unsafe"

	"github.com/valyala/fasthttp"
)

// AutheliaHandlerFunc is used with the NewHTTPToAutheliaHandlerAdaptor to encapsulate a func.
type AutheliaHandlerFunc func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request)

// NewHTTPToAutheliaHandlerAdaptor creates a new adaptor given the AutheliaHandlerFunc.
func NewHTTPToAutheliaHandlerAdaptor(handler AutheliaHandlerFunc) RequestHandler {
	return func(ctx *AutheliaCtx) {
		body := ctx.PostBody()

		proto := bytesToString(ctx.Request.Header.Protocol())

		major, minor, ok := http.ParseHTTPVersion(proto)
		if !ok {
			switch proto {
			case strProtoHTTP2:
				major, minor = 2, 0
			default:
				proto, major, minor = strProtoHTTP11, 1, 1
			}
		}

		r := &http.Request{
			Header:        make(http.Header),
			TLS:           ctx.TLSConnectionState(),
			Proto:         proto,
			ProtoMajor:    major,
			ProtoMinor:    minor,
			ContentLength: int64(len(body)),
			Method:        bytesToString(ctx.Method()),
			Host:          bytesToString(ctx.GetXForwardedHost()),
			RequestURI:    bytesToString(ctx.RequestURI()),
			RemoteAddr:    ctx.RemoteAddr().String(),
			Body:          io.NopCloser(bytes.NewReader(body)),
		}

		for k, v := range ctx.Request.Header.All() {
			key := bytesToString(k)
			value := bytesToString(v)

			switch key {
			case fasthttp.HeaderTransferEncoding:
				r.TransferEncoding = append(r.TransferEncoding, value)
			default:
				if key == fasthttp.HeaderCookie {
					value = strings.Clone(value)
				}

				r.Header.Add(key, value)
			}
		}

		// The host and scheme of the request are those of the X-Forwarded-Host and X-Forwarded-Proto headers, falling
		// back to the Host header and the connection, which are the values the issuer is derived and validated from.
		// The wrapped handlers reconstruct the request URL from these, for example to match the target URI of an RFC
		// 9449 DPoP proof, so they must agree with the issuer rather than with the Host header the proxy forwarded.
		r.Header.Set(fasthttp.HeaderXForwardedProto, string(ctx.XForwardedProto()))

		var (
			uri *url.URL
			err error
		)

		if uri, err = url.ParseRequestURI(r.RequestURI); err != nil {
			ctx.GetLogger().Errorf("Cannot parse requestURI %q: %s", r.RequestURI, err)
			ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)

			return
		}

		r.URL = uri

		var rw netHTTPResponseWriter

		handler(ctx, &rw, r.WithContext(ctx))

		ctx.SetStatusCode(rw.StatusCode())

		for key, values := range rw.Header() {
			for i, value := range values {
				switch i {
				case 0:
					ctx.Response.Header.Set(key, value)
				default:
					ctx.Response.Header.Add(key, value)
				}
			}
		}

		if rw.body != nil {
			_, _ = ctx.Write(rw.body)
		}
	}
}

type netHTTPResponseWriter struct {
	statusCode int
	h          http.Header
	body       []byte
}

// StatusCode returns the status code.
func (w *netHTTPResponseWriter) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}

	return w.statusCode
}

// Header returns the [http.Header].
func (w *netHTTPResponseWriter) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}

	return w.h
}

// WriteHeader writes the status code. Only the first final status code has an effect, and an informational status code
// is replaced by a later one, which matches the net/http semantics the wrapped handlers are written against.
func (w *netHTTPResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode >= http.StatusOK {
		return
	}

	w.statusCode = statusCode
}

// Write writes to the body.
func (w *netHTTPResponseWriter) Write(p []byte) (int, error) {
	w.body = append(w.body, p...)
	return len(p), nil
}

func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
