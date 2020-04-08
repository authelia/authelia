package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ResetPasswordScenario struct {
	*SeleniumSuite
}

func NewResetPasswordScenario() *ResetPasswordScenario {
	return &ResetPasswordScenario{SeleniumSuite: new(SeleniumSuite)}
}

func (s *ResetPasswordScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *ResetPasswordScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *ResetPasswordScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *ResetPasswordScenario) TestShouldResetPassword() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s.doVisit(s.T(), LoginBaseURL)
	s.verifyIsFirstFactorPage(ctx, s.T())

	// Reset the password to abc
	s.doResetPassword(ctx, s.T(), "john", "abc", "abc")

	// Try to login with the old password
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyNotificationDisplayed(ctx, s.T(), "Incorrect username or password.")

	// Try to login with the new password
	s.doLoginOneFactor(ctx, s.T(), "john", "abc", false, "")

	// Logout
	s.doLogout(ctx, s.T())

	// Reset the original password
	s.doResetPassword(ctx, s.T(), "john", "password", "password")
}

func (s *ResetPasswordScenario) TestShouldMakeAttackerThinkPasswordResetIsInitiated() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s.doVisit(s.T(), LoginBaseURL)
	s.verifyIsFirstFactorPage(ctx, s.T())

	// Try to initiate a password reset of an nonexistent user.
	s.doInitiatePasswordReset(ctx, s.T(), "i_dont_exist")

	// Check that the notification make the attacker thinks the process is initiated
	s.verifyMailNotificationDisplayed(ctx, s.T())
}

func (s *ResetPasswordScenario) TestShouldLetUserNoticeThereIsAPasswordMismatch() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s.doVisit(s.T(), LoginBaseURL)
	s.verifyIsFirstFactorPage(ctx, s.T())

	s.doInitiatePasswordReset(ctx, s.T(), "john")
	s.verifyMailNotificationDisplayed(ctx, s.T())

	s.doCompletePasswordReset(ctx, s.T(), "password", "another_password")
	s.verifyNotificationDisplayed(ctx, s.T(), "Passwords do not match.")
}

func TestRunResetPasswordScenario(t *testing.T) {
	suite.Run(t, NewResetPasswordScenario())
}
