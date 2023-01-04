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

	if err = client.ValidateAuthorizationPolicy(requester); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Authorization Request with id '%s' on client with id '%s' failed to validate the authorization policy: %s", requester.GetID(), clientID, rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, err)

		return
	}

	issuer = ctx.RootURL()

	userSession := ctx.GetSession()

	var (
		consent *model.OAuth2ConsentSession
		handled bool
	)

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

	oidcSession := oidc.NewSessionWithAuthorizeRequest(issuer, ctx.Providers.OpenIDConnect.KeyManager.GetActiveKeyID(),
		userSession.Username, userSession.AuthenticationMethodRefs.MarshalRFC8176(), extraClaims, authTime, consent, requester)

	ctx.Logger.Tracef("Authorization Request with id '%s' on client with id '%s' creating session for Authorization Response for subject '%s' with username '%s' with claims: %+v",
		requester.GetID(), oidcSession.ClientID, oidcSession.Subject, oidcSession.Username, oidcSession.Claims)

	if responder, err = ctx.Providers.OpenIDConnect.NewAuthorizeResponse(ctx, requester, oidcSession); err != nil {
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

	ctx.Providers.OpenIDConnect.WriteAuthorizeResponse(ctx, rw, requester, responder)
}
