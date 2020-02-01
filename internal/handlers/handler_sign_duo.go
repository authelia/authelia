package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/duo"
	"github.com/authelia/authelia/internal/middlewares"
)

// SecondFactorDuoPost handler for sending a push notification via duo api.
func SecondFactorDuoPost(duoAPI duo.API) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var requestBody signDuoRequestBody
		err := ctx.ParseBody(&requestBody)

		if err != nil {
			ctx.Error(err, mfaValidationFailedMessage)
			return
		}

		userSession := ctx.GetSession()

		values := url.Values{}
		// { username, ipaddr: clientIP, factor: "push", device: "auto", pushinfo: `target%20url=${targetURL}`}
		values.Set("username", userSession.Username)
		values.Set("ipaddr", ctx.RemoteIP().String())
		values.Set("factor", "push")
		values.Set("device", "auto")
		if requestBody.TargetURL != "" {
			values.Set("pushinfo", fmt.Sprintf("target%%20url=%s", requestBody.TargetURL))
		}

		duoResponse, err := duoAPI.Call(values)
		if err != nil {
			ctx.Error(fmt.Errorf("Duo API errored: %s", err), mfaValidationFailedMessage)
			return
		}

		if duoResponse.Response.Result != "allow" {
			ctx.ReplyUnauthorized()
			return
		}

		userSession.AuthenticationLevel = authentication.TwoFactor
		err = ctx.SaveSession(userSession)

		if err != nil {
			ctx.Error(fmt.Errorf("Unable to update authentication level with Duo: %s", err), mfaValidationFailedMessage)
			return
		}

		HandleAuthResponse(ctx, requestBody.TargetURL)
	}
}
