package server

const embeddedAssets = "public_html/"
const swaggerAssets = embeddedAssets + "api/"
const apiFile = "openapi.yml"
const indexFile = "index.html"

const dev = "dev"

const healthCheckEnv = `# Written by Authelia Process
X_AUTHELIA_HEALTHCHECK_SCHEME=%s
X_AUTHELIA_HEALTHCHECK_HOST=%s
X_AUTHELIA_HEALTHCHECK_PORT=%d
X_AUTHELIA_HEALTHCHECK_PATH=%s
`
