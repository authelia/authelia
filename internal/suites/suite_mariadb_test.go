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

func (s *MariaDBSuite) TestTwoFactorTOTPScenario() {
	suite.Run(s.T(), NewTwoFactorTOTPScenario())
}

func TestMariaDBSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewMariaDBSuite())
}
