// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package notification

import (
	"context"
	"net/mail"

	"github.com/authelia/authelia/v4/internal/templates"
)

// DisabledNotifier is a Notifier which discards every notification. It is used when the notifier is disabled, which
// is only permitted when at least one webhook destination is configured to carry the occurrences in its place.
type DisabledNotifier struct{}

// NewDisabledNotifier returns a Notifier which discards every notification.
func NewDisabledNotifier() *DisabledNotifier {
	return &DisabledNotifier{}
}

// StartupCheck implements the startup check provider interface. There is nothing to check.
func (n *DisabledNotifier) StartupCheck() (err error) {
	return nil
}

// Send discards the notification. It returns no error because a disabled notifier is a deliberate configuration
// rather than a delivery failure, and an error here would surface to the user as a failed flow.
func (n *DisabledNotifier) Send(_ context.Context, _ mail.Address, _ string, _ *templates.EmailTemplate, _ any) (err error) {
	return nil
}
