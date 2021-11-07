package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MySQLSuite struct {
	*RodSuite
}

func NewMySQLSuite() *MySQLSuite {
	return &MySQLSuite{RodSuite: new(RodSuite)}
}

func (s *MySQLSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *MySQLSuite) TestTwoFactorScenario() {
	suite.Run(s.T(), NewTwoFactorScenario())
}

func TestMySQLSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewMySQLSuite())
}
