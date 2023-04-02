// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MySQLSuite struct {
	*RodSuite
}

func NewMySQLSuite() *MySQLSuite {
	return &MySQLSuite{
		RodSuite: NewRodSuite(mysqlSuiteName),
	}
}

func (s *MySQLSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *MySQLSuite) Test2FAScenario() {
	suite.Run(s.T(), New2FAScenario())
}

func TestMySQLSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewMySQLSuite())
}
