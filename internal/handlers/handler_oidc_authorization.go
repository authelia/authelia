package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// OpenIDConnectAuthorization handles GET/POST requests to the OpenID Connect 1.0 Authorization endpoint.
//
// https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint
//
//nolint:gocyclo
func OpenIDConnectAuthorization(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester oauthelia2.AuthorizeRequester
		responder oauthelia2.AuthorizeResponder
		client    oidc.Client
		authTime  time.Time
		issuer    *url.URL
		err       error
	)

	requester, err = ctx.Providers.OpenIDConnect.NewAuthorizeRequest(ctx, r)

	if requester != nil && requester.GetResponseMode() == oidc.ResponseModeFormPost {
		ctx.SetUserValue(middlewares.UserValueKeyOpenIDConnectResponseModeFormPost, true)
	}

	if err != nil {
		ctx.Logger.Errorf("Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if requester.GetResponseMode() == oidc.ResponseModeFormPost {
		ctx.SetUserValue(middlewares.UserValueKeyOpenIDConnectResponseModeFormPost, true)
	}

	clientID := requester.GetClient().GetID()

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' is being processed", requester.GetID(), clientID)

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID); err != nil {
		if errors.Is(err, oauthelia2.ErrNotFound) {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: client was not found", requester.GetID(), clientID)
		} else {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %s", requester.GetID(), clientID, oauthelia2.ErrorToDebugRFC6749Error(err))
		}

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if !oidc.IsPushedAuthorizedRequest(requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)) {
		if err = client.ValidateResponseModePolicy(requester); err != nil {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' failed to validate the Response Mode: %s", requester.GetID(), client.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

			return
		}
	}

	var (
		details     *authentication.UserDetails
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		handled     bool
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred obtaining session information: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the user session."))

		return
	}

	if requester.GetRequestForm().Get(oidc.FormParameterPrompt) == oidc.PromptNone {
		if userSession.IsAnonymous() {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: the 'prompt' type of 'none' was requested but the user is not logged in", requester.GetID(), client.GetID())

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrLoginRequired)

			return
		}

		if client.GetConsentPolicy().Mode == oidc.ClientConsentModeExplicit {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: the 'prompt' type of 'none' was requested but client is configured to require explicit consent", requester.GetID(), client.GetID())

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrConsentRequired)

			return
		}
	}

	issuer = ctx.RootURL()

	if consent, handled = handleOIDCAuthorizationConsent(ctx, issuer, client, userSession, rw, r, requester); handled {
		return
	}

	if details, err = ctx.Providers.UserProvider.GetDetails(userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred retrieving user details for '%s' from the backend", requester.GetID(), client.GetID(), userSession.Username)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the users details."))

		return
	}

	extraClaims := oidcGrantRequests(requester, consent, details)

	if authTime, err = userSession.AuthenticatedTime(client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: details.Username, Groups: details.Groups, IP: ctx.RemoteIP()})); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred checking authentication time: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the authentication time."))

		return
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' was successfully processed, proceeding to build Authorization Response", requester.GetID(), clientID)

	session := oidc.NewSessionWithAuthorizeRequest(ctx, issuer, ctx.Providers.OpenIDConnect.KeyManager.GetKeyID(ctx, client.GetIDTokenSignedResponseKeyID(), client.GetIDTokenSignedResponseAlg()), details.Username, userSession.AuthenticationMethodRefs.MarshalRFC8176(), extraClaims, authTime, consent, requester)

	ctx.Logger.Tracef("Authorization Request with id '%s' on client with id '%s' creating session for Authorization Response for subject '%s' with username '%s' with claims: %+v",
		requester.GetID(), session.ClientID, session.Subject, session.Username, session.Claims)

	ctx.Logger.WithFields(map[string]any{"id": requester.GetID(), "response_type": requester.GetResponseTypes(), "response_mode": requester.GetResponseMode(), "scope": requester.GetRequestedScopes(), "aud": requester.GetRequestedAudience(), "redirect_uri": requester.GetRedirectURI(), "state": requester.GetState()}).Tracef("Authorization Request is using the following request parameters")

	if responder, err = ctx.Providers.OpenIDConnect.NewAuthorizeResponse(ctx, requester, session); err != nil {
		ctx.Logger.Errorf("Authorization Response for Request with id '%s' on client with id '%s' could not be created: %s", requester.GetID(), clientID, oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionGranted(ctx, consent.ID); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return
	}

	responder.GetParameters().Set(oidc.FormParameterIssuer, issuer.String())

	ctx.Providers.OpenIDConnect.WriteAuthorizeResponse(ctx, rw, requester, responder)
}

// OpenIDConnectPushedAuthorizationRequest handles POST requests to the OAuth 2.0 Pushed Authorization Requests endpoint.
//
// RFC9126 https://www.rfc-editor.org/rfc/rfc9126.html
func OpenIDConnectPushedAuthorizationRequest(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester oauthelia2.AuthorizeRequester
		responder oauthelia2.PushedAuthorizeResponder
		err       error
	)

	if requester, err = ctx.Providers.OpenIDConnect.NewPushedAuthorizeRequest(ctx, r); err != nil {
		ctx.Logger.Errorf("Pushed Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	var client oidc.Client

	clientID := requester.GetClient().GetID()

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID); err != nil {
		if errors.Is(err, oauthelia2.ErrNotFound) {
			ctx.Logger.Errorf("Pushed Authorization Request with id '%s' on client with id '%s' could not be processed: client was not found", requester.GetID(), clientID)
		} else {
			ctx.Logger.Errorf("Pushed Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %+v", requester.GetID(), clientID, err)
		}

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = client.ValidateResponseModePolicy(requester); err != nil {
		ctx.Logger.Errorf("Pushed Authorization Request with id '%s' on client with id '%s' failed to validate the Response Mode: %s", requester.GetID(), client.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	if responder, err = ctx.Providers.OpenIDConnect.NewPushedAuthorizeResponse(ctx, requester, oidc.NewSession()); err != nil {
		ctx.Logger.Errorf("Pushed Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	ctx.Providers.OpenIDConnect.WritePushedAuthorizeResponse(ctx, rw, requester, responder)
}
