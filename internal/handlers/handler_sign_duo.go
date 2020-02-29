package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/duo"
	"github.com/authelia/authelia/internal/logging"
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
		log := logging.Logger()
		remoteIP := ctx.RemoteIP().String()

		log.Debugf("Starting 2FA Duo Push Auth Attempt for %s from IP %s", userSession.Username, remoteIP)

		values := url.Values{}
		// { username, ipaddr: clientIP, factor: "push", device: "auto", pushinfo: `target%20url=${targetURL}`}
		values.Set("username", userSession.Username)
		values.Set("ipaddr", remoteIP)
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

		if duoResponse.Stat == "FAIL" {
			if duoResponse.Code == 40002 {
				log.Warnf("Duo failed to process the auth request for %s from %s: %s (%s), error code %d. "+
					"This error often occurs if you've not setup the username in the Admin Dashboard.",
					userSession.Username, remoteIP, duoResponse.Message, duoResponse.MessageDetail, duoResponse.Code)
			} else {
				log.Warnf("Duo failed to process the auth request for %s from %s: %s (%s), error code %d.",
					userSession.Username, remoteIP, duoResponse.Message, duoResponse.MessageDetail, duoResponse.Code)
			}
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

		Handle2FAResponse(ctx, requestBody.TargetURL)
	}
}
