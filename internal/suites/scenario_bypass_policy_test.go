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
	*RodSuite
}

func NewBypassPolicyScenario() *BypassPolicyScenario {
	return &BypassPolicyScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *BypassPolicyScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *BypassPolicyScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *BypassPolicyScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *BypassPolicyScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *BypassPolicyScenario) TestShouldAccessPublicResource() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), AdminBaseURL)
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s/secret.html", PublicBaseURL))
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func TestBypassPolicyScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewBypassPolicyScenario())
}
