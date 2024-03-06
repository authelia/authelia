package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type RedirectionCheckScenario struct {
	*RodSuite
}

func NewRedirectionCheckScenario() *RedirectionCheckScenario {
	return &RedirectionCheckScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *RedirectionCheckScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *RedirectionCheckScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *RedirectionCheckScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *RedirectionCheckScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

var redirectionAuthorizations = map[string]bool{
	// external website.
	"https://www.google.fr": false,
	// Not the right domain.
	"https://public.example.com.a:8080/secret.html": false,
	// Not https.
	"http://secure.example.com:8080/secret.html": false,
	// Domain handled by Authelia.
	"https://secure.example.com:8080/secret.html": true,
}

func (s *RedirectionCheckScenario) TestShouldRedirectOnLoginOnlyWhenDomainIsSafe() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginAndRegisterTOTPThenLogout(s.T(), s.Context(ctx), "john", "password")

	for url, redirected := range redirectionAuthorizations {
		s.T().Run(url, func(t *testing.T) {
			s.doLoginSecondFactorTOTP(t, s.Context(ctx), "john", "password", false, url)

			if redirected {
				s.verifySecretAuthorized(t, s.Context(ctx))
			} else {
				s.verifyIsAuthenticatedPage(t, s.Context(ctx))
			}

			s.doLogout(t, s.Context(ctx))
		})
	}
}

var logoutRedirectionURLs = map[string]bool{
	// external website.
	"https://www.google.fr": false,
	// Not the right domain.
	"https://public.example-not-right.com:8080/index.html": false,
	// Not https.
	"http://public.example.com:8080/index.html": false,
	// Domain handled by Authelia.
	"https://public.example.com:8080/index.html": true,
}

func (s *RedirectionCheckScenario) TestShouldRedirectOnLogoutOnlyWhenDomainIsSafe() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	for url, success := range logoutRedirectionURLs {
		s.T().Run(url, func(t *testing.T) {
			s.doLogoutWithRedirect(t, s.Context(ctx), url, !success)
		})
	}
}

func TestRedirectionCheckScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewRedirectionCheckScenario())
}
