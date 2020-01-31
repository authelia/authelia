package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/utils"
)

func HandleAuthResponse(ctx *middlewares.AutheliaCtx, targetURI string) {
	if targetURI != "" {
		targetURL, err := url.ParseRequestURI(targetURI)

		if err != nil {
			ctx.Error(fmt.Errorf("Unable to parse target URL with U2F: %s", err), mfaValidationFailedMessage)
			return
		}

		if targetURL != nil && utils.IsRedirectionSafe(*targetURL, ctx.Configuration.Session.Domain) {
			ctx.SetJSONBody(redirectResponse{Redirect: targetURI})
		} else {
			ctx.ReplyOK()
		}
	} else {
		if ctx.Configuration.DefaultRedirectionURL != "" {
			ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL})
		} else {
			ctx.ReplyOK()
		}
	}
}
