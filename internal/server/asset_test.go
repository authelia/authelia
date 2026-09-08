package server

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"mime"
	"net"
	"path"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/templates"
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

			assert.Len(t, etagA, 42, "etag should be a quoted 40 character sha1 hex digest")
			assert.Regexp(t, `^"[0-9a-f]{40}"$`, string(etagA), "etag should be a quoted opaque string")
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

	testCases := []struct {
		name              string
		language          string
		namespace         string
		variant           string
		expectedSupported bool
		expectedAsset     string
		expectedEmbedded  bool
	}{
		{
			"ShouldResolveEnglishPortal",
			"en",
			"portal",
			"",
			true,
			"locales/en/portal.json",
			true,
		},
		{
			"ShouldResolveGermanWithVariant",
			"de",
			"portal",
			"DE",
			true,
			"locales/de-DE/portal.json",
			true,
		},
		{
			"ShouldResolveFrenchWithVariant",
			"fr",
			"portal",
			"FR",
			true,
			"locales/fr-FR/portal.json",
			true,
		},
		{
			"ShouldResolveChineseAlias",
			"zh",
			"portal",
			"",
			true,
			"locales/zh-CN/portal.json",
			true,
		},
		{
			"ShouldResolveCzechAlias",
			"cs",
			"portal",
			"",
			true,
			"locales/cs-CZ/portal.json",
			true,
		},
		{
			"ShouldResolveJapaneseAlias",
			"ja",
			"portal",
			"",
			true,
			"locales/ja-JP/portal.json",
			true,
		},
		{
			"ShouldReturnUnsupportedForUnknownLanguage",
			"xx",
			"portal",
			"",
			false,
			"",
			false,
		},
		{
			"ShouldResolveSpanishWithVariant",
			"es",
			"portal",
			"ES",
			true,
			"locales/es-ES/portal.json",
			true,
		},
		{
			"ShouldResolveWithDifferentNamespace",
			"en",
			"common",
			"",
			true,
			"locales/en/common.json",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.SetUserValue("language", tc.language)
			mock.Ctx.SetUserValue("namespace", tc.namespace)

			if tc.variant != "" {
				mock.Ctx.SetUserValue("variant", tc.variant)
			}

			supported, asset, embedded := resolver(mock.Ctx)

			assert.Equal(t, tc.expectedSupported, supported)
			assert.Equal(t, tc.expectedAsset, asset)
			assert.Equal(t, tc.expectedEmbedded, embedded)
		})
	}
}

func TestNewLocalesEmbeddedHandler(t *testing.T) {
	handler, err := newLocalesEmbeddedHandler()

	require.NoError(t, err)
	require.NotNil(t, handler)

	testCases := []struct {
		name               string
		language           string
		namespace          string
		variant            string
		method             string
		ifNoneMatch        string
		expectedStatusCode int
		expectJSON         bool
	}{
		{
			"ShouldServeEnglishPortal",
			"en",
			"portal",
			"",
			fasthttp.MethodGet,
			"",
			fasthttp.StatusOK,
			true,
		},
		{
			"ShouldServeGermanPortalWithVariant",
			"de",
			"portal",
			"DE",
			fasthttp.MethodGet,
			"",
			fasthttp.StatusOK,
			true,
		},
		{
			"ShouldReturn404ForUnsupportedLanguage",
			"xx",
			"portal",
			"",
			fasthttp.MethodGet,
			"",
			fasthttp.StatusNotFound,
			false,
		},
		{
			"ShouldHandleHEADRequest",
			"en",
			"portal",
			"",
			fasthttp.MethodHead,
			"",
			fasthttp.StatusOK,
			false,
		},
		{
			"ShouldServeChineseAlias",
			"zh",
			"portal",
			"",
			fasthttp.MethodGet,
			"",
			fasthttp.StatusOK,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Request.Header.SetMethod(tc.method)
			mock.Ctx.SetUserValue("language", tc.language)
			mock.Ctx.SetUserValue("namespace", tc.namespace)

			if tc.variant != "" {
				mock.Ctx.SetUserValue("variant", tc.variant)
			}

			handler(mock.Ctx)

			assert.Equal(t, tc.expectedStatusCode, mock.Ctx.Response.StatusCode())

			if tc.expectJSON {
				ct := string(mock.Ctx.Response.Header.ContentType())
				assert.Contains(t, ct, "application/json")
			}
		})
	}
}

