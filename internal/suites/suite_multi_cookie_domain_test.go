// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func NewMultiCookieDomainSuite() *MultiCookieDomainSuite {
	return &MultiCookieDomainSuite{
		BaseSuite: &BaseSuite{
			Name: multiCookieDomainSuiteName,
		},
	}
}

type MultiCookieDomainSuite struct {
	*BaseSuite
}

func (s *MultiCookieDomainSuite) TestMultiCookieDomainFirstDomainScenario() {
	suite.Run(s.T(), NewMultiCookieDomainScenario(BaseDomain, Example2DotCom, []string{"__Host-Http-authelia_session", "__Secure-authelia_session_authz", "__Host-authelia_session_csrf", "language"}, true))
}

func (s *MultiCookieDomainSuite) TestMultiCookieDomainSecondDomainScenario() {
	suite.Run(s.T(), NewMultiCookieDomainScenario(Example2DotCom, BaseDomain, []string{"__Host-Http-example2_session", "__Secure-example2_session_authz", "__Host-example2_session_csrf", "language"}, false))
}

func (s *MultiCookieDomainSuite) TestMultiCookieDomainThirdDomainScenario() {
	suite.Run(s.T(), NewMultiCookieDomainScenario(Example3DotCom, BaseDomain, []string{"__Host-Http-authelia_session", "__Secure-authelia_session_authz", "__Host-authelia_session_csrf", "language"}, true))
}

func TestMultiCookieDomainSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewMultiCookieDomainSuite())
}
