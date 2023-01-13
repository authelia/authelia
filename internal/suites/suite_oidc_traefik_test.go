package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type OIDCTraefikSuite struct {
	*RodSuite
}

func NewOIDCTraefikSuite() *OIDCTraefikSuite {
	return &OIDCTraefikSuite{
		RodSuite: &RodSuite{
			Name: oidcTraefikSuiteName,
		},
	}
}

func (s *OIDCTraefikSuite) TestOIDCScenario() {
	suite.Run(s.T(), NewOIDCScenario())
}

func (s *OIDCTraefikSuite) SetupSuite() {
	s.LoadEnvironment()
}

func TestOIDCTraefikSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOIDCTraefikSuite())
}
