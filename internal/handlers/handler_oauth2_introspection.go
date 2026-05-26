package handlers

import (
	"net/http"
	"net/url"

	"github.com/google/uuid"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OAuth2IntrospectionPOST handles POST requests to the OAuth 2.0 Introspection endpoint.
//
// https://datatracker.ietf.org/doc/html/rfc7662
func OAuth2IntrospectionPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		issuer    *url.URL
		requestID uuid.UUID
		responder oauthelia2.IntrospectionResponder
		err       error
	)

	if requestID, err = uuid.NewRandom(); err != nil {
		ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, oauthelia2.ErrServerError)

		return
	}

	ctx.GetLogger().Debugf("Introspection Request with id '%s' is being processed", requestID)

	if issuer, err = ctx.IssuerURL(); err != nil {
		rfc := oidc.ErrEffectiveIssuer.WithWrap(err)

		ctx.GetLogger().WithError(err).Errorf("Introspection Request with id '%s' could not be processed: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(rfc))

		ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, rfc)

		return
	}

	if responder, err = ctx.Providers.OpenIDConnect.NewIntrospectionRequest(ctx, req, oidc.NewSessionWithRequestedAt(ctx.GetClock().Now())); err != nil {
		ctx.GetLogger().WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).Errorf("Introspection Request with id '%s' failed with error", requestID)

		ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, err)

		return
	}

	requester := responder.GetAccessRequester()

	if requester != nil {
		if s := requester.GetSession(); s != nil {
			if session, ok := s.(*oidc.Session); ok {
				if !session.ValidIssuer(issuer) {
					err = oauthelia2.ErrInvalidRequest.WithDebug("The original request and the introspection request occurred at endpoints where the origin or effective issuer did not match.")

					ctx.GetLogger().WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).Errorf("Introspection Request with id '%s' failed with error", requestID)

					ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, err)

					return
				}
			}
		}

		ctx.GetLogger().Tracef("Introspection Request with id '%s' yielded a %s (active: %t) requested at %s created with request id '%s' on client with id '%s'", requestID, responder.GetTokenUse(), responder.IsActive(), requester.GetRequestedAt().String(), requester.GetID(), requester.GetClient().GetID())
	} else {
		ctx.GetLogger().Tracef("Introspection Request with id '%s' yielded a %s (active: %t)", requestID, responder.GetTokenUse(), responder.IsActive())
	}

	ctx.Providers.OpenIDConnect.WriteIntrospectionResponse(ctx, rw, responder)
}
