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
	"path/filepath"
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

func getEmbedETags(embedFS embed.FS, root string, etags map[string][]byte) {
	var (
		err     error
		entries []fs.DirEntry
	)

	if entries, err = embedFS.ReadDir(root); err != nil {
		fmt.Printf("readdir err for dir '%s': %v\n", root, err)

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
			fmt.Printf("readfile err for '%s': %v\n", p, err)
			continue
		}

		sum := sha256.New()

		sum.Write(data)

		etags[p] = []byte(fmt.Sprintf("%x", sum.Sum(nil)))

		fmt.Printf("%s: %s\n", p, etags[p])
	}

}

func newPublicHTMLEmbeddedHandler2() fasthttp.RequestHandler {
	etags := map[string][]byte{}

	getEmbedETags(assets, "public_html", etags)

	fmt.Printf("final etags:\n")

	for key, etag := range etags {
		fmt.Printf("%s: %s\n", key, etag)

	}

	return func(ctx *fasthttp.RequestCtx) {
		p := path.Join("public_html", string(ctx.Path()))

		fmt.Printf("looking for etag for %s\n", p)

		if etag, ok := etags[p]; ok {
			fmt.Printf("etag for %s found: %s\n", p, etag)

			ctx.Response.Header.SetBytesKV(headerETag, etag)
			ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueCacheControlETaggedAssets)

			if bytes.Equal(etag, ctx.Request.Header.PeekBytes(headerIfNoneMatch)) {
				ctx.SetStatusCode(fasthttp.StatusNotModified)

				return
			}
		} else {
			fmt.Printf("etag for %s not found\n", p)
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
