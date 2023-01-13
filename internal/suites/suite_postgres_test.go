package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PostgresSuite struct {
	*RodSuite
}

func NewPostgresSuite() *PostgresSuite {
	return &PostgresSuite{
		RodSuite: NewRodSuite(postgresSuiteName),
	}
}

func (s *PostgresSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *PostgresSuite) Test2FAScenario() {
	suite.Run(s.T(), New2FAScenario())
}

func TestPostgresSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewPostgresSuite())
}
