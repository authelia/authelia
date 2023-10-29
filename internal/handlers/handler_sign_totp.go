package handlers

import (
	"errors"

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
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		ctx.ReplyForbidden()

		return
	}

	var config *model.TOTPConfiguration

	if config, err = ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username); err != nil {
		if errors.Is(err, storage.ErrNoTOTPConfiguration) {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.SetJSONError("Could not find TOTP Configuration for user.")
			ctx.Logger.Errorf("Failed to lookup TOTP configuration for user '%s'", userSession.Username)
		} else {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetJSONError("Could not find TOTP Configuration for user.")
			ctx.Logger.Errorf("Failed to lookup TOTP configuration for user '%s' with unknown error: %v", userSession.Username, err)
		}

		return
	}

	if err = ctx.SetJSONBody(config); err != nil {
		ctx.Logger.Errorf("Unable to perform TOTP configuration response: %s", err)
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

// TimeBasedOneTimePasswordPOST validate the TOTP passcode provided by the user.
func TimeBasedOneTimePasswordPOST(ctx *middlewares.AutheliaCtx) {
	bodyJSON := bodySignTOTPRequest{}

	var (
		userSession session.UserSession
		err         error
	)

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrParseRequestBody, regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	config, err := ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username)
	if err != nil {
		ctx.Logger.Errorf("Failed to load TOTP configuration: %+v", err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	isValid, err := ctx.Providers.TOTP.Validate(bodyJSON.Token, config)
	if err != nil {
		ctx.Logger.Errorf("Failed to perform TOTP verification: %+v", err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if !isValid {
		_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypeTOTP, nil)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if err = markAuthenticationAttempt(ctx, true, nil, userSession.Username, regulation.AuthTypeTOTP, nil); err != nil {
		respondUnauthorized(ctx, messageMFAValidationFailed)
		return
	}

	if err = ctx.RegenerateSession(); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionRegenerate, regulation.AuthTypeTOTP, userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	config.UpdateSignInInfo(ctx.Clock.Now())

	if err = ctx.Providers.StorageProvider.UpdateTOTPConfigurationSignIn(ctx, config.ID, config.LastUsedAt); err != nil {
		ctx.Logger.Errorf("Unable to save %s device sign in metadata for user '%s': %v", regulation.AuthTypeTOTP, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	userSession.SetTwoFactorTOTP(ctx.Clock.Now())

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "authentication time", regulation.AuthTypeTOTP, userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if bodyJSON.Workflow == workflowOpenIDConnect {
		handleOIDCWorkflowResponse(ctx, bodyJSON.TargetURL, bodyJSON.WorkflowID)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}
