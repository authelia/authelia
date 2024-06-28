package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type Traefik3Suite struct {
	*RodSuite
}

func NewTraefik3Suite() *Traefik3Suite {
	return &Traefik3Suite{
		RodSuite: NewRodSuite(traefik3SuiteName),
	}
}

func (s *Traefik3Suite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *Traefik3Suite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *Traefik3Suite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *Traefik3Suite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *Traefik3Suite) TestShouldKeepSessionAfterRedisRestart() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer func() {
		cancel()
		s.collectCoverage(s.Page)
		s.collectScreenshot(ctx.Err(), s.Page)
		s.MustClose()
		err := s.RodSession.Stop()
		s.Require().NoError(err)
	}()

	browser, err := NewRodSession(RodSessionWithCredentials(s))
	s.Require().NoError(err)
	s.RodSession = browser

	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
	s.doLoginAndRegisterTOTPThenLogout(s.T(), s.Context(ctx), "john", "password")

	s.doLoginSecondFactorTOTP(s.T(), s.Context(ctx), "john", "password", false, "")

	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s/secret.html", SecureBaseURL))
	s.verifySecretAuthorized(s.T(), s.Context(ctx))

	err = traefik2DockerEnvironment.Restart("redis")
	s.Require().NoError(err)

	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s/secret.html", SecureBaseURL))
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func TestTraefik3Suite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTraefik3Suite())
}
