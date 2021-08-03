package server

const embeddedAssets = "public_html/"
const swaggerAssets = embeddedAssets + "api/"
const apiFile = "openapi.yml"
const indexFile = "index.html"

const dev = "dev"

const healthCheckEnv = `# Written by Authelia Process
HEALTHCHECK_SCHEME=%s
HEALTHCHECK_HOST=%s
HEALTHCHECK_PORT=%d
HEALTHCHECK_PATH=%s
`
