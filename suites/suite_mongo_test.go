package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MongoSuite struct {
	*SeleniumSuite
}

func NewMongoSuite() *MongoSuite {
	return &MongoSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestMongoSuite(t *testing.T) {
	suite.Run(t, NewOneFactorSuite())
	suite.Run(t, NewTwoFactorSuite())
}
