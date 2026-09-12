// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package notification

import (
	"context"
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisabledNotifierShouldDiscardWithoutError(t *testing.T) {
	notifier := NewDisabledNotifier()

	require.NoError(t, notifier.StartupCheck())

	assert.NotPanics(t, func() {
		assert.NoError(t, notifier.Send(context.Background(), mail.Address{Address: "john@example.com"}, "Password changed successfully", nil, nil))
	})
}

func TestDisabledNotifierShouldImplementNotifier(t *testing.T) {
	assert.Implements(t, (*Notifier)(nil), NewDisabledNotifier())
}
