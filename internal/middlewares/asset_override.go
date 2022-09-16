package middlewares

import (
	"os"
	"path/filepath"

	"github.com/valyala/fasthttp"
)

// AssetOverride allows overriding and serving of specific embedded assets from disk.
func AssetOverride(root string, strip int, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	if root == "" {
		return next
	}

	handler := fasthttp.FSHandler(root, strip)
	stripper := fasthttp.NewPathSlashesStripper(strip)

	return func(ctx *fasthttp.RequestCtx) {
		asset := filepath.Join(root, string(stripper(ctx)))

		if _, err := os.Stat(asset); err != nil {
			next(ctx)

			return
		}

		handler(ctx)
	}
}
