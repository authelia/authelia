package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HAProxySuite struct {
	*RodSuite
}

func NewHAProxySuite() *HAProxySuite {
	return &HAProxySuite{
		RodSuite: NewRodSuite(haproxySuiteName),
	}
}

func (s *HAProxySuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *HAProxySuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *HAProxySuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func TestHAProxySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewHAProxySuite())
}
