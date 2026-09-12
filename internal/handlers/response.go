// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/events"
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
				ctx.GetLogger().Errorf("Unable to set default redirection URL in body: %s", err)
			}
		} else {
			ctx.ReplyOK()
		}

		return
	}

	var object *authorization.Object

	if object, err = authorization.NewObjectMethodURL([]byte(requestMethod), []byte(targetURI)); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred parsing the target URL '%s'", targetURI)
		ctx.SetJSONError(messageAuthenticationFailed)

		return
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
		ctx.ReplyOK()

		return
	}

	if !ctx.IsSafeRedirectionTargetURI(object.URL) {
		ctx.GetLogger().Debugf("Redirection URL %s is not safe", object.URL)

		defaultRedirectionURL := ctx.GetDefaultRedirectionURL()

		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && defaultRedirectionURL != nil {
			if err = ctx.SetJSONBody(redirectResponse{Redirect: defaultRedirectionURL.String()}); err != nil {
				ctx.GetLogger().Errorf("Unable to set default redirection URL in body: %s", err)
			}

			return
		}

		ctx.ReplyOK()

		return
	}

	ctx.GetLogger().Debugf("Redirection URL %s is safe", object.URL)

	if err = ctx.SetJSONBody(redirectResponse{Redirect: targetURI}); err != nil {
		ctx.GetLogger().Errorf("Unable to set redirection URL in body: %s", err)
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

func handleFlowResponse(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, flow, subflow, userCode string) {
	switch flow {
	case flowNameOpenIDConnect:
		handleFlowResponseOpenIDConnect(ctx, userSession, id, subflow, userCode)
	default:
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.GetLogger().
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

		ctx.GetLogger().
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

		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: id, logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Error occurred parsing the consent session flow id")

		return
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, flowID); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Error occurred loading the consent session")

		return
	}

	if consent.Responded() {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Failed to process consent session as it has already been responded to")

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: consent.ClientID}).
			Error("Error occurred loading the client for the consent session")

		return
	}

	if userSession.IsAnonymous() {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID()}).
			Error("Failed to redirect for consent as the user is anonymous")

		return
	}

	var (
		issuer *url.URL
		form   url.Values
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
			Error("Error occurred determining the issuer")

		return
	}

	if form, err = consent.GetForm(); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.GetLogger().
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
			ctx.GetLogger().
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
			ctx.GetLogger().
				WithError(err).
				WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
				Error("Error occurred marshaling JSON response body for consent redirection")
		}
	default:
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
			Info("OpenID Connect 1.0 client requires 2FA")

		ctx.ReplyOK()

		return
	}
}

func handleFlowResponseOpenIDConnectDeviceAuthSubflow(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, id, subflow, userCode string) {
	var (
		issuer    *url.URL
		signature string
		device    *model.OAuth2DeviceCodeSession
		client    oidc.Client
		err       error
	)

	if userSession.IsAnonymous() {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Failed to handle flow response as the user is anonymous")

		return
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow}).
			Error("Error occurred determining the issuer preventing a successful flow response")

		return
	}

	if n := len(userCode); n == 0 {
		handleFlowResponseOpenIDConnectDeviceAuthSubflowResponseNoUserCode(ctx, userSession, id, subflow, level, issuer)

		return
	} else if n > 32 {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username}).
			Error("Failed to handle flow response as the user code is too long")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if signature, err = ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(ctx, userCode); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username}).
			Error("Error occurred determining the signature of the user code session preventing a successful flow response")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if device, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, signature); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username, logging.FieldSignature: signature}).
			Error("Error occurred using the signature of the user code session to retrieve the device code session preventing a successful flow response")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if device.Subject.Valid || device.ChallengeID.Valid || device.Status != int(oauthelia2.DeviceAuthorizeStatusNew) {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldUsername: userSession.Username, logging.FieldSignature: signature, logging.FieldClientID: device.ClientID, logging.FieldSessionID: device.ID, logging.FieldSubject: device.Subject, logging.FieldFlowID: device.ChallengeID, logging.FieldStatus: device.Status}).
			Error("Failed to handle flow response as the device code session is in an invalid state")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, device.ClientID); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.GetLogger().
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
			ctx.GetLogger().
				WithError(err).
				WithFields(map[string]any{logging.FieldFlow: flowNameOpenIDConnect, logging.FieldSubflow: subflow, logging.FieldClientID: client.GetID(), logging.FieldUsername: userSession.Username}).
				Error("Failed to marshal JSON response body for authorization redirection")
		}
	default:
		ctx.GetLogger().
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
			ctx.GetLogger().
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

	doMarkAuthenticationAttemptWithRequest(ctx, successful, ban, ban.Value(), authType, requestURI, requestMethod, errAuth)
}

