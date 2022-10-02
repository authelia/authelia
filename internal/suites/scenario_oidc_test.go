package suites

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OIDCScenario struct {
	*RodSuite
	secret string
}

func NewOIDCScenario() *OIDCScenario {
	return &OIDCScenario{
		RodSuite: new(RodSuite),
	}
}

func (s *OIDCScenario) SetupSuite() {
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
	s.secret = s.doRegisterAndLogin2FA(s.T(), s.Context(ctx), "john", "password", false, AdminBaseURL)
}

func (s *OIDCScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *OIDCScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.Page = s.doCreateTab(s.T(), fmt.Sprintf("%s/logout", OIDCBaseURL))
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (s *OIDCScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *OIDCScenario) TestShouldAuthorizeAccessToOIDCApp() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), OIDCBaseURL)
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.doFillLoginPageAndClick(s.T(), s.Context(ctx), "john", "password", false)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doValidateTOTP(s.T(), s.Context(ctx), s.secret)

	s.waitBodyContains(s.T(), s.Context(ctx), "Not logged yet...")

	// Search for the 'login' link.
	err := s.Page.MustSearch("Log in").Click("left", 1)
	assert.NoError(s.T(), err)

	s.verifyIsConsentPage(s.T(), s.Context(ctx))
	err = s.WaitElementLocatedByID(s.T(), s.Context(ctx), "accept-button").Click("left", 1)
	assert.NoError(s.T(), err)

	// Verify that the app is showing the info related to the user stored in the JWT token.

	rUUID := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	rInteger := regexp.MustCompile(`^\d+$`)
	rBoolean := regexp.MustCompile(`^(true|false)$`)
	rBase64 := regexp.MustCompile(`^[-_A-Za-z0-9+\\/]+([=]{0,3})$`)

	testCases := []struct {
		desc, elementID, elementText string
		pattern                      *regexp.Regexp
	}{
		{"welcome", "welcome", "Logged in as john!", nil},
		{"at_hash", "claim-at_hash", "", rBase64},
		{"jti", "claim-jti", "", rUUID},
		{"iat", "claim-iat", "", rInteger},
		{"nbf", "claim-nbf", "", rInteger},
		{"rat", "claim-rat", "", rInteger},
		{"expires", "claim-exp", "", rInteger},
		{"amr", "claim-amr", "pwd, otp, mfa", nil},
		{"acr", "claim-acr", "", nil},
		{"issuer", "claim-iss", "https://login.example.com:8080", nil},
		{"name", "claim-name", "John Doe", nil},
		{"preferred_username", "claim-preferred_username", "john", nil},
		{"groups", "claim-groups", "admins, dev", nil},
		{"email", "claim-email", "john.doe@authelia.com", nil},
		{"email_verified", "claim-email_verified", "", rBoolean},
	}

	var text string

	for _, tc := range testCases {
		s.T().Run(fmt.Sprintf("check_claims/%s", tc.desc), func(t *testing.T) {
			text, err = s.WaitElementLocatedByID(t, s.Context(ctx), tc.elementID).Text()
			assert.NoError(t, err)
			if tc.pattern == nil {
				assert.Equal(t, tc.elementText, text)
			} else {
				assert.Regexp(t, tc.pattern, text)
			}
		})
	}
}

func (s *OIDCScenario) TestShouldDenyConsent() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), OIDCBaseURL)
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.doFillLoginPageAndClick(s.T(), s.Context(ctx), "john", "password", false)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doValidateTOTP(s.T(), s.Context(ctx), s.secret)

	s.waitBodyContains(s.T(), s.Context(ctx), "Not logged yet...")

	// Search for the 'login' link.
	err := s.Page.MustSearch("Log in").Click("left", 1)
	assert.NoError(s.T(), err)

	s.verifyIsConsentPage(s.T(), s.Context(ctx))

	err = s.WaitElementLocatedByID(s.T(), s.Context(ctx), "deny-button").Click("left", 1)
	assert.NoError(s.T(), err)

	s.verifyIsOIDC(s.T(), s.Context(ctx), "access_denied", "https://oidc.example.com:8080/error?error=access_denied&error_description=The+resource+owner+or+authorization+server+denied+the+request.+Make+sure+that+the+request+you+are+making+is+valid.+Maybe+the+credential+or+request+parameters+you+are+using+are+limited+in+scope+or+otherwise+restricted.&state=random-string-here")

	errorDescription := "The resource owner or authorization server denied the request. Make sure that the request " +
		"you are making is valid. Maybe the credential or request parameters you are using are limited in scope or " +
		"otherwise restricted."

	s.verifyIsOIDCErrorPage(s.T(), s.Context(ctx), "access_denied", errorDescription, "",
		"random-string-here")
}

func TestRunOIDCScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOIDCSuite())
}
