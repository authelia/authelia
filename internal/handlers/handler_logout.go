// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/url"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

type logoutBody struct {
	TargetURL string `json:"targetURL"`
	FlowID    string `json:"flowID,omitempty"`
}

type logoutResponseBody struct {
	SafeTargetURL bool `json:"safeTargetURL"`
}

// LogoutPOST is the handler logging out the user attached to the given cookie.
func LogoutPOST(ctx *middlewares.AutheliaCtx) {
	body := logoutBody{}
	responseBody := logoutResponseBody{SafeTargetURL: false}

	err := ctx.ParseBody(&body)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred parsing the logout request body")
		ctx.SetJSONError(messageOperationFailed)
	}

	if body.FlowID != "" {
		handleLogoutClearConsentSubject(ctx, body.FlowID)
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

func handleLogoutClearConsentSubject(ctx *middlewares.AutheliaCtx, raw string) {
	flowID, err := uuid.Parse(raw)
	if err != nil {
		ctx.GetLogger().WithError(err).WithFields(map[string]any{logging.FieldFlowID: raw}).Debug("Skipped clearing the consent session subject during logout as the flow id is not valid")

		return
	}

	userSession, err := ctx.GetSession()
	if err != nil || userSession.IsAnonymous() {
		return
	}

	consent, err := ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, flowID)
	if err != nil {
		ctx.GetLogger().WithError(err).WithFields(map[string]any{logging.FieldFlowID: raw, logging.FieldUsername: userSession.Username}).Debug("Skipped clearing the consent session subject during logout as the consent session could not be loaded")

		return
	}

	if !consent.Subject.Valid {
		return
	}

	client, err := ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID)
	if err != nil {
		ctx.GetLogger().WithError(err).WithFields(map[string]any{logging.FieldFlowID: raw, logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID}).Error("Error occurred clearing the consent session subject during logout as the client could not be loaded")

		return
	}

	subject, err := ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username)
	if err != nil {
		ctx.GetLogger().WithError(err).WithFields(map[string]any{logging.FieldFlowID: raw, logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID}).Error("Error occurred clearing the consent session subject during logout as the subject could not be determined")

		return
	}

	cleared, err := ctx.Providers.StorageProvider.ClearOAuth2ConsentSessionSubject(ctx, flowID, subject)
	if err != nil {
		ctx.GetLogger().WithError(err).WithFields(map[string]any{logging.FieldFlowID: raw, logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID}).Error("Error occurred clearing the consent session subject during logout")

		return
	}

	if cleared {
		ctx.GetLogger().WithFields(map[string]any{logging.FieldFlowID: raw, logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID}).Debug("Cleared the consent session subject during logout")
	}
}
