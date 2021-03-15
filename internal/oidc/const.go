package oidc

const (
	fallbackOIDCIssuer = "https://login.example.com:8080"
	wellKnownPath      = "/.well-known/openid-configuration"
	jwksPath           = "/.well-known/jwks.json"
	authPath           = "/api/oidc/auth"
	tokenPath          = "/api/oidc/token"
	userinfoPath       = "/api/oidc/userinfo"
	consentPath        = "/api/oidc/consent"
	introspectPath     = "/api/oidc/introspect"
	revokePath         = "/api/oidc/revoke"
)
