package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type BypassAllSuite struct {
	*SeleniumSuite
}

func NewBypassAllSuite() *BypassAllSuite {
	return &BypassAllSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *BypassAllSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *BypassAllSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *BypassAllSuite) TestShouldAccessPublicResource() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doVisit(s.T(), fmt.Sprintf("%s/secret.html", AdminBaseURL))
	s.verifySecretAuthorized(ctx, s.T())

	s.doVisit(s.T(), fmt.Sprintf("%s/secret.html", PublicBaseURL))
	s.verifySecretAuthorized(ctx, s.T())
}

func TestBypassAllSuite(t *testing.T) {
	suite.Run(t, NewBypassAllSuite())
	suite.Run(t, NewCustomHeadersScenario())
}
