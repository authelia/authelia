package handlers

import (
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// Handle1FAResponse handle the redirection upon 1FA authentication.
func Handle1FAResponse(ctx *middlewares.AutheliaCtx, targetURI, requestMethod string, username string, groups []string) {
	var err error

	if len(targetURI) == 0 {
		defaultRedirectionURL := ctx.GetDefaultRedirectionURL()

		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && defaultRedirectionURL != nil {
			if err = ctx.SetJSONBody(redirectResponse{Redirect: defaultRedirectionURL.String()}); err != nil {
				ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
			}
		} else {
			ctx.ReplyOK()
		}

		return
	}

	var targetURL *url.URL

	if targetURL, err = url.ParseRequestURI(targetURI); err != nil {
		ctx.Error(fmt.Errorf("unable to parse target URL %s: %w", targetURI, err), messageAuthenticationFailed)

		return
	}

	_, requiredLevel := ctx.Providers.Authorizer.GetRequiredLevel(
		authorization.Subject{
			Username: username,
			Groups:   groups,
			IP:       ctx.RemoteIP(),
		},
		authorization.NewObject(targetURL, requestMethod))

	ctx.Logger.Debugf("Required level for the URL %s is %s", targetURI, requiredLevel)

	if requiredLevel == authorization.TwoFactor {
		ctx.Logger.Warnf("%s requires 2FA, cannot be redirected yet", targetURI)
		ctx.ReplyOK()

		return
	}

	if !ctx.IsSafeRedirectionTargetURI(targetURL) {
		ctx.Logger.Debugf("Redirection URL %s is not safe", targetURI)

		defaultRedirectionURL := ctx.GetDefaultRedirectionURL()

		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && defaultRedirectionURL != nil {
			if err = ctx.SetJSONBody(redirectResponse{Redirect: defaultRedirectionURL.String()}); err != nil {
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
		defaultRedirectionURL := ctx.GetDefaultRedirectionURL()

		if defaultRedirectionURL == nil {
			ctx.ReplyOK()

			return
		}

		if err = ctx.SetJSONBody(redirectResponse{Redirect: defaultRedirectionURL.String()}); err != nil {
			ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
		}

		return
	}

	var (
		parsedURI *url.URL
		safe      bool
	)

	if parsedURI, err = url.ParseRequestURI(targetURI); err != nil {
		ctx.Error(fmt.Errorf("unable to determine if URI '%s' is safe to redirect to: failed to parse URI '%s': %w", targetURI, targetURI, err), messageMFAValidationFailed)
		return
	}

	safe = ctx.IsSafeRedirectionTargetURI(parsedURI)

	if safe {
		ctx.Logger.Debugf("Redirection URL %s is safe", targetURI)

		if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURI}); err != nil {
			ctx.Logger.Errorf("Unable to set redirection URL in body: %s", err)
		}

		return
	}

	ctx.ReplyOK()
}

func handleOIDCWorkflowResponse(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id string) {
	var (
		workflowID uuid.UUID
		client     oidc.Client
		consent    *model.OAuth2ConsentSession
		err        error
	)

	if workflowID, err = uuid.Parse(id); err != nil {
		ctx.Error(fmt.Errorf("unable to parse consent session challenge id '%s': %w", id, err), messageAuthenticationFailed)

		return
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, workflowID); err != nil {
		ctx.Error(fmt.Errorf("unable to load consent session by challenge id '%s': %w", id, err), messageAuthenticationFailed)

		return
	}

	if consent.Responded() {
		ctx.Error(fmt.Errorf("consent has already been responded to '%s': %w", id, err), messageAuthenticationFailed)

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID); err != nil {
		ctx.Error(fmt.Errorf("unable to get client for client with id '%s' with consent challenge id '%s': %w", id, consent.ChallengeID, err), messageAuthenticationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Error(fmt.Errorf("unable to redirect for authorization/consent for client with id '%s' with consent challenge id '%s': user is anonymous", client.GetID(), consent.ChallengeID), messageAuthenticationFailed)

		return
	}

	if !consent.Subject.Valid {
		if consent.Subject.UUID, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			ctx.Error(fmt.Errorf("unable to determine consent subject for client with id '%s' with consent challenge id '%s': %w", client.GetID(), consent.ChallengeID, err), messageAuthenticationFailed)

			return
		}

		consent.Subject.Valid = true

		if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionSubject(ctx, consent); err != nil {
			ctx.Error(fmt.Errorf("unable to update consent subject for client with id '%s' with consent challenge id '%s': %w", client.GetID(), consent.ChallengeID, err), messageAuthenticationFailed)

			return
		}
	}

	var (
		issuer *url.URL
		form   url.Values
	)

	issuer = ctx.RootURL()

	if form, err = consent.GetForm(); err != nil {
		ctx.Error(fmt.Errorf("unable to get authorization form values from consent session with challenge id '%s': %w", consent.ChallengeID, err), messageAuthenticationFailed)

		return
	}

	if oidc.RequestFormRequiresLogin(form, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		targetURL := issuer.JoinPath(oidc.EndpointPathConsentLogin)

		query := targetURL.Query()
		query.Set(queryArgWorkflow, workflowOpenIDConnect)
		query.Set(queryArgWorkflowID, workflowID.String())

		targetURL.RawQuery = query.Encode()

		if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURL.String()}); err != nil {
			ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
		}

		return
	}

	level := client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()})

	switch {
	case authorization.IsAuthLevelSufficient(userSession.AuthenticationLevel, level), level == authorization.Denied:
		targetURL := issuer.JoinPath(oidc.EndpointPathAuthorization)

		form.Set(queryArgConsentID, workflowID.String())
		targetURL.RawQuery = form.Encode()

		if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURL.String()}); err != nil {
			ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
		}
	default:
		ctx.Logger.Warnf("OpenID Connect client '%s' requires 2FA, cannot be redirected yet", client.GetID())
		ctx.ReplyOK()

		return
	}
}

func markAuthenticationAttempt(ctx *middlewares.AutheliaCtx, successful bool, bannedUntil *time.Time, username string, authType string, errAuth error) (err error) {
	// We only Mark if there was no underlying error.
	ctx.Logger.Debugf("Mark %s authentication attempt made by user '%s'", authType, username)

	var (
		requestURI, requestMethod string
	)

	if referer := ctx.Request.Header.Referer(); referer != nil {
		var refererURL *url.URL

		if refererURL, err = url.ParseRequestURI(string(referer)); err == nil {
			requestURI = refererURL.Query().Get(queryArgRD)
			requestMethod = refererURL.Query().Get(queryArgRM)
		}
	}

	if err = ctx.Providers.Regulator.Mark(ctx, successful, bannedUntil != nil, username, requestURI, requestMethod, authType); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to mark %s authentication attempt by user '%s'", authType, username)

		return err
	}

	if successful {
		ctx.Logger.Debugf("Successful %s authentication attempt made by user '%s'", authType, username)
	} else {
		switch {
		case errAuth != nil:
			ctx.Logger.WithError(errAuth).Errorf("Unsuccessful %s authentication attempt by user '%s'", authType, username)
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
