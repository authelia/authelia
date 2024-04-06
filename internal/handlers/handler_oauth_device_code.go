package handlers

import (
	"net/http"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/x/errorsx"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
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
