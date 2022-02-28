package oidc

// Scope strings.
const (
	ScopeOpenID  = "openid"
	ScopeProfile = "profile"
	ScopeEmail   = "email"
	ScopeGroups  = "groups"
)

// Claim strings.
const (
	ClaimGroups            = "groups"
	ClaimDisplayName       = "name"
	ClaimPreferredUsername = "preferred_username"
	ClaimEmail             = "email"
	ClaimEmailVerified     = "email_verified"
	ClaimEmailAlts         = "alt_emails"
)

// Paths.
const (
	WellKnownOpenIDConfigurationPath      = "/.well-known/openid-configuration"
	WellKnownOAuthAuthorizationServerPath = "/.well-known/oauth-authorization-server"

	JWKsPath          = "/api/oidc/jwks"
	AuthorizationPath = "/api/oidc/authorize"
	TokenPath         = "/api/oidc/token" //nolint:gosec // This is not a hard coded credential, it's a path.
	IntrospectionPath = "/api/oidc/introspect"
	RevocationPath    = "/api/oidc/revoke"
	UserinfoPath      = "/api/oidc/userinfo"
)
