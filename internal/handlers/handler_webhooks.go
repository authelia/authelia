// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/webhooks"
)

// WebhookConfirm handles the callback a webhook destination requests to confirm it accepts deliveries, which is the
// asynchronous form of the CloudEvents abuse protection handshake. It answers both GET and POST.
func WebhookConfirm(ctx *middlewares.AutheliaCtx) {
	confirmer, ok := ctx.Providers.Events.(events.Confirmer)
	if !ok {
		ctx.ReplyStatusCode(fasthttp.StatusNotFound)

		return
	}

	name := string(ctx.QueryArgs().Peek(webhooks.QueryConfirmID))

	if !confirmer.Confirm(name, string(ctx.QueryArgs().Peek(webhooks.QueryConfirmKey)), string(ctx.Request.Header.Peek(events.HeaderWebhookAllowedRate))) {
		ctx.Logger.WithField("destination", name).Error("Webhook destination confirmation was rejected because the destination is unknown, was refused, or the key is incorrect")

		ctx.ReplyStatusCode(fasthttp.StatusNotFound)

		return
	}

	ctx.ReplyOK()
}
