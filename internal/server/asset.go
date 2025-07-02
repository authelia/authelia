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
	"os"
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

		middlewares.SetBaseSecurityHeaders(ctx)
		middlewares.SetSecurityHeadersCSPNone(ctx)

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
func newLocalesPathResolver() (handler func(ctx *middlewares.AutheliaCtx) (supported bool, asset string, embedded bool), err error) {
	var (
		languages, embededDirs []string

		aliases map[string]string
		entries []fs.DirEntry
	)

	if entries, err = locales.ReadDir("locales"); err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			var lng string

			if lng, err = utils.GetLocaleParentOrBaseString(entry.Name()); err != nil {
				continue
			}

			if !utils.IsStringInSlice(entry.Name(), embededDirs) {
				embededDirs = append(embededDirs, entry.Name())
			}

			if utils.IsStringInSlice(lng, languages) {
				continue
			}

			languages = append(languages, lng)
		}
	}

	// generate list of macro to micro locale aliases.
	var languagesInfo *utils.Languages

	if languagesInfo, err = utils.GetEmbeddedLanguages(locales); err != nil {
		return nil, err
	}

	aliases = map[string]string{
		"cs": "cs-CZ",
		"da": "da-DK",
		"el": "el-GR",
		"ja": "ja-JP",
		"nb": "nb-NO",
		"sv": "sv-SE",
		"uk": "uk-UA",
		"zh": "zh-CN",
		"no": "no-NO",
	}

	for _, v := range languagesInfo.Languages {
		if v.Parent == "" {
			continue
		}

		_, ok := aliases[v.Parent]

		if !ok {
			aliases[v.Parent] = v.Locale
		}
	}

	return func(ctx *middlewares.AutheliaCtx) (supported bool, asset string, embedded bool) {
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
				return false, "", false
			}
		}

		if !utils.IsStringInSlice(language, languages) {
			return false, "", false
		}

		switch {
		case useAlias:
			return true, fmt.Sprintf("locales/%s/%s.json", alias, namespace), true
		case utils.IsStringInSlice(locale, embededDirs):
			return true, fmt.Sprintf("locales/%s/%s.json", locale, namespace), true
		case utils.IsStringInSlice(ll, embededDirs):
			return true, fmt.Sprintf("locales/%s-%s/%s.json", language, strings.ToUpper(language), namespace), true
		default:
			return true, fmt.Sprintf("locales/%s/%s.json", locale, namespace), true
		}
	}, nil
}

func newLocalesEmbeddedHandler() (handler func(ctx *middlewares.AutheliaCtx), err error) {
	etags := map[string][]byte{}

	getEmbedETags(locales, "locales", etags)

	var getAssetName func(ctx *middlewares.AutheliaCtx) (supported bool, asset string, embedded bool)

	if getAssetName, err = newLocalesPathResolver(); err != nil {
		return nil, fmt.Errorf("error occurred initializing the embedded locales handler: %w", err)
	}

	return func(ctx *middlewares.AutheliaCtx) {
		supported, asset, useEmbeded := getAssetName(ctx)

		if !supported {
			handlers.SetStatusCodeResponse(ctx.RequestCtx, fasthttp.StatusNotFound)

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

		if useEmbeded {
			if data, err = locales.ReadFile(asset); err != nil {
				data = []byte("{}")
			}
		} else {
			fileSystem := os.DirFS(filepath.Dir(asset))

			if data, err = fs.ReadFile(fileSystem, filepath.Base(asset)); err != nil {
				data = []byte("{}")
			}
		}

		middlewares.SetBaseSecurityHeaders(ctx.RequestCtx)
		middlewares.SetSecurityHeadersCSPNone(ctx.RequestCtx)
		middlewares.SetContentTypeApplicationJSON(ctx.RequestCtx)

		switch {
		case ctx.IsHead():
			ctx.Response.ResetBody()
			ctx.Response.SkipBody = true
			ctx.Response.Header.Set(fasthttp.HeaderContentLength, strconv.Itoa(len(data)))
		default:
			ctx.SetBody(data)
		}
	}, nil
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
func newLocalesListHandler() (handler func(ctx *middlewares.AutheliaCtx), err error) {
	var (
		data []byte
	)

	// preload embedded locales.
	localeInfo, err := utils.GetEmbeddedLanguages(locales)
	if err != nil {
		return nil, fmt.Errorf("error occurred initializing the locale list handler: error occurred loading embedded languages: %w", err)
	}

	// parse embedded locales.
	data, err = json.Marshal(middlewares.OKResponse{Status: "OK", Data: localeInfo})
	if err != nil {
		return nil, fmt.Errorf("error occurred initializing the locale list handler: error occurred marshalling the locale list: %w", err)
	}

	// generate etag for embedded locales.
	etag := generateEtag(data)

	return func(ctx *middlewares.AutheliaCtx) {
		ctx.Response.Header.SetBytesKV(headerETag, etag)
		ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueCacheControlETaggedAssets)

		if bytes.Equal(etag, ctx.Request.Header.PeekBytes(headerIfNoneMatch)) {
			ctx.SetStatusCode(fasthttp.StatusNotModified)
			return
		}

		middlewares.SetStandardSecurityHeaders(ctx.RequestCtx)
		middlewares.SetContentTypeApplicationJSON(ctx.RequestCtx)

		switch {
		case ctx.IsHead():
			ctx.Response.ResetBody()
			ctx.Response.SkipBody = true
			ctx.Response.Header.Set(fasthttp.HeaderContentLength, strconv.Itoa(len(data)))
		default:
			ctx.SetBody(data)
		}
	}, nil
}

// generateEtag generates a unique etag for specified payload.
func generateEtag(payload []byte) []byte {
	sum := sha1.New() //nolint:gosec // Usage is for collision avoidance not security.
	sum.Write(payload)

	return []byte(fmt.Sprintf("%x", sum.Sum(nil)))
}
