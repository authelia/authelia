package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

func WebAuthnWellKnownGET(ctx *middlewares.AutheliaCtx) {
	origin, err := ctx.GetOrigin()
	if err != nil || origin == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		ctx.GetLogger().WithError(err).Error("Error occurred retrieving the origin for the request")

		return
	}

	_, relatedOrigin := webauthn.GetRelatedOriginConfigByOrigin(ctx.GetConfiguration().WebAuthn, origin)

	if relatedOrigin == nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)

		ctx.GetLogger().WithFields(map[string]any{"origin": origin.String()}).Error("WebAuthn Origin not found in related origins configuration")

		return
	}

	response := responseWebAuthnWellKnown{
		Origins: make([]string, len(relatedOrigin.Origins)),
	}

	for i, o := range relatedOrigin.Origins {
		response.Origins[i] = o.String()
	}

	middlewares.SetContentTypeApplicationJSON(ctx.RequestCtx)

	if err = json.NewEncoder(ctx.RequestCtx).Encode(response); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred encoding the response")
	}
}

type responseWebAuthnWellKnown struct {
	Origins []string `json:"origins"`
}
