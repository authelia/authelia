package middlewares

import (
	"io"
	"net/http"
	"net/url"

	"github.com/valyala/fasthttp"
)

// AutheliaHandlerFunc is used with the NewHTTPToAutheliaHandlerAdaptor to encapsulate a func.
type AutheliaHandlerFunc func(ctx *AutheliaCtx, rw http.ResponseWriter, r *http.Request)

// NewHTTPToAutheliaHandlerAdaptor creates a new adaptor given the AutheliaHandlerFunc.
func NewHTTPToAutheliaHandlerAdaptor(h AutheliaHandlerFunc) RequestHandler {
	return func(ctx *AutheliaCtx) {
		var r http.Request

		body := ctx.PostBody()
		r.Method = string(ctx.Method())
		r.Proto = "HTTP/1.1"
		r.ProtoMajor = 1
		r.ProtoMinor = 1
		r.RequestURI = string(ctx.RequestURI())
		r.ContentLength = int64(len(body))
		r.Host = string(ctx.Host())
		r.RemoteAddr = ctx.RemoteAddr().String()

		hdr := make(http.Header)

		for k, v := range ctx.Request.Header.All() {
			sk := string(k)
			sv := string(v)

			switch sk {
			case fasthttp.HeaderTransferEncoding:
				r.TransferEncoding = append(r.TransferEncoding, sv)
			default:
				hdr.Set(sk, sv)
			}
		}

		r.Header = hdr
		r.Body = &netHTTPBody{body}

		rURL, err := url.ParseRequestURI(r.RequestURI)
		if err != nil {
			ctx.Logger.Errorf("Cannot parse requestURI %q: %s", r.RequestURI, err)
			ctx.RequestCtx.Error("Internal Server Error", fasthttp.StatusInternalServerError)

			return
		}

		r.URL = rURL

		var w netHTTPResponseWriter

		h(ctx, &w, r.WithContext(ctx))

		ctx.SetStatusCode(w.StatusCode())

		for k, vv := range w.Header() {
			for _, v := range vv {
				ctx.Response.Header.Set(k, v)
			}
		}

		_, _ = ctx.Write(w.body)
	}
}

type netHTTPBody struct {
	b []byte
}

// Read reads the body.
func (r *netHTTPBody) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}

	n := copy(p, r.b)
	r.b = r.b[n:]

	return n, nil
}

// Close closes the body.
func (r *netHTTPBody) Close() error {
	r.b = r.b[:0]
	return nil
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

// Header returns the http.Header.
func (w *netHTTPResponseWriter) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}

	return w.h
}

// WriteHeader writes the status code.
func (w *netHTTPResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

// Write writes to the body.
func (w *netHTTPResponseWriter) Write(p []byte) (int, error) {
	w.body = append(w.body, p...)
	return len(p), nil
}
