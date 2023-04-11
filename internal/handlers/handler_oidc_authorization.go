package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// OpenIDConnectAuthorization handles GET/POST requests to the OpenID Connect 1.0 Authorization endpoint.
//
// https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint
func OpenIDConnectAuthorization(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester fosite.AuthorizeRequester
		responder fosite.AuthorizeResponder
		client    *oidc.Client
		authTime  time.Time
		issuer    *url.URL
		err       error
	)

	if requester, err = ctx.Providers.OpenIDConnect.NewAuthorizeRequest(ctx, r); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Authorization Request failed with error: %s", rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	clientID := requester.GetClient().GetID()

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' is being processed", requester.GetID(), clientID)

	if client, err = ctx.Providers.OpenIDConnect.GetFullClient(clientID); err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: client was not found", requester.GetID(), clientID)
		} else {
			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %+v", requester.GetID(), clientID, err)
		}

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = client.ValidatePARPolicy(requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' failed to validate the PAR policy: %s", requester.GetID(), clientID, rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if !oidc.IsPushedAuthorizedRequest(requester, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)) {
		if err = client.ValidatePKCEPolicy(requester); err != nil {
			rfc := fosite.ErrorToRFC6749Error(err)

			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' failed to validate the PKCE policy: %s", requester.GetID(), client.GetID(), rfc.WithExposeDebug(true).GetDescription())

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

			return
		}

		if err = client.ValidateResponseModePolicy(requester); err != nil {
			rfc := fosite.ErrorToRFC6749Error(err)

			ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' failed to validate the Response Mode: %s", requester.GetID(), client.GetID(), rfc.WithExposeDebug(true).GetDescription())

			ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

			return
		}
	}

	if err = client.ValidatePKCEPolicy(requester); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' failed to validate the PKCE policy: %s", requester.GetID(), clientID, rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	issuer = ctx.RootURL()

	var (
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		handled     bool
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred obtaining session information: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, fosite.ErrServerError.WithHint("Could not obtain the user session."))

		return
	}

	if consent, handled = handleOIDCAuthorizationConsent(ctx, issuer, client, userSession, rw, r, requester); handled {
		return
	}

	extraClaims := oidcGrantRequests(requester, consent, &userSession)

	if authTime, err = userSession.AuthenticatedTime(client.Policy); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred checking authentication time: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, fosite.ErrServerError.WithHint("Could not obtain the authentication time."))

		return
	}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' was successfully processed, proceeding to build Authorization Response", requester.GetID(), clientID)

	session := oidc.NewSessionWithAuthorizeRequest(issuer, ctx.Providers.OpenIDConnect.KeyManager.GetActiveKeyID(),
		userSession.Username, userSession.AuthenticationMethodRefs.MarshalRFC8176(), extraClaims, authTime, consent, requester)

	ctx.Logger.Tracef("Authorization Request with id '%s' on client with id '%s' creating session for Authorization Response for subject '%s' with username '%s' with claims: %+v",
		requester.GetID(), session.ClientID, session.Subject, session.Username, session.Claims)

	if responder, err = ctx.Providers.OpenIDConnect.NewAuthorizeResponse(ctx, requester, session); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Authorization Response for Request with id '%s' on client with id '%s' could not be created: %s", requester.GetID(), clientID, rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionGranted(ctx, consent.ID); err != nil {
		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return
	}

	if requester.GetResponseMode() == oidc.ResponseModeFormPost {
		ctx.SetUserValueBytes(middlewares.UserValueKeyFormPost, true)
	}

	ctx.Providers.OpenIDConnect.WriteAuthorizeResponse(ctx, rw, requester, responder)
}

// OpenIDConnectPushedAuthorizationRequest handles POST requests to the OAuth 2.0 Pushed Authorization Requests endpoint.
//
// RFC9126 https://www.rfc-editor.org/rfc/rfc9126.html
func OpenIDConnectPushedAuthorizationRequest(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester fosite.AuthorizeRequester
		responder fosite.PushedAuthorizeResponder
		err       error
	)

	if requester, err = ctx.Providers.OpenIDConnect.NewPushedAuthorizeRequest(ctx, r); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Pushed Authorization Request failed with error: %s", rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	var client *oidc.Client

	clientID := requester.GetClient().GetID()

	if client, err = ctx.Providers.OpenIDConnect.GetFullClient(clientID); err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			ctx.Logger.Errorf("Pushed Authorization Request with id '%s' on client with id '%s' could not be processed: client was not found", requester.GetID(), clientID)
		} else {
			ctx.Logger.Errorf("Pushed Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %+v", requester.GetID(), clientID, err)
		}

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = client.ValidatePKCEPolicy(requester); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Pushed Authorization Request with id '%s' on client with id '%s' failed to validate the PKCE policy: %s", requester.GetID(), client.GetID(), rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = client.ValidateResponseModePolicy(requester); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Pushed Authorization Request with id '%s' on client with id '%s' failed to validate the Response Mode: %s", requester.GetID(), client.GetID(), rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	if responder, err = ctx.Providers.OpenIDConnect.NewPushedAuthorizeResponse(ctx, requester, oidc.NewSession()); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Pushed Authorization Request failed with error: %s", rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	ctx.Providers.OpenIDConnect.WritePushedAuthorizeResponse(ctx, rw, requester, responder)
}
