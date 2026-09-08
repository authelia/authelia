package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) doChangePassword(t *testing.T, page *rod.Page, oldPassword, newPassword1, newPassword2, notification string) {
	rs.ClickElementLocatedByID(t, page, "change-password-button")

	rs.doMaybeVerifyIdentity(t, page, "#old-password")

	rs.TypeElementLocatedByID(t, page, "old-password", oldPassword)
	rs.TypeElementLocatedByID(t, page, "new-password", newPassword1)
	rs.TypeElementLocatedByID(t, page, "repeat-new-password", newPassword2)

	rs.ClickElementLocatedByID(t, page, "password-change-dialog-submit")
	rs.verifyNotificationDisplayed(t, page, notification)
}
