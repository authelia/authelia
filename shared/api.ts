/**
 * @apiDefine UserSession
 * @apiHeader {String} Cookie Cookie containing "connect.sid", the user
 * session token.
 */

/**
 * @apiDefine InternalError
 * @apiError (Error 500) {String} error Internal error message.
 */

/**
 * @apiDefine IdentityValidationStart
 *
 * @apiSuccess (Success 204) status Identity validation has been initiated.
 * @apiError (Error 403) AccessDenied Access is denied.
 * @apiError (Error 400) InvalidIdentity User identity is invalid.
 * @apiError (Error 500) {String} error Internal error message.
 *
 * @apiDescription This request issue an identity validation token for the user
 * bound to the session. It sends a challenge to the email address set in the user
 * LDAP entry. The user must visit the sent URL to complete the validation and
 * continue the registration process.
 */

/**
 * @apiDefine IdentityValidationFinish
 * @apiParam {String} identity_token The one-time identity validation token provided in the email.
 * @apiSuccess (Success 200) {String} content The content of the page.
 * @apiError (Error 403) AccessDenied Access is denied.
 * @apiError (Error 500) {String} error Internal error message.
 */

/**
 * @api {post} /api/secondfactor/u2f/register Complete U2F registration
 * @apiName FinishU2FRegistration
 * @apiGroup U2F
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse InternalError
 *
 * @apiSuccess (Success 302) Redirect to the URL that has been stored during last call to /api/verify.
 *
 * @apiDescription Complete U2F registration request.
 */
export const SECOND_FACTOR_U2F_REGISTER_POST = "/api/u2f/register";

/**
 * @api {get} /api/u2f/register_request Start U2F registration
 * @apiName StartU2FRegistration
 * @apiGroup U2F
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse InternalError
 *
 * @apiSuccess (Success 200) authentication_request The U2F registration request.
 * @apiError (Error 403) {none} error Unexpected identity validation challenge.
 *
 * @apiDescription Initiate a U2F device registration request.
 */
export const SECOND_FACTOR_U2F_REGISTER_REQUEST_GET = "/api/u2f/register_request";

/**
 * @api {post} /api/u2f/sign Complete U2F authentication
 * @apiName CompleteU2FAuthentication
 * @apiGroup U2F
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse InternalError
 *
 * @apiSuccess (Success 302) Redirect to the URL that has been stored during last call to /api/verify.
 * @apiError (Error 403) {none} error No authentication request has been provided.
 *
 * @apiDescription Complete authentication request of the U2F device.
 */
export const SECOND_FACTOR_U2F_SIGN_POST = "/api/u2f/sign";

/**
 * @api {get} /api/u2f/sign_request Start U2F authentication
 * @apiName StartU2FAuthentication
 * @apiGroup U2F
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse InternalError
 *
 * @apiSuccess (Success 200) authentication_request The U2F authentication request.
 * @apiError (Error 401) {none} error There is no key registered for user in session.
 *
 * @apiDescription Initiate an authentication request using a U2F device.
 */
export const SECOND_FACTOR_U2F_SIGN_REQUEST_GET = "/api/u2f/sign_request";

/**
 * @api {post} /api/totp Complete TOTP authentication
 * @apiName ValidateTOTPSecondFactor
 * @apiGroup TOTP
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse InternalError
 *
 * @apiParam {String} token TOTP token.
 *
 * @apiSuccess (Success 302) Redirect to the URL that has been stored during last call to /api/verify.
 * @apiError (Error 401) {none} error TOTP token is invalid.
 *
 * @apiDescription Verify TOTP token. The user is authenticated upon success.
 */
export const SECOND_FACTOR_TOTP_POST = "/api/totp";

/**
 * @api {post} /api/duo-push Complete Duo Push Factor
 * @apiName ValidateDuoPushSecondFactor
 * @apiGroup DuoPush
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse InternalError
 *
 * @apiSuccess (Success 302) Redirect to the URL that has been stored during last call to /api/verify.
 * @apiError (Error 401) {none} error TOTP token is invalid.
 *
 * @apiDescription Verify TOTP token. The user is authenticated upon success.
 */
export const SECOND_FACTOR_DUO_PUSH_POST = "/api/duo-push";


/**
 * @api {get} /api/secondfactor/u2f/identity/start Start U2F registration identity validation
 * @apiName RequestU2FRegistration
 * @apiGroup U2F
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse IdentityValidationStart
 */
export const SECOND_FACTOR_U2F_IDENTITY_START_POST = "/api/secondfactor/u2f/identity/start";

/**
 * @api {get} /api/secondfactor/u2f/identity/finish Finish U2F registration identity validation
 * @apiName ServeU2FRegistrationPage
 * @apiGroup U2F
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse IdentityValidationFinish
 *
 * @apiDescription Serves the U2F registration page that asks the user to
 * touch the token of the U2F device.
 */
export const SECOND_FACTOR_U2F_IDENTITY_FINISH_POST = "/api/secondfactor/u2f/identity/finish";



/**
 * @api {get} /api/secondfactor/totp/identity/start Start TOTP registration identity validation
 * @apiName StartTOTPRegistration
 * @apiGroup TOTP
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse IdentityValidationStart
 *
 * @apiDescription Initiates the identity validation
 */
export const SECOND_FACTOR_TOTP_IDENTITY_START_POST = "/api/secondfactor/totp/identity/start";



