package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// WebAuthnWellKnownGET handles the WebAuthn well-known document.
func WebAuthnWellKnownGET(ctx *middlewares.AutheliaCtx) {
	origin, err := ctx.GetOrigin()
	if err != nil || origin == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		ctx.GetLogger().WithError(err).Error("Error occurred retrieving the origin for the request")

		return
	}

	provider, err := ctx.GetWebAuthnProvider()
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		ctx.GetLogger().WithError(err).Error("Error occurred retrieving the webauthn provider for the request")

		return
	}

	related, err := provider.RelatedOrigins()
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		ctx.GetLogger().WithError(err).Error("Error occurred retrieving the related origin for the request")

		return
	}

	middlewares.SetContentTypeApplicationJSON(ctx.RequestCtx)

	if err = json.NewEncoder(ctx.RequestCtx).Encode(related); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred encoding the response")
	}
}
