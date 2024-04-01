package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RegulationScenario struct {
	*RodSuite
}

func NewRegulationScenario() *RegulationScenario {
	return &RegulationScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *RegulationScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *RegulationScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *RegulationScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *RegulationScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *RegulationScenario) TestShouldBanUserAfterTooManyAttempt() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisitLoginPage(s.T(), s.Context(ctx), BaseDomain, "")
	s.doFillLoginPageAndClick(s.T(), s.Context(ctx), "john", "bad-password", false)
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Incorrect username or password")

	for i := 0; i < 3; i++ {
		err := s.WaitElementLocatedByID(s.T(), s.Context(ctx), "password-textfield").Input("bad-password")
		require.NoError(s.T(), err)
		err = s.WaitElementLocatedByID(s.T(), s.Context(ctx), "sign-in-button").Click("left", 1)
		require.NoError(s.T(), err)
	}

	// Enter the correct password and test the regulation lock out.
	err := s.WaitElementLocatedByID(s.T(), s.Context(ctx), "password-textfield").Input("password")
	require.NoError(s.T(), err)
	err = s.WaitElementLocatedByID(s.T(), s.Context(ctx), "sign-in-button").Click("left", 1)
	require.NoError(s.T(), err)
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Incorrect username or password")

	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	time.Sleep(10 * time.Second)

	// Enter the correct password and test a successful login.
	err = s.WaitElementLocatedByID(s.T(), s.Context(ctx), "password-textfield").Input("password")
	require.NoError(s.T(), err)
	err = s.WaitElementLocatedByID(s.T(), s.Context(ctx), "sign-in-button").Click("left", 1)
	require.NoError(s.T(), err)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
}

func TestBlacklistingScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewRegulationScenario())
}
