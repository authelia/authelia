package handlers

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
)

// TimeBasedOneTimePasswordPOST validate the TOTP passcode provided by the user.
func TimeBasedOneTimePasswordPOST(ctx *middlewares.AutheliaCtx) {
	bodyJSON := bodySignTOTPRequest{}

	if err := ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.Errorf(logFmtErrParseRequestBody, regulation.AuthTypeTOTP, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	userSession := ctx.GetSession()

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
		ctx.Logger.Errorf(logFmtErrSessionRegenerate, regulation.AuthTypeTOTP, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	config.UpdateSignInInfo(ctx.Clock.Now())

	if err = ctx.Providers.StorageProvider.UpdateTOTPConfigurationSignIn(ctx, config.ID, config.LastUsedAt); err != nil {
		ctx.Logger.Errorf("Unable to save %s device sign in metadata for user '%s': %v", regulation.AuthTypeTOTP, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	fmt.Println("success")

	userSession.SetTwoFactorTOTP(ctx.Clock.Now())

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "authentication time", regulation.AuthTypeTOTP, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if bodyJSON.Workflow == workflowOpenIDConnect {
		handleOIDCWorkflowResponse(ctx, bodyJSON.TargetURL, bodyJSON.WorkflowID)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}