func TestNewLocalesEmbeddedHandlerETagCaching(t *testing.T) {
	handler, err := newLocalesEmbeddedHandler()

	require.NoError(t, err)

	testCases := []struct {
		name               string
		expectedStatusCode int
	}{
		{"ShouldReturn304WithMatchingETag", fasthttp.StatusNotModified},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock1 := mocks.NewMockAutheliaCtx(t)
			defer mock1.Close()

			mock1.Ctx.SetUserValue("language", "en")
			mock1.Ctx.SetUserValue("namespace", "portal")

			handler(mock1.Ctx)

			etag := mock1.Ctx.Response.Header.Peek("ETag")
			require.NotEmpty(t, etag)

			mock2 := mocks.NewMockAutheliaCtx(t)
			defer mock2.Close()

			mock2.Ctx.SetUserValue("language", "en")
			mock2.Ctx.SetUserValue("namespace", "portal")
			mock2.Ctx.Request.Header.SetBytesKV([]byte("If-None-Match"), etag)

			handler(mock2.Ctx)

			assert.Equal(t, tc.expectedStatusCode, mock2.Ctx.Response.StatusCode())
		})
	}
}

func TestNewLocalesListHandler(t *testing.T) {
	handler, err := newLocalesListHandler()

	require.NoError(t, err)
	require.NotNil(t, handler)

	testCases := []struct {
		name               string
		method             string
		ifNoneMatch        string
		expectedStatusCode int
		expectJSON         bool
	}{
		{
			"ShouldReturnLocaleList",
			fasthttp.MethodGet,
			"",
			fasthttp.StatusOK,
			true,
		},
		{
			"ShouldHandleHEADRequest",
			fasthttp.MethodHead,
			"",
			fasthttp.StatusOK,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Request.Header.SetMethod(tc.method)

			handler(mock.Ctx)

			assert.Equal(t, tc.expectedStatusCode, mock.Ctx.Response.StatusCode())

			if tc.expectJSON {
				ct := string(mock.Ctx.Response.Header.ContentType())
				assert.Contains(t, ct, "application/json")
			}

			etag := mock.Ctx.Response.Header.Peek("ETag")
			assert.NotEmpty(t, etag)
		})
	}
}

func TestNewLocalesListHandlerETagCaching(t *testing.T) {
	handler, err := newLocalesListHandler()

	require.NoError(t, err)

	testCases := []struct {
		name               string
		expectedStatusCode int
	}{
		{"ShouldReturn304WithMatchingETag", fasthttp.StatusNotModified},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock1 := mocks.NewMockAutheliaCtx(t)
			defer mock1.Close()

			handler(mock1.Ctx)

			etag := mock1.Ctx.Response.Header.Peek("ETag")
			require.NotEmpty(t, etag)

			mock2 := mocks.NewMockAutheliaCtx(t)
			defer mock2.Close()

			mock2.Ctx.Request.Header.SetBytesKV([]byte("If-None-Match"), etag)

			handler(mock2.Ctx)

			assert.Equal(t, tc.expectedStatusCode, mock2.Ctx.Response.StatusCode())
		})
	}
}

func TestGetEmbedCompressed(t *testing.T) {
	const asset = "locales/en/portal.json"

	identity, err := locales.ReadFile(asset)
	require.NoError(t, err)

	variants, ok := getEmbedCompressed(locales, "locales")[asset]

	require.True(t, ok)
	require.Len(t, variants, 2)

	assert.Equal(t, encodingBrotli, variants[0].encoding)
	assert.Equal(t, encodingGzip, variants[1].encoding)
	assert.NotEqual(t, variants[0].etag, variants[1].etag)

	for _, variant := range variants {
		assert.Less(t, len(variant.data), len(identity))
		assert.Equal(t, identity, decompressAsset(t, variant.encoding, variant.data))
		assert.Equal(t, generateEtag(variant.data), variant.etag)
		assert.NotEqual(t, generateEtag(identity), variant.etag)
	}
}

func TestGetEmbedCompressedShouldOnlyIncludeCompressibleAssets(t *testing.T) {
	testCases := []struct {
		name    string
		embedFS embed.FS
		root    string
	}{
		{"ShouldOnlyCompressPublicHTML", assets, assetsRoot},
		{"ShouldOnlyCompressLocales", locales, "locales"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for p, variants := range getEmbedCompressed(tc.embedFS, tc.root) {
				assert.Contains(t, extsCompressible, path.Ext(p))
				assert.NotContains(t, templates.AssetPathsTemplated, p)

				identity, err := tc.embedFS.ReadFile(p)
				require.NoError(t, err)

				assert.GreaterOrEqual(t, len(identity), compressionMinSize)

				for _, variant := range variants {
					assert.Less(t, len(variant.data), len(identity))
					assert.Equal(t, identity, decompressAsset(t, variant.encoding, variant.data))
				}
			}
		})
	}
}

