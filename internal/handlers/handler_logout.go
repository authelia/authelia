// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// LogoutGET is the handler reporting whether an RP-Initiated Logout identified by the flow id is awaiting confirmation,
// so the front-end knows whether to ask the user to confirm the logout rather than perform it straight away.
func LogoutGET(ctx *middlewares.AutheliaCtx) {
	responseBody := bodyLogoutPendingResponse{}

	logout, err := logoutPending(ctx, string(ctx.QueryArgs().Peek(queryArgFlowID)))
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred loading the pending logout")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if logout != nil {
		userSession, err := ctx.GetSession()
		if err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred obtaining the user session to determine if a logout is pending")
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		if err = oidcLogoutValidate(ctx, &userSession, logout); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred validating the pending logout")
			ctx.SetJSONError(messageOperationFailed)

			return
		}

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
	if flowID, err := uuid.Parse(string(ctx.QueryArgs().Peek(queryArgFlowID))); err == nil {
		if err = ctx.Providers.StorageProvider.DeleteOAuth2LogoutSession(ctx, flowID); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred cancelling the pending logout")
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

	var err error

	if err = ctx.ParseBody(&body); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred parsing the logout request body")

		body = bodyLogout{}
	}

	pending := logoutLoad(ctx, body.FlowID)
	logout := pending

	userSession, errSession := ctx.GetSession()

	if errSession == nil && logout != nil {
		if err = oidcLogoutValidate(ctx, &userSession, logout); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred validating the pending logout: the user is logged out without it")

			logout = nil
		}

		if logout != nil {
			responseBody.RedirectURL = logoutRedirectURL(logout)
		}
	}

	if err = ctx.DestroySession(); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred destroying the user session during logout")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	// Removed only once the session is destroyed, so the logout can be retried when it fails.
	logoutRemove(ctx, pending)

	// Ended only once the session is destroyed, so a session which survives keeps its OpenID Connect sessions, and a
	// client can't join the session while its participants are being notified.
	if errSession == nil {
		oidcLogout(ctx, &userSession, logout)
	}

	var redirectionURL *url.URL

	if redirectionURL, err = url.ParseRequestURI(body.TargetURL); err != nil {
		ctx.GetLogger().WithError(err).Debug("Error occurred parsing the redirection URL")
	} else {
		responseBody.SafeTargetURL = ctx.IsSafeRedirectionTargetURI(redirectionURL)
	}

	if body.TargetURL != "" {
		ctx.Logger.Debugf("Logout target url is %s, safe %t", body.TargetURL, responseBody.SafeTargetURL)
	}

	if err = ctx.SetJSONBody(responseBody); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred setting the logout response body")
		ctx.SetJSONError(messageOperationFailed)
	}
}

// logoutPending returns the RP-Initiated Logout identified by the flow id which is awaiting confirmation, or nil when
// there is none or it has expired.
func logoutPending(ctx *middlewares.AutheliaCtx, flowID string) (logout *model.OAuth2LogoutSession, err error) {
	id, err := uuid.Parse(flowID)
	if err != nil {
		return nil, nil
	}

	if logout, err = ctx.Providers.StorageProvider.LoadOAuth2LogoutSession(ctx, id); err != nil {
		return nil, err
	}

	if logout == nil || !ctx.GetClock().Now().Before(logout.ExpiresAt) {
		return nil, nil
	}

	return logout, nil
}

func logoutLoad(ctx *middlewares.AutheliaCtx, flowID string) (logout *model.OAuth2LogoutSession) {
	logout, err := logoutPending(ctx, flowID)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred loading the pending logout: the user is logged out without it")

		return nil
	}

	return logout
}

func logoutRemove(ctx *middlewares.AutheliaCtx, logout *model.OAuth2LogoutSession) {
	if logout == nil {
		return
	}

	if err := ctx.Providers.StorageProvider.DeleteOAuth2LogoutSession(ctx, logout.FlowID); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred removing the pending logout")
	}
}

func logoutRedirectURL(logout *model.OAuth2LogoutSession) string {
	if logout.RedirectURI == "" || logout.State == "" {
		return logout.RedirectURI
	}

	separator := "?"

	if strings.Contains(logout.RedirectURI, "?") {
		separator = "&"
	}

	return logout.RedirectURI + separator + url.Values{oidc.FormParameterState: []string{logout.State}}.Encode()
}
