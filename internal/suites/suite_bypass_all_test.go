package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type BypassAllWebDriverSuite struct {
	*RodSuite
}

func NewBypassAllWebDriverSuite() *BypassAllWebDriverSuite {
	return &BypassAllWebDriverSuite{
		RodSuite: NewRodSuite(""),
	}
}

func (s *BypassAllWebDriverSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *BypassAllWebDriverSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *BypassAllWebDriverSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *BypassAllWebDriverSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *BypassAllWebDriverSuite) TestShouldAccessPublicResource() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s/secret.html", AdminBaseURL))
	s.verifySecretAuthorized(s.T(), s.Context(ctx))

	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s/secret.html", PublicBaseURL))
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

type BypassAllSuite struct {
	*BaseSuite
}

func NewBypassAllSuite() *BypassAllSuite {
	return &BypassAllSuite{
		BaseSuite: &BaseSuite{
			Name: bypassAllSuiteName,
		},
	}
}

func (s *BypassAllSuite) TestBypassAllWebDriverSuite() {
	suite.Run(s.T(), NewBypassAllWebDriverSuite())
}

func (s *BypassAllSuite) TestCustomHeadersScenario() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func TestBypassAllSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewBypassAllSuite())
}
