package server

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // Usage is for collision avoidance not security.
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

	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

var (
	//go:embed public_html
	assets embed.FS

	//go:embed locales
	locales embed.FS
)

func newPublicHTMLEmbeddedHandler() fasthttp.RequestHandler {
	etags := map[string][]byte{}

	getEmbedETags(assets, assetsRoot, etags)

	return func(ctx *fasthttp.RequestCtx) {
		p := path.Join(assetsRoot, string(ctx.Path()))

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
	var (
		languages, dirs []string
	)

	etags := map[string][]byte{}

	getEmbedETags(locales, "locales", etags)

	aliases := map[string]string{
		"sv": "sv-SE",
		"zh": "zh-CN",
	}

	entries, err := locales.ReadDir("locales")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				var lng string

				switch len(entry.Name()) {
				case 2:
					lng = entry.Name()
				case 0:
					continue
				default:
					lng = strings.SplitN(entry.Name(), "-", 2)[0]
				}

				if !utils.IsStringInSlice(entry.Name(), dirs) {
					dirs = append(dirs, entry.Name())
				}

				if utils.IsStringInSlice(lng, languages) {
					continue
				}

				languages = append(languages, lng)
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

		var asset string

		if alias, ok := aliases[locale]; ok {
			asset = fmt.Sprintf("locales/%s/%s.json", alias, namespace)
		} else if utils.IsStringInSlice(locale, dirs) {
			asset = fmt.Sprintf("locales/%s/%s.json", locale, namespace)
		} else {
			l := language + "-" + strings.ToUpper(language)

			if utils.IsStringInSlice(l, dirs) {
				asset = fmt.Sprintf("locales/%s-%s/%s.json", language, strings.ToUpper(language), namespace)
			}
		}

		if asset == "" {
			handlers.SetStatusCodeResponse(ctx, fasthttp.StatusNotFound)

			return
		}

		if etag, ok := etags[asset]; ok {
			ctx.Response.Header.SetBytesKV(headerETag, etag)
			ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueCacheControlETaggedAssets)

			if bytes.Equal(etag, ctx.Request.Header.PeekBytes(headerIfNoneMatch)) {
				ctx.SetStatusCode(fasthttp.StatusNotModified)

				return
			}
		}

		var data []byte

		if data, err = locales.ReadFile(asset); err != nil {
			if utils.IsStringInSlice(language, languages) {
				data = []byte("{}")
			} else {
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

		sum := sha1.New() //nolint:gosec // Usage is for collision avoidance not security.

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
