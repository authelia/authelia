package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/valyala/fasthttp"
)

func newFastHTTPHandler(h http.Handler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var r http.Request
		if err := convertRequest(ctx, &r, true); err != nil {
			ctx.Logger().Printf("cannot parse requestURI %q: %s", r.RequestURI, err)
			ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)

			return
		}

		var w netHTTPResponseWriter

		h.ServeHTTP(&w, r.WithContext(ctx))
		ctx.SetStatusCode(w.StatusCode())

		haveContentType := false

		for k, vv := range w.Header() {
			if k == fasthttp.HeaderContentType {
				haveContentType = true

				for _, v := range vv {
					ctx.Response.Header.Set(k, v)
				}

				continue
			}

			for _, v := range vv {
				ctx.Response.Header.Add(k, v)
			}
		}

		if !haveContentType {
			// From net/http.ResponseWriter.Write:
			// If the Header does not contain a Content-Type line, Write adds a Content-Type set
			// to the result of passing the initial 512 bytes of written data to DetectContentType.
			l := 512

			if len(w.body) < 512 {
				l = len(w.body)
			}

			ctx.Response.Header.Set(fasthttp.HeaderContentType, http.DetectContentType(w.body[:l]))
		}

		ctx.Write(w.body) //nolint:errcheck
	}
}

type netHTTPResponseWriter struct {
	statusCode int
	h          http.Header
	body       []byte
}

func (w *netHTTPResponseWriter) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}

	return w.statusCode
}

func (w *netHTTPResponseWriter) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}

	return w.h
}

func (w *netHTTPResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *netHTTPResponseWriter) Write(p []byte) (int, error) {
	w.body = append(w.body, p...)
	return len(p), nil
}

func convertRequest(ctx *fasthttp.RequestCtx, r *http.Request, forServer bool) error {
	rURL, err := url.ParseRequestURI(string(ctx.RequestURI()))
	if err != nil {
		return err
	}

	r.Method = string(ctx.Method())
	r.Proto = "HTTP/1.1"
	r.ProtoMajor = 1
	r.ProtoMinor = 1
	r.ContentLength = int64(len(ctx.PostBody()))
	r.RemoteAddr = ctx.RemoteAddr().String()
	r.Host = string(ctx.Host())

	if forServer {
		r.RequestURI = string(ctx.RequestURI())
	}

	hdr := make(http.Header)
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)
		switch sk {
		case fasthttp.HeaderTransferEncoding:
			r.TransferEncoding = append(r.TransferEncoding, sv)
		case fasthttp.HeaderContentType:
			hdr.Set(sk, sv)
		default:
			hdr.Add(sk, sv)
		}
	})

	r.Header = hdr
	r.Body = ioutil.NopCloser(bytes.NewReader(ctx.PostBody()))
	r.URL = rURL

	return nil
}
