package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MariadbSuite struct {
	*SeleniumSuite
}

func NewMariadbSuite() *MariadbSuite {
	return &MariadbSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *MariadbSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *MariadbSuite) TestTwoFactorScenario() {
	suite.Run(s.T(), NewTwoFactorScenario())
}

func TestMariadbSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewMariadbSuite())
}
