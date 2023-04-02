// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doRegisterTOTP(t *testing.T, page *rod.Page) string {
	err := rs.WaitElementLocatedByID(t, page, "register-link").Click("left", 1)
	require.NoError(t, err)
	rs.verifyMailNotificationDisplayed(t, page)
	link := doGetLinkFromLastMail(t)
	rs.doVisit(t, page, link)
	secretURL, err := page.MustElement("#secret-url").Attribute("value")
	assert.NoError(t, err)

	secret := (*secretURL)[strings.LastIndex(*secretURL, "=")+1:]
	assert.NotEqual(t, "", secret)
	assert.NotNil(t, secret)

	return secret
}

func (rs *RodSession) doEnterOTP(t *testing.T, page *rod.Page, code string) {
	inputs := rs.WaitElementsLocatedByID(t, page, "otp-input input")

	for i := 0; i < len(code); i++ {
		err := inputs[i].Type(input.Key(code[i]))
		require.NoError(t, err)
	}
}

func (rs *RodSession) doValidateTOTP(t *testing.T, page *rod.Page, secret string) {
	code, err := totp.GenerateCode(secret, time.Now())
	assert.NoError(t, err)
	rs.doEnterOTP(t, page, code)
}
