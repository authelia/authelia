// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifyMailNotificationDisplayed(t *testing.T, page *rod.Page) {
	rs.verifyNotificationDisplayed(t, page, "An email has been sent to your address to complete the process.")
}
