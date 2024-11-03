package suites

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"time"
)

type TraefikSuite struct {
	*RodSuite
}

func NewTraefikSuite(name string) *TraefikSuite {
	return &TraefikSuite{
		RodSuite: NewRodSuite(name),
	}
}

func (s *TraefikSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *TraefikSuite) TestTwoFactorTOTPScenario() {
	suite.Run(s.T(), NewTwoFactorTOTPScenario())
}

func (s *TraefikSuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *TraefikSuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *TraefikSuite) TestShouldKeepSessionAfterRedisRestart() {
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

	err = traefik3DockerEnvironment.Restart("redis")
	s.Require().NoError(err)

	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s/secret.html", SecureBaseURL))
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}
