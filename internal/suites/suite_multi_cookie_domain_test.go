package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func NewMultiCookieDomainSuite() *MultiCookieDomainSuite {
	return &MultiCookieDomainSuite{
		BaseSuite: &BaseSuite{
			Name: multiCookieDomainSuiteName,
		},
	}
}

type MultiCookieDomainSuite struct {
	*BaseSuite
}

func (s *MultiCookieDomainSuite) TestMultiCookieDomainFirstDomainScenario() {
	suite.Run(s.T(), NewMultiCookieDomainScenario(BaseDomain, Example2DotCom, []string{"authelia_session", "language"}, true))
}

func (s *MultiCookieDomainSuite) TestMultiCookieDomainSecondDomainScenario() {
	suite.Run(s.T(), NewMultiCookieDomainScenario(Example2DotCom, BaseDomain, []string{"example2_session", "language"}, false))
}

func (s *MultiCookieDomainSuite) TestMultiCookieDomainThirdDomainScenario() {
	suite.Run(s.T(), NewMultiCookieDomainScenario(Example3DotCom, BaseDomain, []string{"authelia_session", "language"}, true))
}

func TestMultiCookieDomainSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewMultiCookieDomainSuite())
}
