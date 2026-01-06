package handlers

import (
	"errors"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/oidc"
)

const (
	// ActionResetPassword is the string representation of the action for which the token has been produced.
	ActionResetPassword = "ResetPassword"
)

const (
	anonymous = "<anonymous>"
)

var (
	headerAuthorization   = []byte(fasthttp.HeaderAuthorization)
	headerWWWAuthenticate = []byte(fasthttp.HeaderWWWAuthenticate)

	headerProxyAuthorization = []byte(fasthttp.HeaderProxyAuthorization)
	headerProxyAuthenticate  = []byte(fasthttp.HeaderProxyAuthenticate)

	headerSessionUsername = []byte("Session-Username")
	headerRemoteUser      = []byte("Remote-User")
	headerRemoteGroups    = []byte("Remote-Groups")
	headerRemoteName      = []byte("Remote-Name")
	headerRemoteEmail     = []byte("Remote-Email")
)

const (
	headerAuthorizationSchemeBasic = "basic"
)

var (
	headerValueAuthenticateBasic = []byte(`Basic realm="Authorization Required"`)
)

const (
	queryArgRD        = "rd"
	queryArgRM        = "rm"
	queryArgAuth      = "auth"
	queryArgConsentID = "consent_id"
	queryArgFlow      = "flow"
	queryArgSubflow   = "subflow"
	queryArgUserCode  = oidc.FormParameterUserCode
	queryArgFlowID    = oidc.FormParameterFlowID
)

var (
	qryArgRD        = []byte(queryArgRD)
	qryArgFlowID    = []byte(oidc.FormParameterFlowID)
	qryArgAuth      = []byte(queryArgAuth)
	qryArgConsentID = []byte(queryArgConsentID)
	qryArgUserCode  = []byte(oidc.FormParameterUserCode)
)

var (
	qryValueBasic = []byte("basic")
	qryValueEmpty = []byte("")
)

const (
	messageOperationFailed                       = "Operation failed."
	messageAuthenticationFailed                  = "Authentication failed. Check your credentials."
	messageUnableToOptionsOneTimePassword        = "Unable to retrieve TOTP registration options."            //nolint:gosec
	messageUnableToRegisterOneTimePassword       = "Unable to set up one-time password."                      //nolint:gosec
	messageUnableToDeleteRegisterOneTimePassword = "Unable to delete one-time password registration session." //nolint:gosec
	messageUnableToDeleteOneTimePassword         = "Unable to delete one-time password."
	messageUnableToRegisterSecurityKey           = "Unable to register your security key."
	messageSecurityKeyDuplicateName              = "Another one of your security keys is already registered with that display name."
	messageUnableToResetPassword                 = "Unable to reset your password."
	messageUnableToChangePassword                = "Unable to change your password."
	messageIncorrectPassword                     = "Incorrect Password"
	messageMFAValidationFailed                   = "Authentication failed, please retry later."
	messagePasswordWeak                          = "Your supplied password does not meet the password policy requirements."
	messageUnableToModifyUser                    = "Unable to modify specified user"
	messageUnableToAddUser                       = "Unable to add specified user"
	messageUnableToDeleteUser                    = "Unable to delete specified user"
	messageNewUserRequiredFields                 = "Username, DisplayName, and Password are required fields"
	messageIncompleteUserCreation                = "The user has been created but metadata creation failed"
	messageUnableToDeleteUserMetadata            = "Unable to delete specified user metadata"
	messageUsernameRequired                      = "username parameter required"
)

const (
	//nolint:unused
	messageUsernameWrongFormat = "Username must be either a valid email address or contain only: letters, numbers, hyphens(-), commas(,), and underscores(_)."
	//nolint:unused
	messageGroupsWrongFormat = "Groups must only contain: letters, numbers, hyphens(-), commas(,), and underscores(_)."
	//nolint:unused
	messageDisplayNameWrongFormat = "Display Name must contain only printable unicode characters (no invisible or control characters)."
	//nolint:unused
	messageEmailWrongFormat = "Username must be a valid email address."
)

const (
	flowNameOpenIDConnect = "openid_connect"

	flowOpenIDConnectSubFlowNameDeviceAuthorization = "device_authorization"
)

const (
	logFmtActionAuthentication = "authentication"
	logFmtActionRegistration   = "registration"

	logFmtErrParseRequestBody     = "Failed to parse %s request body"
	logFmtErrRegulationFail       = "Failed to perform %s authentication regulation for user '%s'"
	logFmtErrSessionRegenerate    = "Could not regenerate session during %s authentication for user '%s'"
	logFmtErrSessionReset         = "Could not reset session during %s authentication for user '%s'"
	logFmtErrSessionSave          = "Could not save session with the %s during %s %s for user '%s'"
	logFmtErrObtainProfileDetails = "Could not obtain profile details during %s authentication for user '%s'"
	logFmtTraceProfileDetails     = "Profile details for user '%s' => groups: %s, emails %s"
	logFmtErrUserNotAdmin         = "user '%s' is not an admin"
)

