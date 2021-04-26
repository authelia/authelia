package handlers

// TOTPRegistrationAction is the string representation of the action for which the token has been produced.
const TOTPRegistrationAction = "RegisterTOTPDevice"

// U2FRegistrationAction is the string representation of the action for which the token has been produced.
const U2FRegistrationAction = "RegisterU2FDevice"

// ResetPasswordAction is the string representation of the action for which the token has been produced.
const ResetPasswordAction = "ResetPassword"

const authPrefix = "Basic "

// ProxyAuthorizationHeader is the basic-auth HTTP header Authelia utilises.
const ProxyAuthorizationHeader = "Proxy-Authorization"

// AuthorizationHeader is the basic-auth HTTP header Authelia utilises with "auth=basic" query param.
const AuthorizationHeader = "Authorization"

// SessionUsernameHeader is used as additional protection to validate a user for things like pam_exec.
const SessionUsernameHeader = "Session-Username"

const remoteUserHeader = "Remote-User"
const remoteNameHeader = "Remote-Name"
const remoteEmailHeader = "Remote-Email"
const remoteGroupsHeader = "Remote-Groups"

const (
	// Forbidden means the user is forbidden the access to a resource.
	Forbidden authorizationMatching = iota
	// NotAuthorized means the user can access the resource with more permissions.
	NotAuthorized authorizationMatching = iota
	// Authorized means the user is authorized given her current permissions.
	Authorized authorizationMatching = iota
)

const operationFailedMessage = "Operation failed."
const authenticationFailedMessage = "Authentication failed. Check your credentials."
const userBannedMessage = "Please retry in a few minutes."
const unableToRegisterOneTimePasswordMessage = "Unable to set up one-time passwords." //nolint:gosec
const unableToRegisterSecurityKeyMessage = "Unable to register your security key."
const unableToResetPasswordMessage = "Unable to reset your password."
const mfaValidationFailedMessage = "Authentication failed, please retry later."

const ldapPasswordComplexityCode = "0000052D."

var ldapPasswordComplexityCodes = []string{"0000052D"}
var ldapPasswordComplexityErrors = []string{"LDAP Result Code 19 \"Constraint Violation\": Password fails quality checking policy"}

const testInactivity = "10"
const testRedirectionURL = "http://redirection.local"
const testResultAllow = "allow"
const testUsername = "john"

const movingAverageWindow = 10
const msMinimumDelay1FA = float64(250)
const msMaximumRandomDelay = int64(85)

// OIDC constants.
const (
	oidcWellKnownPath  = "/.well-known/openid-configuration"
	oidcJWKsPath       = "/api/oidc/jwks"
	oidcAuthorizePath  = "/api/oidc/authorize"
	oidcTokenPath      = "/api/oidc/token"
	oidcIntrospectPath = "/api/oidc/introspect"
	oidcRevokePath     = "/api/oidc/revoke"

	// Note: If you change this const you must also do so in the frontend at web/src/services/Api.ts.
	oidcConsentPath = "/api/oidc/consent"
)

const (
	accept = "accept"
	reject = "reject"
)

var scopeDescriptions = map[string]string{
	"openid":  "Use OpenID to verify your identity",
	"email":   "Access your email addresses",
	"profile": "Access your username",
	"groups":  "Access your group membership",
}

var audienceDescriptions = map[string]string{}
