package server

import (
	"regexp"

	"github.com/valyala/fasthttp"
)

const (
	assetsRoot = "public_html"

	fileLogo = "logo.png"

	extHTML = ".html"
	extJS   = ".js"
	extJSON = ".json"
	extYML  = ".yml"
)

const (
	pathAuthz           = "/api/authz"
	pathAuthzLegacy     = "/api/verify"
	pathParamAuthzEnvoy = "{extauthz:*}"
)

var (
	filesRoot    = []string{"manifest.json", "robots.txt"}
	filesSwagger = []string{
		"favicon-16x16.png",
		"favicon-32x32.png",
		"index.css",
		"oauth2-redirect.html",
		"swagger-ui-bundle.js",
		"swagger-ui-standalone-preset.js",
		"swagger-ui.css",
	}

	// Directories excluded from the not found handler proceeding to the next() handler.
	dirsHTTPServer = []struct {
		name, prefix string
	}{
		{name: "/api", prefix: prefixAPI},
		{name: "/.well-known", prefix: "/.well-known/"},
		{name: "/static", prefix: "/static/"},
		{name: "/locales", prefix: "/locales/"},
	}
)

const (
	strFalse    = "false"
	strTrue     = "true"
	localhost   = "localhost"
	schemeHTTP  = "http"
	schemeHTTPS = "https"
	prefixAPI   = "/api/"
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
	errFmtMessageServerReadBuffer  = "Request from client exceeded the server read buffer. The read buffer can be adjusted by modifying the '%s.buffers.read' configuration value."
	errMessageServerRequestTimeout = "Request timeout occurred while handling request from client."
	errMessageServerNetwork        = "An unknown network error occurred while handling a request from client."
	errFmtMessageServerTLSVersion  = "A %s connection handshake occurred on a non-TLS listener."
	errMessageServerGeneric        = "An unknown error occurred while handling a request from client."
)

const (
	tmplCSPSwagger = "default-src 'self'; img-src 'self' https://validator.swagger.io data:; object-src 'none'; script-src 'self' 'nonce-%s'; style-src 'self' 'sha256-RL3ie0nH+Lzz2YNqQN83mnU0J1ot4QL7b99vMdIX99w='; base-uri 'self'"
)

var (
	reTLSRequestOnPlainTextSocketErr = regexp.MustCompile(`contents: \\x16\\x([a-fA-F0-9]{2})\\x([a-fA-F0-9]{2})`)
	reValidLanguageCodes             = regexp.MustCompile(`^[a-zA-Z0-9-]{1,15}$`)
)
