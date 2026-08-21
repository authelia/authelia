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
	rs.doFillFieldUntilSet(t, usernameElement, username)
	rs.doFillFieldUntilSet(t, passwordElement, password)

	if keepMeLoggedIn {
		rs.ClickElementLocatedByID(t, page, "remember-checkbox")
	}

	rs.ClickElementLocatedByID(t, page, "sign-in-button")
}

func (rs *RodSession) doFillPasswordAndClick(t *testing.T, page *rod.Page, password string) {
	element := rs.WaitElementLocatedByID(t, page, "password-textfield")

	rs.doFillFieldUntilSet(t, element, password)

	rs.ClickElementLocatedByID(t, page, "sign-in-button")
}

func (rs *RodSession) doLoginOneFactor(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool, domain, targetURL string) {
	rs.doVisitLoginPage(t, page, domain, targetURL)
	rs.doFillLoginPageAndClick(t, page, username, password, keepMeLoggedIn)
}

func (rs *RodSession) doLoginPasskey(t *testing.T, page *rod.Page, keepMeLoggedIn bool, domain, targetURL string) {
	rs.doVisitLoginPage(t, page, domain, targetURL)

	if keepMeLoggedIn {
		rs.ClickElementLocatedByID(t, page, "remember-checkbox")
	}

	rs.ClickElementLocatedByID(t, page, "passkey-sign-in-button")
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

func (rs *RodSession) doLoginAndRegisterTOTP(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool) {
	rs.doLoginOneFactor(t, page, username, password, keepMeLoggedIn, BaseDomain, "")
	rs.doOpenSettingsAndRegisterTOTP(t, page, username)

	rs.verifyIsSecondFactorPage(t, page)
}

func (rs *RodSession) doRegisterTOTPAndLogin2FA(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool, targetURL string) { //nolint:unparam
	rs.doLoginAndRegisterTOTPThenLogout(t, page, username, password)
	rs.doLoginSecondFactorTOTP(t, page, username, password, keepMeLoggedIn, targetURL)
}

func (rs *RodSession) doLoginAndRegisterWebAuthn(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool) {
	rs.doLoginOneFactor(t, page, username, password, keepMeLoggedIn, BaseDomain, "")
	require.Greater(t, len(rs.GetWebAuthnAuthenticatorID()), 0)
	rs.doWebAuthnCredentialRegisterAfterVisitSettings(t, page, "testing")

	rs.verifyIsSecondFactorPage(t, page)
}
