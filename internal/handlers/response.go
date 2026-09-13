// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"

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
	uri, def, err := resolve1FATargetURI(ctx, targetURI, requestMethod, username, groups)
	if err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		return
	}

	if len(uri) == 0 {
		ctx.ReplyOK()

		return
	}

	if err = ctx.SetJSONBody(redirectResponse{Redirect: uri}); err != nil {
		if def {
			ctx.GetLogger().Errorf("Unable to set default redirection URL in body: %s", err)
		} else {
			ctx.GetLogger().Errorf("Unable to set redirection URL in body: %s", err)
		}
	}
}

// Handle1FARedirect issues a 302 to the validated post-authentication target for flows which are driven by a browser
// redirect rather than an API call.
func Handle1FARedirect(ctx *middlewares.AutheliaCtx, targetURI, requestMethod, username string, groups []string) {
	uri := determine1FATargetURI(ctx, targetURI, requestMethod, username, groups)

	if len(uri) == 0 {
		uri = handle1FARedirectPortalURI(ctx, targetURI, requestMethod)
	}

	ctx.SpecialRedirect(uri, fasthttp.StatusFound)
}

// handle1FARedirectPortalURI returns the portal URI a browser redirect is sent to when the target can't be redirected
// to directly. A safe target which requires a second factor is the only target resolve1FATargetURI resolves to nothing,
// and it is carried to the portal as the rd and rm query parameters, so the portal asks for the second factor and then
// redirects to the target, exactly as it does after a password sign in. Any other target is not carried at all.
func handle1FARedirectPortalURI(ctx *middlewares.AutheliaCtx, targetURI, requestMethod string) (uri string) {
	root := ctx.TemplateRootURL()

	if len(targetURI) == 0 {
		return root.String()
	}

	target, err := url.ParseRequestURI(targetURI)
	if err != nil || !ctx.IsSafeRedirectionTargetURI(target) {
		return root.String()
	}

	query := url.Values{queryArgRD: []string{targetURI}}

	if len(requestMethod) != 0 {
		query.Set(queryArgRM, requestMethod)
	}

	return root.ResolveReference(&url.URL{RawQuery: query.Encode()}).String()
}

func determine1FATargetURI(ctx *middlewares.AutheliaCtx, targetURI, requestMethod, username string, groups []string) (uri string) {
	uri, _, _ = resolve1FATargetURI(ctx, targetURI, requestMethod, username, groups)

	return uri
}

func resolve1FATargetURI(ctx *middlewares.AutheliaCtx, targetURI, requestMethod, username string, groups []string) (uri string, def bool, err error) {
	if len(targetURI) == 0 {
		uri = default1FATargetURI(ctx)

		return uri, len(uri) != 0, nil
	}

	var object *authorization.Object

	if object, err = authorization.NewObjectMethodURL([]byte(requestMethod), []byte(targetURI)); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred parsing the target URL '%s'", targetURI)

		return "", false, err
	}

	_, requiredLevel := ctx.Providers.Authorizer.GetRequiredLevel(
		authorization.Subject{
			Username: username,
			Groups:   groups,
			IP:       ctx.RemoteIP(),
		},
		*object)

	ctx.GetLogger().Debugf("Required level for the URL %s is %s", targetURI, requiredLevel)

	if requiredLevel == authorization.TwoFactor {
		ctx.GetLogger().Warnf("%s requires 2FA, cannot be redirected yet", object.URL)

		return "", false, nil
	}

	if !ctx.IsSafeRedirectionTargetURI(object.URL) {
		ctx.GetLogger().Debugf("Redirection URL %s is not safe", object.URL)

		uri = default1FATargetURI(ctx)

		return uri, len(uri) != 0, nil
	}

	ctx.GetLogger().Debugf("Redirection URL %s is safe", object.URL)

	return targetURI, false, nil
}

