package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) doInitiatePasswordReset(t *testing.T, page *rod.Page, username string) {
	rs.ClickElementLocatedByID(t, page, "reset-password-button")

	rs.TypeElementLocatedByID(t, page, "username-textfield", username)

	rs.ClickElementLocatedByID(t, page, "reset-button")
}

func (rs *RodSession) doCompletePasswordReset(t *testing.T, page *rod.Page, newPassword1, newPassword2 string) {
	link := doGetResetPasswordJWTLinkFromLastEmail(t)
	rs.doVisit(t, page, link)

	rs.TypeElementLocatedByID(t, page, "password1-textfield", newPassword1)
	rs.TypeElementLocatedByID(t, page, "password2-textfield", newPassword2)

	rs.ClickElementLocatedByID(t, page, "reset-button")
}

func (rs *RodSession) doSuccessfullyCompletePasswordReset(t *testing.T, page *rod.Page, newPassword1, newPassword2 string) {
	rs.doCompletePasswordReset(t, page, newPassword1, newPassword2)
	rs.verifyIsFirstFactorPage(t, page)
}

func (rs *RodSession) doUnsuccessfulPasswordReset(t *testing.T, page *rod.Page, newPassword1, newPassword2 string) {
	rs.doCompletePasswordReset(t, page, newPassword1, newPassword2)
	rs.verifyNotificationDisplayed(t, page, "Your supplied password does not meet the password policy requirements")
}

func (rs *RodSession) doResetPassword(t *testing.T, page *rod.Page, username, newPassword1, newPassword2 string, unsuccessful bool) {
	rs.doInitiatePasswordReset(t, page, username)
	rs.verifyMailNotificationDisplayed(t, page)

	if unsuccessful {
		rs.doUnsuccessfulPasswordReset(t, page, newPassword1, newPassword2)
	} else {
		rs.doSuccessfullyCompletePasswordReset(t, page, newPassword1, newPassword2)
	}
}
