package suites

import (
	"context"
	"fmt"
	"log"
	"time"
)

// MultiCookieDomainScenario represents a set of tests for multi cookie domain suite.
type MultiCookieDomainScenario struct {
	*RodSuite
}

// NewMultiCookieDomainScenario returns a new Multi Cookie Domain Test Scenario.
func NewMultiCookieDomainScenario() *MultiCookieDomainScenario {
	return &MultiCookieDomainScenario{
		RodSuite: new(RodSuite),
	}
}

func (s *MultiCookieDomainScenario) SetupSuite() {
	browser, err := StartRod()

	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *MultiCookieDomainScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *MultiCookieDomainScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *MultiCookieDomainScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *MultiCookieDomainScenario) TestShouldAuthorizeSecretOnFirstDomain() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		s.doLogout(s.T(), s.Page)
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, targetURL)
	s.verifySecretAuthorized(s.T(), s.Page)
}

func (s *MultiCookieDomainScenario) TestShouldAuthorizeSecretOnSecondDomain() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		s.doLogout(s.T(), s.Page)
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURLFmt(Example2Com))

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, Example2Com, targetURL)
	s.verifySecretAuthorized(s.T(), s.Page)
}

func (s *MultiCookieDomainScenario) TestShouldRequestLoginOnSecondDomainAfterLoginOnFirstDomain() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		s.doLogout(s.T(), s.Page)
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	firstDomain := BaseDomain
	secondDomain := Example2Com

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURLFmt(firstDomain))
	targetURL2 := fmt.Sprintf("%s/secret.html", SingleFactorBaseURLFmt(secondDomain))

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, firstDomain, targetURL)
	s.verifySecretAuthorized(s.T(), s.Page)
	s.doVisit(s.T(), s.Page, targetURL2)
	s.verifyIsFirstFactorPage(s.T(), s.Page)
}

func (s *MultiCookieDomainScenario) TestShouldRequestLoginOnFirstDomainAfterLoginOnSecondDomain() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		s.doLogout(s.T(), s.Page)
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	firstDomain := Example2Com
	secondDomain := BaseDomain

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURLFmt(firstDomain))
	targetURL2 := fmt.Sprintf("%s/secret.html", SingleFactorBaseURLFmt(secondDomain))

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, firstDomain, targetURL)
	s.verifySecretAuthorized(s.T(), s.Page)
	s.doVisit(s.T(), s.Page, targetURL2)
	s.verifyIsFirstFactorPage(s.T(), s.Page)
}
