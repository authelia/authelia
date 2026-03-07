package handlers

import (
	"context"
	"fmt"
	"net"
	"net/url"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/logging"
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
		return
	}

	Handle1FAResponse(ctx, targetURI, requestMethod, username, groups)
}

func handleFlowResponse(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, flow, subflow, userCode string) {
	switch flow {
	case flowNameOpenIDConnect:
		handleFlowResponseOpenIDConnect(ctx, userSession, id, subflow, userCode)
	default:
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: id, logging.FieldFlow: flow, logging.FieldSubflow: subflow}).
			Error("Failed to find flow handler for the given flow parameters")
	}
}

func handleFlowResponseOpenIDConnect(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, subflow, userCode string) {
	switch subflow {
	case "":
		handleFlowResponseOpenIDConnectNoSubflow(ctx, userSession, id, subflow)
	case flowOpenIDConnectSubFlowNameDeviceAuthorization:
		handleFlowResponseOpenIDConnectDeviceAuthSubflow(ctx, userSession, id, subflow, userCode)
	default:
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: id, logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Failed to find flow handler for the given flow parameters")
	}
}

func handleFlowResponseOpenIDConnectNoSubflow(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, subflow string) {
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
			WithFields(map[string]any{logging.FieldFlowID: id, logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Error occurred parsing the consent session flow id")

		return
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, flowID); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Error occurred loading the consent session")

		return
	}

	if consent.Responded() {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Failed to process consent session as it has already been responded to")

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: consent.ClientID}).
			Error("Error occurred loading the client for the consent session")

		return
	}

	if userSession.IsAnonymous() {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID()}).
			Error("Failed to redirect for consent as the user is anonymous")

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
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
			Error("Error occurred getting the original form from the consent session")

		return
	}

	if oidc.RequestFormRequiresLogin(form, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		targetURL := issuer.JoinPath(oidc.FrontendEndpointPathConsentDecision)

		query := targetURL.Query()
		query.Set(queryArgFlow, flowNameOpenIDConnect)
		query.Set(queryArgFlowID, flowID.String())

		targetURL.RawQuery = query.Encode()

		if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURL.String()}); err != nil {
			ctx.Logger.
				WithError(err).
				WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
				Error("Error occurred marshaling JSON response body for consent redirection")
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
				WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
				Error("Error occurred marshaling JSON response body for consent redirection")
		}
	default:
		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
			Info("OpenID Connect 1.0 client requires 2FA")

		ctx.ReplyOK()

		return
	}
}

func handleFlowResponseOpenIDConnectDeviceAuthSubflow(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, subflow, userCode string) {
	var (
		signature string
		device    *model.OAuth2DeviceCodeSession
		client    oidc.Client
		err       error
	)

	if userSession.IsAnonymous() {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Failed to handle flow response as the user is anonymous")

		return
	}

	issuer := ctx.RootURL()
	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if n := len(userCode); n == 0 {
		handleFlowResponseOpenIDConnectDeviceAuthSubflowResponseNoUserCode(ctx, userSession, id, subflow, level, issuer)

		return
	} else if n > 32 {
		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username}).
			Error("Failed to handle flow response as the user code is too long")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if signature, err = ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(ctx, userCode); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username}).
			Error("Error occurred determining the signature of the user code session preventing a successful flow response")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if device, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, signature); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username, logging.FieldSignature: signature}).
			Error("Error occurred using the signature of the user code session to retrieve the device code session preventing a successful flow response")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if device.Subject.Valid || device.ChallengeID.Valid || device.Status != int(oauthelia2.DeviceAuthorizeStatusNew) {
		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username, logging.FieldSignature: signature, logging.FieldClientID: device.ClientID, logging.FieldSessionID: device.ID, logging.FieldSubject: device.Subject, logging.FieldFlowID: device.ChallengeID, logging.FieldStatus: device.Status}).
			Error("Failed to handle flow response as the device code session is in an invalid state")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, device.ClientID); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username, logging.FieldSignature: signature, logging.FieldClientID: device.ClientID}).
			Error("Error occurred loading the client for the device code session")

		return
	}

	handleFlowResponseOpenIDConnectDeviceAuthSubflowResponse(ctx, userSession, subflow, userCode, level, client, issuer)
}

func handleFlowResponseOpenIDConnectDeviceAuthSubflowResponse(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, subflow, userCode string, level authentication.Level, client oidc.Client, issuer *url.URL) {
	var err error

	required := client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()})

	switch {
	case authorization.IsAuthLevelSufficient(level, required), required == authorization.Denied:
		targetURL := issuer.JoinPath(oidc.FrontendEndpointPathConsentDecision)

		query := targetURL.Query()

		query.Set(queryArgFlow, flowNameOpenIDConnect)
		query.Set(queryArgSubflow, flowOpenIDConnectSubFlowNameDeviceAuthorization)
		query.Set(queryArgUserCode, userCode)

		targetURL.RawQuery = query.Encode()

		if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURL.String()}); err != nil {
			ctx.Logger.
				WithError(err).
				WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
				Error("Failed to marshal JSON response body for authorization redirection")
		}
	default:
		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
			Info("OpenID Connect 1.0 client requires 2FA")

		ctx.ReplyOK()

		return
	}
}

func handleFlowResponseOpenIDConnectDeviceAuthSubflowResponseNoUserCode(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, subflow string, level authentication.Level, issuer *url.URL) {
	var err error

	switch {
	case level == authentication.TwoFactor, level == authentication.OneFactor && !ctx.Providers.Authorizer.IsSecondFactorEnabled():
		targetURL := issuer.JoinPath(oidc.FrontendEndpointPathConsentDeviceAuthorization)

		query := targetURL.Query()

		query.Set(queryArgFlow, flowNameOpenIDConnect)
		query.Set(queryArgSubflow, flowOpenIDConnectSubFlowNameDeviceAuthorization)

		if len(id) != 0 {
			query.Set(queryArgFlowID, id)
		}

		targetURL.RawQuery = query.Encode()

		if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURL.String()}); err != nil {
			ctx.Logger.
				WithError(err).
				WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username}).
				Error("Failed to marshal JSON response body for flow response redirection")
		}
	default:
		ctx.ReplyOK()
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

func doMarkAuthenticationAttemptWithRequest(ctx markContext, successful bool, ban *regulation.Ban, authType, requestURI, requestMethod string, errAuth error) {
	// We only Mark if there was no underlying error.
	ctx.GetLogger().Debugf("Mark %s authentication attempt made by user '%s'", authType, ban.Value())

	ctx.GetProviders().Regulator.HandleAttempt(ctx, successful, ban.IsBanned(), ban.Value(), requestURI, requestMethod, authType)

	if successful {
		ctx.GetLogger().Debugf("Successful %s authentication attempt made by user '%s'", authType, ban.Value())
	} else {
		switch {
		case errAuth != nil:
			ctx.GetLogger().WithError(errAuth).Errorf("Unsuccessful %s authentication attempt by user '%s'", authType, ban.Value())
		case ban.IsBanned():
			ctx.GetLogger().Errorf("Unsuccessful %s authentication attempt by user '%s' and they are banned until %s", authType, ban.Value(), ban.FormatExpires())
		default:
			ctx.GetLogger().Errorf("Unsuccessful %s authentication attempt by user '%s'", authType, ban.Value())
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

type markContext interface {
	context.Context

	GetLogger() *logrus.Entry
	GetProviders() middlewares.Providers
	RecordAuthn(success bool, banned bool, authType string)
	RemoteIP() (ip net.IP)
}
