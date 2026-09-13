// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/url"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

// CheckSafeRedirectionPOST handler checking whether the redirection to a given URL provided in body is safe.
func CheckSafeRedirectionPOST(ctx *middlewares.AutheliaCtx) {
	checkSafeRedirection(ctx, ctx.IsSafeRedirectionTargetURI)
}

// CheckSafePostLogoutRedirectionPOST handler checking whether the redirection to a given URL provided in body is safe
// as the destination of a logout. It additionally permits the post logout redirect URIs registered by OpenID Connect
// 1.0 clients, which are not safe destinations for any other redirection.
func CheckSafePostLogoutRedirectionPOST(ctx *middlewares.AutheliaCtx) {
	checkSafeRedirection(ctx, ctx.IsSafePostLogoutRedirectionTargetURI)
}

func checkSafeRedirection(ctx *middlewares.AutheliaCtx, safe func(targetURI *url.URL) bool) {
	var (
		s   session.UserSession
		err error
	)
	if s, err = ctx.GetSession(); err != nil {
		ctx.ReplyUnauthorized()
		return
	}

	if s.IsAnonymous() {
		ctx.ReplyUnauthorized()
		return
	}

	var (
		bodyJSON  checkURIWithinDomainRequestBody
		targetURI *url.URL
	)

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred parsing the safe redirection request body")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if targetURI, err = url.ParseRequestURI(bodyJSON.URI); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred determining if the URI '%s' is safe to redirect to as it could not be parsed", bodyJSON.URI)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(checkURIWithinDomainResponseBody{OK: safe(targetURI)}); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred setting the safe redirection response body")
		ctx.SetJSONError(messageOperationFailed)

		return
	}
}
