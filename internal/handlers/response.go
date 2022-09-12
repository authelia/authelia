package handlers

import (
	"fmt"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

// handleOIDCWorkflowResponse handle the redirection upon authentication in the OIDC workflow.
func handleOIDCWorkflowResponse(ctx *middlewares.AutheliaCtx, targetURI string) {
	if len(targetURI) == 0 {
		ctx.Error(fmt.Errorf("unable to parse target URL %s: empty value", targetURI), messageAuthenticationFailed)

		return
	}

	var (
		targetURL *url.URL
		err       error
	)

	if targetURL, err = url.ParseRequestURI(targetURI); err != nil {
		ctx.Error(fmt.Errorf("unable to parse target URL %s: %w", targetURI, err), messageAuthenticationFailed)

		return
	}

	var (
		id     string
		client *oidc.Client
	)

	if id = targetURL.Query().Get("client_id"); len(id) == 0 {
		ctx.Error(fmt.Errorf("unable to get client id from from URL '%s'", targetURL), messageAuthenticationFailed)

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.Store.GetFullClient(id); err != nil {
		ctx.Error(fmt.Errorf("unable to get client for client with id '%s' from URL '%s': %w", id, targetURL, err), messageAuthenticationFailed)

		return
	}

	userSession := ctx.GetSession()

	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		ctx.Logger.Warnf("OpenID Connect client '%s' requires 2FA, cannot be redirected yet", client.ID)
		ctx.ReplyOK()

		return
	}

	if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURL.String()}); err != nil {
		ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
	}
}

// Handle1FAResponse handle the redirection upon 1FA authentication.
func Handle1FAResponse(ctx *middlewares.AutheliaCtx, targetURI, requestMethod string, username string, groups []string) {
	var err error

	if len(targetURI) == 0 {
		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && ctx.Configuration.DefaultRedirectionURL != "" {
			if err = ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL}); err != nil {
				ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
			}
		} else {
			ctx.ReplyOK()
		}

		return
	}

	var targetURL *url.URL

	if targetURL, err = url.ParseRequestURI(targetURI); err != nil {
		ctx.Error(fmt.Errorf("unable to parse target URL %s: %s", targetURI, err), messageAuthenticationFailed)

		return
	}

	_, requiredLevel := ctx.Providers.Authorizer.GetRequiredLevel(
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

	if !utils.IsURISafeRedirection(targetURL, ctx.Configuration.Session.Domain) {
		ctx.Logger.Debugf("Redirection URL %s is not safe", targetURI)

		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && ctx.Configuration.DefaultRedirectionURL != "" {
			if err = ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL}); err != nil {
				ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
			}

			return
		}

		ctx.ReplyOK()

		return
	}

	ctx.Logger.Debugf("Redirection URL %s is safe", targetURI)

	if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURI}); err != nil {
		ctx.Logger.Errorf("Unable to set redirection URL in body: %s", err)
	}
}

// Handle2FAResponse handle the redirection upon 2FA authentication.
func Handle2FAResponse(ctx *middlewares.AutheliaCtx, targetURI string) {
	var err error

	if len(targetURI) == 0 {
		if len(ctx.Configuration.DefaultRedirectionURL) == 0 {
			ctx.ReplyOK()

			return
		}

		if err = ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL}); err != nil {
			ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
		}

		return
	}

	var safe bool

	if safe, err = utils.IsURIStringSafeRedirection(targetURI, ctx.Configuration.Session.Domain); err != nil {
		ctx.Error(fmt.Errorf("unable to check target URL: %s", err), messageMFAValidationFailed)

		return
	}

	if safe {
		ctx.Logger.Debugf("Redirection URL %s is safe", targetURI)

		if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURI}); err != nil {
			ctx.Logger.Errorf("Unable to set redirection URL in body: %s", err)
		}

		return
	}

	ctx.ReplyOK()
}

func markAuthenticationAttempt(ctx *middlewares.AutheliaCtx, successful bool, bannedUntil *time.Time, username string, authType string, errAuth error) (err error) {
	// We only Mark if there was no underlying error.
	ctx.Logger.Debugf("Mark %s authentication attempt made by user '%s'", authType, username)

	var (
		requestURI, requestMethod string
	)

	referer := ctx.Request.Header.Referer()
	if referer != nil {
		refererURL, err := url.ParseRequestURI(string(referer))
		if err == nil {
			requestURI = refererURL.Query().Get("rd")
			requestMethod = refererURL.Query().Get("rm")
		}
	}

	if err = ctx.Providers.Regulator.Mark(ctx, successful, bannedUntil != nil, username, requestURI, requestMethod, authType); err != nil {
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

// SetStatusCodeResponse writes a response status code and an appropriate body on either a
// *fasthttp.RequestCtx or *middlewares.AutheliaCtx.
func SetStatusCodeResponse(ctx *fasthttp.RequestCtx, statusCode int) {
	ctx.Response.Reset()

	middlewares.SetContentTypeTextPlain(ctx)

	ctx.SetStatusCode(statusCode)
	ctx.SetBodyString(fmt.Sprintf("%d %s", statusCode, fasthttp.StatusMessage(statusCode)))
}
