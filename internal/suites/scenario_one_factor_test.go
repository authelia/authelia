package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type OneFactorSuite struct {
	*SeleniumSuite
}

func NewOneFactorScenario() *OneFactorSuite {
	return &OneFactorSuite{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *OneFactorSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *OneFactorSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *OneFactorSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *OneFactorSuite) TestShouldAuthorizeSecretAfterOneFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, targetURL)
	s.verifySecretAuthorized(ctx, s.T())
}

func (s *OneFactorSuite) TestShouldRedirectToSecondFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, targetURL)
	s.verifyIsSecondFactorPage(ctx, s.T())
}

func (s *OneFactorSuite) TestShouldDenyAccessOnBadPassword() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginOneFactor(ctx, s.T(), "john", "bad-password", false, targetURL)
	s.verifyIsFirstFactorPage(ctx, s.T())
	s.verifyNotificationDisplayed(ctx, s.T(), "Incorrect username or password.")
}

func TestRunOneFactor(t *testing.T) {
	suite.Run(t, NewOneFactorScenario())
}
