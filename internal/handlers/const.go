package handlers

const (
	// ActionTOTPRegistration is the string representation of the action for which the token has been produced.
	ActionTOTPRegistration = "RegisterTOTPDevice"

	// ActionU2FRegistration is the string representation of the action for which the token has been produced.
	ActionU2FRegistration = "RegisterU2FDevice"

	// ActionResetPassword is the string representation of the action for which the token has been produced.
	ActionResetPassword = "ResetPassword"
)

const (
	// HeaderProxyAuthorization is the basic-auth HTTP header Authelia utilises.
	HeaderProxyAuthorization = "Proxy-Authorization"

	// HeaderAuthorization is the basic-auth HTTP header Authelia utilises with "auth=basic" query param.
	HeaderAuthorization = "Authorization"

	// HeaderSessionUsername is used as additional protection to validate a user for things like pam_exec.
	HeaderSessionUsername = "Session-Username"

	headerRemoteUser   = "Remote-User"
	headerRemoteName   = "Remote-Name"
	headerRemoteEmail  = "Remote-Email"
	headerRemoteGroups = "Remote-Groups"
)

const (
	// Forbidden means the user is forbidden the access to a resource.
	Forbidden authorizationMatching = iota
	// NotAuthorized means the user can access the resource with more permissions.
	NotAuthorized authorizationMatching = iota
	// Authorized means the user is authorized given her current permissions.
	Authorized authorizationMatching = iota
)

const (
	messageOperationFailed                 = "Operation failed."
	messageAuthenticationFailed            = "Authentication failed. Check your credentials."
	messageUnableToRegisterOneTimePassword = "Unable to set up one-time passwords." //nolint:gosec
	messageUnableToRegisterSecurityKey     = "Unable to register your security key."
	messageUnableToResetPassword           = "Unable to reset your password."
	messageMFAValidationFailed             = "Authentication failed, please retry later."
)

const (
	logFmtErrParseRequestBody     = "Failed to parse %s request body: %+v"
	logFmtErrWriteResponseBody    = "Failed to write %s response body for user '%s': %+v"
	logFmtErrRegulationFail       = "Failed to perform %s authentication regulation for user '%s': %+v"
	logFmtErrSessionRegenerate    = "Could not regenerate session during %s authentication for user '%s': %+v"
	logFmtErrSessionReset         = "Could not reset session during %s authentication for user '%s': %+v"
	logFmtErrSessionSave          = "Could not save session with the %s during %s authentication for user '%s': %+v"
	logFmtErrObtainProfileDetails = "Could not obtain profile details during %s authentication for user '%s': %+v"
	logFmtTraceProfileDetails     = "Profile details for user '%s' => groups: %s, emails %s"
)

const (
	testInactivity     = "10"
	testRedirectionURL = "http://redirection.local"
	testResultAllow    = "allow"
	testUsername       = "john"
)

const (
	loginDelayMovingAverageWindow            = 10
	loginDelayMinimumDelayMilliseconds       = float64(250)
	loginDelayMaximumRandomDelayMilliseconds = int64(85)
)

// OIDC constants.
const (
	pathOpenIDConnectWellKnown = "/.well-known/openid-configuration"

	pathOpenIDConnectJWKs          = "/api/oidc/jwks"
	pathOpenIDConnectAuthorization = "/api/oidc/authorize"
	pathOpenIDConnectToken         = "/api/oidc/token" //nolint:gosec // This is not a hard coded credential, it's a path.
	pathOpenIDConnectIntrospection = "/api/oidc/introspect"
	pathOpenIDConnectRevocation    = "/api/oidc/revoke"
	pathOpenIDConnectUserinfo      = "/api/oidc/userinfo"

	// Note: If you change this const you must also do so in the frontend at web/src/services/Api.ts.
	pathOpenIDConnectConsent = "/api/oidc/consent"
)

const (
	accept = "accept"
	reject = "reject"
)

const authPrefix = "Basic "

const ldapPasswordComplexityCode = "0000052D."

var ldapPasswordComplexityCodes = []string{
	"0000052D", "SynoNumber", "SynoMixedCase", "SynoExcludeNameDesc", "SynoSpecialChar",
}

var ldapPasswordComplexityErrors = []string{
	"LDAP Result Code 19 \"Constraint Violation\": Password fails quality checking policy",
	"LDAP Result Code 19 \"Constraint Violation\": Password is too young to change",
}
