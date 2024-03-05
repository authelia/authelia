package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InactivityScenario struct {
	*RodSuite
}

func NewInactivityScenario() *InactivityScenario {
	return &InactivityScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *InactivityScenario) SetupSuite() {
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
	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doRegisterTOTPAndLogin2FA(s.T(), s.Context(ctx), "john", "password", false, targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func (s *InactivityScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *InactivityScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *InactivityScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *InactivityScenario) TestShouldRequireReauthenticationAfterInactivityPeriod() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)

	s.doLoginSecondFactorTOTP(s.T(), s.Context(ctx), "john", "password", false, "")
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))

	time.Sleep(6 * time.Second)

	s.doVisit(s.T(), s.Context(ctx), targetURL)
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
}

func (s *InactivityScenario) TestShouldRequireReauthenticationAfterCookieExpiration() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)

	s.doLoginSecondFactorTOTP(s.T(), s.Context(ctx), "john", "password", false, "")

	for i := 0; i < 3; i++ {
		s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
		s.verifyIsHome(s.T(), s.Context(ctx))

		time.Sleep(2 * time.Second)

		s.doVisit(s.T(), s.Context(ctx), targetURL)
		s.verifySecretAuthorized(s.T(), s.Context(ctx))
	}

	time.Sleep(2 * time.Second)

	require.NoError(s.T(), s.Context(ctx).Reload())
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
}

func (s *InactivityScenario) TestShouldDisableCookieExpirationAndInactivity() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)

	s.doLoginSecondFactorTOTP(s.T(), s.Context(ctx), "john", "password", true, "")
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))

	time.Sleep(10 * time.Second)

	s.doVisit(s.T(), s.Context(ctx), targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func TestInactivityScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewInactivityScenario())
}
