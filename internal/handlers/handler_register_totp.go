package handlers

import (
	"encoding/json"
	"fmt"
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
	var (
		userSession session.UserSession
		err         error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving TOTP registration options: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToOptionsOneTimePassword)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred retrieving TOTP registration options")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToOptionsOneTimePassword)

		return
	}

	if err = ctx.SetJSONBody(ctx.Providers.TOTP.Options()); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving TOTP registration options for user '%s': %s", userSession.Username, errStrRespBody)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToOptionsOneTimePassword)
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
		ctx.Logger.WithError(err).Errorf("Error occurred generating a TOTP registration session: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred generating a TOTP registration session")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a TOTP registration session for user '%s': %s", userSession.Username, errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	opts := ctx.Providers.TOTP.Options()

	if !utils.IsStringInSlice(bodyJSON.Algorithm, opts.Algorithms) ||
		!utils.IsIntegerInSlice(bodyJSON.Period, opts.Periods) ||
		!utils.IsIntegerInSlice(int(bodyJSON.Length), opts.Lengths) {
		ctx.Logger.WithError(fmt.Errorf("the algorithm '%s', period '%d', or length '%d' was not permitted by configured policy", bodyJSON.Algorithm, bodyJSON.Period, bodyJSON.Length)).Errorf("Error occurred generating a TOTP registration session for user '%s': error occurred validating registration options selection", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	var config *model.TOTPConfiguration

	if config, err = ctx.Providers.TOTP.GenerateCustom(ctx, userSession.Username, bodyJSON.Algorithm, "", uint32(bodyJSON.Length), uint(bodyJSON.Period), 0); err != nil { //nolint:gosec // Validated at runtime.
		ctx.Logger.WithError(err).Errorf("Error occurred generating a TOTP registration session for user '%s': error generating TOTP configuration", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
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
		ctx.Logger.WithError(err).Errorf("Error occurred generating a TOTP registration session for user '%s': %s", userSession.Username, errStrUserSessionDataSave)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	response := TOTPKeyResponse{
		OTPAuthURL:   config.URI(),
		Base32Secret: userSession.TOTP.Secret,
	}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a TOTP registration session for user '%s': %s", userSession.Username, errStrRespBody)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)
	}
}

// TOTPRegisterPOST handles validation that the user has properly registered the configuration.
func TOTPRegisterPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		bodyJSON    bodyRegisterFinishTOTP
		valid       bool
		step        uint64
		err         error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP registration session: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred validating a TOTP registration session")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP registration session for user '%s': %s", userSession.Username, errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	if userSession.TOTP == nil {
		ctx.Logger.Errorf("Error occurred validating a TOTP registration session for user '%s': the user did not initiate a registration session on their current session", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	if ctx.Clock.Now().After(userSession.TOTP.Expires) {
		ctx.Logger.WithError(fmt.Errorf("the registration session is expired")).Errorf("Error occurred validating a TOTP registration session for user '%s': error occurred validating the session", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

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

	if valid, step, err = ctx.Providers.TOTP.Validate(ctx, bodyJSON.Token, &config); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP registration session for user '%s': error occurred validating the user input against the session", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	if !valid {
		ctx.Logger.WithError(fmt.Errorf("user input did not match any expected value")).Errorf("Error occurred validating a TOTP registration session for user '%s'", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	if !ctx.Configuration.TOTP.DisableReuseSecurityPolicy {
		if err = ctx.Providers.StorageProvider.SaveTOTPHistory(ctx, userSession.Username, step*uint64(config.Period)); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP registration session for user '%s': error occurred saving the TOTP history to the storage backend", userSession.Username)

			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

			return
		}
	}

	if err = ctx.Providers.StorageProvider.SaveTOTPConfiguration(ctx, config); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP registration session for user '%s': error occurred saving the TOTP configuration to the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	userSession.TOTP = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP registration session for user '%s': %s", userSession.Username, errStrUserSessionDataSave)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterOneTimePassword)

		return
	}

	body := emailEventBody{
		Prefix: eventEmailAction2FAPrefix,
		Body:   eventEmailAction2FABody,
		Suffix: eventEmailAction2FAAddedSuffix,
	}

	ctxLogEvent(ctx, userSession.Username, eventLogAction2FAAdded, body, map[string]any{eventLogKeyAction: eventLogAction2FAAdded, eventLogKeyCategory: eventLogCategoryOneTimePassword})

	ctx.ReplyOK()
}

// TOTPRegisterDELETE removes a pending TOTP registration.
func TOTPRegisterDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting a TOTP registration session: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToDeleteRegisterOneTimePassword)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred deleting a TOTP registration session")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToDeleteRegisterOneTimePassword)

		return
	}

	if userSession.TOTP == nil {
		ctx.ReplyOK()

		return
	}

	userSession.TOTP = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "deleted pending TOTP configuration", regulation.AuthTypeTOTP, logFmtActionRegistration, userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToDeleteRegisterOneTimePassword)

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
		ctx.Logger.WithError(err).Errorf("Error occurred deleting a TOTP configuration: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToDeleteOneTimePassword)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred deleting a TOTP configuration")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToDeleteOneTimePassword)

		return
	}

	if _, err = ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting a TOTP configuration for user '%s': error occurred loading configuration from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToDeleteOneTimePassword)

		return
	}

	if err = ctx.Providers.StorageProvider.DeleteTOTPConfiguration(ctx, userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting a TOTP configuration for user '%s': error occurred deleting configuration from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToDeleteOneTimePassword)

		return
	}

	body := emailEventBody{
		Prefix: eventEmailAction2FAPrefix,
		Body:   eventEmailAction2FABody,
		Suffix: eventEmailAction2FARemovedSuffix,
	}

	ctxLogEvent(ctx, userSession.Username, eventLogAction2FARemoved, body, map[string]any{eventLogKeyAction: eventLogAction2FARemoved, eventLogKeyCategory: eventLogCategoryOneTimePassword})

	ctx.ReplyOK()
}
