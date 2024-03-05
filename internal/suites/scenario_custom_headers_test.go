package suites

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/stretchr/testify/suite"
)

type CustomHeadersScenario struct {
	*RodSuite
}

func NewCustomHeadersScenario() *CustomHeadersScenario {
	return &CustomHeadersScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *CustomHeadersScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *CustomHeadersScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *CustomHeadersScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *CustomHeadersScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *CustomHeadersScenario) TestShouldNotForwardCustomHeaderForUnauthenticatedUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s/headers", PublicBaseURL))

	body, err := s.Context(ctx).Element("body")
	s.Assert().NoError(err)

	b, err := body.Text()
	s.Assert().NoError(err)
	s.Assert().NotContains(b, "john")
	s.Assert().NotContains(b, "admins")
	s.Assert().NotContains(b, "John Doe")
	s.Assert().NotContains(b, "john.doe@authelia.com")
}

type Headers struct {
	ForwardedEmail  string `json:"Remote-Email"`
	ForwardedGroups string `json:"Remote-Groups"`
	ForwardedName   string `json:"Remote-Name"`
	ForwardedUser   string `json:"Remote-User"`
}

type HeadersPayload struct {
	Headers Headers `json:"headers"`
}

func (s *CustomHeadersScenario) TestShouldForwardCustomHeaderForAuthenticatedUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	expectedGroups := mapset.NewSet("dev", "admins")

	targetURL := fmt.Sprintf("%s/headers", PublicBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, targetURL)
	s.verifyIsPublic(s.T(), s.Context(ctx))

	body, err := s.Context(ctx).Element("body")
	s.Assert().NoError(err)
	s.Assert().NotNil(body)

	content, err := body.Text()
	s.Assert().NoError(err)
	s.Assert().NotNil(content)

	payload := HeadersPayload{}
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		log.Panic(err)
	}

	groups := strings.Split(payload.Headers.ForwardedGroups, ",")
	actualGroups := mapset.NewSet[string]()

	for _, group := range groups {
		actualGroups.Add(group)
	}

	if strings.Contains(payload.Headers.ForwardedUser, "john") && expectedGroups.Equal(actualGroups) &&
		strings.Contains(payload.Headers.ForwardedName, "John Doe") && strings.Contains(payload.Headers.ForwardedEmail, "john.doe@authelia.com") {
		err = nil
	} else {
		err = fmt.Errorf("headers do not include user information")
	}

	s.Require().NoError(err)
}

func TestCustomHeadersScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewCustomHeadersScenario())
}
