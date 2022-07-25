package server

import (
	"bytes"
	"crypto/sha1"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"path/filepath"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

//go:embed locales
var locales embed.FS

//go:embed public_html
var assets embed.FS

func newPublicHTMLEmbeddedHandler() fasthttp.RequestHandler {
	etags := map[string][]byte{}

	getEmbedETags(assets, "public_html", etags)

	return func(ctx *fasthttp.RequestCtx) {
		p := path.Join("public_html", string(ctx.Path()))

		if etag, ok := etags[p]; ok {
			ctx.Response.Header.SetBytesKV(headerETag, etag)
			ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueCacheControlETaggedAssets)

			if bytes.Equal(etag, ctx.Request.Header.PeekBytes(headerIfNoneMatch)) {
				ctx.SetStatusCode(fasthttp.StatusNotModified)

				return
			}
		}

		var (
			data []byte
			err  error
		)

		if data, err = assets.ReadFile(p); err != nil {
			hfsHandleErr(ctx, err)

			return
		}

		contentType := mime.TypeByExtension(path.Ext(p))
		if len(contentType) == 0 {
			contentType = http.DetectContentType(data)
		}

		ctx.SetContentType(contentType)
		ctx.SetBody(data)
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

func getEmbedETags(embedFS embed.FS, root string, etags map[string][]byte) {
	var (
		err     error
		entries []fs.DirEntry
	)

	if entries, err = embedFS.ReadDir(root); err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			getEmbedETags(embedFS, filepath.Join(root, entry.Name()), etags)

			continue
		}

		p := filepath.Join(root, entry.Name())

		var data []byte

		if data, err = embedFS.ReadFile(p); err != nil {
			continue
		}

		sum := sha1.New()

		sum.Write(data)

		etags[p] = []byte(fmt.Sprintf("%x", sum.Sum(nil)))
	}
}

func hfsHandleErr(ctx *fasthttp.RequestCtx, err error) {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		handlers.SetStatusCodeResponse(ctx, fasthttp.StatusNotFound)
	case errors.Is(err, fs.ErrPermission):
		handlers.SetStatusCodeResponse(ctx, fasthttp.StatusForbidden)
	default:
		handlers.SetStatusCodeResponse(ctx, fasthttp.StatusInternalServerError)
	}
}
