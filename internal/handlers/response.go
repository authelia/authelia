package handlers

import (
	"fmt"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

// handleOIDCWorkflowResponse handle the redirection upon authentication in the OIDC workflow.
func handleOIDCWorkflowResponse(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	if userSession.ConsentChallengeID == nil {
		ctx.Logger.Errorf("Unable to handle OIDC workflow response because the user session doesn't contain a consent challenge id")

		respondUnauthorized(ctx, messageOperationFailed)

		return
	}

	externalRootURL, err := ctx.ExternalRootURL()
	if err != nil {
		ctx.Logger.Errorf("Unable to determine external Base URL: %v", err)

		respondUnauthorized(ctx, messageOperationFailed)

		return
	}

	consent, err := ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, *userSession.ConsentChallengeID)
	if err != nil {
		ctx.Logger.Errorf("Unable to load consent session from database: %v", err)

		respondUnauthorized(ctx, messageOperationFailed)

		return
	}

	client, err := ctx.Providers.OpenIDConnect.Store.GetFullClient(consent.ClientID)
	if err != nil {
		ctx.Logger.Errorf("Unable to find client for the consent session: %v", err)

		respondUnauthorized(ctx, messageOperationFailed)

		return
	}

	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		ctx.Logger.Warnf("OpenID Connect client '%s' requires 2FA, cannot be redirected yet", client.ID)
		ctx.ReplyOK()

		return
	}

	if consent.Subject.UUID, err = ctx.Providers.OpenIDConnect.Store.GetSubject(ctx, client.GetSectorIdentifier(), userSession.Username); err != nil {
		ctx.Logger.Errorf("Unable to find subject for the consent session: %v", err)

		respondUnauthorized(ctx, messageOperationFailed)

		return
	}

	consent.Subject.Valid = true

	var preConsent *model.OAuth2ConsentSession

	if preConsent, err = getOIDCPreConfiguredConsentFromClientAndConsent(ctx, client, consent); err != nil {
		ctx.Logger.Errorf("Unable to lookup pre-configured consent for the consent session: %v", err)

		respondUnauthorized(ctx, messageOperationFailed)

		return
	}

	if userSession.ConsentChallengeID != nil && preConsent == nil {
		if err = ctx.SetJSONBody(redirectResponse{Redirect: fmt.Sprintf("%s/consent", externalRootURL)}); err != nil {
			ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
		}

		return
	}

	if userSession.ConsentChallengeID != nil {
		userSession.ConsentChallengeID = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.Errorf("Unable to update user session: %v", err)

			respondUnauthorized(ctx, messageOperationFailed)

			return
		}
	}

	if err = ctx.SetJSONBody(redirectResponse{Redirect: fmt.Sprintf("%s%s?%s", externalRootURL, oidc.AuthorizationPath, consent.Form)}); err != nil {
		ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
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
	ctx.SetContentTypeBytes(headerContentTypeValueTextPlain)
	ctx.SetStatusCode(statusCode)
	ctx.SetBodyString(fmt.Sprintf("%d %s", statusCode, fasthttp.StatusMessage(statusCode)))
}
