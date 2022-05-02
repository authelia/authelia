package suites

import (
	"testing"

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

func (s *Traefik2Suite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func TestTraefik2Suite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTraefik2Suite())
}
