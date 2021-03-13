package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PostgresSuite struct {
	*SeleniumSuite
}

func NewPostgresSuite() *PostgresSuite {
	return &PostgresSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *PostgresSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *PostgresSuite) TestTwoFactorScenario() {
	suite.Run(s.T(), NewTwoFactorScenario())
}

func TestPostgresSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewPostgresSuite())
}