// The repository holds placeholders for the templated assets which the build replaces: two are empty and the third is
// under compressionMinSize, so asserting against the embedded tree would pass for a reason that stops holding once they
// carry the content they ship with.
func TestIsPreCompressibleShouldExcludeTemplatedAssets(t *testing.T) {
	require.NotEmpty(t, templates.AssetPathsTemplated)

	for _, asset := range templates.AssetPathsTemplated {
		t.Run(asset, func(t *testing.T) {
			assert.False(t, isPreCompressible(asset), "a templated asset is rendered per request, so a representation compressed ahead of time could never be sent")
		})
	}

	assert.True(t, isPreCompressible("public_html/static/js/index.js"))
	assert.False(t, isPreCompressible("public_html/favicon.ico"))
}

func TestTemplatedAssetsShouldBeEmbedded(t *testing.T) {
	for _, asset := range templates.AssetPathsTemplated {
		t.Run(asset, func(t *testing.T) {
			_, err := assets.ReadFile(asset)
			assert.NoError(t, err)
		})
	}
}

func TestExtsCompressibleShouldHaveKnownContentTypes(t *testing.T) {
	for _, ext := range extsCompressible {
		t.Run(ext, func(t *testing.T) {
			assert.NotEmpty(t, mime.TypeByExtension(ext), "the content type of a compressed asset can't be detected from its content")
		})
	}
}

func TestGetAcceptedCompressedAsset(t *testing.T) {
	variants := []compressedAsset{
		{encoding: encodingBrotli},
		{encoding: encodingGzip},
	}

	testCases := []struct {
		name             string
		acceptEncoding   string
		expectedEncoding []byte
	}{
		{"ShouldNotMatchWithoutHeader", "", nil},
		{"ShouldNotMatchIdentity", "identity", nil},
		{"ShouldNotMatchUnsupported", "zstd, deflate", nil},
		{"ShouldMatchBrotli", "br", encodingBrotli},
		{"ShouldMatchGzip", "gzip", encodingGzip},
		{"ShouldPreferBrotli", "gzip, deflate, br", encodingBrotli},
		{"ShouldFallbackToGzip", "gzip, deflate", encodingGzip},
		{"ShouldMatchBrotliWithQuality", "br;q=1", encodingBrotli},
		{"ShouldMatchCaseInsensitively", "GZIP", encodingGzip},
		{"ShouldPreferHigherQuality", "br;q=0.5, gzip;q=1", encodingGzip},
		{"ShouldFallbackToGzipWhenBrotliRefused", "gzip;q=1, br;q=0", encodingGzip},
		{"ShouldNotMatchWhenAllRefused", "br;q=0, gzip;q=0", nil},
		{"ShouldMatchWildcard", "*", encodingBrotli},
		{"ShouldMatchWildcardWithQuality", "identity, *;q=0.5", encodingBrotli},
		{"ShouldPreferExplicitCodingOverWildcard", "*, br;q=0", encodingGzip},
		{"ShouldNotMatchRefusedWildcard", "identity, *;q=0", nil},
		{"ShouldNotMatchMalformedQuality", "br;q=nope, gzip;q=2", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			variant := getAcceptedCompressedAsset(newAssetRequestCtx(fasthttp.MethodGet, "/", tc.acceptEncoding), variants)

			if tc.expectedEncoding == nil {
				assert.Nil(t, variant)

				return
			}

			require.NotNil(t, variant)
			assert.Equal(t, tc.expectedEncoding, variant.encoding)
		})
	}
}

func TestNewEmbeddedHandlerCompression(t *testing.T) {
	handler := newEmbeddedHandler(locales, "locales")

	identity, err := locales.ReadFile("locales/en/portal.json")
	require.NoError(t, err)

	testCases := []struct {
		name             string
		acceptEncoding   string
		expectedEncoding []byte
	}{
		{"ShouldServeIdentityWithoutAcceptEncoding", "", nil},
		{"ShouldServeIdentityForUnsupportedEncoding", "zstd", nil},
		{"ShouldServeBrotli", "br", encodingBrotli},
		{"ShouldServeGzip", "gzip", encodingGzip},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newAssetRequestCtx(fasthttp.MethodGet, "/en/portal.json", tc.acceptEncoding)

			handler(ctx)

			assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
			assert.Equal(t, fasthttp.HeaderAcceptEncoding, string(ctx.Response.Header.Peek(fasthttp.HeaderVary)))
			assert.Contains(t, string(ctx.Response.Header.ContentType()), "application/json")

			encoding := ctx.Response.Header.Peek(fasthttp.HeaderContentEncoding)

			if tc.expectedEncoding == nil {
				assert.Empty(t, encoding)
				assert.Equal(t, identity, ctx.Response.Body())

				return
			}

			assert.Equal(t, tc.expectedEncoding, encoding)
			assert.Less(t, len(ctx.Response.Body()), len(identity))
			assert.Equal(t, identity, decompressAsset(t, encoding, ctx.Response.Body()))
		})
	}
}

