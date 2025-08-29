package handlers

import (
	"net/http"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OAuth2IntrospectionPOST handles POST requests to the OAuth 2.0 Introspection endpoint.
//
// https://datatracker.ietf.org/doc/html/rfc7662
func OAuth2IntrospectionPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		requestID uuid.UUID
		responder oauthelia2.IntrospectionResponder
		err       error
	)
	if requestID, err = uuid.NewRandom(); err != nil {
		ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, oauthelia2.ErrServerError)

		return
	}

	ctx.Logger.Debugf("Introspection Request with id '%s' is being processed", requestID)

	if responder, err = ctx.Providers.OpenIDConnect.NewIntrospectionRequest(ctx, req, oidc.NewSessionWithRequestedAt(ctx.GetClock().Now())); err != nil {
		ctx.Logger.Errorf("Introspection Request with id '%s' failed with error: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, err)

		return
	}

	ctx.Logger.Tracef("Introspection Request with id '%s' yielded a %s (active: %t) requested at %s created with request id '%s' on client with id '%s'", requestID, responder.GetTokenUse(), responder.IsActive(), responder.GetAccessRequester().GetRequestedAt().String(), responder.GetAccessRequester().GetID(), responder.GetAccessRequester().GetClient().GetID())

	ctx.Providers.OpenIDConnect.WriteIntrospectionResponse(ctx, rw, responder)
}
