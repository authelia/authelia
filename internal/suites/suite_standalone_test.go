package suites

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

type StandaloneWebDriverSuite struct {
	*RodSuite
}

func NewStandaloneWebDriverSuite() *StandaloneWebDriverSuite {
	return &StandaloneWebDriverSuite{
		RodSuite: NewRodSuite(""),
	}
}

func (s *StandaloneWebDriverSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *StandaloneWebDriverSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *StandaloneWebDriverSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *StandaloneWebDriverSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *StandaloneWebDriverSuite) TestShouldLetUserKnowHeIsAlreadyAuthenticated() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doRegisterTOTPAndLogin2FA(s.T(), s.Context(ctx), "john", "password", false, "")

	// Visit home page to change context.
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))

	// Visit the login page and wait for redirection to 2FA page with success icon displayed.
	s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURL(BaseDomain))
	s.verifyIsAuthenticatedPage(s.T(), s.Context(ctx))
}

func (s *StandaloneWebDriverSuite) TestShouldRedirectAfterOneFactorOnAnotherTab() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)
	page2 := s.Browser().MustPage(targetURL)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		s.collectScreenshot(ctx.Err(), page2)
		page2.MustClose()
	}()

	// Open second tab with secret page.
	page2.MustWaitLoad()

	// Switch to first, visit the login page and wait for redirection to secret page with secret displayed.
	s.Page.MustActivate()
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, targetURL)
	s.verifySecretAuthorized(s.T(), s.Page)

	// Switch to second tab and wait for redirection to secret page with secret displayed.
	page2.MustActivate()
	s.verifySecretAuthorized(s.T(), page2.Context(ctx))
}

func (s *StandaloneWebDriverSuite) TestShouldRedirectAlreadyAuthenticatedUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doRegisterTOTPAndLogin2FA(s.T(), s.Context(ctx), "john", "password", false, "")

	// Visit home page to change context.
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))

	// Visit the login page and wait for redirection to 2FA page with success icon displayed.
	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s?rd=https://secure.example.com:8080", GetLoginBaseURL(BaseDomain)))

	_, err := s.Page.ElementR("h1", "Public resource")
	require.NoError(s.T(), err)
	s.verifyURLIs(s.T(), s.Context(ctx), "https://secure.example.com:8080/")
}

func (s *StandaloneWebDriverSuite) TestShouldNotRedirectAlreadyAuthenticatedUserToUnsafeURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doRegisterTOTPAndLogin2FA(s.T(), s.Context(ctx), "john", "password", false, "")

	// Visit home page to change context.
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))

	// Visit the login page and wait for redirection to 2FA page with success icon displayed.
	s.doVisit(s.T(), s.Context(ctx), fmt.Sprintf("%s?rd=https://secure.example.local:8080", GetLoginBaseURL(BaseDomain)))
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Redirection was determined to be unsafe and aborted ensure the redirection URL is correct")
}

func (s *StandaloneWebDriverSuite) TestShouldCheckUserIsAskedToRegisterDevice() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	username := "john"
	password := "password"

	// Clean up any TOTP secret already in DB.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)

	require.NoError(s.T(), provider.DeleteTOTPConfiguration(ctx, username))

	// Login one factor.
	s.doLoginOneFactor(s.T(), s.Context(ctx), username, password, false, BaseDomain, "")

	// Check the user is asked to register a new device.
	s.WaitElementLocatedByClassName(s.T(), s.Context(ctx), "state-not-registered")

	// Then register the TOTP factor.
	s.doOpenSettingsAndRegisterTOTP(s.T(), s.Context(ctx), username)
	// And logout.
	s.doLogout(s.T(), s.Context(ctx))

	// Login one factor again.
	s.doLoginOneFactor(s.T(), s.Context(ctx), username, password, false, BaseDomain, "")

	// now the user should be asked to perform 2FA.
	s.WaitElementLocatedByClassName(s.T(), s.Context(ctx), "state-method")
}

type StandaloneSuite struct {
	*BaseSuite
}

func NewStandaloneSuite() *StandaloneSuite {
	return &StandaloneSuite{
		BaseSuite: &BaseSuite{
			Name: standaloneSuiteName,
		},
	}
}

func (s *StandaloneSuite) TestShouldRespectMethodsACL() {
	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL(BaseDomain)), nil)
	s.Assert().NoError(err)
	req.Header.Set("X-Forwarded-Method", fasthttp.MethodGet)
	req.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	req.Header.Set(fasthttp.HeaderXForwardedHost, fmt.Sprintf("secure.%s", BaseDomain))
	req.Header.Set("X-Forwarded-URI", "/")
	req.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf8")

	client := NewHTTPClient()
	res, err := client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(fasthttp.StatusFound, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)

	urlEncodedAdminURL := url.QueryEscape(SecureBaseURL + "/")
	s.Assert().Equal(fmt.Sprintf("<a href=\"%s\">302 Found</a>", utils.StringHTMLEscape(fmt.Sprintf("%s/?rd=%s&rm=GET", GetLoginBaseURL(BaseDomain), urlEncodedAdminURL))), string(body))

	req.Header.Set("X-Forwarded-Method", fasthttp.MethodOptions)

	res, err = client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(fasthttp.StatusOK, res.StatusCode)
}

