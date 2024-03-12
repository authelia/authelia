package handlers

import (
	"net/http"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// OAuthRevocationPOST handles POST requests to the OAuth 2.0 Revocation endpoint.
//
// https://datatracker.ietf.org/doc/html/rfc7009
func OAuthRevocationPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		requestID uuid.UUID
		err       error
	)

	if requestID, err = uuid.NewRandom(); err != nil {
		ctx.Providers.OpenIDConnect.WriteRevocationResponse(ctx, rw, oauthelia2.ErrServerError)

		return
	}

	ctx.Logger.Debugf("Revocation Request with id '%s' is being processed", requestID)

	if err = ctx.Providers.OpenIDConnect.NewRevocationRequest(ctx, req); err != nil {
		ctx.Logger.Errorf("Revocation Request with id '%s' failed with error: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(err))
	}

	ctx.Providers.OpenIDConnect.WriteRevocationResponse(ctx, rw, err)

	ctx.Logger.Debugf("Revocation Request with id '%s' was successfully processed", requestID)
}
