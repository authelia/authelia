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

func (s *MultiCookieDomainSuite) TestMultiCookieDomainScenario() {
	suite.Run(s.T(), NewMultiCookieDomainScenario())
}

func TestMultiCookieDomainSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewMultiCookieDomainSuite())
}
