// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TwoFactorSuite struct {
	*RodSuite

	secret string
}

func New2FAScenario() *TwoFactorSuite {
	return &TwoFactorSuite{
		RodSuite: NewRodSuite(""),
	}
}

func (s *TwoFactorSuite) SetupSuite() {
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

func (s *TwoFactorSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *TwoFactorSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *TwoFactorSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *TwoFactorSuite) TestShouldNotAuthorizeSecretBeforeTwoFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)

	s.doVisit(s.T(), s.Context(ctx), targetURL)

	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

	raw := GetLoginBaseURLWithFallbackPrefix(BaseDomain, "/")

	expected, err := url.ParseRequestURI(raw)
	s.Assert().NoError(err)
	s.Require().NotNil(expected)

	query := expected.Query()

	query.Set("rd", targetURL)

	expected.RawQuery = query.Encode()

	rx := regexp.MustCompile(fmt.Sprintf(`^%s(&rm=GET)?$`, regexp.QuoteMeta(expected.String())))

	s.verifyURLIsRegexp(s.T(), s.Context(ctx), rx)
}

func (s *TwoFactorSuite) TestShouldAuthorizeSecretAfterTwoFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	username := testUsername
	password := testPassword

	// Login and register TOTP, logout and login again with 1FA & 2FA.
	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginTwoFactor(s.T(), s.Context(ctx), username, password, false, s.secret, targetURL)

	// And check if the user is redirected to the secret.
	s.verifySecretAuthorized(s.T(), s.Context(ctx))

	// Leave the secret.
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))

	// And try to reload it again to check the session is kept.
	s.doVisit(s.T(), s.Context(ctx), targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func (s *TwoFactorSuite) TestShouldFailTwoFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	wrongPasscode := "123456"

	s.doLoginOneFactor(s.T(), s.Context(ctx), testUsername, testPassword, false, BaseDomain, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doEnterOTP(s.T(), s.Context(ctx), wrongPasscode)
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "The one-time password might be wrong")
}

func TestRunTwoFactor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, New2FAScenario())
}
