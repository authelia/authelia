package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TwoFactorSuite struct {
	*SeleniumSuite
}

func NewTwoFactorSuite() *TwoFactorSuite {
	return &TwoFactorSuite{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *TwoFactorSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.SeleniumSuite.WebDriverSession = wds
}

func (s *TwoFactorSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *TwoFactorSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doLogout(ctx, s.SeleniumSuite)
	doVisit(s.SeleniumSuite, HomeBaseURL)
	verifyURLIs(ctx, s.SeleniumSuite, HomeBaseURL)
}

func (s *TwoFactorSuite) TestShouldAuthorizeSecretAfterTwoFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Register TOTP secret and logout.
	secret := doRegisterThenLogout(ctx, s.SeleniumSuite, "john", "password")

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	doLoginTwoFactor(ctx, s.SeleniumSuite, "john", "password", false, secret, targetURL)

	verifySecretAuthorized(ctx, s.SeleniumSuite)
}

func TestRunTwoFactor(t *testing.T) {
	suite.Run(t, NewTwoFactorSuite())
}
