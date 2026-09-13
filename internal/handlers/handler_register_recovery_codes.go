// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

// RecoveryCodesGenerationPOST generates a fresh batch of recovery codes for the current user, replacing any
// existing codes by soft-revoking them. The plaintext codes are returned exactly once in the response and never
// persisted in any reversible form.
func RecoveryCodesGenerationPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred generating recovery codes: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.GetLogger().WithError(errUserAnonymous).Error("Error occurred generating recovery codes")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	now := ctx.GetClock().Now()

	if err = ctx.Providers.StorageProvider.RevokeRecoveryCodesByUsername(ctx, userSession.Username, model.NewNullIP(ctx.RemoteIP())); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred generating recovery codes for user '%s': error revoking existing codes", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	plaintexts := make([]string, 0, model.RecoveryCodeBatchSize)

	for i := 0; i < model.RecoveryCodeBatchSize; i++ {
		var code *model.RecoveryCode

		if code, err = model.NewRecoveryCode(ctx, userSession.Username); err != nil {
			ctx.GetLogger().WithError(err).Errorf("Error occurred generating recovery codes for user '%s': error generating code %d", userSession.Username, i)

			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		code.CreatedAt = now

		if err = ctx.Providers.StorageProvider.SaveRecoveryCode(ctx, code); err != nil {
			ctx.GetLogger().WithError(err).Errorf("Error occurred generating recovery codes for user '%s': error inserting code %d", userSession.Username, i)

			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		plaintexts = append(plaintexts, code.Plaintext)
	}

	if ctx.Providers.Metrics != nil {
		ctx.Providers.Metrics.RecordRecoveryCodesGenerated(model.RecoveryCodeBatchSize)
	}

	notificationSent := true

	body := emailEventBody{
		Prefix: "A new set of recovery codes was generated for your account.",
		Body:   fmt.Sprintf("Generated on %s. Any previously generated recovery codes have been revoked and can no longer be used.", now.Format("2006-01-02 15:04 MST")),
		Suffix: "If this was not you, change your account password immediately and re-generate the codes from a known good device.",
	}

	if notifyErr := ctxNotifyEvent(ctx, userSession.Username, "Recovery codes generated", body, nil); notifyErr != nil {
		notificationSent = false

		ctx.GetLogger().WithError(notifyErr).Warnf("Failed to send recovery-codes-generated notification email for user '%s'", userSession.Username)
	}

	if err = ctx.SetJSONBody(map[string]any{
		"codes":             plaintexts,
		"notification_sent": notificationSent,
	}); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred generating recovery codes for user '%s': %s", userSession.Username, errStrRespBody)
	}
}

// RecoveryCodesStatusGET returns the user's recovery code status: how many codes remain unused out of how many were
// generated, when the most recent batch was generated, and when one was most recently consumed (if any).
func RecoveryCodesStatusGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred retrieving recovery codes status: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.GetLogger().WithError(errUserAnonymous).Error("Error occurred retrieving recovery codes status")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	codes, err := ctx.Providers.StorageProvider.LoadRecoveryCodesByUsername(ctx, userSession.Username)
	if err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred retrieving recovery codes for user '%s'", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	status := computeRecoveryCodesStatus(codes)

	if err = ctx.SetJSONBody(status); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred retrieving recovery codes status for user '%s': %s", userSession.Username, errStrRespBody)
	}
}

type recoveryCodesStatus struct {
	CodesRemaining int        `json:"codes_remaining"`
	CodesTotal     int        `json:"codes_total"`
	GeneratedAt    *time.Time `json:"generated_at,omitempty"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
}

func computeRecoveryCodesStatus(codes []model.RecoveryCode) recoveryCodesStatus {
	status := recoveryCodesStatus{}

	if len(codes) == 0 {
		return status
	}

	var (
		latestGeneratedAt time.Time
		latestUsedAt      time.Time
	)

	for _, c := range codes {
		if c.RevokedAt.Valid {
			continue
		}

		if c.CreatedAt.After(latestGeneratedAt) {
			latestGeneratedAt = c.CreatedAt
		}

		if c.ConsumedAt.Valid && c.ConsumedAt.Time.After(latestUsedAt) {
			latestUsedAt = c.ConsumedAt.Time
		}
	}

	if latestGeneratedAt.IsZero() {
		return status
	}

	for _, c := range codes {
		if c.RevokedAt.Valid {
			continue
		}

		// Only the latest batch is counted so the displayed totals match the codes the user can actually use today.
		if !c.CreatedAt.Equal(latestGeneratedAt) {
			continue
		}

		status.CodesTotal++

		if !c.ConsumedAt.Valid {
			status.CodesRemaining++
		}
	}

	status.GeneratedAt = &latestGeneratedAt

	if !latestUsedAt.IsZero() {
		status.LastUsedAt = &latestUsedAt
	}

	return status
}
