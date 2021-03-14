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

func NewTwoFactorScenario() *TwoFactorSuite {
	return &TwoFactorSuite{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *TwoFactorSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
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

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *TwoFactorSuite) TestShouldAuthorizeSecretAfterTwoFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	username := testUsername
	password := testPassword

	// Login one factor
	s.doLoginOneFactor(ctx, s.T(), username, password, false, "")

	// Check he reaches the 2FA stage
	s.verifyIsSecondFactorPage(ctx, s.T())

	// Then register the TOTP factor
	secret := s.doRegisterTOTP(ctx, s.T())

	// And logout
	s.doLogout(ctx, s.T())

	// Login again with 1FA & 2FA
	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginTwoFactor(ctx, s.T(), testUsername, testPassword, false, secret, targetURL)

	// And check if the user is redirected to the secret.
	s.verifySecretAuthorized(ctx, s.T())

	// Leave the secret
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())

	// And try to reload it again to check the session is kept
	s.doVisit(s.T(), targetURL)
	s.verifySecretAuthorized(ctx, s.T())
}

func (s *TwoFactorSuite) TestShouldFailTwoFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Register TOTP secret and logout.
	s.doRegisterThenLogout(ctx, s.T(), testUsername, testPassword)

	wrongPasscode := "123456"

	s.doLoginOneFactor(ctx, s.T(), testUsername, testPassword, false, "")
	s.verifyIsSecondFactorPage(ctx, s.T())
	s.doEnterOTP(ctx, s.T(), wrongPasscode)
	s.verifyNotificationDisplayed(ctx, s.T(), "The one-time password might be wrong")
}

func TestRunTwoFactor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTwoFactorScenario())
}
