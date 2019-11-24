package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ShortTimeoutsSuite struct {
	*SeleniumSuite
}

func NewShortTimeoutsSuite() *ShortTimeoutsSuite {
	return &ShortTimeoutsSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestShortTimeoutsSuite(t *testing.T) {
	suite.Run(t, NewInactivityScenario())
	suite.Run(t, NewRegulationScenario())
}
