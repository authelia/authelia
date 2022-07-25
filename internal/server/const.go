package server

import (
	"github.com/valyala/fasthttp"
)

const (
	embeddedAssets = "public_html/"
	swaggerAssets  = embeddedAssets + "api/"
	apiFile        = "openapi.yml"
	indexFile      = "index.html"
	logoFile       = "logo.png"
)

var (
	rootFiles    = []string{"manifest.json", "robots.txt"}
	swaggerFiles = []string{
		"favicon-16x16.png",
		"favicon-32x32.png",
		"index.css",
		"oauth2-redirect.html",
		"swagger-initializer.js",
		"swagger-ui-bundle.js",
		"swagger-ui-bundle.js.map",
		"swagger-ui-es-bundle-core.js",
		"swagger-ui-es-bundle-core.js.map",
		"swagger-ui-es-bundle.js",
		"swagger-ui-es-bundle.js.map",
		"swagger-ui-standalone-preset.js",
		"swagger-ui-standalone-preset.js.map",
		"swagger-ui.css",
		"swagger-ui.css.map",
		"swagger-ui.js",
		"swagger-ui.js.map",
	}

	// Directories excluded from the not found handler proceeding to the next() handler.
	httpServerDirs = []struct {
		name, prefix string
	}{
		{name: "/api", prefix: "/api/"},
		{name: "/.well-known", prefix: "/.well-known/"},
		{name: "/static", prefix: "/static/"},
		{name: "/locales", prefix: "/locales/"},
	}
)

const (
	dev         = "dev"
	f           = "false"
	t           = "true"
	localhost   = "localhost"
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

var (
	headerETag         = []byte(fasthttp.HeaderETag)
	headerIfNoneMatch  = []byte(fasthttp.HeaderIfNoneMatch)
	headerCacheControl = []byte(fasthttp.HeaderCacheControl)

	headerValueCacheControlETaggedAssets = []byte("public, max-age=0, must-revalidate")
)

const healthCheckEnv = `# Written by Authelia Process
X_AUTHELIA_HEALTHCHECK=1
X_AUTHELIA_HEALTHCHECK_SCHEME=%s
X_AUTHELIA_HEALTHCHECK_HOST=%s
X_AUTHELIA_HEALTHCHECK_PORT=%d
X_AUTHELIA_HEALTHCHECK_PATH=%s
`

const (
	cspDefaultTemplate  = "default-src 'self'%s; frame-src 'none'; object-src 'none'; style-src 'self' 'nonce-%s'; frame-ancestors 'none'; base-uri 'self'"
	cspNoncePlaceholder = "${NONCE}"
)
