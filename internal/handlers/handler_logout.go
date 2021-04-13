package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/utils"
)

type logoutBody struct {
	RedirectionURL string `json:"redirection_url"`
}

type logoutResponseBody struct {
	SafeRedirection bool `json:"safe_redirection"`
}

// LogoutPost is the handler logging out the user attached to the given cookie.
func LogoutPost(ctx *middlewares.AutheliaCtx) {
	body := logoutBody{}
	responseBody := logoutResponseBody{SafeRedirection: false}

	ctx.Logger.Tracef("Attempting to decode body")

	err := ctx.ParseBody(&body)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to parse body during logout: %s", err), operationFailedMessage)
	}

	ctx.Logger.Tracef("Attempting to destroy session")

	err = ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to destroy session during logout: %s", err), operationFailedMessage)
	}

	redirectionURL, err := url.Parse(body.RedirectionURL)
	if err == nil {
		responseBody.SafeRedirection = utils.IsRedirectionSafe(*redirectionURL, ctx.Configuration.Session.Domain)
	}

	err = ctx.SetJSONBody(responseBody)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to set body during logout: %s", err), operationFailedMessage)
	}

	ctx.ReplyOK()
}
