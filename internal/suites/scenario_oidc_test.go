// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	err := s.Page.MustSearch("Log in").Click("left", 1)
	assert.NoError(s.T(), err)

	s.verifyIsOpenIDConsentDecisionStage(s.T(), s.Context(ctx))
	s.verifyOpenIDConsentClientLogo(s.T(), s.Context(ctx))

	s.ClickElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-accept")

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

	err := s.Page.MustSearch("Log in").Click("left", 1)
	assert.NoError(s.T(), err)

	s.verifyIsOpenIDConsentDecisionStage(s.T(), s.Context(ctx))

	s.ClickElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-deny")

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
	data.Set("client_secret", clientSecret)
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
	s.verifyOpenIDConsentClientLogo(s.T(), s.Context(ctx))

	s.ClickElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-accept")

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

	assert.NotEmpty(s.T(), token, "Failed to obtain token after polling device authorization endpoint.")
	assert.Equal(s.T(), "bearer", token["token_type"])
	assert.True(s.T(), strings.HasPrefix(token["access_token"].(string), "authelia_at_"))
	assert.Equal(s.T(), scope, token["scope"])
	assert.NotEmpty(s.T(), token["id_token"])
}

func (s *OIDCScenario) TestShouldShowClientLogoForDeviceAuthorizationWithEnteredUserCode() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	c := NewHTTPClient()

	resp, err := c.PostForm(fmt.Sprintf("%s/api/oidc/device-authorization", LoginBaseURL), url.Values{
		"client_id":     []string{"device-code"},
		"client_secret": []string{"foobar"},
		"scope":         []string{"openid profile email groups"},
	})
	s.Require().NoError(err)

	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var device struct {
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
	}

	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&device))
	s.Require().NotEmpty(device.UserCode)
	s.Require().NotEmpty(device.VerificationURI)

	s.doVisit(s.T(), s.Context(ctx), device.VerificationURI)

	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.doFillLoginPageAndClick(s.T(), s.Context(ctx), testUsername, "password", false)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doValidateTOTP(s.T(), s.Context(ctx), testUsername)

	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-device-auth-stage")
	s.doFillFieldUntilSet(s.T(), s.WaitElementLocatedByID(s.T(), s.Context(ctx), "user-code"), device.UserCode)
	s.ClickElementLocatedByID(s.T(), s.Context(ctx), "confirm-button")

	s.verifyIsOpenIDConsentDecisionStage(s.T(), s.Context(ctx))
	s.verifyOpenIDConsentClientLogo(s.T(), s.Context(ctx))

	s.ClickElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-accept")

	s.verifyBodyContains(s.T(), s.Context(ctx), "Consent has been accepted and processed")
}

func (s *OIDCScenario) verifyOpenIDConsentClientLogo(tt *testing.T, page *rod.Page) {
	const src = "https://www.authelia.com/images/branding/logo.png"

	logo := s.WaitElementLocatedByID(tt, page, "openid-consent-client-logo")

	actual, err := logo.Attribute("src")
	require.NoError(tt, err)
	require.NotNil(tt, actual)
	assert.Equal(tt, src, *actual)

	assert.NoError(tt, logo.Timeout(10*time.Second).Wait(rod.Eval(`() => this.complete && this.naturalWidth > 0`)), "the client logo did not load")

	if os.Getenv("CI") != t {
		return
	}

	assert.Contains(tt, s.getEnforcedContentSecurityPolicy(tt, page), "img-src 'self' data: https://www.authelia.com;", "the consent screen does not enforce a policy allowing the logo host")

	info, err := page.Info()
	require.NoError(tt, err)

	consent, err := url.Parse(info.URL)
	require.NoError(tt, err)

	other := *consent
	other.Path = "/"

	c := NewHTTPClient()

	for _, u := range []*url.URL{consent, &other} {
		resp, err := c.Get(u.String())
		require.NoError(tt, err)

		_ = resp.Body.Close()

		policy := resp.Header.Get("Content-Security-Policy")

		require.NotEmpty(tt, policy, "no Content-Security-Policy for %s", u)

		if u == consent {
			assert.Contains(tt, policy, "img-src 'self' data: https://www.authelia.com;", "the logo host is not allowed for %s", u)
		} else {
			assert.NotContains(tt, policy, "https://www.authelia.com", "the logo host is allowed for %s", u)
		}
	}
}

func (s *OIDCScenario) getEnforcedContentSecurityPolicy(tt *testing.T, page *rod.Page) string {
	result, err := page.Eval(`() => new Promise((resolve) => {
		const timer = setTimeout(() => resolve(''), 5000);

		document.addEventListener('securitypolicyviolation', (event) => {
			if (event.effectiveDirective !== 'img-src') {
				return;
			}

			clearTimeout(timer);
			resolve(event.originalPolicy);
		});

		new Image().src = 'https://csp-probe.invalid/probe.png';
	})`)
	require.NoError(tt, err)

	policy := result.Value.Str()

	require.NotEmpty(tt, policy, "the page does not enforce a Content-Security-Policy for images")

	return policy
}

func TestRunOIDCScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOIDCSuite())
}
