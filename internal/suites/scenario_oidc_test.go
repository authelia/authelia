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
	*SeleniumSuite
	secret string
}

func NewOIDCScenario() *OIDCScenario {
	return &OIDCScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *OIDCScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s.secret = s.doRegisterAndLogin2FA(ctx, s.T(), "john", "password", false, AdminBaseURL)
}

func (s *OIDCScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *OIDCScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doVisit(s.T(), fmt.Sprintf("%s/logout", OIDCBaseURL))
	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *OIDCScenario) TestShouldAuthorizeAccessToOIDCApp() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doVisit(s.T(), OIDCBaseURL)
	s.verifyIsFirstFactorPage(ctx, s.T())
	s.doFillLoginPageAndClick(ctx, s.T(), "john", "password", false)
	s.verifyIsSecondFactorPage(ctx, s.T())
	s.doValidateTOTP(ctx, s.T(), s.secret)
	time.Sleep(1 * time.Second)

	s.waitBodyContains(ctx, s.T(), "Not logged yet...")

	// this href represents the 'login' link
	err := s.WaitElementLocatedByTagName(ctx, s.T(), "a").Click()
	assert.NoError(s.T(), err)

	s.verifyIsConsentPage(ctx, s.T())

	err = s.WaitElementLocatedByID(ctx, s.T(), "accept-button").Click()
	assert.NoError(s.T(), err)

	// Verify that the app is showing the info related to the user stored in the JWT token
	time.Sleep(1 * time.Second)
	s.waitBodyContains(ctx, s.T(), "Logged in as john!")
}

func (s *OIDCScenario) TestShouldDenyConsent() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doVisit(s.T(), OIDCBaseURL)
	s.verifyIsFirstFactorPage(ctx, s.T())
	s.doFillLoginPageAndClick(ctx, s.T(), "john", "password", false)
	s.verifyIsSecondFactorPage(ctx, s.T())
	s.doValidateTOTP(ctx, s.T(), s.secret)
	time.Sleep(1 * time.Second)

	s.waitBodyContains(ctx, s.T(), "Not logged yet...")
	// this href represents the 'login' link

	err := s.WaitElementLocatedByTagName(ctx, s.T(), "a").Click()
	assert.NoError(s.T(), err)

	s.verifyIsConsentPage(ctx, s.T())

	err = s.WaitElementLocatedByID(ctx, s.T(), "deny-button").Click()
	assert.NoError(s.T(), err)

	time.Sleep(1 * time.Second)
	s.verifyURLIs(ctx, s.T(), "https://oidc.example.com:8080/oauth2/callback?error=access_denied&error_description=User%20has%20rejected%20the%20scopes")
}

func TestRunOIDCScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOIDCSuite())
}
