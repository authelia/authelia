package suites

import (
	"context"
	"fmt"
	"log"
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
	err := s.Page.MustSearch("Log in").Click("left")
	assert.NoError(s.T(), err)

	s.verifyIsConsentPage(s.T(), s.Context(ctx))
	err = s.WaitElementLocatedByCSSSelector(s.T(), s.Context(ctx), "accept-button").Click("left")
	assert.NoError(s.T(), err)

	// Verify that the app is showing the info related to the user stored in the JWT token.
	s.waitBodyContains(s.T(), s.Context(ctx), "Logged in as john!")
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
	err := s.Page.MustSearch("Log in").Click("left")
	assert.NoError(s.T(), err)

	s.verifyIsConsentPage(s.T(), s.Context(ctx))

	err = s.WaitElementLocatedByCSSSelector(s.T(), s.Context(ctx), "deny-button").Click("left")
	assert.NoError(s.T(), err)

	s.verifyIsOIDC(s.T(), s.Context(ctx), "oauth2:", "https://oidc.example.com:8080/oauth2/callback?error=access_denied&error_description=User%20has%20rejected%20the%20scopes")
}

func TestRunOIDCScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOIDCSuite())
}
