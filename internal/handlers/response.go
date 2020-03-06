package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/utils"
)

// Handle1FAResponse handle the redirection upon 1FA authentication
func Handle1FAResponse(ctx *middlewares.AutheliaCtx, targetURI string, username string, groups []string) {
	if targetURI == "" {
		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && ctx.Configuration.DefaultRedirectionURL != "" {
			ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL})
		} else {
			ctx.ReplyOK()
		}
		return
	}

	targetURL, err := url.ParseRequestURI(targetURI)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to parse target URL %s: %s", targetURI, err), authenticationFailedMessage)
		return
	}

	requiredLevel := ctx.Providers.Authorizer.GetRequiredLevel(authorization.Subject{
		Username: username,
		Groups:   groups,
		IP:       ctx.RemoteIP(),
	}, *targetURL)

	ctx.Logger.Debugf("Required level for the URL %s is %d", targetURI, requiredLevel)

	if requiredLevel > authorization.OneFactor {
		ctx.Logger.Warnf("%s requires more than 1FA, cannot be redirected to", targetURI)
		ctx.ReplyOK()
		return
	}

	safeRedirection := utils.IsRedirectionSafe(*targetURL, ctx.Configuration.Session.Domain)

	if !safeRedirection {
		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && ctx.Configuration.DefaultRedirectionURL != "" {
			ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL})
		} else {
			ctx.ReplyOK()
		}
		return
	}

	ctx.Logger.Debugf("Redirection URL %s is safe", targetURI)
	response := redirectResponse{Redirect: targetURI}
	ctx.SetJSONBody(response)
}

// Handle2FAResponse handle the redirection upon 2FA authentication
func Handle2FAResponse(ctx *middlewares.AutheliaCtx, targetURI string) {
	if targetURI == "" {
		if ctx.Configuration.DefaultRedirectionURL != "" {
			ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL})
		} else {
			ctx.ReplyOK()
		}
		return
	}

	targetURL, err := url.ParseRequestURI(targetURI)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to parse target URL: %s", err), mfaValidationFailedMessage)
		return
	}

	if targetURL != nil && utils.IsRedirectionSafe(*targetURL, ctx.Configuration.Session.Domain) {
		ctx.SetJSONBody(redirectResponse{Redirect: targetURI})
	} else {
		ctx.ReplyOK()
	}
}
