package middlewares

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/valyala/fasthttp"
)

// AssetOverride allows overriding and serving of specific embedded assets from disk.
func AssetOverride(root string, strip int, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	if root == "" {
		return next
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		rootAbs = root
	}

	handler := fasthttp.FSHandler(rootAbs, strip)
	stripper := fasthttp.NewPathSlashesStripper(strip)
	rootPrefix := rootAbs + string(os.PathSeparator)

	return func(ctx *fasthttp.RequestCtx) {
		requestPath := string(stripper(ctx))
		cleaned := filepath.Clean(requestPath)
		cleaned = strings.TrimPrefix(cleaned, string(os.PathSeparator))

		if cleaned == "" || cleaned == "." {
			next(ctx)

			return
		}

		if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusForbidden), fasthttp.StatusForbidden)

			return
		}

		asset := filepath.Join(rootAbs, cleaned)
		if asset != rootAbs && !strings.HasPrefix(asset, rootPrefix) {
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusForbidden), fasthttp.StatusForbidden)

			return
		}

		info, err := os.Stat(asset)
		if err != nil || info.IsDir() {
			next(ctx)

			return
		}

		handler(ctx)
	}
}
