package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DefaultRedirectionURLScenario struct {
	*SeleniumSuite

	secret string
}

func NewDefaultRedirectionURLScenario() *DefaultRedirectionURLScenario {
	return &DefaultRedirectionURLScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (drus *DefaultRedirectionURLScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	drus.WebDriverSession = wds

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	drus.secret = drus.doRegisterAndLogin2FA(ctx, drus.T(), "john", "password", false, targetURL)
	drus.verifySecretAuthorized(ctx, drus.T())
}

func (drus *DefaultRedirectionURLScenario) TearDownSuite() {
	err := drus.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (drus *DefaultRedirectionURLScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	drus.doLogout(ctx, drus.T())
	drus.doVisit(drus.T(), HomeBaseURL)
	drus.verifyIsHome(ctx, drus.T())
}

func (drus *DefaultRedirectionURLScenario) TestUserIsRedirectedToDefaultURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	drus.doLoginTwoFactor(ctx, drus.T(), "john", "password", false, drus.secret, "")
	drus.verifyURLIs(ctx, drus.T(), HomeBaseURL+"/")
}

func TestShouldRunDefaultRedirectionURLScenario(t *testing.T) {
	suite.Run(t, NewDefaultRedirectionURLScenario())
}
