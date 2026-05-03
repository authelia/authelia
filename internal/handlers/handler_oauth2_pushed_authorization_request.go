package handlers

import (
	"errors"
	"net/http"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OAuth2PushedAuthorizationRequest handles POST requests to the OAuth 2.0 Pushed Authorization Requests endpoint.
//
// RFC9126 https://www.rfc-editor.org/rfc/rfc9126.html
func OAuth2PushedAuthorizationRequest(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester oauthelia2.AuthorizeRequester
		responder oauthelia2.PushedAuthorizeResponder
		err       error
	)

	if requester, err = ctx.Providers.OpenIDConnect.NewPushedAuthorizeRequest(ctx, r); err != nil {
		ctx.GetLogger().Errorf("Pushed Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	ctx.GetLogger().Debugf("Pushed Authorization Request with id '%s' is being processed", requester.GetID())

	if _, err = ctx.IssuerURL(); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Pushed Authorization Request with id '%s' could not be processed: %s", requester.GetID(), oidc.ErrTextEffectiveIssuer)

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, oidc.ErrEffectiveIssuer)

		return
	}

	var client oidc.Client

	clientID := requester.GetClient().GetID()

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID); err != nil {
		if errors.Is(err, oauthelia2.ErrNotFound) {
			ctx.GetLogger().Errorf("Pushed Authorization Request with id '%s' on client with id '%s' could not be processed: client was not found", requester.GetID(), clientID)
		} else {
			ctx.GetLogger().Errorf("Pushed Authorization Request with id '%s' on client with id '%s' could not be processed: failed to find client: %+v", requester.GetID(), clientID, err)
		}

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = client.ValidateResponseModePolicy(requester); err != nil {
		ctx.GetLogger().Errorf("Pushed Authorization Request with id '%s' on client with id '%s' failed to validate the Response Modes: %s", requester.GetID(), client.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	if responder, err = ctx.Providers.OpenIDConnect.NewPushedAuthorizeResponse(ctx, requester, oidc.NewSessionWithRequestedAt(ctx.GetClock().Now())); err != nil {
		ctx.GetLogger().Errorf("Pushed Authorization Request failed with error: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WritePushedAuthorizeError(ctx, rw, requester, err)

		return
	}

	ctx.Providers.OpenIDConnect.WritePushedAuthorizeResponse(ctx, rw, requester, responder)
}
