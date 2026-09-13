// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestRedactEmail(t *testing.T) {
	testCases := []struct {
		testName string
		input    string
		expected string
	}{
		{"ShouldRedactEmail", "james.dean@authelia.com", "j********n@authelia.com"},
		{"ShouldRedactShortEmail", "me@authelia.com", "**@authelia.com"},
		{"ShouldRedactInvalidEmail", "invalidEmail.com", ""},
		{"ShouldRedactUnicode", "søren@example.com", "s***n@example.com"},
		{"ShouldReturnUnicodeInRedactedEmail", "øpenme@example.com", "ø****e@example.com"},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			require.Equal(t, tc.expected, redactEmail(tc.input))
		})
	}
}

func condAuthentication(mock *mocks.MockAutheliaCtx, username, eventType, stage, method, reason string) func(event *events.Event) bool {
	return func(event *events.Event) bool {
		data, ok := event.Data.(*events.DataAuthentication)
		if !ok {
			return false
		}

		return event.Type == eventType && data.Username == username &&
			data.RemoteIP == mock.Ctx.RemoteIP().String() &&
			data.Stage == stage && data.Method == method && data.Reason == reason
	}
}

func expectAuthnSuccess(mock *mocks.MockAutheliaCtx, username, stage, method string) *gomock.Call {
	return mock.EventsMock.EXPECT().
		Emit(mock.Ctx, gomock.Cond(condAuthentication(mock, username, events.TypeSecurityAuthenticationSucceeded, stage, method, ""))).
		Times(1)
}

func expectAuthnFailure(mock *mocks.MockAutheliaCtx, username, stage, method, reason string) {
	mock.EventsMock.EXPECT().
		Emit(mock.Ctx, gomock.Cond(condAuthentication(mock, username, events.TypeSecurityAuthenticationFailed, stage, method, reason))).
		Times(1)
}

func condNotification(notification *events.Notification, sent bool) bool {
	if notification == nil {
		return false
	}

	// A suppressed notification is neither of these: it was never handed to a notifier and nothing failed. Asserting
	// it here keeps a contradictory state, such as one both sent and suppressed, from satisfying either branch.
	if notification.Suppressed {
		return false
	}

	if sent {
		return notification.Sent && notification.Error == ""
	}

	return !notification.Sent && notification.Error != ""
}

func condUserPasswordSuppressed(eventType string) func(event *events.Event) bool {
	return func(event *events.Event) bool {
		data, ok := event.Data.(*events.DataUserPassword)
		if !ok || data.Notification == nil {
			return false
		}

		n := data.Notification

		return event.Type == eventType && n.Suppressed && !n.Sent && n.Error == "" &&
			len(n.Recipients) == 1 && n.Recipients[0].Email == testEmail && n.Values != nil
	}
}

func condUserCredential(eventType, description string, sent bool) func(event *events.Event) bool {
	return func(event *events.Event) bool {
		data, ok := event.Data.(*events.DataUserCredential)
		if !ok {
			return false
		}

		return event.Type == eventType && data.Description == description && condNotification(data.Notification, sent)
	}
}

func condUserPassword(eventType string, sent bool) func(event *events.Event) bool {
	return func(event *events.Event) bool {
		data, ok := event.Data.(*events.DataUserPassword)
		if !ok {
			return false
		}

		return event.Type == eventType && condNotification(data.Notification, sent)
	}
}

func condSessionElevation(username, email string, sent bool) func(event *events.Event) bool {
	return func(event *events.Event) bool {
		data, ok := event.Data.(*events.DataSessionElevation)
		if !ok {
			return false
		}

		if !condNotification(data.Notification, sent) {
			return false
		}

		recipients := data.Notification.Recipients

		return event.Type == events.TypeUserSessionElevationRequested && data.Username == username &&
			len(recipients) == 1 && recipients[0].Email == email
	}
}
