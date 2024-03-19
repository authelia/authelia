package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type OneFactorOnlySuite struct {
	*RodSuite
}

func NewOneFactorOnlySuite() *OneFactorOnlySuite {
	return &OneFactorOnlySuite{
		RodSuite: NewRodSuite(oneFactorOnlySuiteName),
	}
}

func (s *OneFactorOnlySuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *OneFactorOnlySuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *OneFactorOnlySuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *OneFactorOnlySuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

// No target url is provided, then the user should be redirect to the default url.
func (s *OneFactorOnlySuite) TestShouldRedirectUserToDefaultURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyIsHome(s.T(), s.Context(ctx))
}

// Unsafe URL is provided, then the user should be redirect to the default url.
func (s *OneFactorOnlySuite) TestShouldRedirectUserToDefaultURLWhenURLIsUnsafe() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "http://unsafe.local")
	s.verifyIsHome(s.T(), s.Context(ctx))
}

// When use logged in and visit the portal again, she gets redirect to the authenticated view.
func (s *OneFactorOnlySuite) TestShouldDisplayAuthenticatedView() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyIsHome(s.T(), s.Context(ctx))
	s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURL(BaseDomain))
	s.verifyIsAuthenticatedPage(s.T(), s.Context(ctx))
}

func (s *OneFactorOnlySuite) TestShouldRedirectAlreadyAuthenticatedUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyIsHome(s.T(), s.Context(ctx))

	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s?rd=https://singlefactor.example.com:8080/secret.html", GetLoginBaseURL(BaseDomain)))
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
	s.verifyURLIs(s.T(), s.Context(ctx), "https://singlefactor.example.com:8080/secret.html")
}

func (s *OneFactorOnlySuite) TestShouldNotRedirectAlreadyAuthenticatedUserToUnsafeURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyIsHome(s.T(), s.Context(ctx))

	// Visit the login page and wait for redirection to 2FA page with success icon displayed.
	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s?rd=https://secure.example.local:8080", GetLoginBaseURL(BaseDomain)))
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Redirection was determined to be unsafe and aborted ensure the redirection URL is correct")
}

func TestOneFactorOnlySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOneFactorOnlySuite())
}
