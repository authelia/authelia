package handlers

import (
	"fmt"
	"net/url"

	"github.com/clems4ever/authelia/authentication"
	"github.com/clems4ever/authelia/middlewares"
)

// SecondFactorU2FSignPost handler for completing a signing request.
func SecondFactorU2FSignPost(ctx *middlewares.AutheliaCtx) {
	var requestBody signU2FRequestBody
	err := ctx.ParseBody(&requestBody)

	if err != nil {
		ctx.Error(err, mfaValidationFailedMessage)
		return
	}

	userSession := ctx.GetSession()
	if userSession.U2FChallenge == nil {
		ctx.Error(fmt.Errorf("U2F signing has not been initiated yet (no challenge)"), mfaValidationFailedMessage)
		return
	}

	if userSession.U2FRegistration == nil {
		ctx.Error(fmt.Errorf("U2F signing has not been initiated yet (no registration)"), mfaValidationFailedMessage)
		return
	}

	// TODO(c.michaud): store the counter to help detecting cloned U2F keys.
	_, err = userSession.U2FRegistration.Authenticate(
		requestBody.SignResponse, *userSession.U2FChallenge, 0)

	if err != nil {
		ctx.Error(err, mfaValidationFailedMessage)
		return
	}

	userSession.AuthenticationLevel = authentication.TwoFactor
	err = ctx.SaveSession(userSession)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to update authentication level with U2F: %s", err), mfaValidationFailedMessage)
		return
	}

	if requestBody.TargetURL != "" {
		targetURL, err := url.ParseRequestURI(requestBody.TargetURL)

		if err != nil {
			ctx.Error(fmt.Errorf("Unable to parse target URL with U2F: %s", err), mfaValidationFailedMessage)
			return
		}

		if targetURL != nil && isRedirectionSafe(*targetURL, ctx.Configuration.Session.Domain) {
			ctx.SetJSONBody(redirectResponse{Redirect: requestBody.TargetURL})
		} else {
			ctx.ReplyOK()
		}
	} else {
		ctx.ReplyOK()
	}
}
