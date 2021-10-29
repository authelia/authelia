package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
)

func (rs *RodSession) doInitiatePasswordReset(t *testing.T, page *rod.Page, username string) {
	rs.WaitElementLocatedByID(t, page, "reset-password-button").MustClick()
	// Fill in username.
	rs.WaitElementLocatedByID(t, page, "username-textfield").MustInput(username)
	// And click on the reset button.
	rs.WaitElementLocatedByID(t, page, "reset-button").MustClick()
}

func (rs *RodSession) doCompletePasswordReset(t *testing.T, page *rod.Page, newPassword1, newPassword2 string) {
	link := doGetLinkFromLastMail(t)
	rs.doVisit(page, link)

	time.Sleep(1 * time.Second)
	rs.WaitElementLocatedByID(t, page, "password1-textfield").MustInput(newPassword1)

	time.Sleep(1 * time.Second)
	rs.WaitElementLocatedByID(t, page, "password2-textfield").MustInput(newPassword2)

	rs.WaitElementLocatedByID(t, page, "reset-button").MustClick()
}

func (rs *RodSession) doSuccessfullyCompletePasswordReset(t *testing.T, page *rod.Page, newPassword1, newPassword2 string) {
	rs.doCompletePasswordReset(t, page, newPassword1, newPassword2)
	rs.verifyIsFirstFactorPage(t, page)
}

func (rs *RodSession) doUnsuccessfulPasswordReset(t *testing.T, page *rod.Page, newPassword1, newPassword2 string) {
	rs.doCompletePasswordReset(t, page, newPassword1, newPassword2)
	rs.verifyNotificationDisplayed(t, page, "Your supplied password does not meet the password policy requirements.")
}

func (rs *RodSession) doResetPassword(t *testing.T, page *rod.Page, username, newPassword1, newPassword2 string, unsuccessful bool) {
	rs.doInitiatePasswordReset(t, page, username)
	// then wait for the "email sent notification".
	rs.verifyMailNotificationDisplayed(t, page)

	if unsuccessful {
		rs.doUnsuccessfulPasswordReset(t, page, newPassword1, newPassword2)
	} else {
		rs.doSuccessfullyCompletePasswordReset(t, page, newPassword1, newPassword2)
	}
}
