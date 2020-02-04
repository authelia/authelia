package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type RedirectionCheckScenario struct {
	*SeleniumSuite
}

func NewRedirectionCheckScenario() *RedirectionCheckScenario {
	return &RedirectionCheckScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *RedirectionCheckScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *RedirectionCheckScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *RedirectionCheckScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

var redirectionAuthorizations = map[string]bool{
	// external website
	"https://www.google.fr": false,
	// Not the right domain
	"https://public.example.com.a:8080/secret.html": false,
	// Not https
	"http://secure.example.com:8080/secret.html": false,
	// Domain handled by Authelia
	"https://secure.example.com:8080/secret.html": true,
}

func (s *RedirectionCheckScenario) TestShouldRedirectOnlyWhenDomainIsHandledByAuthelia() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secret := s.doRegisterThenLogout(ctx, s.T(), "john", "password")

	for url, redirected := range redirectionAuthorizations {
		s.T().Run(url, func(t *testing.T) {
			s.doLoginTwoFactor(ctx, t, "john", "password", false, secret, url)
			time.Sleep(1 * time.Second)
			if redirected {
				s.verifySecretAuthorized(ctx, t)
			} else {
				s.verifyIsAuthenticatedPage(ctx, t)
			}
			s.doLogout(ctx, t)
		})
	}
}

func TestRedirectionCheckScenario(t *testing.T) {
	suite.Run(t, NewRedirectionCheckScenario())
}
