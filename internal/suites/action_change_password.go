package suites

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doChangePassword(t *testing.T, page *rod.Page, oldPassword, newPassword1, newPassword2, notification string) {
	rs.ClickElementLocatedByID(t, page, "change-password-button")

	rs.doMaybeVerifyIdentity(t, page)

	oldPasswordInput := rs.WaitElementLocatedByID(t, page, "old-password")
	newPasswordInput := rs.WaitElementLocatedByID(t, page, "new-password")
	repeatNewPasswordInput := rs.WaitElementLocatedByID(t, page, "repeat-new-password")

	require.NoError(t, oldPasswordInput.Type(rs.toInputs(oldPassword)...))
	require.NoError(t, newPasswordInput.Type(rs.toInputs(newPassword1)...))
	require.NoError(t, repeatNewPasswordInput.Type(rs.toInputs(newPassword2)...))

	rs.ClickElementLocatedByID(t, page, "password-change-dialog-submit")
	rs.verifyNotificationDisplayed(t, page, notification)
}
