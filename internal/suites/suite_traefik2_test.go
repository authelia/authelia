package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type Traefik2Suite struct {
	*SeleniumSuite
}

func NewTraefik2Suite() *Traefik2Suite {
	return &Traefik2Suite{SeleniumSuite: new(SeleniumSuite)}
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wds, err := StartWebDriver()
	s.Require().NoError(err)

	defer func() {
		err = wds.Stop()
		s.Require().NoError(err)
	}()

	secret := wds.doRegisterThenLogout(ctx, s.T(), "john", "password")

	wds.doLoginTwoFactor(ctx, s.T(), "john", "password", false, secret, "")

	wds.doVisit(s.T(), fmt.Sprintf("%s/secret.html", SecureBaseURL))
	wds.verifySecretAuthorized(ctx, s.T())

	err = traefik2DockerEnvironment.Restart("redis")
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)

	wds.doVisit(s.T(), fmt.Sprintf("%s/secret.html", SecureBaseURL))
	wds.verifySecretAuthorized(ctx, s.T())
}

func TestTraefik2Suite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTraefik2Suite())
}
