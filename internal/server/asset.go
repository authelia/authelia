package server

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // Usage is for collision avoidance not security.
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
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

		middlewares.SetStandardSecurityHeaders(ctx)

		contentType := mime.TypeByExtension(path.Ext(p))
		if len(contentType) == 0 {
			contentType = http.DetectContentType(data)
		}

		ctx.SetContentType(contentType)

		switch {
		case ctx.IsHead():
			ctx.Response.ResetBody()
			ctx.Response.SkipBody = true
			ctx.Response.Header.Set(fasthttp.HeaderContentLength, strconv.Itoa(len(data)))
		default:
			ctx.SetBody(data)
		}
	}
}

//nolint:gocyclo
func newLocalesPathResolver() func(ctx *fasthttp.RequestCtx) (supported bool, asset string) {
	var (
		languages, dirs []string
		aliases         = map[string]string{}
	)

	entries, err := locales.ReadDir("locales")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				var lng string

				if lng, err = utils.GetLocaleParentOrBaseString(entry.Name()); err != nil {
					continue
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

	// generate list of macro to micro locale aliases.
	if localeInfo, err := utils.GetLanguagesFromEmbedFS(locales); err == nil {
		for _, v := range localeInfo.Languages {
			if v.Parent == "" {
				continue
			}

			_, ok := aliases[v.Parent]

			if !ok {
				aliases[v.Parent] = v.Locale
			}
		}
	}

	aliases["no"] = "nb"

	return func(ctx *fasthttp.RequestCtx) (supported bool, asset string) {
		var language, namespace, variant, locale string

		language, namespace = ctx.UserValue("language").(string), ctx.UserValue("namespace").(string)

		if v := ctx.UserValue("variant"); v != nil {
			variant = v.(string)
			locale = fmt.Sprintf("%s-%s", language, variant)
		} else {
			locale = language
		}

		ll := language + "-" + strings.ToUpper(language)

		alias, useAlias := aliases[locale]
		if useAlias {
			if language, err = utils.GetLocaleParentOrBaseString(alias); err != nil {
				return false, ""
			}
		}

		if !utils.IsStringInSlice(language, languages) {
			return false, ""
		}

		switch {
		case useAlias:
			return true, fmt.Sprintf("locales/%s/%s.json", alias, namespace)
		case utils.IsStringInSlice(locale, dirs):
			return true, fmt.Sprintf("locales/%s/%s.json", locale, namespace)
		case utils.IsStringInSlice(ll, dirs):
			return true, fmt.Sprintf("locales/%s-%s/%s.json", language, strings.ToUpper(language), namespace)
		default:
			return true, fmt.Sprintf("locales/%s/%s.json", locale, namespace)
		}
	}
}

func newLocalesEmbeddedHandler() (handler fasthttp.RequestHandler) {
	etags := map[string][]byte{}

	getEmbedETags(locales, "locales", etags)

	getAssetName := newLocalesPathResolver()

	return func(ctx *fasthttp.RequestCtx) {
		supported, asset := getAssetName(ctx)

		if !supported {
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

		var (
			data []byte
			err  error
		)

		if data, err = locales.ReadFile(asset); err != nil {
			data = []byte("{}")
		}

		middlewares.SetStandardSecurityHeaders(ctx)
		middlewares.SetContentTypeApplicationJSON(ctx)

		switch {
		case ctx.IsHead():
			ctx.Response.ResetBody()
			ctx.Response.SkipBody = true
			ctx.Response.Header.Set(fasthttp.HeaderContentLength, strconv.Itoa(len(data)))
		default:
			ctx.SetBody(data)
		}
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

		etags[p] = generateEtag(data)
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

// newLocalesListHandler handles request for obtaining the available locales in backend.
func newLocalesListHandler() (handler func(ctx *middlewares.AutheliaCtx)) {
	var (
		data []byte
		err  error
	)

	// preload embedded locales.
	localeInfo, err := utils.GetLanguagesFromEmbedFS(locales)
	if err != nil {
		panic(errCantLoadLocaleInfo + err.Error())
	}

	return func(ctx *middlewares.AutheliaCtx) {
		if ctx.Configuration.CustomLocales.Enabled {
			customLocaleInfo, err := utils.GetLanguagesFromPath(ctx.Configuration.CustomLocales.Path)

			if err != nil {
				ctx.Logger.Errorf("Unable to load custom locales: %s", err)
			}

			mergeLocaleInfo(localeInfo, customLocaleInfo)
		}
		// parse embedded locales.
		data, err = json.Marshal(middlewares.OKResponse{Status: "OK", Data: localeInfo})

		if err != nil {
			ctx.Logger.Errorf("Unable to marshal locales: %s", err)
		}

		// generate etag for embedded locales.
		etag := generateEtag(data)

		ctx.Response.Header.SetBytesKV(headerETag, etag)
		ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueCacheControlETaggedAssets)

		if bytes.Equal(etag, ctx.Request.Header.PeekBytes(headerIfNoneMatch)) {
			ctx.SetStatusCode(fasthttp.StatusNotModified)
			return
		}

		switch {
		case ctx.IsHead():
			ctx.Response.ResetBody()
			ctx.Response.SkipBody = true
			ctx.Response.Header.Set(fasthttp.HeaderContentLength, strconv.Itoa(len(data)))
		default:
			ctx.SetBody(data)
		}
	}
}

// generateEtag generates a unique etag for specified payload.
func generateEtag(payload []byte) []byte {
	sum := sha1.New() //nolint:gosec // Usage is for collision avoidance not security.
	sum.Write(payload)

	return []byte(fmt.Sprintf("%x", sum.Sum(nil)))
}

func mergeLocaleInfo(embedded *utils.Languages, custom *utils.Languages) {

}
