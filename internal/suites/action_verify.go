package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) isVerifyIdentityShowing(t *testing.T, page *rod.Page) bool {
	require.NoError(t, page.WaitStable(time.Millisecond*100))

	has, _, err := page.Has("#dialog-verify-one-time-code")
	require.NoError(t, err)

	return has
}

func (rs *RodSession) doMaybeVerifyIdentity(t *testing.T, page *rod.Page) {
	if !rs.isVerifyIdentityShowing(t, page) {
		return
	}

	rs.doMustVerifyIdentity(t, page)
}

func (rs *RodSession) doMustVerifyIdentity(t *testing.T, page *rod.Page) {
	element := rs.WaitElementLocatedByID(t, page, "one-time-code")
	code := doGetOneTimeCodeFromLastMail(t)

	require.NoError(t, element.Type(rs.toInputs(code)...))

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "dialog-verify").Click("left", 1))
}

func (rs *RodSession) doMustVerifyIdentityBadCode(t *testing.T, page *rod.Page) {
	element := rs.WaitElementLocatedByID(t, page, "one-time-code")

	require.NoError(t, element.Type(rs.toInputs("BADCODE")...))

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "dialog-verify").Click("left", 1))

	rs.verifyNotificationDisplayed(t, page, "The One-Time Code either doesn't match the one generated or an unknown error occurred")
}
