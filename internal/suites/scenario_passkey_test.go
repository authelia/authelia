package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PasskeyScenario struct {
	*RodSuite
}

func NewPasskeyScenario() *PasskeyScenario {
	return &PasskeyScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *PasskeyScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)

		s.collectCoverage(s.Page)
		s.MustClose()
	}()

	s.Page = s.doCreateTab(s.T(), HomeBaseURL)

	s.doWebAuthnInitialize(s.T(), s.Page, false)

	s.doLoginAndRegisterWebAuthn(s.T(), s.Context(ctx), "john", "password", false)
	s.doLogout(s.T(), s.Page)
}

func (s *PasskeyScenario) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *PasskeyScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)

	s.doWebAuthnInitialize(s.T(), s.Page, false)
	s.doWebAuthnRestoreCredentials(s.T(), s.Page)
}

func (s *PasskeyScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *PasskeyScenario) TestShouldAuthorizeAfterPasskeyLogin() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginPasskey(s.T(), s.Context(ctx), false, BaseDomain, targetURL)
	s.verifyIsSecondFactorPasswordPage(s.T(), s.Context(ctx))
	s.doFillPasswordAndClick(s.T(), s.Context(ctx), "bad-password")

	s.verifyIsSecondFactorPasswordPage(s.T(), s.Context(ctx))
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Incorrect password")
	s.doFillPasswordAndClick(s.T(), s.Context(ctx), "password")

	// And check if the user is redirected to the secret.
	s.verifySecretAuthorized(s.T(), s.Context(ctx))

	// Leave the secret.
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))

	// And try to reload it again to check the session is kept.
	s.doVisit(s.T(), s.Context(ctx), targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func TestRunPasskey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTwoFactorWebAuthnScenario())
}
