// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmailAddressedTo(t *testing.T) {
	message := EmailMessage{To: []Address{{Address: "Harry.Potter@authelia.com"}}}

	assert.True(t, isEmailAddressedTo(message, "harry.potter@authelia.com"))
	assert.False(t, isEmailAddressedTo(message, "bob.dylan@authelia.com"))
	assert.False(t, isEmailAddressedTo(EmailMessage{}, "harry.potter@authelia.com"))
}
