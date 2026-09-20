// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifyIsPasswordChangeRequiredPage(t *testing.T, page *rod.Page) {
	rs.WaitElementLocatedByID(t, page, "password-change-required")
}
