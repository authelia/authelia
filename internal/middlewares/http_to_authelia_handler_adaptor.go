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

		r := &http.Request{
			Header:        make(http.Header),
			TLS:           ctx.TLSConnectionState(),
			Proto:         bytesToString(ctx.Request.Header.Protocol()),
			ProtoMinor:    1,
			ContentLength: int64(len(body)),
			Method:        bytesToString(ctx.Method()),
			Host:          bytesToString(ctx.Host()),
			RequestURI:    bytesToString(ctx.RequestURI()),
			RemoteAddr:    ctx.RemoteAddr().String(),
			Body:          io.NopCloser(bytes.NewReader(body)),
		}

		if r.Proto == strProtoHTTP2 {
			r.ProtoMajor = 2
		} else {
			r.ProtoMajor = 1
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

		var (
			uri *url.URL
			err error
		)

		if uri, err = url.ParseRequestURI(r.RequestURI); err != nil {
			ctx.GetLogger().Errorf("Cannot parse requestURI %q: %s", r.RequestURI, err)
			ctx.RequestCtx.Error("Internal Server Error", fasthttp.StatusInternalServerError)

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

// WriteHeader writes the status code.
func (w *netHTTPResponseWriter) WriteHeader(statusCode int) {
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
