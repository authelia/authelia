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
	return &Traefik2Suite{
		RodSuite: NewRodSuite(traefik2SuiteName),
	}
}

func (s *Traefik2Suite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *Traefik2Suite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *Traefik2Suite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *Traefik2Suite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
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

func TestTraefik2Suite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTraefik2Suite())
}
