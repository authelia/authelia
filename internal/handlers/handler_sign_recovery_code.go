// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

// RecoveryCodePOST validates a recovery code and, on success, bumps the session to TwoFactor.
// Recovery code authentication does NOT mint an elevated session, so re-enrolling a primary 2FA factor
// still requires the existing email-based identity verification flow.
//
//nolint:gocyclo
func RecoveryCodePOST(ctx *middlewares.AutheliaCtx) {
	bodyJSON := bodySignRecoveryCodeRequest{}

	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred validating a recovery code authentication: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.GetLogger().WithError(errUserAnonymous).Error("Error occurred validating a recovery code authentication")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred validating a recovery code authentication for user '%s': %s", userSession.Username, errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if model.NormalizeRecoveryCode(bodyJSON.Code) == "" {
		ctx.GetLogger().Errorf("Error occurred validating a recovery code authentication for user '%s': empty code after normalization", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	code, err := ctx.Providers.StorageProvider.LoadRecoveryCode(ctx, userSession.Username, bodyJSON.Code)
	if err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred validating a recovery code authentication for user '%s': error occurred retrieving the recovery code from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if code == nil || code.ConsumedAt.Valid || code.RevokedAt.Valid {
		var reason error

		switch {
		case code == nil:
			reason = errors.New("the user input did not match any active recovery code")
		case code.ConsumedAt.Valid:
			reason = errors.New("the recovery code has already been consumed")
		case code.RevokedAt.Valid:
			reason = errors.New("the recovery code has been revoked")
		}

		ctx.GetLogger().WithError(reason).Errorf("Error occurred validating a recovery code authentication for user '%s'", userSession.Username)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypeRecoveryCode, nil)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.ConsumeRecoveryCode(ctx, code.ID, model.NewNullIP(ctx.RemoteIP())); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred validating a recovery code authentication for user '%s': error occurred consuming the recovery code", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypeRecoveryCode, nil)

	if err = ctx.RegenerateSession(); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred validating a recovery code authentication for user '%s': error regenerating the user session", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	userSession.SetTwoFactorRecoveryCode(ctx.GetClock().Now())

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred validating a recovery code authentication for user '%s': %s", userSession.Username, errStrUserSessionDataSave)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if remaining, errRemaining := ctx.Providers.StorageProvider.CountUnusedRecoveryCodesByUsername(ctx, userSession.Username); errRemaining == nil && remaining <= 2 {
		ctx.GetLogger().WithFields(map[string]any{"username": userSession.Username, "remaining": remaining}).Warn("User has signed in with a recovery code and is running low on remaining codes")
	}

	ctxLogEvent(ctx, userSession.Username, "Recovery code used", emailEventBody{
		Prefix: "A recovery code was used to sign in to your account.",
		Body:   fmt.Sprintf("The recovery code (generated %s) was just consumed and can no longer be used.", code.CreatedAt.Format("2006-01-02 15:04 MST")),
		Suffix: "If this was not you, change your account password immediately and revoke any remaining recovery codes from the Two-Factor Authentication settings page.",
	}, nil)

	if len(bodyJSON.Flow) > 0 {
		handleFlowResponse(ctx, &userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}
