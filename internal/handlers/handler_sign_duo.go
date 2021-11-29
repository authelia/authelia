package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
)

// SecondFactorDuoPost handler for sending a push notification via duo api.
func SecondFactorDuoPost(duoAPI duo.API) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var requestBody signDuoRequestBody

		if err := ctx.ParseBody(&requestBody); err != nil {
			ctx.Logger.Errorf(logFmtErrParseRequestBody, regulation.AuthTypeDUO, err)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		userSession := ctx.GetSession()
		remoteIP := ctx.RemoteIP().String()

		ctx.Logger.Debugf("Starting Duo Push Auth Attempt for user '%s' with IP '%s'", userSession.Username, remoteIP)

		values := url.Values{}

		values.Set("username", userSession.Username)
		values.Set("ipaddr", remoteIP)
		values.Set("factor", "push")
		values.Set("device", "auto")

		if requestBody.TargetURL != "" {
			values.Set("pushinfo", fmt.Sprintf("target%%20url=%s", requestBody.TargetURL))
		}

		duoResponse, err := duoAPI.Call(values, ctx)
		if err != nil {
			ctx.Logger.Errorf("Failed to perform DUO call for user '%s': %+v", userSession.Username, err)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		if duoResponse.Stat == "FAIL" {
			if duoResponse.Code == 40002 {
				ctx.Logger.Warnf("Duo Push Auth failed to process the auth request for %s from %s: %s (%s), error code %d. "+
					"This error often occurs if you've not setup the username in the Admin Dashboard.",
					userSession.Username, remoteIP, duoResponse.Message, duoResponse.MessageDetail, duoResponse.Code)
			} else {
				ctx.Logger.Warnf("Duo Push Auth failed to process the auth request for %s from %s: %s (%s), error code %d.",
					userSession.Username, remoteIP, duoResponse.Message, duoResponse.MessageDetail, duoResponse.Code)
			}
		}

		if duoResponse.Response.Result != testResultAllow {
			_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypeDUO,
				fmt.Errorf("result: %s, code: %d, message: %s (%s)", duoResponse.Response.Result, duoResponse.Code,
					duoResponse.Message, duoResponse.MessageDetail))

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		if err = markAuthenticationAttempt(ctx, true, nil, userSession.Username, regulation.AuthTypeDUO, nil); err != nil {
			respondUnauthorized(ctx, messageMFAValidationFailed)
			return
		}

		if err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx); err != nil {
			ctx.Logger.Errorf(logFmtErrSessionRegenerate, regulation.AuthTypeDUO, userSession.Username, err)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		userSession.SetTwoFactor(ctx.Clock.Now())

		err = ctx.SaveSession(userSession)
		if err != nil {
			ctx.Logger.Errorf(logFmtErrSessionSave, "authentication time", regulation.AuthTypeTOTP, userSession.Username, err)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		if userSession.OIDCWorkflowSession != nil {
			handleOIDCWorkflowResponse(ctx)
		} else {
			Handle2FAResponse(ctx, requestBody.TargetURL)
		}
	}
}
