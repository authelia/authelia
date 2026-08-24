package handlers

import (
	"net/url"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

type logoutBody struct {
	TargetURL string `json:"targetURL"`
}

type logoutResponseBody struct {
	SafeTargetURL bool `json:"safeTargetURL"`
}

// LogoutPOST is the handler logging out the user attached to the given cookie.
func LogoutPOST(ctx *middlewares.AutheliaCtx) {
	body := logoutBody{}
	responseBody := logoutResponseBody{SafeTargetURL: false}

	err := ctx.ParseBody(&body)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred parsing the logout request body")
		ctx.SetJSONError(messageOperationFailed)
	}

	err = ctx.DestroySession()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred destroying the user session during logout")
		ctx.SetJSONError(messageOperationFailed)
	}

	redirectionURL, err := url.ParseRequestURI(body.TargetURL)
	if err == nil {
		responseBody.SafeTargetURL = ctx.IsSafeRedirectionTargetURI(redirectionURL)
	}

	if body.TargetURL != "" {
		ctx.Logger.Debugf("Logout target url is %s, safe %t", body.TargetURL, responseBody.SafeTargetURL)
	}

	err = ctx.SetJSONBody(responseBody)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred setting the logout response body")
		ctx.SetJSONError(messageOperationFailed)
	}
}
