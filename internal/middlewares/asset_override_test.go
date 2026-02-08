package middlewares

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestAssetOverride(t *testing.T) {
	var example fasthttp.RequestHandler = func(ctx *fasthttp.RequestCtx) {
		_, _ = ctx.WriteString("example")
	}

	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.txt"), []byte("test"), 0600))

	testCases := []struct {
		name     string
		root     string
		path     string
		strip    int
		next     fasthttp.RequestHandler
		expected string
	}{
		{
			"ShouldReturnEmpty",
			"",
			"",
			0,
			example,
			"example",
		},
		{
			"ShouldReturnedOverrideAsset",
			dir,
			"index.txt",
			0,
			example,
			"test",
		},
		{
			"ShouldNextAsset",
			dir,
			"index.html",
			0,
			example,
			"example",
		},
		{
			"ShouldReturnStrippedOverrideAsset",
			dir,
			"path/index.txt",
			1,
			example,
			"test",
		},
		{
			"ShouldReturnOverrideAssetWithLeadingSlash",
			dir,
			"/index.txt",
			0,
			example,
			"test",
		},
		{
			"ShouldNextAssetWithLeadingSlash",
			dir,
			"/index.html",
			0,
			example,
			"example",
		},
		{
			"ShouldReturnStrippedOverrideAssetMulti",
			dir,
			"/a/b/index.txt",
			2,
			example,
			"test",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := AssetOverride(tc.root, tc.strip, tc.next)

			ctx := &fasthttp.RequestCtx{}

			ctx.Request.SetRequestURI(tc.path)

			handler(ctx)

			assert.Equal(t, tc.expected, string(ctx.Response.Body()))
		})
	}
}
