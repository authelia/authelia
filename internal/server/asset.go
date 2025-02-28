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

func newLocalesPathResolver() func(ctx *fasthttp.RequestCtx) (supported bool, asset string) {
	var (
		languages, dirs []string
	)

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

	aliases := map[string]string{
		"cs": "cs-CZ",
		"da": "da-DK",
		"el": "el-GR",
		"ja": "ja-JP",
		"nb": "nb-NO",
		"sv": "sv-SE",
		"uk": "uk-UA",
		"zh": "zh-CN",
	}

	return func(ctx *fasthttp.RequestCtx) (supported bool, asset string) {
		var language, namespace, variant, locale string

		language, namespace = ctx.UserValue("language").(string), ctx.UserValue("namespace").(string)

		if !utils.IsStringInSlice(language, languages) {
			return false, ""
		}

		if v := ctx.UserValue("variant"); v != nil {
			variant = v.(string)
			locale = fmt.Sprintf("%s-%s", language, variant)
		} else {
			locale = language
		}

		ll := language + "-" + strings.ToUpper(language)
		alias, ok := aliases[locale]

		switch {
		case ok:
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

		middlewares.SetBaseSecurityHeaders(ctx)
		middlewares.SetSecurityHeadersCSPNone(ctx)
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

// ValidateTranslations checks that translations contain required placeholders.
func ValidateTranslations() error {
	criticalTranslations := map[string]map[string][]string{
		"portal.json": {
			"You must view and accept the Privacy Policy before using {{authelia}}": {"authelia"},
			"Powered by {{authelia}}": {"authelia"},
		},
	}

	allLocales, err := locales.ReadDir("locales")
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	var validationErrors []string

	for _, currentLocaleDirectory := range allLocales {
		localeToCheck := currentLocaleDirectory.Name()

		for fileName, keysToCheck := range criticalTranslations {
			if len(keysToCheck) == 0 {
				continue
			}

			translationFile := fmt.Sprintf("locales/%s/%s", localeToCheck, fileName)
			data, err := locales.ReadFile(translationFile)

			if err != nil {
				return fmt.Errorf("failed to read required translation file %s: %w", translationFile, err)
			}

			var translations map[string]interface{}
			if err := json.Unmarshal(data, &translations); err != nil {
				return fmt.Errorf("failed to parse translation file %s: %w", translationFile, err)
			}

			for key, requiredPlaceholders := range keysToCheck {
				translationValue, exists := translations[key]
				if !exists {
					continue
				}

				translation, ok := translationValue.(string)
				if !ok {
					validationErrors = append(validationErrors,
						fmt.Sprintf("%s locale, file %s: key %s is not a string", localeToCheck, fileName, key))
					continue
				}

				for _, placeholder := range requiredPlaceholders {
					if !strings.Contains(translation, fmt.Sprintf("{{%s}}", placeholder)) {
						validationErrors = append(validationErrors,
							fmt.Sprintf("%s locale, file %s: missing placeholder {{%s}} in key %s",
								localeToCheck, fileName, placeholder, key))
					}
				}
			}
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("translation validation failed with %d errors:\n%s",
			len(validationErrors), strings.Join(validationErrors, "\n"))
	}

	return nil
}
