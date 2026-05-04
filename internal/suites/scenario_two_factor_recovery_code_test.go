// SPDX-FileCopyrightText: 2026 Authelia
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

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TwoFactorRecoveryCodeScenario struct {
	*RodSuite

	codes []string
}

func New2FARecoveryCodeScenario() *TwoFactorRecoveryCodeScenario {
	return &TwoFactorRecoveryCodeScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *TwoFactorRecoveryCodeScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		s.collectCoverage(s.Page)
		s.MustClose()
	}()

	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.doLoginAndRegisterTOTP(s.T(), s.Context(ctx), "john", "password", false)

	s.codes = s.doRegisterRecoveryCodes(s.T(), s.Context(ctx))

	s.doLogout(s.T(), s.Context(ctx))
}

func (s *TwoFactorRecoveryCodeScenario) TearDownSuite() {
	if err := s.Stop(); err != nil {
		log.Fatal(err)
	}
}

func (s *TwoFactorRecoveryCodeScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *TwoFactorRecoveryCodeScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *TwoFactorRecoveryCodeScenario) TestShouldNotAuthorizeSecretBeforeRecoveryCode2FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	q := expected.Query()
	q.Set("rd", targetURL)
	expected.RawQuery = q.Encode()

	rx := regexp.MustCompile(fmt.Sprintf(`^%s(&rm=GET)?$`, regexp.QuoteMeta(expected.String())))
	s.verifyURLIsRegexp(s.T(), s.Context(ctx), rx)
}

func (s *TwoFactorRecoveryCodeScenario) TestShouldAuthorizeSecretAfterRecoveryCode2FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	require.NotEmpty(s.T(), s.codes, "recovery codes should have been generated in SetupSuite")

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)

	code := s.popRecoveryCode()
	s.doLoginSecondFactorRecoveryCode(s.T(), s.Context(ctx), "john", "password", false, code, targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))

	s.doLogout(s.T(), s.Context(ctx))
}

func (s *TwoFactorRecoveryCodeScenario) TestShouldFailWithUsedRecoveryCode() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	require.NotEmpty(s.T(), s.codes, "recovery codes should have been generated in SetupSuite")

	used := s.popRecoveryCode()

	s.doLoginSecondFactorRecoveryCode(s.T(), s.Context(ctx), "john", "password", false, used, "")
	s.doLogout(s.T(), s.Context(ctx))

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))

	s.openRecoveryCodeForm(s.T(), s.Context(ctx))
	s.doEnterRecoveryCode(s.T(), s.Context(ctx), used)

	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "The recovery code might be wrong")
}

func (s *TwoFactorRecoveryCodeScenario) TestShouldFailWithGarbageRecoveryCode() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))

	s.openRecoveryCodeForm(s.T(), s.Context(ctx))
	s.doEnterRecoveryCode(s.T(), s.Context(ctx), "BADBA-DBADB")

	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "The recovery code might be wrong")
}

func (s *TwoFactorRecoveryCodeScenario) popRecoveryCode() string {
	require.NotEmpty(s.T(), s.codes)

	code := s.codes[0]
	s.codes = s.codes[1:]

	return code
}

// doRegisterRecoveryCodes opens the Recovery Codes panel under settings, completes the elevation flow, captures the
// 10 codes from the generation modal, satisfies the retype-confirm gate, and returns the codes for use by the suite.
func (s *TwoFactorRecoveryCodeScenario) doRegisterRecoveryCodes(t *testing.T, page *rod.Page) []string {
	t.Helper()

	s.doVisit(t, page, GetLoginBaseURL(BaseDomain)+"/settings/two-factor-authentication")

	addBtn := s.WaitElementLocatedByID(t, page, "recovery-codes-add")
	require.NoError(t, addBtn.Click("left", 1))

	// The dialog auto-fires generation on open. Wait for the codes to render. Each code is shown on its own row;
	// we read the body text and split on lines that match the XXXXX-XXXXX format.
	codes := s.scrapeRecoveryCodes(t, page)

	if len(codes) == 0 {
		t.Fatalf("expected to scrape recovery codes from the generation dialog, got none")
	}

	confirmInput := s.WaitElementLocatedByID(t, page, "recovery-codes-confirm-input")
	require.NoError(t, confirmInput.Input(codes[0]))

	doneBtn := s.WaitElementLocatedByID(t, page, "recovery-codes-done")
	require.NoError(t, doneBtn.Click("left", 1))

	return codes
}

func (s *TwoFactorRecoveryCodeScenario) scrapeRecoveryCodes(t *testing.T, page *rod.Page) []string {
	t.Helper()

	require.NoError(t, page.WaitStable(2*time.Second))

	body, err := page.Element("body")
	require.NoError(t, err)

	text, err := body.Text()
	require.NoError(t, err)

	// Codes are 5-char + hyphen + 5-char from random.CharSetUnambiguousUpper (no 0/O/I/L/1/5).
	re := regexp.MustCompile(`\b[ABCDEFGHJKLMNPQRTUVWYXZ2346789]{5}-[ABCDEFGHJKLMNPQRTUVWYXZ2346789]{5}\b`)
	matches := re.FindAllString(text, -1)

	seen := make(map[string]bool, len(matches))
	out := make([]string, 0, len(matches))

	for _, m := range matches {
		if seen[m] {
			continue
		}

		seen[m] = true
		out = append(out, m)
	}

	return out
}

func (s *TwoFactorRecoveryCodeScenario) openRecoveryCodeForm(t *testing.T, page *rod.Page) {
	t.Helper()

	link := s.WaitElementLocatedByID(t, page, "recovery-code-link")
	require.NoError(t, link.Click("left", 1))
}

func (s *TwoFactorRecoveryCodeScenario) doEnterRecoveryCode(t *testing.T, page *rod.Page, code string) {
	t.Helper()

	field := s.WaitElementLocatedByID(t, page, "recovery-code-input")
	require.NoError(t, field.Input(code))

	signInBtn := s.WaitElementLocatedByID(t, page, "recovery-code-submit")
	require.NoError(t, signInBtn.Click("left", 1))
}

func (s *TwoFactorRecoveryCodeScenario) doLoginSecondFactorRecoveryCode(t *testing.T, page *rod.Page, username, password string, keepMeLoggedIn bool, code, targetURL string) {
	t.Helper()

	s.doLoginOneFactor(t, page, username, password, keepMeLoggedIn, BaseDomain, targetURL)
	s.verifyIsSecondFactorPage(t, page)

	s.openRecoveryCodeForm(t, page)
	s.doEnterRecoveryCode(t, page, code)

	if targetURL == "" {
		require.NoError(t, page.WaitStable(time.Second))
	}
}

func TestRunTwoFactorRecoveryCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, New2FARecoveryCodeScenario())
}
