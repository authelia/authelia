// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AppAuthenticatorSuite struct {
	*BaseSuite
}

func NewAppAuthenticatorSuite() *AppAuthenticatorSuite {
	return &AppAuthenticatorSuite{
		BaseSuite: &BaseSuite{
			Name: appAuthenticatorSuiteName,
		},
	}
}

func (s *AppAuthenticatorSuite) TestAppAuthenticatorScenario() {
	suite.Run(s.T(), NewAppAuthenticatorScenario(2))
}

func (s *AppAuthenticatorSuite) TestAvailableMethodsScenario() {
	suite.Run(s.T(), NewAvailableMethodsScenario([]string{
		"TIME-BASED ONE-TIME PASSWORD",
		"SECURITY KEY - WEBAUTHN",
		"AUTHELIA APP",
	}))
}

func TestAppAuthenticatorSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewAppAuthenticatorSuite())
}
