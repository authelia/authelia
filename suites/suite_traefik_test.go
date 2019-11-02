package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TraefikSuite struct {
	*SeleniumSuite
}

func NewTraefikSuite() *TraefikSuite {
	return &TraefikSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestTraefikSuite(t *testing.T) {
	suite.Run(t, NewOneFactorSuite())
	suite.Run(t, NewTwoFactorSuite())
}
