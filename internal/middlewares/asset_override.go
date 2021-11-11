package middlewares

import (
	"os"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/utils"
)

// AssetOverrideMiddleware allows overriding and serving of specific embedded assets from disk.
func AssetOverrideMiddleware(assetPath string, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		uri := string(ctx.RequestURI())
		file := uri[strings.LastIndex(uri, "/")+1:]

		if assetPath != "" && utils.IsStringInSlice(file, validOverrideAssets) {
			_, err := os.Stat(assetPath + file)
			if err != nil {
				next(ctx)
			} else {
				fasthttp.FSHandler(assetPath, strings.Count(uri, "/")-1)(ctx)
			}
		} else {
			next(ctx)
		}
	}
}