func TestNewEmbeddedHandlerCompressionHEAD(t *testing.T) {
	handler := newEmbeddedHandler(locales, "locales")

	identity, err := locales.ReadFile("locales/en/portal.json")
	require.NoError(t, err)

	ctx := newAssetRequestCtx(fasthttp.MethodHead, "/en/portal.json", "br")

	handler(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.True(t, ctx.Response.SkipBody)
	assert.Equal(t, encodingBrotli, ctx.Response.Header.Peek(fasthttp.HeaderContentEncoding))

	contentLength, err := strconv.Atoi(string(ctx.Response.Header.Peek(fasthttp.HeaderContentLength)))
	require.NoError(t, err)

	assert.Positive(t, contentLength)
	assert.Less(t, contentLength, len(identity))
}

func TestNewEmbeddedHandlerCompressionETag(t *testing.T) {
	handler := newEmbeddedHandler(locales, "locales")

	ctxIdentity := newAssetRequestCtx(fasthttp.MethodGet, "/en/portal.json", "")

	handler(ctxIdentity)

	ctxBrotli := newAssetRequestCtx(fasthttp.MethodGet, "/en/portal.json", "br")

	handler(ctxBrotli)

	etagIdentity := bytes.Clone(ctxIdentity.Response.Header.Peek(fasthttp.HeaderETag))
	etagBrotli := bytes.Clone(ctxBrotli.Response.Header.Peek(fasthttp.HeaderETag))

	require.NotEmpty(t, etagIdentity)
	require.NotEmpty(t, etagBrotli)

	assert.Regexp(t, `^"[0-9a-f]{40}"$`, string(etagIdentity))
	assert.Regexp(t, `^"[0-9a-f]{40}"$`, string(etagBrotli))
	assert.NotEqual(t, etagIdentity, etagBrotli)

	testCases := []struct {
		name               string
		acceptEncoding     string
		ifNoneMatch        []byte
		expectedStatusCode int
	}{
		{"ShouldReturn304ForMatchingIdentityETag", "", etagIdentity, fasthttp.StatusNotModified},
		{"ShouldReturn304ForMatchingBrotliETag", "br", etagBrotli, fasthttp.StatusNotModified},
		{"ShouldReturn200ForBrotliETagWithoutAcceptEncoding", "", etagBrotli, fasthttp.StatusOK},
		{"ShouldReturn200ForIdentityETagWithAcceptEncoding", "br", etagIdentity, fasthttp.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newAssetRequestCtx(fasthttp.MethodGet, "/en/portal.json", tc.acceptEncoding)

			ctx.Request.Header.SetBytesKV(headerIfNoneMatch, tc.ifNoneMatch)

			handler(ctx)

			assert.Equal(t, tc.expectedStatusCode, ctx.Response.StatusCode())
			assert.Equal(t, fasthttp.HeaderAcceptEncoding, string(ctx.Response.Header.Peek(fasthttp.HeaderVary)))
		})
	}
}

func newAssetRequestCtx(method, uri, acceptEncoding string) (ctx *fasthttp.RequestCtx) {
	var req fasthttp.Request

	req.Header.SetMethod(method)
	req.SetRequestURI(uri)

	if acceptEncoding != "" {
		req.Header.Set(fasthttp.HeaderAcceptEncoding, acceptEncoding)
	}

	ctx = &fasthttp.RequestCtx{}

	ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

	return ctx
}

func decompressAsset(t *testing.T, encoding, data []byte) (decompressed []byte) {
	t.Helper()

	var err error

	switch {
	case bytes.Equal(encoding, encodingBrotli):
		decompressed, err = fasthttp.AppendUnbrotliBytes(nil, data)
	case bytes.Equal(encoding, encodingGzip):
		decompressed, err = fasthttp.AppendGunzipBytes(nil, data)
	default:
		t.Fatalf("unexpected content encoding '%s'", encoding)
	}

	require.NoError(t, err)

	return decompressed
}
