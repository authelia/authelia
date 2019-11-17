package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PostgresSuite struct {
	*SeleniumSuite
}

func NewPostgresSuite() *PostgresSuite {
	return &PostgresSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, NewOneFactorSuite())
	suite.Run(t, NewTwoFactorSuite())
}
