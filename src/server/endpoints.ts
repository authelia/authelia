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
 * @apiSuccess (Success 302) Redirect to the URL that has been stored during last call to /verify.
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
 * @apiSuccess (Success 302) Redirect to the URL that has been stored during last call to /verify.
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
 * @apiSuccess (Success 302) Redirect to the URL that has been stored during last call to /verify.
 * @apiError (Error 401) {none} error TOTP token is invalid.
 *
 * @apiDescription Verify TOTP token. The user is authenticated upon success.
 */
export const SECOND_FACTOR_TOTP_POST = "/api/totp";


/**
 * @api {get} /secondfactor/u2f/identity/start Start U2F registration identity validation
 * @apiName RequestU2FRegistration
 * @apiGroup U2F
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse IdentityValidationStart
 */
export const SECOND_FACTOR_U2F_IDENTITY_START_GET = "/secondfactor/u2f/identity/start";

/**
 * @api {get} /secondfactor/u2f/identity/finish Finish U2F registration identity validation
 * @apiName ServeU2FRegistrationPage
 * @apiGroup U2F
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse IdentityValidationFinish
 *
 * @apiDescription Serves the U2F registration page that asks the user to
 * touch the token of the U2F device.
 */
export const SECOND_FACTOR_U2F_IDENTITY_FINISH_GET = "/secondfactor/u2f/identity/finish";



/**
 * @api {get} /secondfactor/totp/identity/start Start TOTP registration identity validation
 * @apiName StartTOTPRegistration
 * @apiGroup TOTP
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse IdentityValidationStart
 *
 * @apiDescription Initiates the identity validation
 */
export const SECOND_FACTOR_TOTP_IDENTITY_START_GET = "/secondfactor/totp/identity/start";



/**
 * @api {get} /secondfactor/totp/identity/finish Finish TOTP registration identity validation
 * @apiName FinishTOTPRegistration
 * @apiGroup TOTP
 * @apiVersion 1.0.0
 * @apiUse UserSession
 * @apiUse IdentityValidationFinish
 *
 *
 * @apiDescription Serves the TOTP registration page that displays the secret.
 * The secret is a QRCode and a base32 secret.
 */
export const SECOND_FACTOR_TOTP_IDENTITY_FINISH_GET = "/secondfactor/totp/identity/finish";



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
export const RESET_PASSWORD_REQUEST_GET = "/password-reset/request";



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
export const RESET_PASSWORD_IDENTITY_START_GET = "/password-reset/identity/start";



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
export const RESET_PASSWORD_IDENTITY_FINISH_GET = "/password-reset/identity/finish";



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
 * @api {get} / First factor page
 * @apiName Login
 * @apiGroup Authentication
 * @apiVersion 1.0.0
 *
 * @apiSuccess (Success 200) {String} Content The content of the first factor page.
 *
 * @apiDescription Serves the login page and create a create a cookie for the client.
 */
export const FIRST_FACTOR_GET = "/";

/**
 * @api {get} /secondfactor Second factor page
 * @apiName SecondFactor
 * @apiGroup Authentication
 * @apiVersion 1.0.0
 *
 * @apiSuccess (Success 200) {String} Content The content of second factor page.
 *
 * @apiDescription Serves the second factor page
 */
export const SECOND_FACTOR_GET = "/secondfactor";

/**
 * @api {get} /verify Verify user authentication
 * @apiName VerifyAuthentication
 * @apiGroup Verification
 * @apiVersion 1.0.0
 * @apiUse UserSession
 *
 * @apiSuccess (Success 204) status The user is authenticated.
 * @apiError (Error 401) status The user is not authenticated.
 *
 * @apiDescription Verify that the user is authenticated, i.e., the two
 * factors have been validated
 */
export const VERIFY_GET = "/verify";

/**
 * @api {get} /logout Serves logout page
 * @apiName Logout
 * @apiGroup Authentication
 * @apiVersion 1.0.0
 *
 * @apiParam {String} redirect Redirect to this URL when user is deauthenticated.
 * @apiSuccess (Success 302) redirect Redirect to the URL.
 *
 * @apiDescription Log out the user and redirect to the URL.
 */
export const LOGOUT_GET = "/logout";

export const ERROR_401_GET = "/error/401";
export const ERROR_403_GET = "/error/403";
export const ERROR_404_GET = "/error/404";
