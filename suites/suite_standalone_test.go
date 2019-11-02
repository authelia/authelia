package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type StandaloneSuite struct {
	*SeleniumSuite
}

func NewStandaloneSuite() *StandaloneSuite {
	return &StandaloneSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestStandaloneSuite(t *testing.T) {
	suite.Run(t, NewOneFactorSuite())
	suite.Run(t, NewTwoFactorSuite())
}
