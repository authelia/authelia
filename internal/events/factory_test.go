// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/events"
)

func TestNewNotification(t *testing.T) {
	recipients := []events.Recipient{
		{Email: "john@example.com", Username: "john", DisplayName: "John Smith"},
	}

	values := &events.NotificationValues{BodyEvent: "Password Change"}

	testCases := []struct {
		name       string
		err        error
		suppressed bool
		title      string
		recipients []events.Recipient
		values     *events.NotificationValues
		expected   *events.Notification
	}{
		{
			"ShouldRecordDelivery",
			nil,
			false,
			"Password changed successfully",
			recipients,
			values,
			&events.Notification{
				Sent:       true,
				Recipients: recipients,
				Title:      "Password changed successfully",
				Values:     values,
			},
		},
		{
			"ShouldRecordFailure",
			fmt.Errorf("connection refused"),
			false,
			"Password changed successfully",
			recipients,
			values,
			&events.Notification{
				Sent:       false,
				Recipients: recipients,
				Title:      "Password changed successfully",
				Error:      "connection refused",
				Values:     values,
			},
		},
		{
			"ShouldHandleNoRecipientsOrValues",
			nil,
			false,
			"Confirm your identity",
			nil,
			nil,
			&events.Notification{
				Sent:  true,
				Title: "Confirm your identity",
			},
		},
		{
			"ShouldRecordSuppressionWhenTheNotifierIsDisabled",
			nil,
			true,
			"Password changed successfully",
			recipients,
			values,
			&events.Notification{
				Sent:       false,
				Suppressed: true,
				Recipients: recipients,
				Title:      "Password changed successfully",
				Values:     values,
			},
		},
		{
			"ShouldNotRecordAnErrorWhenSuppressed",
			fmt.Errorf("connection refused"),
			true,
			"Password changed successfully",
			recipients,
			values,
			&events.Notification{
				Sent:       false,
				Suppressed: true,
				Recipients: recipients,
				Title:      "Password changed successfully",
				Values:     values,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := events.NewNotification(tc.err, tc.suppressed, tc.title, tc.recipients, tc.values)

			require.NotNil(t, actual)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
