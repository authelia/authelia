package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MariadbSuite struct {
	*SeleniumSuite
}

func NewMariadbSuite() *MariadbSuite {
	return &MariadbSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestMariadbSuite(t *testing.T) {
	suite.Run(t, NewOneFactorSuite())
	suite.Run(t, NewTwoFactorSuite())
}
