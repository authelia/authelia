package middleware

import (
	"os"
	"path/filepath"

	"github.com/valyala/fasthttp"
)

// AssetOverride allows overriding and serving of specific embedded assets from disk.
func AssetOverride(root string, strip int, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if root == "" {
			next(ctx)

			return
		}

		_, err := os.Stat(filepath.Join(root, string(fasthttp.NewPathSlashesStripper(strip)(ctx))))
		if err != nil {
			next(ctx)

			return
		}

		fasthttp.FSHandler(root, strip)(ctx)
	}
}
