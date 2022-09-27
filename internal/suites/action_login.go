package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doFillLoginPageAndClick(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool) {
	usernameElement := rs.WaitElementLocatedByID(t, page, "username-textfield")
	err := usernameElement.Input(username)
	require.NoError(t, err)

	passwordElement := rs.WaitElementLocatedByID(t, page, "password-textfield")
	err = passwordElement.Input(password)
	require.NoError(t, err)

	if keepMeLoggedIn {
		keepMeLoggedInElement := rs.WaitElementLocatedByID(t, page, "remember-checkbox")
		err = keepMeLoggedInElement.Click("left", 1)
		require.NoError(t, err)
	}

	buttonElement := rs.WaitElementLocatedByID(t, page, "sign-in-button")
	err = buttonElement.Click("left", 1)
	require.NoError(t, err)
}

// Login 1FA.
func (rs *RodSession) doLoginOneFactor(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool, targetURL string) {
	rs.doVisitLoginPage(t, page, targetURL)
	rs.doFillLoginPageAndClick(t, page, username, password, keepMeLoggedIn)
}

// Login 1FA and 2FA subsequently (must already be registered).
func (rs *RodSession) doLoginTwoFactor(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool, otpSecret, targetURL string) {
	rs.doLoginOneFactor(t, page, username, password, keepMeLoggedIn, targetURL)
	rs.verifyIsSecondFactorPage(t, page)
	rs.doValidateTOTP(t, page, otpSecret)
	// timeout when targetURL is not defined to prevent a show stopping redirect when visiting a protected domain.
	if targetURL == "" {
		time.Sleep(1 * time.Second)
	}
}

// Login 1FA and register 2FA.
func (rs *RodSession) doLoginAndRegisterTOTP(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool) string {
	rs.doLoginOneFactor(t, page, username, password, keepMeLoggedIn, "")
	secret := rs.doRegisterTOTP(t, page)
	rs.doVisit(t, page, GetLoginBaseURL())
	rs.verifyIsSecondFactorPage(t, page)

	return secret
}

// Register a user with TOTP, logout and then authenticate until TOTP-2FA.
func (rs *RodSession) doRegisterAndLogin2FA(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool, targetURL string) string { //nolint:unparam
	// Register TOTP secret and logout.
	secret := rs.doRegisterThenLogout(t, page, username, password)
	rs.doLoginTwoFactor(t, page, username, password, keepMeLoggedIn, secret, targetURL)

	return secret
}
