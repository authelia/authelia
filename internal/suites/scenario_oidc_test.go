package suites

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/oidc"
)

type OIDCScenario struct {
	*RodSuite
}

func NewOIDCScenario() *OIDCScenario {
	return &OIDCScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *OIDCScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)

		s.collectCoverage(s.Page)
		s.MustClose()
	}()

	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.doRegisterTOTPAndLogin2FA(s.T(), s.Context(ctx), "john", "password", false, AdminBaseURL)
}

func (s *OIDCScenario) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *OIDCScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.Page = s.doCreateTab(s.T(), fmt.Sprintf("%s/logout", OIDCBaseURL))
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (s *OIDCScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *OIDCScenario) TestShouldAuthorizeAccessToOIDCApp() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), OIDCBaseURL)
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.doFillLoginPageAndClick(s.T(), s.Context(ctx), testUsername, "password", false)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doValidateTOTP(s.T(), s.Context(ctx), testUsername)

	s.waitBodyContains(s.T(), s.Context(ctx), "Not logged yet...")

	// Search for the 'login' link.
	err := s.Page.MustSearch("Log in").Click("left", 1)
	assert.NoError(s.T(), err)

	s.verifyIsOpenIDConsentDecisionStage(s.T(), s.Context(ctx))
	err = s.WaitElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-accept").Click("left", 1)
	assert.NoError(s.T(), err)

	// Verify that the app is showing the info related to the user stored in the JWT token.

	rAuthCodeURL := regexp.MustCompile(`/oauth2/callback\?code=authelia_ac_([^&=]+)&iss=https%3A%2F%2Flogin\.example\.com%3A8080&scope=openid\+profile\+email\+groups&state=random-string-here$`)
	rUUID := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	rInteger := regexp.MustCompile(`^\d+$`)
	rBoolean := regexp.MustCompile(`^(true|false)$`)
	rBase64 := regexp.MustCompile(`^[-_A-Za-z0-9+\\/]+([=]{0,3})$`)

	testCases := []struct {
		desc, elementID string
		expected        any
	}{
		{"welcome", "welcome", "Logged in as john!"},
		{"AuthorizeCodeURL", "auth-code-url", rAuthCodeURL},
		{oidc.ClaimAccessTokenHash, "", rBase64},
		{oidc.ClaimJWTID, "", rUUID},
		{oidc.ClaimIssuedAt, "", rInteger},
		{oidc.ClaimSubject, "", rUUID},
		{oidc.ClaimNotBefore, "", rInteger},
		{oidc.ClaimRequestedAt, "", rInteger},
		{oidc.ClaimExpirationTime, "", rInteger},
		{oidc.ClaimAuthenticationMethodsReference, "", "pwd, kba, otp, mfa"},
		{oidc.ClaimAuthenticationContextClassReference, "", ""},
		{oidc.ClaimIssuer, "", "https://login.example.com:8080"},
		{oidc.ClaimFullName, "", "John Doe"},
		{oidc.ClaimPreferredUsername, "", "john"},
		{oidc.ClaimGroups, "", "admins, dev"},
		{oidc.ClaimEmail, "", "john.doe@authelia.com"},
		{oidc.ClaimEmailVerified, "", rBoolean},
	}

	var actual string

	for _, tc := range testCases {
		s.T().Run(fmt.Sprintf("check_claims/%s", tc.desc), func(t *testing.T) {
			switch tc.elementID {
			case "":
				actual, err = s.WaitElementLocatedByID(t, s.Context(ctx), "claim-"+tc.desc).Text()
			default:
				actual, err = s.WaitElementLocatedByID(t, s.Context(ctx), tc.elementID).Text()
			}

			assert.NoError(t, err)

			switch expected := tc.expected.(type) {
			case *regexp.Regexp:
				assert.Regexp(t, expected, actual)
			default:
				assert.Equal(t, expected, actual)
			}
		})
	}
}