func (s *StandaloneSuite) TestShouldRespondWithCorrectStatusCode() {
	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL(BaseDomain)), nil)
	s.Assert().NoError(err)
	req.Header.Set("X-Forwarded-Method", fasthttp.MethodGet)
	req.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	req.Header.Set(fasthttp.HeaderXForwardedHost, fmt.Sprintf("secure.%s", BaseDomain))
	req.Header.Set("X-Forwarded-URI", "/")
	req.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf8")

	client := NewHTTPClient()
	res, err := client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(fasthttp.StatusFound, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)

	urlEncodedAdminURL := url.QueryEscape(SecureBaseURL + "/")
	s.Assert().Equal(fmt.Sprintf("<a href=\"%s\">302 Found</a>", utils.StringHTMLEscape(fmt.Sprintf("%s/?rd=%s&rm=GET", GetLoginBaseURL(BaseDomain), urlEncodedAdminURL))), string(body))

	req.Header.Set("X-Forwarded-Method", fasthttp.MethodPost)

	res, err = client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(fasthttp.StatusSeeOther, res.StatusCode)
	body, err = io.ReadAll(res.Body)
	s.Assert().NoError(err)

	urlEncodedAdminURL = url.QueryEscape(SecureBaseURL + "/")
	s.Assert().Equal(fmt.Sprintf("<a href=\"%s\">303 See Other</a>", utils.StringHTMLEscape(fmt.Sprintf("%s/?rd=%s&rm=POST", GetLoginBaseURL(BaseDomain), urlEncodedAdminURL))), string(body))
}

// Standard case using nginx.
func (s *StandaloneSuite) TestShouldVerifyAPIVerifyUnauthorized() {
	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("%s/api/verify", AutheliaBaseURL), nil)
	s.Assert().NoError(err)
	req.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	req.Header.Set("X-Original-URL", AdminBaseURL)
	req.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf8")

	client := NewHTTPClient()
	res, err := client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(fasthttp.StatusUnauthorized, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Equal("401 Unauthorized", string(body))
}

// Standard case using Kubernetes.
func (s *StandaloneSuite) TestShouldVerifyAPIVerifyRedirectFromXOriginalURL() {
	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL(BaseDomain)), nil)
	s.Assert().NoError(err)
	req.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	req.Header.Set("X-Original-URL", AdminBaseURL)
	req.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf8")

	client := NewHTTPClient()
	res, err := client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(fasthttp.StatusFound, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)

	urlEncodedAdminURL := url.QueryEscape(AdminBaseURL)
	s.Assert().Equal(fmt.Sprintf("<a href=\"%s\">302 Found</a>", utils.StringHTMLEscape(fmt.Sprintf("%s/?rd=%s&rm=GET", GetLoginBaseURL(BaseDomain), urlEncodedAdminURL))), string(body))
}

func (s *StandaloneSuite) TestShouldVerifyAPIVerifyRedirectFromXOriginalHostURI() {
	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL(BaseDomain)), nil)
	s.Assert().NoError(err)
	req.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	req.Header.Set(fasthttp.HeaderXForwardedHost, "secure.example.com:8080")
	req.Header.Set("X-Forwarded-URI", "/")
	req.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf8")

	client := NewHTTPClient()
	res, err := client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(fasthttp.StatusFound, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)

	urlEncodedAdminURL := url.QueryEscape(SecureBaseURL + "/")
	s.Assert().Equal(fmt.Sprintf("<a href=\"%s\">302 Found</a>", utils.StringHTMLEscape(fmt.Sprintf("%s/?rd=%s&rm=GET", GetLoginBaseURL(BaseDomain), urlEncodedAdminURL))), string(body))
}

func (s *StandaloneSuite) TestShouldRecordMetrics() {
	client := NewHTTPClient()

	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("%s/api/health", LoginBaseURL), nil)
	s.Require().NoError(err)

	res, err := client.Do(req)
	s.Require().NoError(err)
	s.Assert().Equal(fasthttp.StatusOK, fasthttp.StatusOK, res.StatusCode)

	req, err = http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("%s/metrics", LoginBaseURL), nil)
	s.Require().NoError(err)

	res, err = client.Do(req)
	s.Require().NoError(err)
	s.Assert().Equal(fasthttp.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	s.Require().NoError(err)

	metrics := string(body)

	s.Assert().Contains(metrics, "authelia_request_duration_bucket{")
	s.Assert().Contains(metrics, "authelia_request_duration_sum{")
}

func (s *StandaloneSuite) TestStandaloneWebDriverScenario() {
	suite.Run(s.T(), NewStandaloneWebDriverSuite())
}

func (s *StandaloneSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *StandaloneSuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *StandaloneSuite) TestBypassPolicyScenario() {
	suite.Run(s.T(), NewBypassPolicyScenario())
}

func (s *StandaloneSuite) TestBackendProtectionScenario() {
	suite.Run(s.T(), NewBackendProtectionScenario())
}

func (s *StandaloneSuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *StandaloneSuite) TestRequestMethodScenario() {
	suite.Run(s.T(), NewRequestMethodScenario())
}

func (s *StandaloneSuite) TestAvailableMethodsScenario() {
	suite.Run(s.T(), NewAvailableMethodsScenario([]string{"TIME-BASED ONE-TIME PASSWORD", "SECURITY KEY - WEBAUTHN"}))
}

func (s *StandaloneSuite) TestRedirectionURLScenario() {
	suite.Run(s.T(), NewRedirectionURLScenario())
}

func (s *StandaloneSuite) TestRedirectionCheckScenario() {
	suite.Run(s.T(), NewRedirectionCheckScenario())
}

func TestStandaloneSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewStandaloneSuite())
}
