package suites

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func (wds *WebDriverSession) doFillLoginPageAndClick(ctx context.Context, t *testing.T, username, password string, keepMeLoggedIn bool) {
	usernameElement := wds.WaitElementLocatedByID(ctx, t, "username-textfield")
	err := usernameElement.SendKeys(username)
	require.NoError(t, err)

	passwordElement := wds.WaitElementLocatedByID(ctx, t, "password-textfield")
	err = passwordElement.SendKeys(password)
	require.NoError(t, err)

	if keepMeLoggedIn {
		keepMeLoggedInElement := wds.WaitElementLocatedByID(ctx, t, "remember-checkbox")
		err = keepMeLoggedInElement.Click()
		require.NoError(t, err)
	}

	buttonElement := wds.WaitElementLocatedByID(ctx, t, "sign-in-button")
	err = buttonElement.Click()
	require.NoError(t, err)
}

// Login 1FA.
func (wds *WebDriverSession) doLoginOneFactor(ctx context.Context, t *testing.T, username, password string, keepMeLoggedIn bool, targetURL string) {
	wds.doVisitLoginPage(ctx, t, targetURL)
	wds.doFillLoginPageAndClick(ctx, t, username, password, keepMeLoggedIn)
}

// Login 1FA and 2FA subsequently (must already be registered).
func (wds *WebDriverSession) doLoginTwoFactor(ctx context.Context, t *testing.T, username, password string, keepMeLoggedIn bool, otpSecret, targetURL string) {
	wds.doLoginOneFactor(ctx, t, username, password, keepMeLoggedIn, targetURL)
	wds.verifyIsSecondFactorPage(ctx, t)
	wds.doValidateTOTP(ctx, t, otpSecret)
	// timeout when targetURL is not defined to prevent a show stopping redirect when visiting a protected domain
	if targetURL == "" {
		time.Sleep(1 * time.Second)
	}
}

// Login 1FA and register 2FA.
func (wds *WebDriverSession) doLoginAndRegisterTOTP(ctx context.Context, t *testing.T, username, password string, keepMeLoggedIn bool) string {
	wds.doLoginOneFactor(ctx, t, username, password, keepMeLoggedIn, "")
	secret := wds.doRegisterTOTP(ctx, t)
	wds.doVisit(t, LoginBaseURL)
	wds.verifyIsSecondFactorPage(ctx, t)
	return secret
}

// Register a user with TOTP, logout and then authenticate until TOTP-2FA.
func (wds *WebDriverSession) doRegisterAndLogin2FA(ctx context.Context, t *testing.T, username, password string, keepMeLoggedIn bool, targetURL string) string { //nolint:unparam
	// Register TOTP secret and logout.
	secret := wds.doRegisterThenLogout(ctx, t, username, password)
	wds.doLoginTwoFactor(ctx, t, username, password, keepMeLoggedIn, secret, targetURL)
	return secret
}
