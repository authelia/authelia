// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DefaultRedirectionURLScenario struct {
	*RodSuite

	secret string
}

func NewDefaultRedirectionURLScenario() *DefaultRedirectionURLScenario {
	return &DefaultRedirectionURLScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *DefaultRedirectionURLScenario) SetupSuite() {
	browser, err := StartRod()

	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)

		s.collectCoverage(s.Page)
		s.MustClose()
	}()

	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.secret = s.doLoginAndRegisterTOTP(s.T(), s.Context(ctx), "john", "password", false)
}

func (s *DefaultRedirectionURLScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *DefaultRedirectionURLScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *DefaultRedirectionURLScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *DefaultRedirectionURLScenario) TestUserIsRedirectedToDefaultURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)

	s.doLoginTwoFactor(s.T(), s.Context(ctx), "john", "password", false, s.secret, targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
	s.doLogout(s.T(), s.Context(ctx))

	s.doLoginTwoFactor(s.T(), s.Context(ctx), "john", "password", false, s.secret, "")
	s.verifyIsHome(s.T(), s.Page)
}

func TestShouldRunDefaultRedirectionURLScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewDefaultRedirectionURLScenario())
}
