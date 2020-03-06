package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MySQLSuite struct {
	*SeleniumSuite
}

func NewMySQLSuite() *MySQLSuite {
	return &MySQLSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *MySQLSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *MySQLSuite) TestTwoFactorScenario() {
	suite.Run(s.T(), NewTwoFactorScenario())
}

func TestMySQLSuite(t *testing.T) {
	suite.Run(t, NewMySQLSuite())
}
