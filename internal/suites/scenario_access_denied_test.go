// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"log"
	"time"
)

type AccessDeniedScenario struct {
	*RodSuite
}

func NewAccessDeniedScenario() *AccessDeniedScenario {
	return &AccessDeniedScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *AccessDeniedScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *AccessDeniedScenario) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *AccessDeniedScenario) SetupTest() {
	s.doSetupTest(HomeBaseURL)
}

func (s *AccessDeniedScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *AccessDeniedScenario) TestShouldShowAccessDeniedPage() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", DenyBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, targetURL)
	s.verifyIsDeny(s.T(), s.Context(ctx))
}

func (s *AccessDeniedScenario) TestShouldSwitchUserFromAccessDeniedPage() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", DenyBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, targetURL)
	s.verifyIsDeny(s.T(), s.Context(ctx))

	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "switch-user-button").MustClick()
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
}