/**
 * @api {get} /api/secondfactor/totp/identity/finish Finish TOTP registration identity validation
 * @apiName FinishTOTPRegistration
 * @apiGroup TOTP
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse IdentityValidationFinish
 *
 * @apiDescription Serves the TOTP registration page that displays the secret.
 * The secret is a QRCode and a base32 secret.
 */
export const SECOND_FACTOR_TOTP_IDENTITY_FINISH_POST = "/api/secondfactor/totp/identity/finish";

/**
 * @api {get} /api/secondfactor/preferences Retrieve the user preferences.
 * @apiName GetUserPreferences
 * @apiGroup 2FA
 * @apiVersion 1.0.0
 * @apiUse UserSession
 *
 * @apiDescription Retrieve the user preferences sucha as the prefered method to use (TOTP or U2F).
 */
export const SECOND_FACTOR_PREFERENCES_GET = "/api/secondfactor/preferences";

/**
 * @api {post} /api/secondfactor/preferences Set the user preferences.
 * @apiName SetUserPreferences
 * @apiGroup 2FA
 * @apiVersion 1.0.0
 * @apiUse UserSession
 *
 * @apiDescription Set the user preferences sucha as the prefered method to use  (TOTP or U2F).
 */
export const SECOND_FACTOR_PREFERENCES_POST = "/api/secondfactor/preferences";

/**
 * @api {post} /api/secondfactor/available List the available methods.
 * @apiName GetAvailableMethods
 * @apiGroup 2FA
 * @apiVersion 1.0.0
 *
 * @apiDescription Get the available 2FA methods.
 */
export const SECOND_FACTOR_AVAILABLE_GET = "/api/secondfactor/available";


/**
 * @api {post} /api/password-reset Set new password
 * @apiName SetNewLDAPPassword
 * @apiGroup PasswordReset
 * @apiVersion 1.0.0
 * @apiUse UserSession
 *
 * @apiParam {String} password New password
 *
 * @apiDescription Set a new password for the user.
 */
export const RESET_PASSWORD_FORM_POST = "/api/password-reset";



/**
 * @api {get} /password-reset/request Request username
 * @apiName ServePasswordResetPage
 * @apiGroup PasswordReset
 * @apiVersion 1.0.0
 * @apiUse UserSession
 *
 * @apiDescription Serve a page that requires the username.
 */
export const RESET_PASSWORD_REQUEST_GET = "/api/password-reset/request";



/**
 * @api {get} /password-reset/identity/start Start password reset request
 * @apiName StartPasswordResetRequest
 * @apiGroup PasswordReset
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse IdentityValidationStart
 *
 * @apiDescription Start password reset request.
 */
export const RESET_PASSWORD_IDENTITY_START_GET = "/api/password-reset/identity/start";



/**
 * @api {post} /reset-password/request Finish password reset request
 * @apiName FinishPasswordResetRequest
 * @apiGroup PasswordReset
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse IdentityValidationFinish
 *
 * @apiDescription Start password reset request.
 */
export const RESET_PASSWORD_IDENTITY_FINISH_GET = "/api/password-reset/identity/finish";



/**
 * @api {post} /1stfactor Bind user against LDAP
 * @apiName ValidateFirstFactor
 * @apiGroup Authentication
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse InternalError
 *
 * @apiParam {String} username User username.
 * @apiParam {String} password User password.
 *
 * @apiSuccess (Success 204) status 1st factor is validated.
 * @apiError (Error 401) {none} error 1st factor is not validated.
 * @apiError (Error 401) {none} error Access has been restricted after too
 * many authentication attempts
 *
 * @apiDescription Verify credentials against the LDAP.
 */
export const FIRST_FACTOR_POST = "/api/firstfactor";

/**
 * @api {get} /state Authentication state
 * @apiName State
 * @apiGroup Authentication
 * @apiVersion 1.0.0
 *
 * @apiSuccess (Success 200) A dict containing the username and the authentication
 * level
 *
 * @apiDescription Get the authentication state of the user based on the cookie.
 */
export const STATE_GET = "/api/state";

/**
 * @api {get} /api/verify Verify user authentication
 * @apiName VerifyAuthentication
 * @apiGroup Verification
 * @apiVersion 1.0.0
 * @apiUse UserSession
 *
 * @apiParam {String} redirect Optional parameter set to the url where the user
 * is redirected if access is refused. It is mainly used by Traefik that does
 * not control the redirection itself.
 *
 * @apiSuccess (Success 204) status The user is authenticated.
 * @apiError (Error 302) redirect The user is redirected if redirect parameter is provided.
 * @apiError (Error 401) status The user get an error if access failed
 *
 * @apiDescription Verify that the user is authenticated, i.e., the two
 * factors have been validated.
 * If the user is authenticated the response headers Remote-User and Remote-Groups
 * are set. Remote-User contains the user id of the currently logged in user and Remote-Groups
 * a comma separated list of assigned groups.
 */
export const VERIFY_GET = "/api/verify";

/**
 * @api {post} /api/logout Logout procedure
 * @apiName Logout
 * @apiGroup Authentication
 * @apiVersion 1.0.0
 *
 * @apiSuccess (Success 200)
 *
 * @apiDescription Resets the session to logout the user.
 */
export const LOGOUT_POST = "/api/logout";