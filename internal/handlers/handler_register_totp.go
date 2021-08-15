package handlers

import (
	"fmt"

	"github.com/pquerna/otp/totp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

// identityRetrieverFromSession retriever computing the identity from the cookie session.
func identityRetrieverFromSession(ctx *middlewares.AutheliaCtx) (*session.Identity, error) {
	userSession := ctx.GetSession()

	if len(userSession.Emails) == 0 {
		return nil, fmt.Errorf("user %s does not have any email address", userSession.Username)
	}

	return &session.Identity{
		Username: userSession.Username,
		Email:    userSession.Emails[0],
	}, nil
}

func isTokenUserValidFor2FARegistration(ctx *middlewares.AutheliaCtx, username string) bool {
	return ctx.GetSession().Username == username
}

// SecondFactorTOTPIdentityStart the handler for initiating the identity validation.
var SecondFactorTOTPIdentityStart = middlewares.IdentityVerificationStart(middlewares.IdentityVerificationStartArgs{
	MailTitle:             "Register your mobile",
	MailButtonContent:     "Register",
	TargetEndpoint:        "/one-time-password/register",
	ActionClaim:           ActionTOTPRegistration,
	IdentityRetrieverFunc: identityRetrieverFromSession,
})

func secondFactorTOTPIdentityFinish(ctx *middlewares.AutheliaCtx, username string) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      ctx.Configuration.TOTP.Issuer,
		AccountName: username,
		SecretSize:  32,
		Period:      uint(ctx.Configuration.TOTP.Period),
	})

	if err != nil {
		ctx.Error(fmt.Errorf("unable to generate TOTP key: %s", err), messageUnableToRegisterOneTimePassword)
		return
	}

	err = ctx.Providers.StorageProvider.SaveTOTPSecret(username, key.Secret())
	if err != nil {
		ctx.Error(fmt.Errorf("unable to save TOTP secret in DB: %s", err), messageUnableToRegisterOneTimePassword)
		return
	}

	response := TOTPKeyResponse{
		OTPAuthURL:   key.URL(),
		Base32Secret: key.Secret(),
	}

	err = ctx.SetJSONBody(response)
	if err != nil {
		ctx.Logger.Errorf("unable to set TOTP key response in body: %s", err)
	}
}

// SecondFactorTOTPIdentityFinish the handler for finishing the identity validation.
var SecondFactorTOTPIdentityFinish = middlewares.IdentityVerificationFinish(
	middlewares.IdentityVerificationFinishArgs{
		ActionClaim:          ActionTOTPRegistration,
		IsTokenUserValidFunc: isTokenUserValidFor2FARegistration,
	}, secondFactorTOTPIdentityFinish)