func doMarkAuthenticationAttemptWithRequest(ctx markContext, successful bool, ban *regulation.Ban, username, authType, requestURI, requestMethod string, errAuth error) {
	ctx.GetLogger().Debugf("Mark %s authentication attempt made by user '%s'", authType, ban.Value())

	ctx.GetProviders().Regulator.HandleAttempt(ctx, successful, ban, requestURI, requestMethod, authType)

	stage, method := doGetAuthenticationEventStageAndMethod(authType)

	ctx.GetProviders().Events.Emit(ctx, events.NewEvent(&events.DataAuthentication{
		Type:     doGetAuthenticationEventType(successful),
		Username: username,
		RemoteIP: ctx.RemoteIP().String(),
		Stage:    stage,
		Method:   method,
		Reason:   doGetAuthenticationEventReason(successful, ban, errAuth),
	}))

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

// doGetAuthenticationEventType returns the event type name for an authentication outcome.
func doGetAuthenticationEventType(successful bool) string {
	if successful {
		return events.TypeSecurityAuthenticationSucceeded
	}

	return events.TypeSecurityAuthenticationFailed
}

// doGetAuthenticationEventStageAndMethod maps a regulation.AuthType to the event stage and method it represents. The
// passkey type is a first factor stage because it replaces the password entirely, and a WebAuthn method because the
// credential presented is a WebAuthn credential. Every declared regulation.AuthType is covered, which
// TestAuthenticationEventStageAndMethodCoversEveryAuthType enforces.
func doGetAuthenticationEventStageAndMethod(authType string) (stage, method string) {
	switch authType {
	case regulation.AuthType1FA:
		return events.StageFirstFactor, events.MethodPassword
	case regulation.AuthTypePasskey:
		return events.StageFirstFactor, events.MethodWebAuthn
	case regulation.AuthTypePassword:
		return events.StageSecondFactor, events.MethodPassword
	case regulation.AuthTypeTOTP:
		return events.StageSecondFactor, events.MethodTOTP
	case regulation.AuthTypeWebAuthn:
		return events.StageSecondFactor, events.MethodWebAuthn
	case regulation.AuthTypeDuo:
		return events.StageSecondFactor, events.MethodDuo
	default:
		return events.StageSecondFactor, strings.ToLower(authType)
	}
}

// doGetAuthenticationEventReason classifies an authentication failure. The ban takes precedence so that the reason
// always agrees with the banned dimension recorded by the metrics. An error which a site has marked as a rejection is
// the user or their authenticator refusing the attempt, and is classified ahead of the backend sentinel because only
// the raising site knows that distinction and it states it deliberately, whereas a wrapped sentinel could be an
// accident of wrapping. The authentication backend's not found sentinel is then carved out of the remaining errors so
// that a submitted username which does not exist is not reported as a fault, leaving internal for the errors which
// genuinely are one, such as an unreachable authentication backend. An unreachable backend does not match the sentinel
// and so is still classified as internal.
func doGetAuthenticationEventReason(successful bool, ban *regulation.Ban, errAuth error) string {
	var rejected *errAuthenticationRejected

	switch {
	case successful:
		return ""
	case ban.IsBanned():
		return events.ReasonBanned
	case errors.As(errAuth, &rejected):
		return events.ReasonInvalidCredentials
	case errors.Is(errAuth, authentication.ErrUserNotFound):
		return events.ReasonUserNotFound
	case errAuth != nil:
		return events.ReasonInternalError
	default:
		return events.ReasonInvalidCredentials
	}
}

// newErrAuthenticationRejected marks an authentication error as caused by the user or their authenticator rather than
// by a fault, i.e. a declined push, a failed assertion, or a credential which is not theirs. Only the site raising the
// error knows which of the two it is, and every site hands the choke point the same non-nil error, so the distinction
// has to be carried rather than inferred there.
func newErrAuthenticationRejected(err error) error {
	return &errAuthenticationRejected{err: err}
}

// errAuthenticationRejected is an authentication error caused by the user or their authenticator. It renders exactly
// as the error it wraps so that the log lines which already display these errors are unchanged.
type errAuthenticationRejected struct {
	err error
}

// Error returns the message of the wrapped error.
func (e *errAuthenticationRejected) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error.
func (e *errAuthenticationRejected) Unwrap() error {
	return e.err
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
	EmitEvent(event *events.Event)
	RemoteIP() (ip net.IP)
}
