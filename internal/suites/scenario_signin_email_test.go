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
	*RodSuite
}

func NewSigninEmailScenario() *SigninEmailScenario {
	return &SigninEmailScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *SigninEmailScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *SigninEmailScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *SigninEmailScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *SigninEmailScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *SigninEmailScenario) TestShouldSignInWithUserEmail() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john.doe@authelia.com", "password", false, BaseDomain, targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func TestSigninEmailScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewSigninEmailScenario())
}
