// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/events"
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
	eventEmailReasonPasswordReuse    = "was unsuccessful because the new password must be different to the password currently set on your account and any previously used passwords."
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

	ctxLogEvent(ctx, events.TypeUserPasswordResetFailed, username, eventLogActionPasswordResetFailure, emailEventBody{
		Prefix: eventEmailActionPasswordModifyPrefix,
		Body:   eventEmailActionPasswordReset,
		Suffix: reason,
	}, map[string]any{eventLogKeyAction: eventEmailActionPasswordReset})
}

func ctxLogEvent(ctx *middlewares.AutheliaCtx, eventType, username, description string, body emailEventBody, eventDetails map[string]any) {
	var (
		details *authentication.UserDetails
		err     error
	)

	ctx.Logger.Debugf("Getting user details for notification")

	if details, err = ctx.Providers.UserProvider.GetDetailsCached(username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred looking up user details for user '%s' while attempting to alert them of an important event", username)

		ctx.EmitEvent(events.NewEvent(newUserEventData(eventType, events.Subject{Username: username, RemoteIP: ctx.RemoteIP().String()},
			eventCredentialDescription(eventDetails), ctx.NewNotification(err, description, nil, nil))))

		return
	}

	notified := len(details.Emails) != 0

	if !notified {
		err = fmt.Errorf("no email address was found for user")

		ctx.Logger.WithError(err).Errorf("Error occurred looking up user details for user '%s' while attempting to alert them of an important event", username)
	} else {
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

		err = ctx.Providers.Notifier.Send(ctx, addresses[0], description, ctx.Providers.Templates.GetEventEmailTemplate(), data)
	}

	ctx.EmitEvent(events.NewEvent(newUserEventData(eventType, subjectFromDetails(ctx, username, details), eventCredentialDescription(eventDetails),
		ctx.NewNotification(err, description, recipientsFromDetails(username, details), &events.NotificationValues{
			BodyPrefix: body.Prefix,
			BodyEvent:  body.Body,
			BodySuffix: body.Suffix,
			Details:    eventDetails,
		}))))

	if notified && err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred sending notification to user '%s' while attempting to alert them of an important event", username)
	}
}

func eventCredentialDescription(eventDetails map[string]any) string {
	value, ok := eventDetails[eventLogKeyDescription]
	if !ok {
		return ""
	}

	description, ok := value.(string)
	if !ok {
		return ""
	}

	return description
}

func recipientsFromDetails(username string, details *authentication.UserDetails) (recipients []events.Recipient) {
	if details == nil || len(details.Emails) == 0 {
		return nil
	}

	return []events.Recipient{
		{
			Email:       details.Emails[0],
			Username:    username,
			DisplayName: details.DisplayName,
		},
	}
}

func newUserEventData(eventType string, subject events.Subject, description string, notification *events.Notification) events.Data {
	switch eventType {
	case events.TypeUserPasswordChanged, events.TypeUserPasswordReset, events.TypeUserPasswordResetFailed:
		return &events.DataUserPassword{Type: eventType, Subject: subject, Notification: notification}
	default:
		return &events.DataUserCredential{Type: eventType, Subject: subject, Description: description, Notification: notification}
	}
}

func subjectFromDetails(ctx *middlewares.AutheliaCtx, username string, details *authentication.UserDetails) events.Subject {
	return events.Subject{
		Username:    username,
		DisplayName: details.DisplayName,
		Emails:      details.Emails,
		RemoteIP:    ctx.RemoteIP().String(),
	}
}

func ctxNotifyEvent(ctx *middlewares.AutheliaCtx, eventType, username, description string, body emailEventBody, eventDetails map[string]any) (err error) {
	var details *authentication.UserDetails

	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil {
		ctx.EmitEvent(events.NewEvent(newUserEventData(eventType, events.Subject{Username: username, RemoteIP: ctx.RemoteIP().String()},
			eventCredentialDescription(eventDetails), ctx.NewNotification(err, description, nil, nil))))

		return fmt.Errorf("failed to look up user details: %w", err)
	}

	values := &events.NotificationValues{
		BodyPrefix: body.Prefix,
		BodyEvent:  body.Body,
		BodySuffix: body.Suffix,
		Details:    eventDetails,
	}

	if len(details.Emails) == 0 {
		err = errors.New("user has no email address")

		ctx.EmitEvent(events.NewEvent(newUserEventData(eventType, subjectFromDetails(ctx, username, details),
			eventCredentialDescription(eventDetails), ctx.NewNotification(err, description, nil, values))))

		return err
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

	addresses := details.Addresses()

	err = ctx.Providers.Notifier.Send(ctx, addresses[0], description, ctx.Providers.Templates.GetEventEmailTemplate(), data)

	ctx.EmitEvent(events.NewEvent(newUserEventData(eventType, subjectFromDetails(ctx, username, details),
		eventCredentialDescription(eventDetails), ctx.NewNotification(err, description, recipientsFromDetails(username, details), values))))

	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
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

func replyPasswordModified(ctx *middlewares.AutheliaCtx, failed bool) {
	if failed {
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}
