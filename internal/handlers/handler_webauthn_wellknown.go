// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

// WebAuthnWellKnownGET handles the Related Origin Requests well-known document which clients fetch from the relying
// party id host to discover the other origins permitted to perform ceremonies against that relying party.
func WebAuthnWellKnownGET(ctx *middlewares.AutheliaCtx) {
	origin, err := ctx.GetOrigin()
	if err != nil || origin == nil {
		ctx.GetLogger().WithError(err).Error("Error occurred retrieving the origin for the request")

		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	// The document is only served by the origins of a configured relying party. Any other origin has no related
	// origins to declare, so it doesn't serve the resource at all.
	if _, relyingParty := webauthn.GetRelatedOriginConfigByOrigin(ctx.GetConfiguration().WebAuthn, origin); relyingParty == nil {
		ctx.GetLogger().Debugf("Origin '%s' does not match any configured WebAuthn relying party", origin.String())

		ctx.SetStatusCode(fasthttp.StatusNotFound)

		return
	}

	provider, err := ctx.GetWebAuthnProvider()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred retrieving the webauthn provider for the request")

		ctx.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	related, err := provider.RelatedOrigins()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred retrieving the related origins for the request")

		ctx.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	middlewares.SetContentTypeApplicationJSON(ctx.RequestCtx)

	if err = json.NewEncoder(ctx.RequestCtx).Encode(related); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred encoding the response")
	}
}
