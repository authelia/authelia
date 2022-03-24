package asset

import (
	"embed"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

//go:embed locales
var locales embed.FS

var blank = []byte("{}")

func NewLocalesEmbeddedFS() (handler middlewares.RequestHandler) {
	var languages []string

	entries, err := locales.ReadDir("locales")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.Contains(entry.Name(), "-") {
				languages = append(languages, entry.Name())
			}
		}
	}

	return func(ctx *middlewares.AutheliaCtx) {
		lng := ctx.RequestCtx.QueryArgs().Peek("lng")
		namespace := ctx.RequestCtx.QueryArgs().Peek("ns")
		locale := string(lng)

		var language, variant string

		parts := strings.SplitN(locale, "-", 2)

		if len(parts) == 0 {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}

		language = parts[0]

		if len(parts) == 2 {
			variant = parts[1]
		}

		if len(language) != 2 || (len(variant) != 0 && len(variant) != 2) {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}

		var data []byte

		if data, err = locales.ReadFile(fmt.Sprintf("locales/%s/%s.json", locale, namespace)); err != nil {
			if variant != "" && utils.IsStringInSliceFold(language, languages) {
				data = blank
			}

			if len(data) == 0 {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
				return
			}
		}

		ctx.SetContentType("application/json")
		ctx.SetBody(data)
	}
}
