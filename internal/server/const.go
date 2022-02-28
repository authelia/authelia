package server

const (
	embeddedAssets = "public_html/"
	swaggerAssets  = embeddedAssets + "api/"
	apiFile        = "openapi.yml"
	indexFile      = "index.html"
	logoFile       = "logo.png"
)

var rootFiles = []string{"favicon.ico", "manifest.json", "robots.txt"}

const (
	dev = "dev"
	f   = "false"
	t   = "true"
)

const healthCheckEnv = `# Written by Authelia Process
X_AUTHELIA_HEALTHCHECK=1
X_AUTHELIA_HEALTHCHECK_SCHEME=%s
X_AUTHELIA_HEALTHCHECK_HOST=%s
X_AUTHELIA_HEALTHCHECK_PORT=%d
X_AUTHELIA_HEALTHCHECK_PATH=%s
`

const (
	cspDefaultTemplate    = "default-src 'self'; object-src 'none'; style-src 'self' 'nonce-%s'"
	cspDefaultDevTemplate = "default-src 'self' 'unsafe-eval'; object-src 'none'; style-src 'self' 'nonce-%s'"
	cspNoncePlaceholder   = "${NONCE}"
)
