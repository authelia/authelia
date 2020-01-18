package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type RedirectionURLScenario struct {
	*SeleniumSuite
}

func NewRedirectionURLScenario() *RedirectionURLScenario {
	return &RedirectionURLScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (rus *RedirectionURLScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	rus.WebDriverSession = wds
}

func (rus *RedirectionURLScenario) TearDownSuite() {
	err := rus.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (rus *RedirectionURLScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rus.doLogout(ctx, rus.T())
	rus.doVisit(rus.T(), HomeBaseURL)
	rus.verifyIsHome(ctx, rus.T())
}

func (rus *RedirectionURLScenario) TestShouldVerifyCustomURLParametersArePropagatedAfterRedirection() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html?myparam=test", SingleFactorBaseURL)
	rus.doLoginOneFactor(ctx, rus.T(), "john", "password", false, targetURL)
	rus.verifySecretAuthorized(ctx, rus.T())
	rus.verifyURLIs(ctx, rus.T(), targetURL)
}

func TestRedirectionURLScenario(t *testing.T) {
	suite.Run(t, NewRedirectionURLScenario())
}
