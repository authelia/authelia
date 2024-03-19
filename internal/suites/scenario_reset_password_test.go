package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ResetPasswordScenario struct {
	*RodSuite
}

func NewResetPasswordScenario() *ResetPasswordScenario {
	return &ResetPasswordScenario{RodSuite: NewRodSuite("")}
}

func (s *ResetPasswordScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *ResetPasswordScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *ResetPasswordScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *ResetPasswordScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *ResetPasswordScenario) TestShouldResetPassword() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURLWithFallbackPrefix(BaseDomain, "/"))
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

	// Reset the password to abc.
	s.doResetPassword(s.T(), s.Context(ctx), "john", "abc", "abc", false)

	// Try to login with the old password.
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Incorrect username or password")

	// Try to login with the new password.
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "abc", false, BaseDomain, "")

	// Logout.
	s.doLogout(s.T(), s.Context(ctx))

	// Reset the original password.
	s.doResetPassword(s.T(), s.Context(ctx), "john", "password", "password", false)
}

func (s *ResetPasswordScenario) TestShouldMakeAttackerThinkPasswordResetIsInitiated() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURLWithFallbackPrefix(BaseDomain, "/"))
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

	// Try to initiate a password reset of an nonexistent user.
	s.doInitiatePasswordReset(s.T(), s.Context(ctx), "i_dont_exist")

	// Check that the notification make the attacker thinks the process is initiated.
	s.verifyMailNotificationDisplayed(s.T(), s.Context(ctx))
}

func (s *ResetPasswordScenario) TestShouldLetUserNoticeThereIsAPasswordMismatch() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURLWithFallbackPrefix(BaseDomain, "/"))
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

	s.doInitiatePasswordReset(s.T(), s.Context(ctx), "john")
	s.verifyMailNotificationDisplayed(s.T(), s.Context(ctx))

	s.doCompletePasswordReset(s.T(), s.Context(ctx), "password", "another_password")
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Passwords do not match")
}

func TestRunResetPasswordScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewResetPasswordScenario())
}
