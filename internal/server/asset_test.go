package server

import (
	"errors"
	"io/fs"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestGenerateEtag(t *testing.T) {
	testCases := []struct {
		name      string
		payloadA  []byte
		payloadB  []byte
		wantEqual bool
	}{
		{
			name:      "ShouldReturnSameEtagForSamePayload",
			payloadA:  []byte("hello world"),
			payloadB:  []byte("hello world"),
			wantEqual: true,
		},
		{
			name:      "ShouldReturnDifferentEtagForDifferentPayload",
			payloadA:  []byte("hello world"),
			payloadB:  []byte("HELLO WORLD"),
			wantEqual: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			etagA := generateEtag(tc.payloadA)
			etagB := generateEtag(tc.payloadB)

			if tc.wantEqual {
				assert.Equal(t, etagA, etagB, "etags should be equal for identical payloads")
			} else {
				assert.NotEqual(t, etagA, etagB, "etags should differ for different payloads")
			}

			assert.Len(t, etagA, 40, "etag should be 40 characters (sha1 hex)")
		})
	}
}

func TestGetEmbedETags(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "ShouldComputeETagsForEmbeddedLocalesRecursively",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			etags := map[string][]byte{}

			getEmbedETags(locales, "locales", etags)

			assert.Greater(t, len(etags), 0, "expected at least one embedded locale file to have an etag")

			for p, etag := range etags {
				data, err := locales.ReadFile(p)
				assert.NoError(t, err, "should be able to read embedded file %s", p)
				assert.Equal(t, generateEtag(data), etag, "etag for %s should match generateEtag(data)", p)

				break
			}
		})
	}
}

func TestHFSHandleErr(t *testing.T) {
	testCases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "ShouldSet404ForNotExist",
			err:        fs.ErrNotExist,
			wantStatus: fasthttp.StatusNotFound,
		},
		{
			name:       "ShouldSet403ForPermission",
			err:        fs.ErrPermission,
			wantStatus: fasthttp.StatusForbidden,
		},
		{
			name:       "ShouldSet500ForOtherErrors",
			err:        errors.New("some other error"),
			wantStatus: fasthttp.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ctx fasthttp.RequestCtx

			hfsHandleErr(&ctx, tc.err)

			assert.Equal(t, tc.wantStatus, ctx.Response.StatusCode())
		})
	}
}

func TestNewPublicHTMLEmbeddedHandler(t *testing.T) {
	handler := newPublicHTMLEmbeddedHandler()

	require.NotNil(t, handler)

	testCases := []struct {
		name               string
		path               string
		method             string
		expectedStatusCode int
	}{
		{"ShouldServeExistingFile", "/api/openapi.yml", fasthttp.MethodGet, fasthttp.StatusOK},
		{"ShouldServeIndexHTML", "/api/index.html", fasthttp.MethodGet, fasthttp.StatusOK},
		{"ShouldReturn404ForMissing", "/nonexistent.file", fasthttp.MethodGet, fasthttp.StatusNotFound},
		{"ShouldHandleHEADRequest", "/api/openapi.yml", fasthttp.MethodHead, fasthttp.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				ctx fasthttp.RequestCtx
				req fasthttp.Request
			)

			req.Header.SetMethod(tc.method)
			req.SetRequestURI(tc.path)
			ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

			handler(&ctx)

			assert.Equal(t, tc.expectedStatusCode, ctx.Response.StatusCode())
		})
	}
}

func TestNewPublicHTMLEmbeddedHandlerETagCaching(t *testing.T) {
	handler := newPublicHTMLEmbeddedHandler()

	testCases := []struct {
		name               string
		path               string
		sendETag           bool
		expectedStatusCode int
	}{
		{"ShouldReturn200WithoutETag", "/api/openapi.yml", false, fasthttp.StatusOK},
		{"ShouldReturn304WithMatchingETag", "/api/openapi.yml", true, fasthttp.StatusNotModified},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				ctx1 fasthttp.RequestCtx
				req1 fasthttp.Request
			)

			req1.SetRequestURI(tc.path)
			ctx1.Init(&req1, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

			handler(&ctx1)

			etag := ctx1.Response.Header.Peek("ETag")

			if tc.sendETag && len(etag) > 0 {
				var (
					ctx2 fasthttp.RequestCtx
					req2 fasthttp.Request
				)

				req2.SetRequestURI(tc.path)
				req2.Header.SetBytesKV([]byte("If-None-Match"), etag)
				ctx2.Init(&req2, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

				handler(&ctx2)

				assert.Equal(t, tc.expectedStatusCode, ctx2.Response.StatusCode())
			} else {
				assert.Equal(t, tc.expectedStatusCode, ctx1.Response.StatusCode())
			}
		})
	}
}

func TestNewLocalesPathResolver(t *testing.T) {
	resolver, err := newLocalesPathResolver()

	require.NoError(t, err)
	require.NotNil(t, resolver)
}

func TestNewLocalesEmbeddedHandler(t *testing.T) {
	handler, err := newLocalesEmbeddedHandler()

	require.NoError(t, err)
	require.NotNil(t, handler)
}

func TestNewLocalesListHandler(t *testing.T) {
	handler, err := newLocalesListHandler()

	require.NoError(t, err)
	require.NotNil(t, handler)
}
