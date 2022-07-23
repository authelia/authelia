package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
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
		ctx.Error(fmt.Errorf("unable to parse body during logout: %s", err), messageOperationFailed)
	}

	domain, err := ctx.GetCurrentSessionDomain()
	if err != nil {
		ctx.Error(fmt.Errorf(logFmtErrObtainSessionProvider, domain, err), messageOperationFailed)
		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	sessionProvider, err := ctx.Providers.SessionProvider.Get(domain)
	if err != nil {
		ctx.Error(fmt.Errorf(logFmtErrObtainSessionProvider, domain, err), messageOperationFailed)
		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	err = sessionProvider.DestroySession(ctx.RequestCtx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to destroy session during logout: %s", err), messageOperationFailed)
	}

	redirectionURL, err := url.Parse(body.TargetURL)
	if err == nil {
		responseBody.SafeTargetURL = utils.IsRedirectionSafe(*redirectionURL, ctx.Configuration.Session.GetProtectedDomains()...)
	}

	if body.TargetURL != "" {
		ctx.Logger.Debugf("Logout target url is %s, safe %t", body.TargetURL, responseBody.SafeTargetURL)
	}

	err = ctx.SetJSONBody(responseBody)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to set body during logout: %s", err), messageOperationFailed)
	}
}
