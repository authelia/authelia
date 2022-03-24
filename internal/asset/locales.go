package asset

import (
	"embed"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
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
		language := ctx.UserValue("language").(string)
		variant := ctx.UserValue("variant")
		namespace := ctx.UserValue("namespace").(string)

		fmt.Printf("lang: %s, variant: %v, namespace: %s\n", language, variant, namespace)

		locale := language
		if variant != nil {
			locale += "-" + variant.(string)
		}

		var data []byte

		if data, err = locales.ReadFile(fmt.Sprintf("locales/%s/%s.json", locale, namespace)); err != nil {
			if variant != nil {
				for _, lang := range languages {
					if strings.EqualFold(language, lang) {
						data = blank
					}
				}
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
