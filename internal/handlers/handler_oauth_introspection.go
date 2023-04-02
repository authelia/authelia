// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OAuthIntrospectionPOST handles POST requests to the OAuth 2.0 Introspection endpoint.
//
// https://datatracker.ietf.org/doc/html/rfc7662
func OAuthIntrospectionPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		responder fosite.IntrospectionResponder
		err       error
	)

	oidcSession := oidc.NewSession()

	if responder, err = ctx.Providers.OpenIDConnect.NewIntrospectionRequest(ctx, req, oidcSession); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Introspection Request failed with error: %s", rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WriteIntrospectionError(ctx, rw, err)

		return
	}

	requester := responder.GetAccessRequester()

	ctx.Logger.Tracef("Introspection Request yeilded a %s (active: %t) requested at %s created with request id '%s' on client with id '%s'", responder.GetTokenUse(), responder.IsActive(), requester.GetRequestedAt().String(), requester.GetID(), requester.GetClient().GetID())

	ctx.Providers.OpenIDConnect.WriteIntrospectionResponse(ctx, rw, responder)
}
