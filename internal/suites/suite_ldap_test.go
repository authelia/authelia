package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LDAPSuite struct {
	*SeleniumSuite
}

func NewLDAPSuite() *LDAPSuite {
	return &LDAPSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestLDAPSuite(t *testing.T) {
	suite.Run(t, NewOneFactorScenario())
	suite.Run(t, NewTwoFactorScenario())
}
