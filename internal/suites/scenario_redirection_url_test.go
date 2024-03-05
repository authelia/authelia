package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type RedirectionURLScenario struct {
	*RodSuite
}

func NewRedirectionURLScenario() *RedirectionURLScenario {
	return &RedirectionURLScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *RedirectionURLScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *RedirectionURLScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *RedirectionURLScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *RedirectionURLScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *RedirectionURLScenario) TestShouldVerifyCustomURLParametersArePropagatedAfterRedirection() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html?myparam=test", SingleFactorBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
	s.verifyURLIs(s.T(), s.Context(ctx), targetURL)
}

func TestRedirectionURLScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewRedirectionURLScenario())
}
