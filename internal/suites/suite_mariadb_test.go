package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MariaDBSuite struct {
	*RodSuite
}

func NewMariaDBSuite() *MariaDBSuite {
	return &MariaDBSuite{
		RodSuite: NewRodSuite(mariadbSuiteName),
	}
}

func (s *MariaDBSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *MariaDBSuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func TestMariaDBSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewMariaDBSuite())
}
