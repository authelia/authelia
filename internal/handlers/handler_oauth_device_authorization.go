package handlers

import (
	"errors"
	"net/http"
	"net/url"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/x/errorsx"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func OAuthDeviceAuthorizationPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		request  oauthelia2.DeviceAuthorizeRequester
		response oauthelia2.DeviceAuthorizeResponder

		err error
	)

	if request, err = ctx.Providers.OpenIDConnect.NewRFC862DeviceAuthorizeRequest(ctx, req); err != nil {
		ctx.Logger.Errorf("Device Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		errorsx.WriteJSONError(rw, req, err)

		return
	}

	if response, err = ctx.Providers.OpenIDConnect.NewRFC862DeviceAuthorizeResponse(ctx, request, oidc.NewSession()); err != nil {
		ctx.Logger.Errorf("Device Authorization Request with id '%s' on client with id '%s'  failed with error: %s", request.GetID(), request.GetClient().GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		errorsx.WriteJSONError(rw, req, err)

		return
	}

	ctx.Providers.OpenIDConnect.WriteRFC862DeviceAuthorizeResponse(ctx, rw, request, response)
}

func OAuthDeviceAuthorizationPUT(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester oauthelia2.DeviceAuthorizeRequester
		responder oauthelia2.DeviceUserAuthorizeResponder
		client    oidc.Client

		err error
	)

	if requester, err = ctx.Providers.OpenIDConnect.NewRFC8628UserAuthorizeRequest(ctx, r); err != nil {
		ctx.Logger.Errorf("Device Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, err)

		return
	}

	clientID := requester.GetClient().GetID()

	ctx.Logger.Debugf("Device Authorization Request with id '%s' on client with id '%s' is being processed", requester.GetID(), clientID)

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID); err != nil {
		if errors.Is(err, oauthelia2.ErrNotFound) {
			ctx.Logger.Errorf("Device Authorization Request with id '%s' on client with id '%s' could not be processed: client was not found", requester.GetID(), clientID)
		} else {
			ctx.Logger.Errorf("Device Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %s", requester.GetID(), clientID, oauthelia2.ErrorToDebugRFC6749Error(err))
		}

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, err)

		return
	}

	var (
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		issuer      *url.URL
		handled     bool
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.Errorf("Device Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred obtaining session information: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the user session."))

		return
	}

	issuer = ctx.RootURL()

	if consent, handled = handleOIDCAuthorizationConsent(ctx, issuer, client, userSession, rw, r, requester); handled {
		return
	}

	var details *authentication.UserDetailsExtended

	if details, err = ctx.Providers.UserProvider.GetDetailsExtended(userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Device Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred retrieving user details for '%s' from the backend", requester.GetID(), client.GetID(), userSession.Username)

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the users details."))

		return
	}

	oidc.GrantScopeAudienceConsent(requester, consent)

	extra := map[string]any{}

	ctx.Logger.Debugf("Authorization Request with id '%s' on client with id '%s' was successfully processed, proceeding to build Authorization Response", requester.GetID(), clientID)

	session := oidc.NewSessionWithRequester(ctx, issuer, ctx.Providers.OpenIDConnect.Issuer.GetKeyID(ctx, client.GetIDTokenSignedResponseKeyID(), client.GetIDTokenSignedResponseAlg()), details.Username, userSession.AuthenticationMethodRefs.MarshalRFC8176(), extra, userSession.LastAuthenticatedTime(), consent, requester, nil)

	if responder, err = ctx.Providers.OpenIDConnect.NewRFC8628UserAuthorizeResponse(ctx, requester, session); err != nil {
		ctx.Logger.Errorf("Device Authorization Request with id '%s' failed with error: %s", requester.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionGranted(ctx, consent.ID); err != nil {
		ctx.Logger.Errorf("Device Authorization Request with id '%s' on client with id '%s' could not be processed: error occurred saving consent session: %+v", requester.GetID(), client.GetID(), err)

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return
	}

	ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeResponse(ctx, rw, requester, responder)
}
