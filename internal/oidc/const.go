package oidc

// Scope strings.
const (
	ScopeOfflineAccess = "offline_access"
	ScopeOpenID        = "openid"
	ScopeProfile       = "profile"
	ScopeEmail         = "email"
	ScopeGroups        = "groups"
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

// Endpoints.
const (
	AuthorizationEndpoint = "authorization"
	TokenEndpoint         = "token"
	UserinfoEndpoint      = "userinfo"
	IntrospectionEndpoint = "introspection"
	RevocationEndpoint    = "revocation"
)

// Paths.
const (
	WellKnownOpenIDConfigurationPath      = "/.well-known/openid-configuration"
	WellKnownOAuthAuthorizationServerPath = "/.well-known/oauth-authorization-server"
	JWKsPath                              = "/jwks.json"

	RootPath = "/api/oidc"

	AuthorizationPath = RootPath + "/" + AuthorizationEndpoint
	TokenPath         = RootPath + "/" + TokenEndpoint
	UserinfoPath      = RootPath + "/" + UserinfoEndpoint
	IntrospectionPath = RootPath + "/" + IntrospectionEndpoint
	RevocationPath    = RootPath + "/" + RevocationEndpoint
)