func (s *OIDCScenario) TestShouldDenyConsent() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisit(s.T(), s.Context(ctx), OIDCBaseURL)
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.doFillLoginPageAndClick(s.T(), s.Context(ctx), testUsername, "password", false)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doValidateTOTP(s.T(), s.Context(ctx), testUsername)

	s.waitBodyContains(s.T(), s.Context(ctx), "Not logged yet...")

	// Search for the 'login' link.
	err := s.Page.MustSearch("Log in").Click("left", 1)
	assert.NoError(s.T(), err)

	s.verifyIsOpenIDConsentDecisionStage(s.T(), s.Context(ctx))

	err = s.WaitElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-deny").Click("left", 1)
	assert.NoError(s.T(), err)

	s.verifyIsOIDC(s.T(), s.Context(ctx), "access_denied", "https://oidc.example.com:8080/error?error=access_denied&error_description=The+resource+owner+or+authorization+server+denied+the+request.+Make+sure+that+the+request+you+are+making+is+valid.+Maybe+the+credential+or+request+parameters+you+are+using+are+limited+in+scope+or+otherwise+restricted.&iss=https%3A%2F%2Flogin.example.com%3A8080&state=random-string-here")

	errorDescription := "The resource owner or authorization server denied the request. Make sure that the request " +
		"you are making is valid. Maybe the credential or request parameters you are using are limited in scope or " +
		"otherwise restricted."

	s.verifyIsOIDCErrorPage(s.T(), s.Context(ctx), "access_denied", errorDescription, "",
		"random-string-here")
}

func (s *OIDCScenario) TestShouldIssueDeviceAuthorizationBearerToken() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	c := NewHTTPClient()
	clientID := "device-code"
	clientSecret := "foobar"
	scope := "openid profile email groups"

	metadataURL := fmt.Sprintf("%s/.well-known/openid-configuration", LoginBaseURL)
	resp, err := c.Get(metadataURL)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	var metadata map[string]interface{}

	err = json.Unmarshal(body, &metadata)
	assert.NoError(s.T(), err)

	deviceAuthEndpoint, ok := metadata["device_authorization_endpoint"].(string)
	assert.True(s.T(), ok)

	tokenEndpoint, ok := metadata["token_endpoint"].(string)
	assert.True(s.T(), ok)

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", scope)

	deviceResp, err := c.PostForm(deviceAuthEndpoint, data)
	assert.NoError(s.T(), err)

	defer deviceResp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, deviceResp.StatusCode)

	deviceBody, err := io.ReadAll(deviceResp.Body)
	assert.NoError(s.T(), err)

	var deviceData map[string]interface{}

	err = json.Unmarshal(deviceBody, &deviceData)
	assert.NoError(s.T(), err)

	deviceCode, ok := deviceData["device_code"].(string)
	assert.True(s.T(), ok)

	_, ok = deviceData["user_code"].(string)
	assert.True(s.T(), ok)

	_, ok = deviceData["verification_uri"].(string)
	assert.True(s.T(), ok)

	verificationURIComplete, ok := deviceData["verification_uri_complete"].(string)
	assert.True(s.T(), ok)

	s.doVisit(s.T(), s.Context(ctx), verificationURIComplete)

	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.doFillLoginPageAndClick(s.T(), s.Context(ctx), testUsername, "password", false)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doValidateTOTP(s.T(), s.Context(ctx), testUsername)

	s.verifyIsOpenIDConsentDecisionStage(s.T(), s.Context(ctx))
	err = s.WaitElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-accept").Click("left", 1)
	assert.NoError(s.T(), err)

	s.verifyBodyContains(s.T(), s.Context(ctx), "Consent has been accepted and processed")

	var token map[string]interface{}

	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)

		tokenData := url.Values{}
		tokenData.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
		tokenData.Set("device_code", deviceCode)
		tokenData.Set("client_id", clientID)
		tokenData.Set("client_secret", clientSecret)

		tokenResp, err := c.PostForm(tokenEndpoint, tokenData)
		if err != nil {
			continue
		}

		tokenBody, err := io.ReadAll(tokenResp.Body)
		tokenResp.Body.Close()

		if err != nil {
			continue
		}

		if tokenResp.StatusCode == http.StatusOK {
			err = json.Unmarshal(tokenBody, &token)
			if err != nil {
				continue
			}

			break
		}
	}

	assert.Equal(s.T(), "bearer", token["token_type"])
	assert.True(s.T(), strings.HasPrefix(token["access_token"].(string), "authelia_at_"))
	assert.Equal(s.T(), scope, token["scope"])
	assert.NotEmpty(s.T(), token["id_token"])
}

func TestRunOIDCScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOIDCSuite())
}
