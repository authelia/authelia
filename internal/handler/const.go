package handler

import (
	"github.com/valyala/fasthttp"
)

const (
	// ActionTOTPRegistration is the string representation of the action for which the token has been produced.
	ActionTOTPRegistration = "RegisterTOTPDevice"

	// ActionWebauthnRegistration is the string representation of the action for which the token has been produced.
	ActionWebauthnRegistration = "RegisterWebauthnDevice"

	// ActionResetPassword is the string representation of the action for which the token has been produced.
	ActionResetPassword = "ResetPassword"
)

var (
	headerAuthorization   = []byte(fasthttp.HeaderAuthorization)
	headerWWWAuthenticate = []byte(fasthttp.HeaderWWWAuthenticate)

	headerProxyAuthorization = []byte(fasthttp.HeaderProxyAuthorization)

	headerSessionUsername = []byte("Session-Username")

	headerRemoteUser   = []byte("Remote-User")
	headerRemoteGroups = []byte("Remote-Groups")
	headerRemoteName   = []byte("Remote-Name")
	headerRemoteEmail  = []byte("Remote-Email")
)

var (
	headerWWWAuthenticateValueBasic = []byte(`Basic realm="Authentication Required"`)
)

const (
	headerAuthorizationSchemeBasic = "Basic"
)

const (
	headerAcceptsMIMETextHTML = "text/html"
)

var (
	queryArgRD    = []byte("rd")
	queryArgProxy = []byte("proxy")
	queryArgAuth  = []byte("auth")
)

var (
	valueEmpty = []byte("")
	valueBasic = []byte("basic")
)

const (
	queryArgProxyValueNGINX = "nginx"
)

const (
	// Forbidden means the user is forbidden the access to a resource.
	Forbidden authorizationMatching = iota

	// NotAuthorized means the user can access the resource with more permissions.
	NotAuthorized

	// Authorized means the user is authorized given her current permissions.
	Authorized
)

const (
	messageOperationFailed                 = "Operation failed."
	messageAuthenticationFailed            = "Authentication failed. Check your credentials."
	messageUnableToRegisterOneTimePassword = "Unable to set up one-time passwords." //nolint:gosec
	messageUnableToRegisterSecurityKey     = "Unable to register your security key."
	messageUnableToResetPassword           = "Unable to reset your password."
	messageMFAValidationFailed             = "Authentication failed, please retry later."
	messagePasswordWeak                    = "Your supplied password does not meet the password policy requirements"
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

// Duo constants.
const (
	allow  = "allow"
	deny   = "deny"
	enroll = "enroll"
	auth   = "auth"
)

const (
	accept = "accept"
	reject = "reject"
)

const ldapPasswordComplexityCode = "0000052D."

var ldapPasswordComplexityCodes = []string{
	"0000052D", "SynoNumber", "SynoMixedCase", "SynoExcludeNameDesc", "SynoSpecialChar",
}

var ldapPasswordComplexityErrors = []string{
	"LDAP Result Code 19 \"Constraint Violation\": Password fails quality checking policy",
	"LDAP Result Code 19 \"Constraint Violation\": Password is too young to change",
}
