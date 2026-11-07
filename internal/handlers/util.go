// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

const (
	eventLogKeyAction      = "Action"
	eventLogKeyCategory    = "Category"
	eventLogKeyDescription = "Description"

	eventEmailAction2FABody  = "Second Factor Method"
	eventLogAction2FAAdded   = "Second Factor Method Added"
	eventLogAction2FARemoved = "Second Factor Method Removed"

	eventEmailAction2FAPrefix        = "a"
	eventEmailAction2FAAddedSuffix   = "was added to your account."
	eventEmailAction2FARemovedSuffix = "was removed from your account."

	eventEmailActionPasswordModifyPrefix = "your"
	eventEmailActionPasswordReset        = "Password Reset"
	eventEmailActionPasswordChange       = "Password Change"
	eventEmailActionPasswordModifySuffix = "was successful."

	eventLogActionPasswordResetFailure = "Password reset was unsuccessful"

	eventEmailReasonPasswordPolicy   = "was unsuccessful because the new password does not meet the password policy."
	eventEmailReasonPasswordBackend  = "was unsuccessful because the new password does not meet the requirements of the authentication backend."
	eventEmailReasonPasswordReuse    = "was unsuccessful because the new password must be different to the password currently set on your account."
	eventEmailReasonPasswordTooYoung = "was unsuccessful because your password was changed too recently to be changed again."

	eventLogCategoryOneTimePassword    = "One-Time Password"
	eventLogCategoryWebAuthnCredential = "WebAuthn Credential" //nolint:gosec
)

func passwordResetFailureReason(err error) string {
	if err == nil {
		return ""
	}

	switch msg := err.Error(); {
	case utils.IsStringInSliceContains(msg, ldapPasswordReuseErrors):
		return eventEmailReasonPasswordReuse
	case utils.IsStringInSliceContains(msg, ldapPasswordTooYoungErrors):
		return eventEmailReasonPasswordTooYoung
	case utils.IsStringInSliceContains(msg, ldapPasswordComplexityCodes),
		utils.IsStringInSliceContains(msg, ldapPasswordComplexityErrors):
		return eventEmailReasonPasswordBackend
	default:
		return ""
	}
}

type emailEventBody struct {
	Prefix string
	Body   string
	Suffix string
}

func ctxLogEventPasswordResetFailure(ctx *middlewares.AutheliaCtx, username, reason string) {
	if reason == "" {
		return
	}

	ctxLogEvent(ctx, username, eventLogActionPasswordResetFailure, emailEventBody{
		Prefix: eventEmailActionPasswordModifyPrefix,
		Body:   eventEmailActionPasswordReset,
		Suffix: reason,
	}, map[string]any{eventLogKeyAction: eventEmailActionPasswordReset})
}

func ctxLogEvent(ctx *middlewares.AutheliaCtx, username, description string, body emailEventBody, eventDetails map[string]any) {
	var (
		details *authentication.UserDetails
		err     error
	)

	ctx.Logger.Debugf("Getting user details for notification")

	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred looking up user details for user '%s' while attempting to alert them of an important event", username)
		return
	}

	if len(details.Emails) == 0 {
		ctx.Logger.WithError(fmt.Errorf("no email address was found for user")).Errorf("Error occurred looking up user details for user '%s' while attempting to alert them of an important event", username)
		return
	}

	data := templates.EmailEventValues{
		Title:       description,
		DisplayName: details.DisplayName,
		RemoteIP:    ctx.RemoteIP().String(),
		Details:     eventDetails,
		BodyPrefix:  body.Prefix,
		BodyEvent:   body.Body,
		BodySuffix:  body.Suffix,
	}

	ctx.Logger.Debugf("Getting user addresses for notification")

	addresses := details.Addresses()

	ctx.Logger.Debugf("Sending an email to user %s (%s) to inform them of an important event.", username, addresses[0].String())

	if err = ctx.Providers.Notifier.Send(ctx, addresses[0], description, ctx.Providers.Templates.GetEventEmailTemplate(), data); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred sending notification to user '%s' while attempting to alert them of an important event", username)
		return
	}
}

func redactEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	localRunes := []rune(parts[0])
	domain := parts[1]

	if len(localRunes) <= 2 {
		return strings.Repeat("*", len(localRunes)) + "@" + domain
	}

	first := string(localRunes[0])
	last := string(localRunes[len(localRunes)-1])
	middle := strings.Repeat("*", len(localRunes)-2)

	return first + middle + last + "@" + domain
}

func isRegulatorSkippedErr(err error) bool {
	var e *authentication.PoolErr

	if errors.As(err, &e) {
		return e.IsDeadlineError()
	}

	return false
}
