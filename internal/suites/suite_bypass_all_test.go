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
	*SeleniumSuite
}

func NewBypassAllWebDriverSuite() *BypassAllWebDriverSuite {
	return &BypassAllWebDriverSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *BypassAllWebDriverSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *BypassAllWebDriverSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *BypassAllWebDriverSuite) TestShouldAccessPublicResource() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doVisit(s.T(), fmt.Sprintf("%s/secret.html", AdminBaseURL))
	s.verifySecretAuthorized(ctx, s.T())

	s.doVisit(s.T(), fmt.Sprintf("%s/secret.html", PublicBaseURL))
	s.verifySecretAuthorized(ctx, s.T())
}

type BypassAllSuite struct {
	suite.Suite
}

func NewBypassAllSuite() *BypassAllSuite {
	return &BypassAllSuite{}
}

func (s *BypassAllSuite) TestBypassAllWebDriverSuite() {
	suite.Run(s.T(), NewBypassAllWebDriverSuite())
}

func (s *BypassAllSuite) TestCustomHeadersScenario() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func TestBypassAllSuite(t *testing.T) {
	suite.Run(t, NewBypassAllSuite())
}
