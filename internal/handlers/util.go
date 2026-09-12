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

	eventLogCategoryOneTimePassword    = "One-Time Password"
	eventLogCategoryWebAuthnCredential = "WebAuthn Credential" //nolint:gosec
)

type emailEventBody struct {
	Prefix string
	Body   string
	Suffix string
}

func ctxLogEvent(ctx *middlewares.AutheliaCtx, eventType, username, description string, body emailEventBody, eventDetails map[string]any) {
	var (
		details *authentication.UserDetails
		err     error
	)

	ctx.Logger.Debugf("Getting user details for notification")

	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred looking up user details for user '%s' while attempting to alert them of an important event", username)

		ctx.Providers.Events.Emit(ctx, events.NewEvent(&events.DataUserCredential{
			Type:         eventType,
			Subject:      events.Subject{Username: username, RemoteIP: ctx.RemoteIP().String()},
			Description:  eventCredentialDescription(eventDetails),
			Notification: events.NewNotification(err, ctx.GetConfiguration().Notifier.Disable, description, nil, nil),
		}))

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

	ctx.Providers.Events.Emit(ctx, events.NewEvent(&events.DataUserCredential{
		Type:        eventType,
		Username:    username,
		DisplayName: details.DisplayName,
		Emails:      details.Emails,
		RemoteIP:    ctx.RemoteIP().String(),
		Description: eventCredentialDescription(eventDetails),
		Notification: events.NewNotification(err, ctx.GetConfiguration().Notifier.Disable, description, recipientsFromDetails(username, details), &events.NotificationValues{
			BodyPrefix: body.Prefix,
			BodyEvent:  body.Body,
			BodySuffix: body.Suffix,
			Details:    eventDetails,
		}),
	}))

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

	// Only the first address is notified, because UserDetails.Addresses preserves the order of Emails and every send
	// site passes addresses[0]. Listing the rest would name recipients which received nothing.
	return []events.Recipient{
		{
			Email:       details.Emails[0],
			Username:    username,
			DisplayName: details.DisplayName,
		},
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
