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

func NewOneFactorSuite() *OneFactorSuite {
	return &OneFactorSuite{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *OneFactorSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.SeleniumSuite.WebDriverSession = wds
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

	doLogout(ctx, s.SeleniumSuite)
	doVisit(s.SeleniumSuite, HomeBaseURL)
	verifyURLIs(ctx, s.SeleniumSuite, HomeBaseURL)
}

func (s *OneFactorSuite) TestShouldAuthorizeSecretAfterOneFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)
	doLoginOneFactor(ctx, s.SeleniumSuite, "john", "password", false, targetURL)
	verifySecretAuthorized(ctx, s.SeleniumSuite)
}

func (s *OneFactorSuite) TestShouldRedirectToSecondFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	doLoginOneFactor(ctx, s.SeleniumSuite, "john", "password", false, targetURL)
	verifyIsSecondFactorPage(ctx, s.SeleniumSuite)
}

func (s *OneFactorSuite) TestShouldDenyAccessOnBadPassword() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	doLoginOneFactor(ctx, s.SeleniumSuite, "john", "bad-password", false, targetURL)
	verifyIsFirstFactorPage(ctx, s.SeleniumSuite)
	verifyNotificationDisplayed(ctx, s.SeleniumSuite, "Authentication failed. Check your credentials.")
}

func TestRunOneFactor(t *testing.T) {
	suite.Run(t, NewOneFactorSuite())
}
