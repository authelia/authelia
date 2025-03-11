package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MySQLSuite struct {
	*RodSuite
}

func NewMySQLSuite() *MySQLSuite {
	return &MySQLSuite{
		RodSuite: NewRodSuite(mysqlSuiteName),
	}
}

func (s *MySQLSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *MySQLSuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func TestMySQLSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewMySQLSuite())
}
