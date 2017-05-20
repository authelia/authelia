
import express = require("express");

const routes = require("./routes");
const identity_check = require("./identity_check");

export default class RestApi {
  static setup(app: express.Application): void {
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
     * @apiDefine IdentityValidationPost
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
     * @apiDefine IdentityValidationGet
     * @apiParam {String} identity_token The one-time identity validation token provided in the email.
     * @apiSuccess (Success 200) {String} content The content of the page.
     * @apiError (Error 403) AccessDenied Access is denied.
     * @apiError (Error 500) {String} error Internal error message.
     */

    /**
     * @api {get} /login Serve login page
     * @apiName Login
     * @apiGroup Pages
     * @apiVersion 1.0.0
     *
     * @apiParam {String} redirect Redirect to this URL when user is authenticated.
     * @apiSuccess (Success 200) {String} Content The content of the login page.
     *
     * @apiDescription Create a user session and serve the login page along with
     * a cookie.
     */
    app.get("/login", routes.login);

    /**
     * @api {get} /logout Server logout page
     * @apiName Logout
     * @apiGroup Pages
     * @apiVersion 1.0.0
     *
     * @apiParam {String} redirect Redirect to this URL when user is deauthenticated.
     * @apiSuccess (Success 301) redirect Redirect to the URL.
     *
     * @apiDescription Deauthenticate the user and redirect him.
     */
    app.get("/logout", routes.logout);

    /**
     * @api {post} /totp-register Request TOTP registration
     * @apiName RequestTOTPRegistration
     * @apiGroup Registration
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse IdentityValidationPost
     */
    /**
     * @api {get} /totp-register Serve TOTP registration page
     * @apiName ServeTOTPRegistrationPage
     * @apiGroup Registration
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse IdentityValidationGet
     *
     *
     * @apiDescription Serves the TOTP registration page that displays the secret.
     * The secret is a QRCode and a base32 secret.
     */
    identity_check(app, "/totp-register", routes.totp_register.icheck_interface);


    /**
     * @api {post} /u2f-register Request U2F registration
     * @apiName RequestU2FRegistration
     * @apiGroup Registration
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse IdentityValidationPost
     */
    /**
     * @api {get} /u2f-register Serve U2F registration page
     * @apiName ServeU2FRegistrationPage
     * @apiGroup Pages
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse IdentityValidationGet
     *
     * @apiDescription Serves the U2F registration page that asks the user to
     * touch the token of the U2F device.
     */
    identity_check(app, "/u2f-register", routes.u2f_register.icheck_interface);

    /**
     * @api {post} /reset-password Request for password reset
     * @apiName RequestPasswordReset
     * @apiGroup Registration
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse IdentityValidationPost
     */
    /**
     * @api {get} /reset-password Serve password reset form.
     * @apiName ServePasswordResetForm
     * @apiGroup Pages
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse IdentityValidationGet
     *
     * @apiDescription Serves password reset form that allow the user to provide
     * the new password.
     */
    identity_check(app, "/reset-password", routes.reset_password.icheck_interface);

    app.get("/reset-password-form", function (req, res) { res.render("reset-password-form"); });

    /**
     * @api {post} /new-password Set LDAP password
     * @apiName SetLDAPPassword
     * @apiGroup Registration
     * @apiVersion 1.0.0
     * @apiUse UserSession
     *
     * @apiParam {String} password New password
     *
     * @apiDescription Set a new password for the user.
     */
    app.post("/new-password", routes.reset_password.post);

    /**
     * @api {post} /new-totp-secret Generate TOTP secret
     * @apiName GenerateTOTPSecret
     * @apiGroup Registration
     * @apiVersion 1.0.0
     * @apiUse UserSession
     *
     * @apiSuccess (Success 200) {String} base32 The base32 representation of the secret.
     * @apiSuccess (Success 200) {String} ascii The ASCII representation of the secret.
     * @apiSuccess (Success 200) {String} qrcode The QRCode of the secret in URI format.
     *
     * @apiError (Error 403) {String} error No user provided in the session or
     * unexpected identity validation challenge in the session.
     * @apiError (Error 500) {String} error Internal error message
     *
     * @apiDescription Generate a new TOTP secret and returns it.
     */
    app.post("/new-totp-secret", routes.totp_register.post);

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
    app.get("/verify", routes.verify);

    /**
     * @api {post} /1stfactor LDAP authentication
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
     * @apiError (Error 403) {none} error Access has been restricted after too
     * many authentication attempts
     *
     * @apiDescription Verify credentials against the LDAP.
     */
    app.post("/1stfactor", routes.first_factor);

    /**
     * @api {post} /2ndfactor/totp TOTP authentication
     * @apiName ValidateTOTPSecondFactor
     * @apiGroup Authentication
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse InternalError
     *
     * @apiParam {String} token TOTP token.
     *
     * @apiSuccess (Success 204) status TOTP token is valid.
     * @apiError (Error 401) {none} error TOTP token is invalid.
     *
     * @apiDescription Verify TOTP token. The user is authenticated upon success.
     */
    app.post("/2ndfactor/totp", routes.second_factor.totp);

    /**
     * @api {get} /2ndfactor/u2f/sign_request U2F Start authentication
     * @apiName StartU2FAuthentication
     * @apiGroup Authentication
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse InternalError
     *
     * @apiSuccess (Success 200) authentication_request The U2F authentication request.
     * @apiError (Error 401) {none} error There is no key registered for user in session.
     *
     * @apiDescription Initiate an authentication request using a U2F device.
     */
    app.get("/2ndfactor/u2f/sign_request", routes.second_factor.u2f.sign_request);

    /**
     * @api {post} /2ndfactor/u2f/sign U2F Complete authentication
     * @apiName CompleteU2FAuthentication
     * @apiGroup Authentication
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse InternalError
     *
     * @apiSuccess (Success 204) status The U2F authentication succeeded.
     * @apiError (Error 403) {none} error No authentication request has been provided.
     *
     * @apiDescription Complete authentication request of the U2F device.
     */
    app.post("/2ndfactor/u2f/sign", routes.second_factor.u2f.sign);

    /**
     * @api {get} /2ndfactor/u2f/register_request U2F Start device registration
     * @apiName StartU2FRegistration
     * @apiGroup Registration
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse InternalError
     *
     * @apiSuccess (Success 200) authentication_request The U2F registration request.
     * @apiError (Error 403) {none} error Unexpected identity validation challenge.
     *
     * @apiDescription Initiate a U2F device registration request.
     */
    app.get("/2ndfactor/u2f/register_request", routes.second_factor.u2f.register_request);

    /**
     * @api {post} /2ndfactor/u2f/register U2F Complete device registration
     * @apiName CompleteU2FRegistration
     * @apiGroup Registration
     * @apiVersion 1.0.0
     * @apiUse UserSession
     * @apiUse InternalError
     *
     * @apiSuccess (Success 204) status The U2F registration succeeded.
     * @apiError (Error 403) {none} error Unexpected identity validation challenge.
     * @apiError (Error 403) {none} error No registration request has been provided.
     *
     * @apiDescription Complete U2F registration request.
     */
    app.post("/2ndfactor/u2f/register", routes.second_factor.u2f.register);
  }
}