func default1FATargetURI(ctx *middlewares.AutheliaCtx) (uri string) {
	defaultRedirectionURL := ctx.GetDefaultRedirectionURL()

	if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && defaultRedirectionURL != nil {
		return defaultRedirectionURL.String()
	}

	return ""
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
			ctx.GetLogger().Errorf("Unable to set default redirection URL in body: %s", err)
		}

		return
	}

	var (
		parsedURI *url.URL
		safe      bool
	)

	if parsedURI, err = url.ParseRequestURI(targetURI); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred determining if the URI '%s' is safe to redirect to as it could not be parsed", targetURI)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	safe = ctx.IsSafeRedirectionTargetURI(parsedURI)

	if safe {
		ctx.GetLogger().Debugf("Redirection URL %s is safe", targetURI)

		if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURI}); err != nil {
			ctx.GetLogger().Errorf("Unable to set redirection URL in body: %s", err)
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

// flowResponseKind is the kind of response which resumes a flow once the user has authenticated.
type flowResponseKind int

const (
	// flowResponseKindFailure is a flow which could not be resumed.
	flowResponseKindFailure flowResponseKind = iota

	// flowResponseKindRedirect is a flow which is resumed by sending the user to another location.
	flowResponseKindRedirect

	// flowResponseKindSecondFactor is a flow which can only be resumed once the user has completed a second factor.
	flowResponseKindSecondFactor
)

// flowResponse is the decision of how a flow is resumed. It is kept separate from how the decision is written, so the
// flows resumed by an API call of the portal and the flows resumed by a browser redirect share the same decision.
type flowResponse struct {
	kind     flowResponseKind
	redirect string
	message  string
}

func newFlowResponseFailure(message string) flowResponse {
	return flowResponse{kind: flowResponseKindFailure, message: message}
}

func newFlowResponseRedirect(redirect string) flowResponse {
	return flowResponse{kind: flowResponseKindRedirect, redirect: redirect}
}

func newFlowResponseSecondFactor() flowResponse {
	return flowResponse{kind: flowResponseKindSecondFactor}
}

// handleFlowResponse resumes a flow once the user has authenticated, responding to an API call of the portal with JSON.
func handleFlowResponse(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, flow, subflow, userCode string) {
	response := resolveFlowResponse(ctx, userSession, id, flow, subflow, userCode)

	switch response.kind {
	case flowResponseKindRedirect:
		if err := ctx.SetJSONBody(redirectResponse{Redirect: response.redirect}); err != nil {
			ctx.GetLogger().
				WithError(err).
				WithFields(map[string]any{logging.FieldFlowID: id, logging.FieldFlow: flow, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username}).
				Error("Error occurred marshaling JSON response body for flow response redirection")
		}
	case flowResponseKindSecondFactor:
		ctx.ReplyOK()
	default:
		ctx.SetJSONError(response.message)
	}
}

func resolveFlowResponse(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, flow, subflow, userCode string) flowResponse {
	switch flow {
	case flowNameOpenIDConnect:
		return resolveFlowResponseOpenIDConnect(ctx, userSession, id, subflow, userCode)
	default:
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: id, logging.FieldFlow: flow, logging.FieldSubflow: subflow}).
			Error("Failed to find flow handler for the given flow parameters")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}
}

func resolveFlowResponseOpenIDConnect(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, subflow, userCode string) flowResponse {
	switch subflow {
	case "":
		return resolveFlowResponseOpenIDConnectNoSubflow(ctx, userSession, id, subflow)
	case flowOpenIDConnectSubFlowNameDeviceAuthorization:
		return resolveFlowResponseOpenIDConnectDeviceAuthSubflow(ctx, userSession, id, subflow, userCode)
	default:
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: id, logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Failed to find flow handler for the given flow parameters")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}
}

func resolveFlowResponseOpenIDConnectNoSubflow(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, subflow string) flowResponse {
	var (
		flowID  uuid.UUID
		client  oidc.Client
		consent *model.OAuth2ConsentSession
		err     error
	)

	if flowID, err = uuid.Parse(id); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: id, logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Error occurred parsing the consent session flow id")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, flowID); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Error occurred loading the consent session")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}

	if consent.Responded() {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Failed to process consent session as it has already been responded to")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: consent.ClientID}).
			Error("Error occurred loading the client for the consent session")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}

	if userSession.IsAnonymous() {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID()}).
			Error("Failed to redirect for consent as the user is anonymous")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}

	var (
		issuer *url.URL
		form   url.Values
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
			Error("Error occurred determining the issuer")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}

	if form, err = consent.GetForm(); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
			Error("Error occurred getting the original form from the consent session")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}

	if oidc.RequestFormRequiresLogin(form, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		targetURL := issuer.JoinPath(oidc.FrontendEndpointPathConsentDecision)

		query := targetURL.Query()
		query.Set(queryArgFlow, flowNameOpenIDConnect)
		query.Set(queryArgFlowID, flowID.String())

		targetURL.RawQuery = query.Encode()

		return newFlowResponseRedirect(targetURL.String())
	}

	level := client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()})

	switch {
	case authorization.IsAuthLevelSufficient(userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA), level), level == authorization.Denied:
		targetURL := issuer.JoinPath(oidc.EndpointPathAuthorization)

		form.Set(queryArgConsentID, flowID.String())
		targetURL.RawQuery = form.Encode()

		return newFlowResponseRedirect(targetURL.String())
	default:
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
			Info("OpenID Connect 1.0 client requires 2FA")

		return newFlowResponseSecondFactor()
	}
}

