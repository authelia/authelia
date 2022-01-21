package handlers

import (
	"fmt"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

// handleOIDCWorkflowResponse handle the redirection upon authentication in the OIDC workflow.
func handleOIDCWorkflowResponse(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	if !authorization.IsAuthLevelSufficient(userSession.AuthenticationLevel, userSession.OIDCWorkflowSession.RequiredAuthorizationLevel) {
		ctx.Logger.Warnf("OpenID Connect client '%s' requires 2FA, cannot be redirected yet", userSession.OIDCWorkflowSession.ClientID)
		ctx.ReplyOK()

		return
	}

	uri, err := ctx.ExternalRootURL()
	if err != nil {
		ctx.Logger.Errorf("Unable to determine external Base URL: %v", err)

		respondUnauthorized(ctx, messageOperationFailed)

		return
	}

	if isConsentMissing(
		userSession.OIDCWorkflowSession,
		userSession.OIDCWorkflowSession.RequestedScopes,
		userSession.OIDCWorkflowSession.RequestedAudience) {
		err = ctx.SetJSONBody(redirectResponse{Redirect: fmt.Sprintf("%s/consent", uri)})

		if err != nil {
			ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
		}
	} else {
		err = ctx.SetJSONBody(redirectResponse{Redirect: userSession.OIDCWorkflowSession.AuthURI})
		if err != nil {
			ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
		}
	}
}

// Handle1FAResponse handle the redirection upon 1FA authentication.
func Handle1FAResponse(ctx *middlewares.AutheliaCtx, targetURI, requestMethod string, username string, groups []string) {
	if targetURI == "" {
		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && ctx.Configuration.DefaultRedirectionURL != "" {
			err := ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL})
			if err != nil {
				ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
			}
		} else {
			ctx.ReplyOK()
		}

		return
	}

	targetURL, err := url.ParseRequestURI(targetURI)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to parse target URL %s: %s", targetURI, err), messageAuthenticationFailed)
		return
	}

	requiredLevel := ctx.Providers.Authorizer.GetRequiredLevel(
		authorization.Subject{
			Username: username,
			Groups:   groups,
			IP:       ctx.RemoteIP(),
		},
		authorization.NewObject(targetURL, requestMethod))

	ctx.Logger.Debugf("Required level for the URL %s is %d", targetURI, requiredLevel)

	if requiredLevel == authorization.TwoFactor {
		ctx.Logger.Warnf("%s requires 2FA, cannot be redirected yet", targetURI)
		ctx.ReplyOK()

		return
	}

	safeRedirection := utils.IsRedirectionSafe(*targetURL, ctx.Configuration.Session.Domain)

	if !safeRedirection {
		ctx.Logger.Debugf("Redirection URL %s is not safe", targetURI)

		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && ctx.Configuration.DefaultRedirectionURL != "" {
			err := ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL})
			if err != nil {
				ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
			}
		} else {
			ctx.ReplyOK()
		}

		return
	}

	ctx.Logger.Debugf("Redirection URL %s is safe", targetURI)
	err = ctx.SetJSONBody(redirectResponse{Redirect: targetURI})

	if err != nil {
		ctx.Logger.Errorf("Unable to set redirection URL in body: %s", err)
	}
}

// Handle2FAResponse handle the redirection upon 2FA authentication.
func Handle2FAResponse(ctx *middlewares.AutheliaCtx, targetURI string) {
	if targetURI == "" {
		if ctx.Configuration.DefaultRedirectionURL != "" {
			err := ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL})
			if err != nil {
				ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
			}
		} else {
			ctx.ReplyOK()
		}

		return
	}

	safe, err := utils.IsRedirectionURISafe(targetURI, ctx.Configuration.Session.Domain)

	if err != nil {
		ctx.Error(fmt.Errorf("unable to check target URL: %s", err), messageMFAValidationFailed)
		return
	}

	if safe {
		ctx.Logger.Debugf("Redirection URL %s is safe", targetURI)
		err := ctx.SetJSONBody(redirectResponse{Redirect: targetURI})

		if err != nil {
			ctx.Logger.Errorf("Unable to set redirection URL in body: %s", err)
		}
	} else {
		ctx.ReplyOK()
	}
}

func markAuthenticationAttempt(ctx *middlewares.AutheliaCtx, successful bool, bannedUntil *time.Time, username string, authType string, errAuth error) (err error) {
	// We only Mark if there was no underlying error.
	ctx.Logger.Debugf("Mark %s authentication attempt made by user '%s'", authType, username)

	var (
		requestURI, requestMethod string
	)

	referer := ctx.Request.Header.Referer()
	if referer != nil {
		refererURL, err := url.Parse(string(referer))
		if err == nil {
			requestURI = refererURL.Query().Get("rd")
			requestMethod = refererURL.Query().Get("rm")
		}
	}

	if err = ctx.Providers.Regulator.Mark(ctx, successful, bannedUntil != nil, username, requestURI, requestMethod, authType, ctx.RemoteIP()); err != nil {
		ctx.Logger.Errorf("Unable to mark %s authentication attempt by user '%s': %+v", authType, username, err)

		return err
	}

	if successful {
		ctx.Logger.Debugf("Successful %s authentication attempt made by user '%s'", authType, username)
	} else {
		switch {
		case errAuth != nil:
			ctx.Logger.Errorf("Unsuccessful %s authentication attempt by user '%s': %+v", authType, username, errAuth)
		case bannedUntil != nil:
			ctx.Logger.Errorf("Unsuccessful %s authentication attempt by user '%s' and they are banned until %s", authType, username, bannedUntil)
		default:
			ctx.Logger.Errorf("Unsuccessful %s authentication attempt by user '%s'", authType, username)
		}
	}

	return nil
}

func respondUnauthorized(ctx *middlewares.AutheliaCtx, message string) {
	ctx.SetStatusCode(fasthttp.StatusUnauthorized)
	ctx.SetJSONError(message)
}