const (
	logFmtAuthzRedirect = "Access to %s (method %s) is not authorized to user %s, responding with status code %d with location redirect to %s"

	logFmtAuthorizationPrefix = "Authorization Request with id '%s' on client with id '%s' "

	logFmtErrConsentCantDetermineConsentMode = logFmtAuthorizationPrefix + "could not be processed: error occurred generating consent: client consent mode could not be reliably determined"

	logFmtConsentPrefix = logFmtAuthorizationPrefix + "using consent mode '%s' "

	logFmtErrConsentParseChallengeID = logFmtConsentPrefix + "could not be processed: error occurred parsing the consent id (challenge) '%s': %+v"
	logFmtErrConsentPreConfLookup    = logFmtConsentPrefix + "had error looking up pre-configured consent sessions: %+v"
	logFmtErrConsentPreConfRowsClose = logFmtConsentPrefix + "had error closing rows while looking up pre-configured consent sessions: %+v"
	logFmtErrConsentZeroID           = logFmtConsentPrefix + "could not be processed: the consent id had a zero value"
	logFmtErrConsentCantGetSubject   = logFmtConsentPrefix + "could not be processed: error occurred retrieving subject identifier for user '%s' and sector identifier '%s': %+v"
	logFmtErrConsentGenerateError    = logFmtConsentPrefix + "could not be processed: error occurred %s consent: %+v"

	logFmtDbgConsentGenerate                  = logFmtConsentPrefix + "proceeding to generate a new consent session"
	logFmtDbgConsentAuthenticationSufficiency = logFmtConsentPrefix + "authentication level '%s' is %s for client level '%s'"
	logFmtDbgConsentRedirect                  = logFmtConsentPrefix + "is being redirected to '%s'"
	logFmtDbgConsentPreConfSuccessfulLookup   = logFmtConsentPrefix + "successfully looked up pre-configured consent with signature of client id '%s' and subject '%s' and scopes '%s' with id '%d'"
	logFmtDbgConsentPreConfUnsuccessfulLookup = logFmtConsentPrefix + "unsuccessfully looked up pre-configured consent with signature of client id '%s' and subject '%s' and scopes '%s' and audience '%s'"
	logFmtDbgConsentPreConfTryingLookup       = logFmtConsentPrefix + "attempting to discover pre-configurations with signature of client id '%s' and subject '%s' and scopes '%s'"

	logFmtErrConsentWithIDCouldNotBeProcessed = logFmtConsentPrefix + "could not be processed: error occurred performing consent for consent session with id '%s': "

	logFmtErrConsentLookupLoadingSession        = logFmtErrConsentWithIDCouldNotBeProcessed + "error occurred while loading session: %+v"
	logFmtErrConsentSessionSubjectNotAuthorized = logFmtErrConsentWithIDCouldNotBeProcessed + "user '%s' with subject '%s' is not authorized to consent for subject '%s'"
	logFmtErrConsentCantGrant                   = logFmtErrConsentWithIDCouldNotBeProcessed + "the session does not appear to be valid for %s consent: either the subject is null, the consent has already been granted, or the consent session is a pre-configured session"
	logFmtErrConsentCantGrantPreConf            = logFmtErrConsentWithIDCouldNotBeProcessed + "the session does not appear to be valid for pre-configured consent: either the subject is null, the consent has been granted and is either not pre-configured, or the pre-configuration is expired"
	logFmtErrConsentCantGrantRejected           = logFmtErrConsentWithIDCouldNotBeProcessed + "the user explicitly rejected this consent session"
	logFmtErrConsentSaveSessionResponse         = logFmtErrConsentWithIDCouldNotBeProcessed + "error occurred saving consent session response: %+v"
	logFmtErrConsentSaveSession                 = logFmtErrConsentWithIDCouldNotBeProcessed + "error occurred saving consent session: %+v"
)

const (
	logFmtErrPasskeyAuthenticationChallengeGenerate     = "Error occurred generating a WebAuthn passkey authentication challenge: %s"
	logFmtErrPasskeyAuthenticationChallengeValidate     = "Error occurred validating a WebAuthn passkey authentication challenge: %s"
	logFmtErrPasskeyAuthenticationChallengeValidateUser = "Error occurred validating a WebAuthn passkey authentication challenge for user '%s': %s"
)

// Duo constants.
const (
	allow  = "allow"
	deny   = "deny"
	enroll = "enroll"
	auth   = "auth"
)

const ldapPasswordComplexityCode = "0000052D."

var ldapPasswordComplexityCodes = []string{
	"0000052D", "SynoNumber", "SynoMixedCase", "SynoExcludeNameDesc", "SynoSpecialChar",
}

var ldapPasswordComplexityErrors = []string{
	"LDAP Result Code 19 \"Constraint Violation\": Password fails quality checking policy",
	"LDAP Result Code 19 \"Constraint Violation\": Password is too young to change",
}

const (
	errStrReqBodyParse        = "error parsing the request body"
	errStrRespBody            = "error occurred writing the response body"
	errStrUserSessionData     = "error occurred retrieving the user session data"
	errStrUserSessionDataSave = "error occurred saving the user session data"
)

var (
	errUserAnonymous              = errors.New("user is anonymous")
	errUserIsAlreadyAuthenticated = errors.New("user is already authenticated")
)
