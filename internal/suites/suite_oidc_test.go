// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type OIDCSuite struct {
	*RodSuite
}

func NewOIDCSuite() *OIDCSuite {
	return &OIDCSuite{
		RodSuite: NewRodSuite(oidcSuiteName),
	}
}

func (s *OIDCSuite) TestOIDCScenario() {
	suite.Run(s.T(), NewOIDCScenario())
}

func TestOIDCSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOIDCSuite())
}
