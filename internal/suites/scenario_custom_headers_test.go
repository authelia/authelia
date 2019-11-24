package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tebeka/selenium"
)

type CustomHeadersScenario struct {
	*SeleniumSuite
}

func NewCustomHeadersScenario() *CustomHeadersScenario {
	return &CustomHeadersScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *CustomHeadersScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *CustomHeadersScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *CustomHeadersScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *CustomHeadersScenario) TestShouldNotForwardCustomHeaderForUnauthenticatedUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doVisit(s.T(), fmt.Sprintf("%s/headers", PublicBaseURL))

	body, err := s.WebDriver().FindElement(selenium.ByTagName, "body")
	s.Assert().NoError(err)
	s.WaitElementTextContains(ctx, s.T(), body, "httpbin:8000")
}

func (s *CustomHeadersScenario) TestShouldForwardCustomHeaderForAuthenticatedUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/headers", PublicBaseURL)
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, targetURL)
	s.verifyURLIs(ctx, s.T(), targetURL)

	body, err := s.WebDriver().FindElement(selenium.ByTagName, "body")
	s.Assert().NoError(err)
	s.WaitElementTextContains(ctx, s.T(), body, "\"Custom-Forwarded-User\": \"john\"")
	s.WaitElementTextContains(ctx, s.T(), body, "\"Custom-Forwarded-Groups\": \"admins,dev\"")
}

func TestCustomHeadersScenario(t *testing.T) {
	suite.Run(t, NewCustomHeadersScenario())
}
