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
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

var (
	//go:embed public_html
	assets embed.FS

	//go:embed locales
	locales embed.FS
)

func newPublicHTMLEmbeddedHandler() fasthttp.RequestHandler {
	return newEmbeddedHandler(assets, assetsRoot)
}

func newEmbeddedHandler(embedFS embed.FS, root string) fasthttp.RequestHandler {
	etags := map[string][]byte{}

	getEmbedETags(embedFS, root, etags)

	encoded := getEmbedCompressed(embedFS, root)

	return func(ctx *fasthttp.RequestCtx) {
		p := path.Join(root, string(ctx.Path()))

		var variant *compressedAsset

		if variants, ok := encoded[p]; ok {
			ctx.Response.Header.SetBytesKV(headerVary, headerValueVaryAcceptEncoding)

			variant = getAcceptedCompressedAsset(ctx, variants)
		}

		etag, ok := etags[p]

		if variant != nil {
			etag, ok = variant.etag, true
		}

		if ok {
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

		if variant != nil {
			data = variant.data
		} else if data, err = embedFS.ReadFile(p); err != nil {
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

		if variant != nil {
			ctx.Response.Header.SetBytesKV(headerContentEncoding, variant.encoding)
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

//nolint:gocyclo
func newLocalesPathResolver() (handler func(ctx *middlewares.AutheliaCtx) (supported bool, asset string, embedded bool), err error) {
	var (
		languages, embeddedDirs []string

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

			if !utils.IsStringInSlice(entry.Name(), embeddedDirs) {
				embeddedDirs = append(embeddedDirs, entry.Name())
			}

			if utils.IsStringInSlice(lng, languages) {
				continue
			}

			languages = append(languages, lng)
		}
	}

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
		case utils.IsStringInSlice(locale, embeddedDirs):
			return true, fmt.Sprintf("locales/%s/%s.json", locale, namespace), true
		case utils.IsStringInSlice(ll, embeddedDirs):
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
		supported, asset, useEmbedded := getAssetName(ctx)

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

		if useEmbedded {
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

type compressedAsset struct {
	encoding []byte
	etag     []byte
	data     []byte
}

func isPreCompressible(p string) bool {
	return utils.IsStringInSlice(path.Ext(p), extsCompressible) && !utils.IsStringInSlice(p, templates.AssetPathsTemplated)
}

func getEmbedCompressed(embedFS embed.FS, root string) (encoded map[string][]compressedAsset) {
	encoded = map[string][]compressedAsset{}

	_ = fs.WalkDir(embedFS, root, func(p string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !isPreCompressible(p) {
			return nil
		}

		var data []byte

		// Files which are unreadable are handled by the request handler, and files which are tiny aren't worth
		// compressing as the framing overhead outweighs the saving.
		if data, err = embedFS.ReadFile(p); err != nil || len(data) < compressionMinSize {
			return nil
		}

		// Ordered by preference, since the first variant the client accepts is the one served.
		for _, variant := range []compressedAsset{
			{encoding: encodingBrotli, data: fasthttp.AppendBrotliBytesLevel(nil, data, compressionLevelBrotli)},
			{encoding: encodingGzip, data: fasthttp.AppendGzipBytesLevel(nil, data, compressionLevelGzip)},
		} {
			if len(variant.data) >= len(data) {
				continue
			}

			variant.etag = generateEtag(variant.data)

			encoded[p] = append(encoded[p], variant)
		}

		return nil
	})

	return encoded
}

func getAcceptedCompressedAsset(ctx *fasthttp.RequestCtx, variants []compressedAsset) (variant *compressedAsset) {
	header := ctx.Request.Header.PeekBytes(headerAcceptEncoding)

	if len(header) == 0 {
		return nil
	}

	// A quality of zero is a refusal, so the comparison also excludes the codings the client has explicitly ruled out.
	// Ties keep the earlier variant since the slice is ordered by our own preference.
	var quality float64

	for i := range variants {
		if q := getAcceptEncodingQuality(header, variants[i].encoding); q > quality {
			variant, quality = &variants[i], q
		}
	}

	return variant
}

func getAcceptEncodingQuality(header, coding []byte) (quality float64) {
	var element []byte

	for len(header) != 0 {
		if i := bytes.IndexByte(header, ','); i >= 0 {
			element, header = header[:i], header[i+1:]
		} else {
			element, header = header, nil
		}

		name, q := element, 1.0

		if i := bytes.IndexByte(element, ';'); i >= 0 {
			name, q = element[:i], getAcceptEncodingParamsQuality(element[i+1:])
		}

		switch name = bytes.TrimSpace(name); {
		case bytes.EqualFold(name, coding):
			return q
		case bytes.Equal(name, encodingWildcard):
			// A coding the header names explicitly takes precedence over the wildcard, so the wildcard is only
			// remembered as a fallback and the scan continues in case the coding is named later on.
			quality = q
		}
	}

	return quality
}

func getAcceptEncodingParamsQuality(params []byte) float64 {
	var param []byte

	for len(params) != 0 {
		if i := bytes.IndexByte(params, ';'); i >= 0 {
			param, params = params[:i], params[i+1:]
		} else {
			param, params = params, nil
		}

		i := bytes.IndexByte(param, '=')

		if i < 0 || !bytes.EqualFold(bytes.TrimSpace(param[:i]), paramQuality) {
			continue
		}

		// A malformed or out of range quality is treated as a refusal rather than guessed at, which just means the
		// identity representation is served.
		quality, err := strconv.ParseFloat(string(bytes.TrimSpace(param[i+1:])), 64)

		if err != nil || quality < 0 || quality > 1 {
			return 0
		}

		return quality
	}

	return 1
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

func newLocalesListHandler() (handler func(ctx *middlewares.AutheliaCtx), err error) {
	var (
		data []byte
	)

	localeInfo, err := utils.GetEmbeddedLanguages(locales)
	if err != nil {
		return nil, fmt.Errorf("error occurred initializing the locale list handler: error occurred loading embedded languages: %w", err)
	}

	data, err = json.Marshal(middlewares.OKResponse{Status: "OK", Data: localeInfo})
	if err != nil {
		return nil, fmt.Errorf("error occurred initializing the locale list handler: error occurred marshaling the locale list: %w", err)
	}

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

func generateEtag(payload []byte) []byte {
	sum := sha1.New() //nolint:gosec // Usage is for collision avoidance not security.
	sum.Write(payload)

	// The digest is wrapped in double quotes as an ETag is an opaque quoted string, and an unquoted value is liable
	// to be dropped or rewritten by intermediaries.
	return []byte(fmt.Sprintf(`"%x"`, sum.Sum(nil)))
}
