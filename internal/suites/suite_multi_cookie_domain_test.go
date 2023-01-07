package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func NewMultiCookieDomainSuite() *MultiCookieDomainSuite {
	return &MultiCookieDomainSuite{}
}

type MultiCookieDomainSuite struct {
	suite.Suite
}

func (s *MultiCookieDomainSuite) TestMultiCookieDomainFirstDomainScenario() {
	suite.Run(s.T(), NewMultiCookieDomainScenario(BaseDomain, Example2Com, true))
}

func (s *MultiCookieDomainSuite) TestMultiCookieDomainSecondDomainScenario() {
	suite.Run(s.T(), NewMultiCookieDomainScenario(Example2Com, BaseDomain, false))
}

func TestMultiCookieDomainSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewMultiCookieDomainSuite())
}
