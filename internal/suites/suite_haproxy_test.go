package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HAProxySuite struct {
	*SeleniumSuite
}

func NewHAProxySuite() *HAProxySuite {
	return &HAProxySuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *HAProxySuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *HAProxySuite) TestTwoFactorScenario() {
	suite.Run(s.T(), NewTwoFactorScenario())
}

func (s *HAProxySuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func TestHAProxySuite(t *testing.T) {
	suite.Run(t, NewHAProxySuite())
}
