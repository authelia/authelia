package handlers

import (
	"net/http"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OpenIDConnectTokenPOST handles POST requests to the OpenID Connect 1.0 Token endpoint.
//
// https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint
func OpenIDConnectTokenPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		requester fosite.AccessRequester
		responder fosite.AccessResponder
		err       error
	)

	session := oidc.NewSession()

	if requester, err = ctx.Providers.OpenIDConnect.NewAccessRequest(ctx, req, session); err != nil {
		ctx.Logger.Errorf("Access Request failed with error: %s", oidc.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

		return
	}

	client := requester.GetClient()

	ctx.Logger.Debugf("Access Request with id '%s' on client with id '%s' is being processed", requester.GetID(), client.GetID())

	if requester.GetGrantTypes().ExactOne(oidc.GrantTypeClientCredentials) {
		if err = oidc.PopulateClientCredentialsFlowSessionWithAccessRequest(ctx, client, session); err != nil {
			ctx.Logger.Errorf("Access Response for Request with id '%s' failed to be created with error: %s", requester.GetID(), oidc.ErrorToDebugRFC6749Error(err))

			ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

			return
		}

		if err = oidc.PopulateClientCredentialsFlowRequester(ctx, ctx.Providers.OpenIDConnect, client, requester); err != nil {
			ctx.Logger.Errorf("Access Response for Request with id '%s' failed to be created with error: %s", requester.GetID(), oidc.ErrorToDebugRFC6749Error(err))

			ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

			return
		}
	}

	ctx.Logger.Tracef("Access Request with id '%s' on client with id '%s' response is being generated for session with type '%T'", requester.GetID(), client.GetID(), requester.GetSession())

	if responder, err = ctx.Providers.OpenIDConnect.NewAccessResponse(ctx, requester); err != nil {
		ctx.Logger.Errorf("Access Response for Request with id '%s' failed to be created with error: %s", requester.GetID(), oidc.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteAccessError(ctx, rw, requester, err)

		return
	}

	ctx.Logger.Debugf("Access Request with id '%s' on client with id '%s' has successfully been processed", requester.GetID(), client.GetID())

	ctx.Logger.Tracef("Access Request with id '%s' on client with id '%s' produced the following claims: %+v", requester.GetID(), client.GetID(), oidc.AccessResponderToClearMap(responder))

	ctx.Providers.OpenIDConnect.WriteAccessResponse(ctx, rw, requester, responder)
}
