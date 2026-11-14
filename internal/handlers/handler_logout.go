// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

type logoutBody struct {
	TargetURL string `json:"targetURL"`
}

type logoutResponseBody struct {
	SafeTargetURL bool `json:"safeTargetURL"`
}

// LogoutPOST is the handler logging out the user attached to the given cookie.
func LogoutPOST(ctx *middlewares.AutheliaCtx) {
	body := logoutBody{}
	responseBody := logoutResponseBody{SafeTargetURL: false}

	// The session cookie is never sent on a cross-site request, but the authz cookie may be depending on its configured
	// SameSite value, and it alone is enough to end the session, so a cross-site logout is refused to prevent CSRF.
	if strings.EqualFold(string(ctx.Request.Header.Peek(headerSecFetchSite)), valueSecFetchSiteCrossSite) {
		ctx.GetLogger().Error("Error occurred during logout: the request was cross-site")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	err := ctx.ParseBody(&body)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred parsing the logout request body")
		ctx.SetJSONError(messageOperationFailed)
	}

	err = ctx.DestroySession()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred destroying the user session during logout")
		ctx.SetJSONError(messageOperationFailed)
	}

	redirectionURL, err := url.ParseRequestURI(body.TargetURL)
	if err == nil {
		responseBody.SafeTargetURL = ctx.IsSafeRedirectionTargetURI(redirectionURL)
	}

	if body.TargetURL != "" {
		ctx.Logger.Debugf("Logout target url is %s, safe %t", body.TargetURL, responseBody.SafeTargetURL)
	}

	err = ctx.SetJSONBody(responseBody)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred setting the logout response body")
		ctx.SetJSONError(messageOperationFailed)
	}
}
