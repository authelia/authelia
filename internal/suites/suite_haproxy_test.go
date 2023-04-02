// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HAProxySuite struct {
	*RodSuite
}

func NewHAProxySuite() *HAProxySuite {
	return &HAProxySuite{
		RodSuite: NewRodSuite(haproxySuiteName),
	}
}

func (s *HAProxySuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *HAProxySuite) Test2FAScenario() {
	suite.Run(s.T(), New2FAScenario())
}

func (s *HAProxySuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func TestHAProxySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewHAProxySuite())
}
