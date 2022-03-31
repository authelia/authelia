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
		path := ctx.Path()
		stripPath := fasthttp.NewPathSlashesStripper(strip)(ctx)
		fullPath := filepath.Join(root, string(stripPath))

		fmt.Printf("path: %s, stripped: %s, root: %s, full: %s\n", path, stripPath, root, fullPath)

		if root == "" {
			next(ctx)

			return
		}

		_, err := os.Stat(filepath.Join(root, string(fasthttp.NewPathSlashesStripper(strip)(ctx))))
		if err != nil {
			fmt.Println(err)

			next(ctx)

			return
		}

		fasthttp.FSHandler(root, strip)(ctx)
	}
}
