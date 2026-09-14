// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"fmt"
	"net/url"
	"regexp"
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) verifyIsDeny(t *testing.T, page *rod.Page) {
	rs.WaitElementLocatedByID(t, page, "access-denied-stage")

	base, err := url.ParseRequestURI(GetLoginBaseURLWithFallbackPrefix(BaseDomain, "/"))
	require.NoError(t, err)

	rs.verifyURLIsRegexp(t, page, regexp.MustCompile(fmt.Sprintf(`^%s\?ec=forbidden&rd=`, regexp.QuoteMeta(base.JoinPath("error").String()))))
}
