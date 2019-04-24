package handlers

import (
	"fmt"
	"net/url"

	"github.com/clems4ever/authelia/authentication"
	"github.com/clems4ever/authelia/middlewares"
	"github.com/pquerna/otp/totp"
)

// SecondFactorTOTPPost validate the TOTP passcode provided by the user.
func SecondFactorTOTPPost(ctx *middlewares.AutheliaCtx) {
	bodyJSON := signTOTPRequestBody{}
	err := ctx.ParseBody(&bodyJSON)

	if err != nil {
		ctx.Error(err, mfaValidationFailedMessage)
		return
	}

	userSession := ctx.GetSession()
	secret, err := ctx.Providers.StorageProvider.LoadTOTPSecret(userSession.Username)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to load TOTP secret: %s", err), mfaValidationFailedMessage)
		return
	}

	isValid := totp.Validate(bodyJSON.Token, secret)

	if !isValid {
		ctx.Error(fmt.Errorf("Wrong passcode during TOTP validation for user %s", userSession.Username), mfaValidationFailedMessage)
		return
	}

	userSession.AuthenticationLevel = authentication.TwoFactor
	err = ctx.SaveSession(userSession)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to update the authentication level with TOTP: %s", err), mfaValidationFailedMessage)
		return
	}

	if bodyJSON.TargetURL != "" {
		targetURL, err := url.ParseRequestURI(bodyJSON.TargetURL)
		if err != nil {
			ctx.Error(fmt.Errorf("Unable to parse URL with TOTP: %s", err), mfaValidationFailedMessage)
			return
		}

		if targetURL != nil && isRedirectionSafe(*targetURL, ctx.Configuration.Session.Domain) {
			ctx.SetJSONBody(redirectResponse{bodyJSON.TargetURL})
		} else {
			ctx.ReplyOK()
		}
	} else {
		ctx.ReplyOK()
	}
}
