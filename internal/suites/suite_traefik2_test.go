package suites

import (
	"testing"

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

func TestTraefik2Suite(t *testing.T) {
	suite.Run(t, NewTraefik2Suite())
}
