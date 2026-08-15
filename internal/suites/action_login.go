package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

const fillFieldAttempts = 10

func (rs *RodSession) doFillFieldUntilSet(t *testing.T, element *rod.Element, value string) {
	var (
		text string
		err  error
	)

	for i := 0; i < fillFieldAttempts; i++ {
		if err = element.SelectAllText(); err != nil {
			break
		}

		if err = element.Input(value); err != nil {
			break
		}

		if text, err = element.Text(); err != nil {
			break
		}

		if text == value {
			return
		}
	}

	require.NoError(t, err)
	require.Equal(t, value, text)
}

func (rs *RodSession) doFillLoginPageAndClick(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool) {
	usernameElement := rs.WaitElementLocatedByID(t, page, "username-textfield")
	passwordElement := rs.WaitElementLocatedByID(t, page, "password-textfield")
	buttonElement := rs.WaitElementLocatedByID(t, page, "sign-in-button")

	rs.doFillFieldUntilSet(t, usernameElement, username)
	rs.doFillFieldUntilSet(t, passwordElement, password)

	if keepMeLoggedIn {
		keepMeLoggedInElement := rs.WaitElementLocatedByID(t, page, "remember-checkbox")
		require.NoError(t, keepMeLoggedInElement.Click("left", 1))
	}

	require.NoError(t, buttonElement.Click("left", 1))
}

func (rs *RodSession) doFillPasswordAndClick(t *testing.T, page *rod.Page, password string) {
	element := rs.WaitElementLocatedByID(t, page, "password-textfield")
	button := rs.WaitElementLocatedByID(t, page, "sign-in-button")

	rs.doFillFieldUntilSet(t, element, password)

	require.NoError(t, button.Click("left", 1))
}

// Login 1FA.
func (rs *RodSession) doLoginOneFactor(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool, domain, targetURL string) {
	rs.doVisitLoginPage(t, page, domain, targetURL)
	rs.doFillLoginPageAndClick(t, page, username, password, keepMeLoggedIn)
}

func (rs *RodSession) doLoginPasskey(t *testing.T, page *rod.Page, keepMeLoggedIn bool, domain, targetURL string) {
	rs.doVisitLoginPage(t, page, domain, targetURL)

	passkeyElement := rs.WaitElementLocatedByID(t, page, "passkey-sign-in-button")

	require.NoError(t, passkeyElement.Click("left", 1))

	rs.doAnswerPasskeyRememberMe(t, page, keepMeLoggedIn)
}

func (rs *RodSession) doAnswerPasskeyRememberMe(t *testing.T, page *rod.Page, keepMeLoggedIn bool) {
	id := "dialog-remember-me-no"

	if keepMeLoggedIn {
		id = "dialog-remember-me-yes"
	}

	element := rs.WaitElementLocatedByID(t, page, id)

	require.NoError(t, element.Click("left", 1))
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
