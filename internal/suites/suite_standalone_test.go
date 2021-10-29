package suites

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/poy/onpar"

	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestStandaloneSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	s := setupTest(t, "", true)
	teardownTest(s)

	o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
		s := setupTest(t, "", false)
		return t, s
	})

	o.AfterEach(func(t *testing.T, s RodSuite) {
		teardownTest(s)
	})

	o.Spec("TestShouldLetUserKnowHeIsAlreadyAuthenticated", func(t *testing.T, s RodSuite) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
		}()

		s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, "")

		// Visit home page to change context.
		s.doVisit(s.Context(ctx), HomeBaseURL)
		s.verifyIsHome(t, s.Context(ctx))

		// Visit the login page and wait for redirection to 2FA page with success icon displayed.
		s.doVisit(s.Context(ctx), GetLoginBaseURL())
		s.verifyIsAuthenticatedPage(t, s.Context(ctx))
	})

	o.Spec("TestShouldRedirectAlreadyAuthenticatedUser", func(t *testing.T, s RodSuite) {
		is := is.New(t)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
		}()

		s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, "")

		// Visit home page to change context.
		s.doVisit(s.Context(ctx), HomeBaseURL)
		s.verifyIsHome(t, s.Context(ctx))

		// Visit the login page and wait for redirection to 2FA page with success icon displayed.
		s.doVisit(s.Context(ctx), fmt.Sprintf("%s?rd=https://secure.example.com:8080", GetLoginBaseURL()))

		_, err := s.Page.ElementR("h1", "Public resource")
		is.NoErr(err)
		s.verifyURLIs(t, s.Context(ctx), "https://secure.example.com:8080/")
	})

	o.Spec("TestShouldNotRedirectAlreadyAuthenticatedUserToUnsafeURL", func(t *testing.T, s RodSuite) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
		}()

		s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, "")

		// Visit home page to change context.
		s.doVisit(s.Context(ctx), HomeBaseURL)
		s.verifyIsHome(t, s.Context(ctx))

		// Visit the login page and wait for redirection to 2FA page with success icon displayed.
		s.doVisit(s.Context(ctx), fmt.Sprintf("%s?rd=https://secure.example.local:8080", GetLoginBaseURL()))
		s.verifyNotificationDisplayed(t, s.Context(ctx), "Redirection was determined to be unsafe and aborted. Ensure the redirection URL is correct.")
	})

	o.Spec("TestShouldRespectMethodsACL", func(t *testing.T, s RodSuite) {
		is := is.New(t)
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL()), nil)
		is.NoErr(err)
		req.Header.Set("X-Forwarded-Method", "GET")
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", fmt.Sprintf("secure.%s", BaseDomain))
		req.Header.Set("X-Forwarded-URI", "/")
		req.Header.Set("Accept", "text/html; charset=utf8")

		client := NewHTTPClient()
		res, err := client.Do(req)
		is.NoErr(err)
		is.Equal(res.StatusCode, 302)
		body, err := io.ReadAll(res.Body)
		is.NoErr(err)

		urlEncodedAdminURL := url.QueryEscape(SecureBaseURL + "/")
		is.Equal(fmt.Sprintf("<a href=\"%s\">Found</a>", utils.StringHTMLEscape(fmt.Sprintf("%s/?rd=%s&rm=GET", GetLoginBaseURL(), urlEncodedAdminURL))), string(body))

		req.Header.Set("X-Forwarded-Method", "OPTIONS")

		res, err = client.Do(req)
		is.NoErr(err)
		is.Equal(res.StatusCode, 200)
	})

	o.Spec("TestShouldRespondWithCorrectStatusCode", func(t *testing.T, s RodSuite) {
		is := is.New(t)
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL()), nil)
		is.NoErr(err)
		req.Header.Set("X-Forwarded-Method", "GET")
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", fmt.Sprintf("secure.%s", BaseDomain))
		req.Header.Set("X-Forwarded-URI", "/")
		req.Header.Set("Accept", "text/html; charset=utf8")

		client := NewHTTPClient()
		res, err := client.Do(req)
		is.NoErr(err)
		is.Equal(res.StatusCode, 302)
		body, err := io.ReadAll(res.Body)
		is.NoErr(err)

		urlEncodedAdminURL := url.QueryEscape(SecureBaseURL + "/")
		is.Equal(fmt.Sprintf("<a href=\"%s\">Found</a>", utils.StringHTMLEscape(fmt.Sprintf("%s/?rd=%s&rm=GET", GetLoginBaseURL(), urlEncodedAdminURL))), string(body))

		req.Header.Set("X-Forwarded-Method", "POST")

		res, err = client.Do(req)
		is.NoErr(err)
		is.Equal(res.StatusCode, 303)
		body, err = io.ReadAll(res.Body)
		is.NoErr(err)

		urlEncodedAdminURL = url.QueryEscape(SecureBaseURL + "/")
		is.Equal(fmt.Sprintf("<a href=\"%s\">See Other</a>", utils.StringHTMLEscape(fmt.Sprintf("%s/?rd=%s&rm=POST", GetLoginBaseURL(), urlEncodedAdminURL))), string(body))
	})

	// Standard case using nginx.
	o.Spec("TestShouldVerifyAPIVerifyUnauthorized", func(t *testing.T, s RodSuite) {
		is := is.New(t)
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/verify", AutheliaBaseURL), nil)
		is.NoErr(err)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Original-URL", AdminBaseURL)
		req.Header.Set("Accept", "text/html; charset=utf8")

		client := NewHTTPClient()
		res, err := client.Do(req)
		is.NoErr(err)
		is.Equal(res.StatusCode, 401)
		body, err := io.ReadAll(res.Body)
		is.NoErr(err)
		is.Equal("Unauthorized", string(body))
	})

	// Standard case using Kubernetes.
	o.Spec("TestShouldVerifyAPIVerifyRedirectFromXOriginalURL", func(t *testing.T, s RodSuite) {
		is := is.New(t)
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL()), nil)
		is.NoErr(err)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Original-URL", AdminBaseURL)
		req.Header.Set("Accept", "text/html; charset=utf8")

		client := NewHTTPClient()
		res, err := client.Do(req)
		is.NoErr(err)
		is.Equal(res.StatusCode, 302)
		body, err := io.ReadAll(res.Body)
		is.NoErr(err)

		urlEncodedAdminURL := url.QueryEscape(AdminBaseURL)
		is.Equal(fmt.Sprintf("<a href=\"%s\">Found</a>", utils.StringHTMLEscape(fmt.Sprintf("%s/?rd=%s", GetLoginBaseURL(), urlEncodedAdminURL))), string(body))
	})

	o.Spec("TestShouldVerifyAPIVerifyRedirectFromXOriginalHostURI", func(t *testing.T, s RodSuite) {
		is := is.New(t)
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL()), nil)
		is.NoErr(err)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "secure.example.com:8080")
		req.Header.Set("X-Forwarded-URI", "/")
		req.Header.Set("Accept", "text/html; charset=utf8")

		client := NewHTTPClient()
		res, err := client.Do(req)
		is.NoErr(err)
		is.Equal(res.StatusCode, 302)
		body, err := io.ReadAll(res.Body)
		is.NoErr(err)

		urlEncodedAdminURL := url.QueryEscape(SecureBaseURL + "/")
		is.Equal(fmt.Sprintf("<a href=\"%s\">Found</a>", utils.StringHTMLEscape(fmt.Sprintf("%s/?rd=%s", GetLoginBaseURL(), urlEncodedAdminURL))), string(body))
	})

	methods = []string{"TIME-BASED ONE-TIME PASSWORD", "SECURITY KEY - WEBAUTHN"}

	TestRun1FAScenario(t)
	TestRun2FAScenario(t)
	TestRunBypassPolicyScenario(t)
	TestRunBackendProtectionScenario(t)
	TestRunResetPasswordScenario(t)
	TestRunAvailableMethodsScenario(t)
	TestRunRedirectionURLScenario(t)
	TestRunRedirectionCheckScenario(t)
	t.Run("TestShouldCheckUserIsAskedToRegisterDevice", TestShouldCheckUserIsAskedToRegisterDevice)
}

func TestShouldCheckUserIsAskedToRegisterDevice(t *testing.T) {
	s := setupTest(t, "", false)
	is := is.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownTest(s)
	}()

	// Clean up any TOTP secret already in DB.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	is.NoErr(provider.DeleteTOTPConfiguration(ctx, testUsername))

	// Login one factor.
	s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")

	// Check the user is asked to register a new device.
	s.WaitElementLocatedByClassName(t, s.Context(ctx), "state-not-registered")

	// Then register the TOTP factor.
	secret = s.doRegisterTOTP(t, s.Context(ctx))
	// And logout.
	s.doLogout(t, s.Context(ctx))

	// Login one factor again.
	s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")

	// now the user should be asked to perform 2FA.
	s.WaitElementLocatedByClassName(t, s.Context(ctx), "state-method")
}
