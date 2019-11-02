package suites

import (
	"context"

	"github.com/stretchr/testify/assert"
)

func doFillLoginPageAndClick(ctx context.Context, s *SeleniumSuite, username, password string, keepMeLoggedIn bool) {
	usernameElement := WaitElementLocatedByID(ctx, s, "username")
	err := usernameElement.SendKeys(username)
	assert.NoError(s.T(), err)

	passwordElement := WaitElementLocatedByID(ctx, s, "password")
	err = passwordElement.SendKeys(password)
	assert.NoError(s.T(), err)

	if keepMeLoggedIn {
		keepMeLoggedInElement := WaitElementLocatedByID(ctx, s, "remember-checkbox")
		err = keepMeLoggedInElement.Click()
		assert.NoError(s.T(), err)
	}

	buttonElement := WaitElementLocatedByTagName(ctx, s, "button")
	err = buttonElement.Click()
	assert.NoError(s.T(), err)
}

func doLoginOneFactor(ctx context.Context, s *SeleniumSuite, username, password string, keepMeLoggedIn bool, targetURL string) {
	doVisitLoginPage(ctx, s, targetURL)
	doFillLoginPageAndClick(ctx, s, username, password, keepMeLoggedIn)
}

func doLoginTwoFactor(ctx context.Context, s *SeleniumSuite, username, password string, keepMeLoggedIn bool, otpSecret, targetURL string) {
	doLoginOneFactor(ctx, s, username, password, keepMeLoggedIn, targetURL)
	verifyIsSecondFactorPage(ctx, s)
	doValidateTOTP(ctx, s, otpSecret)
}

func doLoginAndRegisterTOTP(ctx context.Context, s *SeleniumSuite, username, password string, keepMeLoggedIn bool) string {
	doLoginOneFactor(ctx, s, username, password, keepMeLoggedIn, "")
	secret := doRegisterTOTP(ctx, s)
	s.Assert().NotNil(secret)
	doVisit(s, LoginBaseURL)
	verifyIsSecondFactorPage(ctx, s)
	return secret
}
