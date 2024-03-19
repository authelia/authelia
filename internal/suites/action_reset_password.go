package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doInitiatePasswordReset(t *testing.T, page *rod.Page, username string) {
	err := rs.WaitElementLocatedByID(t, page, "reset-password-button").Click("left", 1)
	require.NoError(t, err)

	require.NoError(t, page.WaitStable(time.Millisecond*100))

	// Fill in username.
	err = rs.WaitElementLocatedByID(t, page, "username-textfield").Input(username)
	require.NoError(t, err)
	// And click on the reset button.
	err = rs.WaitElementLocatedByID(t, page, "reset-button").Click("left", 1)
	require.NoError(t, err)
}

func (rs *RodSession) doCompletePasswordReset(t *testing.T, page *rod.Page, newPassword1, newPassword2 string) {
	link := doGetResetPasswordJWTLinkFromLastEmail(t)
	rs.doVisit(t, page, link)

	password1 := rs.WaitElementLocatedByID(t, page, "password1-textfield")
	password2 := rs.WaitElementLocatedByID(t, page, "password2-textfield")

password1:
	err := password1.MustSelectAllText().Input(newPassword1)
	require.NoError(t, err)

	if password1.MustText() != newPassword1 {
		goto password1
	}

password2:
	err = password2.MustSelectAllText().Input(newPassword2)
	require.NoError(t, err)

	if password2.MustText() != newPassword2 {
		goto password2
	}

	err = rs.WaitElementLocatedByID(t, page, "reset-button").Click("left", 1)
	require.NoError(t, err)
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
	// then wait for the "email sent notification".
	rs.verifyMailNotificationDisplayed(t, page)

	if unsuccessful {
		rs.doUnsuccessfulPasswordReset(t, page, newPassword1, newPassword2)
	} else {
		rs.doSuccessfullyCompletePasswordReset(t, page, newPassword1, newPassword2)
	}
}
