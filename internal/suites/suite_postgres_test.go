package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PostgresSuite struct {
	*RodSuite
}

func NewPostgresSuite() *PostgresSuite {
	return &PostgresSuite{
		RodSuite: NewRodSuite(postgresSuiteName),
	}
}

func (s *PostgresSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *PostgresSuite) TestTwoFactorTOTPScenario() {
	suite.Run(s.T(), NewTwoFactorTOTPScenario())
}

func TestPostgresSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewPostgresSuite())
}
