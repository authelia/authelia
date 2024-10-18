package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SecuritySettingsSuite struct {
	*BaseSuite
}

func NewSecuritySettingsSuite() *SecuritySettingsSuite {
	return &SecuritySettingsSuite{
		BaseSuite: &BaseSuite{
			Name: securitySettingsSuiteName,
		},
	}
}

func (s *SecuritySettingsSuite) TestChangePasswordScenario() {
	suite.Run(s.T(), NewChangePasswordScenario())
}

func TestSecuritySettingsSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewSecuritySettingsSuite())
}
