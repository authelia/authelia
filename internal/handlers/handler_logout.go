// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"crypto/subtle"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// LogoutGET is the handler reporting whether the session has an RP-Initiated Logout awaiting confirmation, so the
// front-end knows whether to ask the user to confirm the logout rather than perform it straight away.
func LogoutGET(ctx *middlewares.AutheliaCtx) {
	responseBody := bodyLogoutPendingResponse{}

	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred obtaining the user session to determine if a logout is pending")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if logout := logoutPending(ctx, userSession, string(ctx.QueryArgs().Peek(queryArgFlowID))); logout != nil {
		responseBody.Pending = true
		responseBody.ClientID = logout.ClientID

		if logout.ClientID != "" && ctx.Providers.OpenIDConnect != nil {
			if client, errClient := ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, logout.ClientID); errClient == nil {
				responseBody.ClientName = client.GetName()
			}
		}
	}

	if err = ctx.SetJSONBody(responseBody); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred setting the logout response body")
		ctx.SetJSONError(messageOperationFailed)
	}
}

// LogoutDELETE is the handler cancelling an RP-Initiated Logout awaiting confirmation. The user remains logged in.
func LogoutDELETE(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred obtaining the user session to cancel the pending logout")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if logoutFlowMatches(userSession.OpenIDConnectLogout, string(ctx.QueryArgs().Peek(queryArgFlowID))) {
		userSession.OpenIDConnectLogout = nil

		if err = ctx.SaveSession(&userSession); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred saving the user session after cancelling the pending logout")
			ctx.SetJSONError(messageOperationFailed)

			return
		}
	}

	ctx.ReplyOK()
}

// LogoutPOST is the handler logging out the user attached to the given cookie.
func LogoutPOST(ctx *middlewares.AutheliaCtx) {
	body := bodyLogout{}
	responseBody := bodyLogoutResponse{SafeTargetURL: false}

	err := ctx.ParseBody(&body)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred parsing the logout request body")
		ctx.SetJSONError(messageOperationFailed)
	}

	if userSession, errSession := ctx.GetSession(); errSession == nil {
		if logout := logoutPending(ctx, userSession, body.FlowID); logout != nil {
			responseBody.RedirectURL = logoutRedirectURL(logout)
		}

		oidcBackChannelLogout(ctx, &userSession)
	}

	logoutRemoveOAuth2SessionIDs(ctx)

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

func logoutRemoveOAuth2SessionIDs(ctx *middlewares.AutheliaCtx) {
	if ctx.Configuration.IdentityProviders.OIDC == nil {
		return
	}

	provider, err := ctx.GetSessionProvider()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred obtaining the session provider during logout")

		return
	}

	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred obtaining the user session during logout")

		return
	}

	if userSession.PublicID == "" {
		return
	}

	if err = ctx.Providers.StorageProvider.DeleteOAuth2SessionIDByPublicID(ctx, provider.GetIssuer(), userSession.PublicID); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred removing the OpenID Connect session identifiers during logout")
	}
}

func logoutPending(ctx *middlewares.AutheliaCtx, userSession session.UserSession, flowID string) *session.OpenIDConnectLogout {
	if !logoutFlowMatches(userSession.OpenIDConnectLogout, flowID) || !ctx.GetClock().Now().Before(userSession.OpenIDConnectLogout.Expires) {
		return nil
	}

	return userSession.OpenIDConnectLogout
}

func logoutFlowMatches(logout *session.OpenIDConnectLogout, flowID string) bool {
	if logout == nil || logout.FlowID == "" || flowID == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(logout.FlowID), []byte(flowID)) == 1
}

func logoutRedirectURL(logout *session.OpenIDConnectLogout) string {
	if logout.RedirectURI == "" || logout.State == "" {
		return logout.RedirectURI
	}

	separator := "?"

	if strings.Contains(logout.RedirectURI, "?") {
		separator = "&"
	}

	return logout.RedirectURI + separator + url.Values{oidc.FormParameterState: []string{logout.State}}.Encode()
}
