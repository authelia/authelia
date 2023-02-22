package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

func identityRetrieverFromStorage(ctx *middlewares.AutheliaCtx) (*session.Identity, error) {
	var requestBody resetPasswordStep1RequestBody
	err := json.Unmarshal(ctx.PostBody(), &requestBody)

	if err != nil {
		return nil, err
	}

	details, err := ctx.Providers.UserProvider.GetDetails(requestBody.Username)

	if err != nil {
		return nil, err
	}

	if len(details.Emails) == 0 {
		return nil, fmt.Errorf("user %s has no email address configured", requestBody.Username)
	}

	return &session.Identity{
		Username:    requestBody.Username,
		DisplayName: details.DisplayName,
		Email:       details.Emails[0],
	}, nil
}

// ResetPasswordIdentityStart the handler for initiating the identity validation for resetting a password.
// We need to ensure the attacker cannot perform user enumeration by always replying with 200 whatever what happens in backend.
var ResetPasswordIdentityStart = middlewares.IdentityVerificationStart(middlewares.IdentityVerificationStartArgs{
	MailTitle:             "Reset your password",
	MailButtonContent:     "Reset",
	TargetEndpoint:        "/reset-password/step2",
	ActionClaim:           ActionResetPassword,
	IdentityRetrieverFunc: identityRetrieverFromStorage,
}, middlewares.TimingAttackDelay(10, 250, 85, time.Millisecond*500, false))

func resetPasswordIdentityFinish(ctx *middlewares.AutheliaCtx, username string) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.Errorf("Unable to get session to clear password reset flag in session for user %s: %s", userSession.Username, err)
	}

	// TODO(c.michaud): use JWT tokens to expire the request in only few seconds for better security.
	userSession.PasswordResetUsername = &username

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf("Unable to clear password reset flag in session for user %s: %s", userSession.Username, err)
	}

	ctx.ReplyOK()
}

// ResetPasswordIdentityFinish the handler for finishing the identity validation.
var ResetPasswordIdentityFinish = middlewares.IdentityVerificationFinish(
	middlewares.IdentityVerificationFinishArgs{ActionClaim: ActionResetPassword}, resetPasswordIdentityFinish)
