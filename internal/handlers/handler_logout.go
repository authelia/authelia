package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/utils"
)

type logoutBody struct {
	TargetURL string `json:"targetURL"`
}

type logoutResponseBody struct {
	SafeTargetURL bool `json:"safeTargetURL"`
}

// LogoutPost is the handler logging out the user attached to the given cookie.
func LogoutPost(ctx *middlewares.AutheliaCtx) {
	body := logoutBody{}
	responseBody := logoutResponseBody{SafeTargetURL: false}

	ctx.Logger.Tracef("Attempting to decode body")

	err := ctx.ParseBody(&body)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to parse body during logout: %s", err), messageOperationFailed)
	}

	ctx.Logger.Tracef("Attempting to destroy session")

	err = ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to destroy session during logout: %s", err), messageOperationFailed)
	}

	redirectionURL, err := url.Parse(body.TargetURL)
	if err == nil {
		responseBody.SafeTargetURL = utils.IsRedirectionSafe(*redirectionURL, ctx.Configuration.Session.Domain)
	}

	if body.TargetURL != "" {
		ctx.Logger.Debugf("Logout target url is %s, safe %t", body.TargetURL, responseBody.SafeTargetURL)
	}

	err = ctx.SetJSONBody(responseBody)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to set body during logout: %s", err), messageOperationFailed)
	}
}
