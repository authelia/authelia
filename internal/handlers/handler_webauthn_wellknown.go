// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/json"

	"github.com/go-webauthn/webauthn/protocol"
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

	config := ctx.GetConfiguration().WebAuthn

	// Clients fetch the document from the relying party id host, so the document is only served when the hostname is
	// the id of a configured relying party. Any other host has no related origins to declare.
	relyingParty := webauthn.GetRelatedOriginConfigByRPID(config, origin.Hostname())
	if config.Disable || relyingParty == nil {
		ctx.GetLogger().Debugf("Hostname '%s' does not match any configured WebAuthn relying party id", origin.Hostname())

		ctx.SetStatusCode(fasthttp.StatusNotFound)

		return
	}

	related, err := protocol.NewRelatedOrigins(relyingParty.StringOrigins()...)
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
