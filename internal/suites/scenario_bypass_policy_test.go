package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type BypassPolicyScenario struct {
	*SeleniumSuite
}

func NewBypassPolicyScenario() *BypassPolicyScenario {
	return &BypassPolicyScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *BypassPolicyScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *BypassPolicyScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *BypassPolicyScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *BypassPolicyScenario) TestShouldAccessPublicResource() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doVisit(s.T(), AdminBaseURL)
	s.verifyIsFirstFactorPage(ctx, s.T())

	s.doVisit(s.T(), fmt.Sprintf("%s/secret.html", PublicBaseURL))
	s.verifySecretAuthorized(ctx, s.T())
}

func TestBypassPolicyScenario(t *testing.T) {
	suite.Run(t, NewBypassPolicyScenario())
}
