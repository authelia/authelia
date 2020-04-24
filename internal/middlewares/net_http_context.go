package middlewares

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (n NetHTTPCtx) ResponseWriter() *NetHTTPResponseWriter {
	if n.responseWriter == nil {
		n.responseWriter = &NetHTTPResponseWriter{AutheliaCtx: n.AutheliaCtx}
	}
	return n.responseWriter
}

func (n *NetHTTPCtx) GetRequest() http.Request {
	request := http.Request{
		Method:     string(n.AutheliaCtx.Method()),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 0,
		RequestURI: string(n.AutheliaCtx.RequestURI()),
		Host:       string(n.AutheliaCtx.Host()),
		RemoteAddr: n.AutheliaCtx.RemoteAddr().String(),
	}
	body := n.AutheliaCtx.PostBody()
	request.ContentLength = int64(len(body))
	headers := make(http.Header)
	n.AutheliaCtx.Request.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)
		switch sk {
		case "Transfer-Encoding":
			request.TransferEncoding = append(request.TransferEncoding, sv)
		default:
			headers.Set(sk, sv)
		}
	})
	request.Header = headers
	request.Body = &netHTTPBody{body}
	reqUrl, err := url.ParseRequestURI(n.AutheliaCtx.URI().String())
	if err == nil {
		request.URL = reqUrl
	}

	return request
}

func (wr *NetHTTPResponseWriter) Headers() http.Header {
	if wr.headers == nil {
		wr.headers = make(http.Header)
	}
	return wr.headers
}
func (wr *NetHTTPResponseWriter) Write(data []byte) (int, error) {
	for key, value := range wr.headers {
		if len(value) <= 1 {
			wr.AutheliaCtx.Response.Header.Set(key, value[0])
		} else {
			wr.AutheliaCtx.Response.Header.Set(key, strings.Join(value, ";"))
		}
	}
	wr.AutheliaCtx.SetStatusCode(wr.StatusCode())
	return wr.AutheliaCtx.RequestCtx.Write(data)
}
func (wr *NetHTTPResponseWriter) WriteHeader(statusCode int) {
	wr.statusCode = statusCode
}
func (wr *NetHTTPResponseWriter) StatusCode() int {
	if wr.statusCode == 0 {
		return http.StatusOK
	}
	return wr.statusCode
}

func (r *netHTTPBody) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

func (r *netHTTPBody) Close() error {
	r.b = r.b[:0]
	return nil
}
