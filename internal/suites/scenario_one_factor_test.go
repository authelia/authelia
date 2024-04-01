package suites

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type OneFactorSuite struct {
	*RodSuite
}

func New1FAScenario() *OneFactorSuite {
	return &OneFactorSuite{
		RodSuite: NewRodSuite(""),
	}
}

func (s *OneFactorSuite) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *OneFactorSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *OneFactorSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *OneFactorSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *OneFactorSuite) TestShouldNotAuthorizeSecretBeforeOneFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)

	s.doVisit(s.T(), s.Context(ctx), targetURL)

	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

	raw := GetLoginBaseURLWithFallbackPrefix(BaseDomain, "/")

	expected, err := url.ParseRequestURI(raw)
	s.Assert().NoError(err)
	s.Require().NotNil(expected)

	query := expected.Query()

	query.Set("rd", targetURL)

	expected.RawQuery = query.Encode()

	rx := regexp.MustCompile(fmt.Sprintf(`^%s(&rm=GET)?$`, regexp.QuoteMeta(expected.String())))

	s.verifyURLIsRegexp(s.T(), s.Context(ctx), rx)
}

func (s *OneFactorSuite) TestShouldAuthorizeSecretAfterOneFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, targetURL)
	s.verifySecretAuthorized(s.T(), s.Page)
}

func (s *OneFactorSuite) TestShouldRedirectToSecondFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, targetURL)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
}

func (s *OneFactorSuite) TestShouldDenyAccessOnBadPassword() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "bad-password", false, BaseDomain, targetURL)
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Incorrect username or password")
}

func (s *OneFactorSuite) TestShouldDenyAccessOnForbidden() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", DenyBaseURL)
	s.doVisit(s.T(), s.Context(ctx), targetURL)
	s.NoError(s.WaitStable(time.Millisecond * 10))

	s.verifyURLIs(s.T(), s.Context(ctx), targetURL)
	s.verifyBodyContains(s.T(), s.Context(ctx), "403 Forbidden")
}

func TestRunOneFactor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, New1FAScenario())
}
