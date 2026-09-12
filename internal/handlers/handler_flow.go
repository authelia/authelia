// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

// FlowContinuePOST handles requests to continue a flow for a user who is already authenticated.
func FlowContinuePOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	bodyJSON := bodyFlowContinueRequest{}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.WithError(err).Error("Failed to continue the flow as the body could not be parsed")

		return
	}

	if len(bodyJSON.Flow) == 0 {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: bodyJSON.FlowID, logging.FieldSubflow: bodyJSON.SubFlow}).
			Error("Failed to continue the flow as no flow was provided")

		return
	}

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.WithError(err).Error("Failed to continue the flow as the user session could not be obtained")

		return
	}

	if userSession.IsAnonymous() {
		ctx.SetJSONError(messageAuthenticationFailed)

		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: bodyJSON.FlowID, logging.FieldFlow: bodyJSON.Flow, logging.FieldSubflow: bodyJSON.SubFlow}).
			Error("Failed to continue the flow as the user is anonymous")

		return
	}

	handleFlowResponse(ctx, &userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
}
