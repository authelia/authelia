package asset

import (
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

//go:embed locales
var locales embed.FS

// NewLocalesEmbeddedFS creates a handler for the locales assets.
func NewLocalesEmbeddedFS(path string) (handler middlewares.RequestHandler) {
	// TODO: implement a method to load from the assets path.

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
		var (
			lng, ns                              []byte
			language, variant, locale, namespace string
		)

		lng = ctx.RequestCtx.QueryArgs().Peek("lng")
		ns = ctx.RequestCtx.QueryArgs().Peek("ns")

		if language, variant, locale, namespace, err = localeDecodeLngAndNS(lng, ns); err != nil {
			fmt.Printf("%v, %v\n", lng, ns)

			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}

		fmt.Printf("%s %s %s %s", language, variant, locale, namespace)

		var data []byte

		if data, err = locales.ReadFile(fmt.Sprintf("locales/%s/%s.json", locale, namespace)); err != nil {
			if variant != "" && utils.IsStringInSliceFold(language, languages) {
				data = jsonEmptyObject
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

func localeDecodeLngAndNS(lng, ns []byte) (language, variant, locale, namespace string, err error) {
	locale, namespace = string(lng), strings.ToLower(string(ns))

	if len(namespace) == 0 {
		namespace = defaultLocaleNS
	}

	parts := strings.SplitN(locale, "-", 2)

	if len(parts) == 0 {
		return defaultLocaleLanguage, variant, defaultLocaleLanguage, namespace, nil
	}

	language = strings.ToLower(parts[0])

	if len(parts) == 2 {
		variant = strings.ToUpper(parts[1])
	}

	if len(language) != 2 || len(variant) != 0 && len(variant) != 2 {
		return "", "", "", "", errors.New("invalid lng")
	}

	return language, variant, locale, namespace, nil
}
