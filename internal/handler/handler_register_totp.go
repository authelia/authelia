package handler

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/middleware"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

// identityRetrieverFromSession retriever computing the identity from the cookie session.
func identityRetrieverFromSession(ctx *middleware.AutheliaCtx) (*session.Identity, error) {
	userSession := ctx.GetSession()

	if len(userSession.Emails) == 0 {
		return nil, fmt.Errorf("user %s does not have any email address", userSession.Username)
	}

	return &session.Identity{
		Username: userSession.Username,
		Email:    userSession.Emails[0],
	}, nil
}

func isTokenUserValidFor2FARegistration(ctx *middleware.AutheliaCtx, username string) bool {
	return ctx.GetSession().Username == username
}

// TOTPIdentityStart the handler for initiating the identity validation.
var TOTPIdentityStart = middleware.IdentityVerificationStart(middleware.IdentityVerificationStartArgs{
	MailTitle:             "Register your mobile",
	MailButtonContent:     "Register",
	TargetEndpoint:        "/one-time-password/register",
	ActionClaim:           ActionTOTPRegistration,
	IdentityRetrieverFunc: identityRetrieverFromSession,
}, nil)

func totpIdentityFinish(ctx *middleware.AutheliaCtx, username string) {
	var (
		config *model.TOTPConfiguration
		err    error
	)

	if config, err = ctx.Providers.TOTP.Generate(username); err != nil {
		ctx.Error(fmt.Errorf("unable to generate TOTP key: %s", err), messageUnableToRegisterOneTimePassword)
	}

	err = ctx.Providers.StorageProvider.SaveTOTPConfiguration(ctx, *config)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to save TOTP secret in DB: %s", err), messageUnableToRegisterOneTimePassword)
		return
	}

	response := TOTPKeyResponse{
		OTPAuthURL:   config.URI(),
		Base32Secret: string(config.Secret),
	}

	err = ctx.SetJSONBody(response)
	if err != nil {
		ctx.Logger.Errorf("Unable to set TOTP key response in body: %s", err)
	}
}

// TOTPIdentityFinish the handler for finishing the identity validation.
var TOTPIdentityFinish = middleware.IdentityVerificationFinish(
	middleware.IdentityVerificationFinishArgs{
		ActionClaim:          ActionTOTPRegistration,
		IsTokenUserValidFunc: isTokenUserValidFor2FARegistration,
	}, totpIdentityFinish)
