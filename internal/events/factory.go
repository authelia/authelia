// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

// NewNotification returns the notification object embedded in an event for an occurrence which notified a user. The
// error is recorded when delivery failed so that a receiver can observe notifications which never arrived.
//
// When suppressed is true the notifier is disabled, so nothing was handed to it and nothing failed: sent is false and
// the error is left empty. The recipients are still recorded as the addresses which would have been notified, and the
// values are still carried, since a deployment which disables the notifier relies on the webhook to convey them.
func NewNotification(err error, suppressed bool, title string, recipients []Recipient, values *NotificationValues) (notification *Notification) {
	notification = &Notification{
		Sent:       !suppressed && err == nil,
		Suppressed: suppressed,
		Recipients: recipients,
		Title:      title,
		Values:     values,
	}

	if suppressed {
		return notification
	}

	if err != nil {
		notification.Error = err.Error()
	}

	return notification
}