func resolveFlowResponseOpenIDConnectDeviceAuthSubflow(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, subflow, userCode string) flowResponse {
	var (
		issuer    *url.URL
		signature string
		device    *model.OAuth2DeviceCodeSession
		client    oidc.Client
		err       error
	)

	if userSession.IsAnonymous() {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Failed to handle flow response as the user is anonymous")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Error occurred determining the issuer preventing a successful flow response")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}

	if n := len(userCode); n == 0 {
		return resolveFlowResponseOpenIDConnectDeviceAuthSubflowNoUserCode(ctx, id, level, issuer)
	} else if n > 32 {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username}).
			Error("Failed to handle flow response as the user code is too long")

		return newFlowResponseFailure(messageOperationFailed)
	}

	if signature, err = ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(ctx, userCode); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username}).
			Error("Error occurred determining the signature of the user code session preventing a successful flow response")

		return newFlowResponseFailure(messageOperationFailed)
	}

	if device, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, signature); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username, logging.FieldSignature: signature}).
			Error("Error occurred using the signature of the user code session to retrieve the device code session preventing a successful flow response")

		return newFlowResponseFailure(messageOperationFailed)
	}

	if device.Subject.Valid || device.ChallengeID.Valid || device.Status != int(oauthelia2.DeviceAuthorizeStatusNew) {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username, logging.FieldSignature: signature, logging.FieldClientID: device.ClientID, logging.FieldSessionID: device.ID, logging.FieldSubject: device.Subject, logging.FieldFlowID: device.ChallengeID, logging.FieldStatus: device.Status}).
			Error("Failed to handle flow response as the device code session is in an invalid state")

		return newFlowResponseFailure(messageOperationFailed)
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, device.ClientID); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username, logging.FieldSignature: signature, logging.FieldClientID: device.ClientID}).
			Error("Error occurred loading the client for the device code session")

		return newFlowResponseFailure(messageAuthenticationFailed)
	}

	return resolveFlowResponseOpenIDConnectDeviceAuthSubflowUserCode(ctx, userSession, subflow, userCode, level, client, issuer)
}

func resolveFlowResponseOpenIDConnectDeviceAuthSubflowUserCode(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, subflow, userCode string, level authentication.Level, client oidc.Client, issuer *url.URL) flowResponse {
	required := client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()})

	switch {
	case authorization.IsAuthLevelSufficient(level, required), required == authorization.Denied:
		targetURL := issuer.JoinPath(oidc.FrontendEndpointPathConsentDecision)

		query := targetURL.Query()

		query.Set(queryArgFlow, flowNameOpenIDConnect)
		query.Set(queryArgSubflow, flowOpenIDConnectSubFlowNameDeviceAuthorization)
		query.Set(queryArgUserCode, userCode)

		targetURL.RawQuery = query.Encode()

		return newFlowResponseRedirect(targetURL.String())
	default:
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
			Info("OpenID Connect 1.0 client requires 2FA")

		return newFlowResponseSecondFactor()
	}
}

func resolveFlowResponseOpenIDConnectDeviceAuthSubflowNoUserCode(ctx *middlewares.AutheliaCtx, id string, level authentication.Level, issuer *url.URL) flowResponse {
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

		return newFlowResponseRedirect(targetURL.String())
	default:
		return newFlowResponseSecondFactor()
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
	ctx.GetLogger().Debugf("Mark %s authentication attempt made by user '%s'", authType, ban.Value())

	ctx.GetProviders().Regulator.HandleAttempt(ctx, successful, ban, requestURI, requestMethod, authType)

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
