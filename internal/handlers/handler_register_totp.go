package handlers

import (
	"encoding/json"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// TOTPRegisterGET returns the registration specific options.
func TOTPRegisterGET(ctx *middlewares.AutheliaCtx) {
	if err := ctx.SetJSONBody(ctx.Providers.TOTP.Options()); err != nil {
		ctx.Logger.Errorf("Unable to set TOTP options response in body: %s", err)
	}
}

// TOTPRegisterPUT handles the users choice of registration specific options and returns the generated configuration.
func TOTPRegisterPUT(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		bodyJSON    bodyRegisterTOTP
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.Errorf("Error occurred handling request: anonymous user attempted %s registration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred unmarshaling body %s registration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	opts := ctx.Providers.TOTP.Options()

	if !utils.IsStringInSlice(bodyJSON.Algorithm, opts.Algorithms) ||
		!utils.IsIntegerInSlice(bodyJSON.Period, opts.Periods) ||
		!utils.IsIntegerInSlice(bodyJSON.Length, opts.Lengths) {
		ctx.Logger.Errorf("Validation failed for %s registration because the input options were not permitted by the configuration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	var config *model.TOTPConfiguration

	if config, err = ctx.Providers.TOTP.GenerateCustom(userSession.Username, bodyJSON.Algorithm, "", uint(bodyJSON.Length), uint(bodyJSON.Period), 0); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating TOTP configuration")

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	userSession.TOTP = &session.TOTP{
		Issuer:    config.Issuer,
		Algorithm: config.Algorithm,
		Digits:    config.Digits,
		Period:    config.Period,
		Secret:    string(config.Secret),
		Expires:   ctx.Clock.Now().Add(time.Minute * 10),
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "pending TOTP configuration", regulation.AuthTypeTOTP, logFmtActionRegistration, userSession.Username)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	response := TOTPKeyResponse{
		OTPAuthURL:   config.URI(),
		Base32Secret: userSession.TOTP.Secret,
	}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to set TOTP key response in body")
	}
}

// TOTPRegisterPOST handles validation that the user has properly registered the configuration.
func TOTPRegisterPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		bodyJSON    bodyRegisterFinishTOTP
		valid       bool
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.Errorf("Error occurred handling request: anonymous user attempted %s registration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.TOTP == nil {
		ctx.Logger.Errorf("Error occurred during %s registration: the user did not initiate a registration on their current session", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if ctx.Clock.Now().After(userSession.TOTP.Expires) {
		ctx.Logger.Errorf("Error occurred during %s registration: the registration is expired", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred unmarshaling body %s registration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	config := model.TOTPConfiguration{
		CreatedAt: ctx.Clock.Now(),
		Username:  userSession.Username,
		Issuer:    userSession.TOTP.Issuer,
		Algorithm: userSession.TOTP.Algorithm,
		Period:    userSession.TOTP.Period,
		Digits:    userSession.TOTP.Digits,
		Secret:    []byte(userSession.TOTP.Secret),
	}

	if valid, err = ctx.Providers.TOTP.Validate(bodyJSON.Token, &config); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating %s registration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	} else if !valid {
		ctx.Logger.Errorf("Error occurred validating %s registration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveTOTPConfiguration(ctx, config); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred saving %s registration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	userSession.TOTP = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "completed TOTP configuration", regulation.AuthTypeTOTP, logFmtActionRegistration, userSession.Username)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	ctxLogEvent(ctx, userSession.Username, eventLogAction2FAAdded, map[string]any{eventLogKeyAction: eventLogAction2FAAdded, eventLogKeyCategory: eventLogCategoryOneTimePassword})

	ctx.ReplyOK()
}

// TOTPRegisterDELETE removes a pending TOTP registration.
func TOTPRegisterDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration cancel", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.Errorf("Error occurred handling request: anonymous user attempted %s registration", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.TOTP == nil {
		ctx.ReplyOK()

		return
	}

	userSession.TOTP = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "deleted pending TOTP configuration", regulation.AuthTypeTOTP, logFmtActionRegistration, userSession.Username)

		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	ctx.ReplyOK()
}

// TOTPConfigurationDELETE removes a registered TOTP configuration.
func TOTPConfigurationDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s configuration delete operation", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToDeleteOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.Errorf("Error occurred handling request: anonymous user attempted %s removal", regulation.AuthTypeTOTP)

		ctx.SetJSONError(messageUnableToDeleteOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if _, err = ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred loading from storage for %s configuration delete operation for user '%s'", regulation.AuthTypeTOTP, userSession.Username)

		ctx.SetJSONError(messageUnableToDeleteOneTimePassword)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if err = ctx.Providers.StorageProvider.DeleteTOTPConfiguration(ctx, userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting from storage for %s configuration delete operation for user '%s'", regulation.AuthTypeTOTP, userSession.Username)

		ctx.SetJSONError(messageUnableToDeleteOneTimePassword)

		return
	}

	ctxLogEvent(ctx, userSession.Username, eventLogAction2FARemoved, map[string]any{eventLogKeyAction: eventLogAction2FARemoved, eventLogKeyCategory: eventLogCategoryOneTimePassword})

	ctx.ReplyOK()
}
