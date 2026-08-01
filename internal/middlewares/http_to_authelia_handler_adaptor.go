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
func NewHTTPToAutheliaHandlerAdaptor(handler AutheliaHandlerFunc) RequestHandler {
	return func(ctx *AutheliaCtx) {
		body := ctx.PostBody()

		proto := string(ctx.Request.Header.Protocol())

		major, minor, ok := http.ParseHTTPVersion(proto)
		if !ok {
			proto, major, minor = strProtoHTTP11, 1, 1
		}

		r := &http.Request{
			Header:        make(http.Header),
			TLS:           ctx.TLSConnectionState(),
			Proto:         proto,
			ProtoMajor:    major,
			ProtoMinor:    minor,
			ContentLength: int64(len(body)),
			Method:        string(ctx.Method()),
			Host:          string(ctx.Host()),
			RequestURI:    string(ctx.RequestURI()),
			RemoteAddr:    ctx.RemoteAddr().String(),
		}

		for k, v := range ctx.Request.Header.All() {
			header := string(k)
			value := string(v)

			switch header {
			case fasthttp.HeaderTransferEncoding:
				r.TransferEncoding = append(r.TransferEncoding, value)
			default:
				r.Header.Add(header, value)
			}
		}

		r.Body = &netHTTPBody{body}

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

// WriteHeader writes the status code. Only the first call has an effect which matches the net/http semantics the
// wrapped handlers are written against.
func (w *netHTTPResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode != 0 {
		return
	}

	w.statusCode = statusCode
}

// Write writes to the body.
func (w *netHTTPResponseWriter) Write(p []byte) (int, error) {
	w.body = append(w.body, p...)
	return len(p), nil
}
