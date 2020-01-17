package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type InactivityScenario struct {
	*SeleniumSuite
	secret string
}

func NewInactivityScenario() *InactivityScenario {
	return &InactivityScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *InactivityScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.secret = s.doRegisterAndLogin2FA(ctx, s.T(), "john", "password", false, targetURL)
	s.verifySecretAuthorized(ctx, s.T())
}

func (s *InactivityScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *InactivityScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *InactivityScenario) TestShouldRequireReauthenticationAfterInactivityPeriod() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginTwoFactor(ctx, s.T(), "john", "password", false, s.secret, "")

	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())

	time.Sleep(6 * time.Second)

	s.doVisit(s.T(), targetURL)
	s.verifyIsFirstFactorPage(ctx, s.T())
}

func (s *InactivityScenario) TestShouldRequireReauthenticationAfterCookieExpiration() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginTwoFactor(ctx, s.T(), "john", "password", false, s.secret, "")

	for i := 0; i < 3; i++ {
		s.doVisit(s.T(), HomeBaseURL)
		s.verifyIsHome(ctx, s.T())

		time.Sleep(2 * time.Second)
		s.doVisit(s.T(), targetURL)
		s.verifySecretAuthorized(ctx, s.T())
	}

	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())

	time.Sleep(2 * time.Second)

	s.doVisit(s.T(), targetURL)
	s.verifyIsFirstFactorPage(ctx, s.T())
}

func (s *InactivityScenario) TestShouldDisableCookieExpirationAndInactivity() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginTwoFactor(ctx, s.T(), "john", "password", true, s.secret, "")

	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())

	time.Sleep(10 * time.Second)

	s.doVisit(s.T(), targetURL)
	s.verifySecretAuthorized(ctx, s.T())
}

func TestInactivityScenario(t *testing.T) {
	suite.Run(t, NewInactivityScenario())
}
