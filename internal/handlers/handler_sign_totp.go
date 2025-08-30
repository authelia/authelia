package handlers

import (
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

// TimeBasedOneTimePasswordGET returns the users TOTP configuration.
func TimeBasedOneTimePasswordGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving TOTP configuration: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred retrieving TOTP configuration")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	var config *model.TOTPConfiguration

	if config, err = ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username); err != nil {

		if errors.Is(err, storage.ErrNoTOTPConfiguration) {
			ctx.Logger.WithError(err).Debugf("Error occurred retrieving TOTP configuration for user '%s'", userSession.Username)
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.SetJSONError("Could not find TOTP Configuration for user.")
		} else {
			ctx.Logger.WithError(err).Errorf("Error occurred retrieving TOTP configuration for user '%s': error occurred retrieving the configuration from the storage backend", userSession.Username)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetJSONError("Could not find TOTP Configuration for user.")
		}

		return
	}

	if err = ctx.SetJSONBody(config); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving TOTP configuration for user '%s': %s", userSession.Username, errStrRespBody)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}
}

// TimeBasedOneTimePasswordPOST validate the TOTP passcode provided by the user.
//
//nolint:gocyclo
func TimeBasedOneTimePasswordPOST(ctx *middlewares.AutheliaCtx) {
	bodyJSON := bodySignTOTPRequest{}

	var (
		userSession   session.UserSession
		config        *model.TOTPConfiguration
		valid, exists bool
		step          uint64
		err           error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP authentication: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred validating a TOTP authentication")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP authentication for user '%s': %s", userSession.Username, errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if n := len(bodyJSON.Token); n != 6 && n != 8 {
		ctx.Logger.Errorf("Error occurred validating a TOTP authentication for user '%s': expected code length is 6 or 8 but the user provided code was %d characters in length", userSession.Username, n)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if config, err = ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP authentication for user '%s': error occurred retrieving the configuration from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if valid, step, err = ctx.Providers.TOTP.Validate(ctx, bodyJSON.Token, config); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP authentication for user '%s': error occurred validating the user input", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if !valid {
		ctx.Logger.WithError(fmt.Errorf("the user input wasn't valid")).Errorf("Error occurred validating a TOTP authentication for user '%s': error occurred validating the user input", userSession.Username)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypeTOTP, nil)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if exists, err = ctx.Providers.StorageProvider.ExistsTOTPHistory(ctx, userSession.Username, step*uint64(config.Period)); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP authentication for user '%s': error occurred checking the TOTP history", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if exists {
		if ctx.Configuration.TOTP.DisableReuseSecurityPolicy {
			ctx.Logger.WithFields(map[string]any{"username": userSession.Username}).Warn("User has reused a Time-based One Time Password with the given step but the policy to disable reuse is disabled")
		} else {
			ctx.Logger.WithError(fmt.Errorf("the user has already used this code recently and will not be permitted to reuse it")).Errorf("Error occurred validating a TOTP authentication for user '%s': error occurred satisfying security policies", userSession.Username)

			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetJSONError(messageMFAValidationFailed)

			return
		}
	} else if err = ctx.Providers.StorageProvider.SaveTOTPHistory(ctx, userSession.Username, step*uint64(config.Period)); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP authentication for user '%s': error occurred saving the TOTP history to the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypeTOTP, nil)

	if err = ctx.RegenerateSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP authentication for user '%s': error regenerating the user session", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	config.UpdateSignInInfo(ctx.Clock.Now())

	if err = ctx.Providers.StorageProvider.UpdateTOTPConfigurationSignIn(ctx, config.ID, config.LastUsedAt); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP authentication for user '%s': error occurred saving the credential sign-in information to the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	userSession.SetTwoFactorTOTP(ctx.Clock.Now())

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a TOTP authentication for user '%s': %s", userSession.Username, errStrUserSessionDataSave)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if len(bodyJSON.Flow) > 0 {
		handleFlowResponse(ctx, &userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}
