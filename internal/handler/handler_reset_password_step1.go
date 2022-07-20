package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/middleware"
	"github.com/authelia/authelia/v4/internal/session"
)

func identityRetrieverFromStorage(ctx *middleware.AutheliaCtx) (*session.Identity, error) {
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
		Email:       details.Emails[0],
		DisplayName: details.DisplayName,
	}, nil
}

// ResetPasswordIdentityStart the handler for initiating the identity validation for resetting a password.
// We need to ensure the attacker cannot perform user enumeration by always replying with 200 whatever what happens in backend.
var ResetPasswordIdentityStart = middleware.IdentityVerificationStart(middleware.IdentityVerificationStartArgs{
	MailTitle:             "Reset your password",
	MailButtonContent:     "Reset",
	TargetEndpoint:        "/reset-password/step2",
	ActionClaim:           ActionResetPassword,
	IdentityRetrieverFunc: identityRetrieverFromStorage,
}, middleware.TimingAttackDelay(10, 250, 85, time.Millisecond*500, false))

func resetPasswordIdentityFinish(ctx *middleware.AutheliaCtx, username string) {
	userSession := ctx.GetSession()
	// TODO(c.michaud): use JWT tokens to expire the request in only few seconds for better security.
	userSession.PasswordResetUsername = &username

	err := ctx.SaveSession(userSession)
	if err != nil {
		ctx.Logger.Errorf("Unable to clear password reset flag in session for user %s: %s", userSession.Username, err)
	}

	ctx.ReplyOK()
}

// ResetPasswordIdentityFinish the handler for finishing the identity validation.
var ResetPasswordIdentityFinish = middleware.IdentityVerificationFinish(
	middleware.IdentityVerificationFinishArgs{ActionClaim: ActionResetPassword}, resetPasswordIdentityFinish)
