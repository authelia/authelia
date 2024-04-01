package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PasswordComplexityScenario struct {
	*RodSuite
}

func NewPasswordComplexityScenario() *PasswordComplexityScenario {
	return &PasswordComplexityScenario{RodSuite: NewRodSuite("")}
}

func (s *PasswordComplexityScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *PasswordComplexityScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *PasswordComplexityScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *PasswordComplexityScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *PasswordComplexityScenario) TestShouldRejectPasswordReset() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURL(BaseDomain))
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

	// Attempt to reset the password to a.
	s.doResetPassword(s.T(), s.Context(ctx), "john", "a", "a", true)
}

func TestRunPasswordComplexityScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewPasswordComplexityScenario())
}
