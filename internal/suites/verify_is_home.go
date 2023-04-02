// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"fmt"
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifyIsHome(t *testing.T, page *rod.Page) {
	page.MustElementR("h1", "Access the secret")
	rs.verifyURLIs(t, page, fmt.Sprintf("%s/", HomeBaseURL))
}
