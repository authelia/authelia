package oidc

var scopeDescriptions = map[string]string{
	"openid":  "Use OpenID to verify your identity",
	"email":   "Access your email addresses",
	"profile": "Access your display name",
	"groups":  "Access your group membership",
}

var audienceDescriptions = map[string]string{}

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
	ClaimAltEmails         = "alt_emails"
)
