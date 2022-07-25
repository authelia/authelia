package server

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

//go:embed locales
var locales embed.FS

//go:embed public_html
var assets embed.FS

func getEmbedETags(embedFS embed.FS) (etags map[string][]byte) {
	var (
		err     error
		entries []fs.DirEntry
	)

	if entries, err = embedFS.ReadDir("public_html"); err != nil {
		return nil
	}

	var data []byte

	etags = map[string][]byte{}

	sum := sha256.New()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if data, err = embedFS.ReadFile(entry.Name()); err != nil {
			continue
		}

		sum.Write(data)

		etags[entry.Name()] = []byte(fmt.Sprintf("%x", sum.Sum(nil)))

		sum.Reset()
	}

	return etags
}

func newPublicHTMLEmbeddedHandler2() fasthttp.RequestHandler {
	etags := getEmbedETags(assets)

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

func newPublicHTMLEmbeddedHandler() fasthttp.RequestHandler {
	embeddedPath, _ := fs.Sub(assets, "public_html")

	handler := fasthttpadaptor.NewFastHTTPHandler(http.FileServer(http.FS(embeddedPath)))

	return func(ctx *fasthttp.RequestCtx) {
		handler(ctx)

		setCacheControl(ctx)
	}
}

func setCacheControl(ctx *fasthttp.RequestCtx) {
	uri := path.Base(string(ctx.Path()))

	if strings.HasPrefix(uri, "index.") {
		ext := path.Ext(uri)

		switch ext {
		case css, js:
			ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueCacheControlReact)
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
		handlers.SetStatusCodeResponse(ctx, fasthttp.StatusNotFound)
	case errors.Is(err, fs.ErrPermission):
		handlers.SetStatusCodeResponse(ctx, fasthttp.StatusForbidden)
	default:
		handlers.SetStatusCodeResponse(ctx, fasthttp.StatusInternalServerError)
	}
}
