package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type OpenIDConnectRelyingPartySuite struct {
	*RodSuite
}

func NewOpenIDConnectRelyingPartySuite() *OpenIDConnectRelyingPartySuite {
	return &OpenIDConnectRelyingPartySuite{
		RodSuite: NewRodSuite(openIDConnectRelyingPartySuiteName),
	}
}

func (s *OpenIDConnectRelyingPartySuite) TestOpenIDConnectRelyingPartyScenario() {
	suite.Run(s.T(), NewOpenIDConnectRelyingPartyScenario())
}

func TestOpenIDConnectRelyingPartySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOpenIDConnectRelyingPartySuite())
}
