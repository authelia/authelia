package server

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

//go:embed locales
var locales embed.FS

//go:embed public_html
var assets embed.FS

func newPublicHTMLEmbeddedHandler() fasthttp.RequestHandler {
	embeddedPath, _ := fs.Sub(assets, "public_html")

	header := []byte(fasthttp.HeaderCacheControl)
	headerValue := []byte("public, max-age=31536000, immutable")

	handler := fasthttpadaptor.NewFastHTTPHandler(http.FileServer(http.FS(embeddedPath)))

	return func(ctx *fasthttp.RequestCtx) {
		handler(ctx)

		uri := path.Base(string(ctx.Path()))

		if strings.HasPrefix(uri, "index.") {
			ext := path.Ext(uri)

			switch ext {
			case css, js:
				ctx.Response.Header.SetBytesKV(header, headerValue)
			}
		}
	}
}

func newLocalesEmbeddedHandler() (handler fasthttp.RequestHandler) {
	var languages []string

	entries, err := locales.ReadDir("locales")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && len(entry.Name()) == 2 {
				languages = append(languages, entry.Name())
			}
		}
	}

	return func(ctx *fasthttp.RequestCtx) {
		var (
			language, variant, locale, namespace string
		)

		language = ctx.UserValue("language").(string)
		namespace = ctx.UserValue("namespace").(string)
		locale = language

		if v := ctx.UserValue("variant"); v != nil {
			variant = v.(string)
			locale = fmt.Sprintf("%s-%s", language, variant)
		}

		var data []byte

		if data, err = locales.ReadFile(fmt.Sprintf("locales/%s/%s.json", locale, namespace)); err != nil {
			if variant != "" && utils.IsStringInSliceFold(language, languages) {
				data = []byte("{}")
			}

			if len(data) == 0 {
				hfsHandleErr(ctx, err)

				return
			}
		}

		middlewares.SetContentTypeApplicationJSON(ctx)

		ctx.SetBody(data)
	}
}

func hfsHandleErr(ctx *fasthttp.RequestCtx, err error) {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		writeStatus(ctx, fasthttp.StatusNotFound)
	case errors.Is(err, fs.ErrPermission):
		writeStatus(ctx, fasthttp.StatusForbidden)
	default:
		writeStatus(ctx, fasthttp.StatusInternalServerError)
	}
}

func writeStatus(ctx *fasthttp.RequestCtx, status int) {
	ctx.SetStatusCode(status)
	ctx.SetBodyString(fmt.Sprintf("%d %s", status, fasthttp.StatusMessage(status)))
}
