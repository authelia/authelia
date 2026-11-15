// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) doChangeRequiredPassword(t *testing.T, page *rod.Page, oldPassword, newPassword1, newPassword2 string) {
	rs.verifyIsPasswordChangeRequiredPage(t, page)

	rs.TypeElementLocatedByID(t, page, "old-password", oldPassword)
	rs.TypeElementLocatedByID(t, page, "new-password", newPassword1)
	rs.TypeElementLocatedByID(t, page, "repeat-new-password", newPassword2)

	rs.ClickElementLocatedByID(t, page, "password-change-button")
}
