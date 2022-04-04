package middlewares

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/valyala/fasthttp"
)

// AssetOverrideMiddleware allows overriding and serving of specific embedded assets from disk.
func AssetOverrideMiddleware(root string, strip int, next fasthttp.RequestHandler) fasthttp.RequestHandler {
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
