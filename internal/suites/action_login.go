package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doFillLoginPageAndClick(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool) {
	usernameElement := rs.WaitElementLocatedByID(t, page, "username-textfield")
	passwordElement := rs.WaitElementLocatedByID(t, page, "password-textfield")
	buttonElement := rs.WaitElementLocatedByID(t, page, "sign-in-button")

username:
	err := usernameElement.MustSelectAllText().Input(username)
	require.NoError(t, err)

	if usernameElement.MustText() != username {
		goto username
	}

password:
	err = passwordElement.MustSelectAllText().Input(password)
	require.NoError(t, err)

	if passwordElement.MustText() != password {
		goto password
	}

	if keepMeLoggedIn {
		keepMeLoggedInElement := rs.WaitElementLocatedByID(t, page, "remember-checkbox")
		err = keepMeLoggedInElement.Click("left", 1)
		require.NoError(t, err)
	}

click:
	err = buttonElement.Click("left", 1)
	require.NoError(t, err)

	if buttonElement.MustInteractable() {
		goto click
	}
}

// Login 1FA.
func (rs *RodSession) doLoginOneFactor(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool, domain string, targetURL string) {
	rs.doVisitLoginPage(t, page, domain, targetURL)
	rs.doFillLoginPageAndClick(t, page, username, password, keepMeLoggedIn)
}

// Login 1FA and 2FA subsequently (must already be registered).
func (rs *RodSession) doLoginSecondFactorTOTP(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool, targetURL string) {
	rs.doLoginOneFactor(t, page, username, password, keepMeLoggedIn, BaseDomain, targetURL)
	rs.verifyIsSecondFactorPage(t, page)
	rs.doValidateTOTP(t, page, username)
	// timeout when targetURL is not defined to prevent a show stopping redirect when visiting a protected domain.
	if targetURL == "" {
		require.NoError(t, page.WaitStable(time.Second))
	}
}

// Login 1FA and register 2FA.
func (rs *RodSession) doLoginAndRegisterTOTP(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool) {
	rs.doLoginOneFactor(t, page, username, password, keepMeLoggedIn, BaseDomain, "")
	rs.doOpenSettingsAndRegisterTOTP(t, page, username)

	rs.verifyIsSecondFactorPage(t, page)
}

// Register a user with TOTP, logout and then authenticate until TOTP-2FA.
func (rs *RodSession) doRegisterTOTPAndLogin2FA(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool, targetURL string) { //nolint:unparam
	// Register TOTP secret and logout.
	rs.doLoginAndRegisterTOTPThenLogout(t, page, username, password)
	rs.doLoginSecondFactorTOTP(t, page, username, password, keepMeLoggedIn, targetURL)
}

func (rs *RodSession) doLoginAndRegisterWebAuthn(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool) {
	rs.doLoginOneFactor(t, page, username, password, keepMeLoggedIn, BaseDomain, "")
	require.Greater(t, len(rs.GetWebAuthnAuthenticatorID()), 0)
	rs.doWebAuthnCredentialRegisterAfterVisitSettings(t, page, "testing")

	rs.verifyIsSecondFactorPage(t, page)
}
