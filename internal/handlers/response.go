package handlers

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

// Handle1FAResponse handle the redirection upon 1FA authentication.
func Handle1FAResponse(ctx *middlewares.AutheliaCtx, targetURI, requestMethod, username string, groups []string) {
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

// HandlePasskeyResponse is a specialized handler for the Passkey login flow which switches adaptively between the 1FA and 2FA response handlers respectively.
func HandlePasskeyResponse(ctx *middlewares.AutheliaCtx, targetURI, requestMethod, username string, groups []string, isTwoFactor bool) {
	if isTwoFactor {
		Handle2FAResponse(ctx, targetURI)
	}

	Handle1FAResponse(ctx, targetURI, requestMethod, username, groups)
}

func handleFlowResponse(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, flow, subflow string) {
	switch flow {
	case flowNameOpenIDConnect:
		handleFlowResponseOpenIDConnect(ctx, userSession, id, subflow)
	default:
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.WithFields(map[string]any{"flow_id": id, "flow": flow, "subflow": subflow}).Error("Failed to find flow handler for the given flow parameters.")
	}
}

func handleFlowResponseOpenIDConnect(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, subflow string) {
	var (
		flowID  uuid.UUID
		client  oidc.Client
		consent *model.OAuth2ConsentSession
		err     error
	)

	if flowID, err = uuid.Parse(id); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"flow_id": id, "flow": flowNameOpenIDConnect, "subflow": subflow}).
			Error("Failed to parse flow id for consent session.")

		return
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, flowID); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"flow_id": flowID.String(), "flow": flowNameOpenIDConnect, "subflow": subflow}).
			Error("Failed load consent session with the provided flow id.")

		return
	}

	if consent.Responded() {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithFields(map[string]any{"flow_id": flowID.String(), "flow": flowNameOpenIDConnect, "subflow": subflow}).
			Error("Consent session has already been responded to.")

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"flow_id": flowID.String(), "flow": flowNameOpenIDConnect, "subflow": subflow, "client_id": consent.ClientID}).
			Error("Failed to get client by 'client_id' for consent session.")

		return
	}

	if userSession.IsAnonymous() {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"flow_id": flowID.String(), "flow": flowNameOpenIDConnect, "subflow": subflow, "client_id": client.GetID()}).
			Error("Failed to redirect for consent as the user is anonymous.")

		return
	}

	var (
		issuer *url.URL
		form   url.Values
	)

	issuer = ctx.RootURL()

	if form, err = consent.GetForm(); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"flow_id": flowID.String(), "flow": flowNameOpenIDConnect, "subflow": subflow, "client_id": client.GetID(), "username": userSession.Username}).
			Error("Failed to get authorization form values from consent session.")

		return
	}

	if oidc.RequestFormRequiresLogin(form, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		targetURL := issuer.JoinPath(oidc.FrontendEndpointPathConsentLogin)

		query := targetURL.Query()
		query.Set(queryArgFlow, flowNameOpenIDConnect)
		query.Set(queryArgFlowID, flowID.String())

		targetURL.RawQuery = query.Encode()

		if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURL.String()}); err != nil {
			ctx.Logger.
				WithError(err).
				WithFields(map[string]any{"flow_id": flowID.String(), "flow": flowNameOpenIDConnect, "subflow": subflow, "client_id": client.GetID(), "username": userSession.Username}).
				Error("Failed to marshal JSON response body for consent redirection.")
		}

		return
	}

	level := client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()})

	switch {
	case authorization.IsAuthLevelSufficient(userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA), level), level == authorization.Denied:
		targetURL := issuer.JoinPath(oidc.EndpointPathAuthorization)

		form.Set(queryArgConsentID, flowID.String())
		targetURL.RawQuery = form.Encode()

		if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURL.String()}); err != nil {
			ctx.Logger.
				WithError(err).
				WithFields(map[string]any{"flow_id": flowID.String(), "flow": flowNameOpenIDConnect, "subflow": subflow, "client_id": client.GetID(), "username": userSession.Username}).
				Error("Failed to marshal JSON response body for authorization redirection.")
		}
	default:
		ctx.Logger.Warnf("OpenID Connect client '%s' requires 2FA, cannot be redirected yet", client.GetID())

		ctx.Logger.
			WithFields(map[string]any{"flow_id": flowID.String(), "flow": flowNameOpenIDConnect, "subflow": subflow, "client_id": client.GetID(), "username": userSession.Username}).
			Info("OpenID Connect 1.0 client requires 2FA.")

		ctx.ReplyOK()

		return
	}
}

func doMarkAuthenticationAttempt(ctx *middlewares.AutheliaCtx, successful bool, ban *regulation.Ban, authType string, errAuth error) {
	var (
		requestURI, requestMethod string
		err                       error
	)

	if referer := ctx.Request.Header.Referer(); referer != nil {
		var refererURL *url.URL

		if refererURL, err = url.ParseRequestURI(string(referer)); err == nil {
			requestURI = refererURL.Query().Get(queryArgRD)
			requestMethod = refererURL.Query().Get(queryArgRM)
		}
	}

	doMarkAuthenticationAttemptWithRequest(ctx, successful, ban, authType, requestURI, requestMethod, errAuth)
}

func doMarkAuthenticationAttemptWithRequest(ctx *middlewares.AutheliaCtx, successful bool, ban *regulation.Ban, authType, requestURI, requestMethod string, errAuth error) {
	// We only Mark if there was no underlying error.
	ctx.Logger.Debugf("Mark %s authentication attempt made by user '%s'", authType, ban.Value())

	ctx.Providers.Regulator.HandleAttempt(ctx, successful, ban.IsBanned(), ban.Value(), requestURI, requestMethod, authType)

	if successful {
		ctx.Logger.Debugf("Successful %s authentication attempt made by user '%s'", authType, ban.Value())
	} else {
		switch {
		case errAuth != nil:
			ctx.Logger.WithError(errAuth).Errorf("Unsuccessful %s authentication attempt by user '%s'", authType, ban.Value())
		case ban.IsBanned():
			ctx.Logger.Errorf("Unsuccessful %s authentication attempt by user '%s' and they are banned until %s", authType, ban.Value(), ban.FormatExpires())
		default:
			ctx.Logger.Errorf("Unsuccessful %s authentication attempt by user '%s'", authType, ban.Value())
		}
	}
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
