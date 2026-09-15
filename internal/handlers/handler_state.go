// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

// StateGET is the handler serving the user state.
func StateGET(ctx *middlewares.AutheliaCtx) {
	var (
		provider    session.Strategy
		userSession session.UserSession
		err         error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		ctx.ReplyForbidden()

		return
	}

	// The portal requests the state before any other request, so the CSRF token is delivered here for a session which
	// was established before the token cookie existed or which outlived it.
	if provider, err = ctx.GetSessionProvider(); err == nil {
		provider.SetCSRFCookie(ctx)
	}

	stateResponse := StateResponse{
		Username:            userSession.Username,
		AuthenticationLevel: userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA),
		FactorKnowledge:     userSession.AuthenticationMethodRefs.FactorKnowledge(),
	}

	if uri := ctx.GetDefaultRedirectionURL(); uri != nil {
		stateResponse.DefaultRedirectionURL = uri.String()
	}

	if err = ctx.SetJSONBody(stateResponse); err != nil {
		ctx.Logger.Errorf("Unable to set state response in body: %s", err)
	}
}
