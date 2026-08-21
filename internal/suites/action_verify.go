package suites

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) isVerifyIdentityShowing(t *testing.T, page *rod.Page, otherwise string) bool {
	_, err := page.Race().
		Element("#dialog-verify-one-time-code").
		Element(otherwise).
		Do()

	require.NoError(t, err)

	has, _, err := page.Has("#dialog-verify-one-time-code")
	require.NoError(t, err)

	return has
}

func (rs *RodSession) doMaybeVerifyIdentity(t *testing.T, page *rod.Page, otherwise string) {
	if !rs.isVerifyIdentityShowing(t, page, otherwise) {
		return
	}

	rs.doMustVerifyIdentity(t, page)
}

func (rs *RodSession) doMustVerifyIdentity(t *testing.T, page *rod.Page) {
	code := doGetOneTimeCodeFromLastMail(t)

	rs.TypeElementLocatedByID(t, page, "one-time-code", code)

	rs.ClickElementLocatedByID(t, page, "dialog-verify")
}

func (rs *RodSession) doMustVerifyIdentityBadCode(t *testing.T, page *rod.Page) {
	rs.TypeElementLocatedByID(t, page, "one-time-code", "BADCODE")

	rs.ClickElementLocatedByID(t, page, "dialog-verify")

	rs.verifyNotificationDisplayed(t, page, "The One-Time Code either doesn't match the one generated or an unknown error occurred")
}
