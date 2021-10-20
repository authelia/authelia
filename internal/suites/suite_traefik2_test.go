package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type Traefik2Suite struct {
	*RodSuite
}

func NewTraefik2Suite() *Traefik2Suite {
	return &Traefik2Suite{RodSuite: new(RodSuite)}
}

func (s *Traefik2Suite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *Traefik2Suite) TestTwoFactorScenario() {
	suite.Run(s.T(), NewTwoFactorScenario())
}

func (s *Traefik2Suite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *Traefik2Suite) TestShouldKeepSessionAfterRedisRestart() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer func() {
		cancel()
		s.collectCoverage(s.Page)
		s.collectScreenshot(ctx.Err(), s.Page)
		s.MustClose()
		err := s.RodSession.Stop()
		s.Require().NoError(err)
	}()

	browser, err := StartRod()
	s.Require().NoError(err)
	s.RodSession = browser

	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
	secret := s.doRegisterThenLogout(s.T(), s.Context(ctx), "john", "password")

	s.doLoginTwoFactor(s.T(), s.Context(ctx), "john", "password", false, secret, "")

	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s/secret.html", SecureBaseURL))
	s.verifySecretAuthorized(s.T(), s.Context(ctx))

	err = traefik2DockerEnvironment.Restart("redis")
	s.Require().NoError(err)

	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s/secret.html", SecureBaseURL))
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func TestTraefik2Suite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTraefik2Suite())
}
