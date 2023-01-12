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
	suite.Suite
}

type OneFactorOnlyWebSuite struct {
	*RodSuite
}

func NewOneFactorOnlyWebSuite() *OneFactorOnlyWebSuite {
	return &OneFactorOnlyWebSuite{RodSuite: new(RodSuite)}
}

func (s *OneFactorOnlyWebSuite) SetupSuite() {
	browser, err := StartRod()

	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *OneFactorOnlyWebSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *OneFactorOnlyWebSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *OneFactorOnlyWebSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

// No target url is provided, then the user should be redirect to the default url.
func (s *OneFactorOnlyWebSuite) TestShouldRedirectUserToDefaultURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyIsHome(s.T(), s.Context(ctx))
}

// Unsafe URL is provided, then the user should be redirect to the default url.
func (s *OneFactorOnlyWebSuite) TestShouldRedirectUserToDefaultURLWhenURLIsUnsafe() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "http://unsafe.local")
	s.verifyIsHome(s.T(), s.Context(ctx))
}

// When use logged in and visit the portal again, she gets redirect to the authenticated view.
func (s *OneFactorOnlyWebSuite) TestShouldDisplayAuthenticatedView() {
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

func (s *OneFactorOnlyWebSuite) TestShouldRedirectAlreadyAuthenticatedUser() {
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

func (s *OneFactorOnlyWebSuite) TestShouldNotRedirectAlreadyAuthenticatedUserToUnsafeURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyIsHome(s.T(), s.Context(ctx))

	// Visit the login page and wait for redirection to 2FA page with success icon displayed.
	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s?rd=https://secure.example.local:8080", GetLoginBaseURL(BaseDomain)))
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Redirection was determined to be unsafe and aborted. Ensure the redirection URL is correct.")
}

func (s *OneFactorOnlySuite) TestWeb() {
	suite.Run(s.T(), NewOneFactorOnlyWebSuite())
}

func TestOneFactorOnlySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, new(OneFactorOnlySuite))
}
