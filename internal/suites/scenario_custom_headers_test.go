package suites

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tebeka/selenium"
)

type CustomHeadersScenario struct {
	*SeleniumSuite
}

func NewCustomHeadersScenario() *CustomHeadersScenario {
	return &CustomHeadersScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *CustomHeadersScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *CustomHeadersScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *CustomHeadersScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *CustomHeadersScenario) TestShouldNotForwardCustomHeaderForUnauthenticatedUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doVisit(s.T(), fmt.Sprintf("%s/headers", PublicBaseURL))

	body, err := s.WebDriver().FindElement(selenium.ByTagName, "body")
	s.Assert().NoError(err)
	s.WaitElementTextContains(ctx, s.T(), body, "\"Host\"")

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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	expectedGroups := mapset.NewSetWith("dev", "admins")

	targetURL := fmt.Sprintf("%s/headers", PublicBaseURL)
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, targetURL)
	s.verifyURLIs(ctx, s.T(), targetURL)

	err := s.Wait(ctx, func(d selenium.WebDriver) (bool, error) {
		body, err := s.WebDriver().FindElement(selenium.ByTagName, "body")
		if err != nil {
			return false, err
		}

		if body == nil {
			return false, nil
		}

		content, err := body.Text()
		if err != nil {
			return false, err
		}

		payload := HeadersPayload{}
		if err := json.Unmarshal([]byte(content), &payload); err != nil {
			return false, err
		}

		groups := strings.Split(payload.Headers.ForwardedGroups, ",")
		actualGroups := mapset.NewSet()
		for _, group := range groups {
			actualGroups.Add(group)
		}

		return strings.Contains(payload.Headers.ForwardedUser, "john") && expectedGroups.Equal(actualGroups) &&
			strings.Contains(payload.Headers.ForwardedName, "John Doe") && strings.Contains(payload.Headers.ForwardedEmail, "john.doe@authelia.com"), nil
	})

	require.NoError(s.T(), err)
}

func TestCustomHeadersScenario(t *testing.T) {
	suite.Run(t, NewCustomHeadersScenario())
}
