package suites

// This scenario is used to test sign in using the user email address.

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type SigninEmailScenario struct {
	*SeleniumSuite
}

func NewSigninEmailScenario() *SigninEmailScenario {
	return &SigninEmailScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *SigninEmailScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *SigninEmailScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *SigninEmailScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *SigninEmailScenario) TestShouldSignInWithUserEmail() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)
	s.doLoginOneFactor(ctx, s.T(), "john.doe@authelia.com", "password", false, targetURL)
	s.verifySecretAuthorized(ctx, s.T())
}

func TestSigninEmailScenario(t *testing.T) {
	suite.Run(t, NewSigninEmailScenario())
}
