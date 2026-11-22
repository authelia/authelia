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

	endpoints := s.getOIDCEndpoints(s.T(), c)

	device := s.doRequestDeviceAuthorization(s.T(), c, endpoints)

	s.doVisit(s.T(), s.Context(ctx), device.VerificationURIComplete)

	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.doFillLoginPageAndClick(s.T(), s.Context(ctx), testUsername, "password", false)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doValidateTOTP(s.T(), s.Context(ctx), testUsername)

	s.verifyIsOpenIDConsentDecisionStage(s.T(), s.Context(ctx))
	s.ClickElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-accept")

	s.verifyBodyContains(s.T(), s.Context(ctx), "Consent has been accepted and processed")

	s.verifyDeviceAuthorizationBearerToken(s.T(), s.doPollDeviceAuthorizationToken(s.T(), c, endpoints, device))
}

func (s *OIDCScenario) TestShouldAuthorizeConcurrentDeviceAuthorizationsInSeparateTabs() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)

	var page *rod.Page

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)

		if page != nil {
			s.collectScreenshot(ctx.Err(), page)
			page.MustClose()
		}
	}()

	c := NewHTTPClient()

	endpoints := s.getOIDCEndpoints(s.T(), c)

	first := s.doRequestDeviceAuthorization(s.T(), c, endpoints)
	second := s.doRequestDeviceAuthorization(s.T(), c, endpoints)

	require.NotEqual(s.T(), first.UserCode, second.UserCode)

	s.doVisit(s.T(), s.Context(ctx), first.VerificationURI)

	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.doFillLoginPageAndClick(s.T(), s.Context(ctx), testUsername, "password", false)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doValidateTOTP(s.T(), s.Context(ctx), testUsername)

	page = s.Browser().MustPage(second.VerificationURI)

	s.doEnterDeviceAuthorizationUserCode(s.T(), s.Context(ctx), first.UserCode)

	page.MustActivate()
	s.doEnterDeviceAuthorizationUserCode(s.T(), page.Context(ctx), second.UserCode)

	s.MustActivate()
	s.ClickElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-accept")
	s.verifyBodyContains(s.T(), s.Context(ctx), "Consent has been accepted and processed")

	s.verifyDeviceAuthorizationBearerToken(s.T(), s.doPollDeviceAuthorizationToken(s.T(), c, endpoints, first))

	status, pending := s.doRequestDeviceAuthorizationToken(s.T(), c, endpoints, second)

	assert.Equal(s.T(), http.StatusBadRequest, status)
	assert.Equal(s.T(), "authorization_pending", pending["error"])

	page.MustActivate()
	s.ClickElementLocatedByID(s.T(), page.Context(ctx), "openid-consent-accept")
	s.verifyBodyContains(s.T(), page.Context(ctx), "Consent has been accepted and processed")

	s.verifyDeviceAuthorizationBearerToken(s.T(), s.doPollDeviceAuthorizationToken(s.T(), c, endpoints, second))
}

func (s *OIDCScenario) TestShouldAuthorizeConcurrentAuthorizationRequestsInSeparateTabs() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	var page *rod.Page

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)

		if page != nil {
			s.collectScreenshot(ctx.Err(), page)
			page.MustClose()
		}
	}()

	s.doVisit(s.T(), s.Context(ctx), OIDCBaseURL)
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.doFillLoginPageAndClick(s.T(), s.Context(ctx), testUsername, "password", false)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doValidateTOTP(s.T(), s.Context(ctx), testUsername)

	s.waitBodyContains(s.T(), s.Context(ctx), "Not logged yet...")

	page = s.Browser().MustPage(OIDCBaseURL)

	s.waitBodyContains(s.T(), page.Context(ctx), "Not logged yet...")

	s.MustActivate()
	require.NoError(s.T(), s.Page.MustSearch("Log in").Click("left", 1))
	s.verifyIsOpenIDConsentDecisionStage(s.T(), s.Context(ctx))

	page.MustActivate()
	require.NoError(s.T(), page.MustSearch("Log in").Click("left", 1))
	s.verifyIsOpenIDConsentDecisionStage(s.T(), page.Context(ctx))

	s.MustActivate()
	s.ClickElementLocatedByID(s.T(), s.Context(ctx), "openid-consent-accept")

	first := s.verifyOIDCAppAuthorized(s.T(), s.Context(ctx))

	page.MustActivate()
	s.ClickElementLocatedByID(s.T(), page.Context(ctx), "openid-consent-accept")

	second := s.verifyOIDCAppAuthorized(s.T(), page.Context(ctx))

	assert.NotEqual(s.T(), first, second)
}

type oidcEndpoints struct {
	DeviceAuthorization string
	Token               string
}

type oidcDeviceAuthorization struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
}

const (
	oidcDeviceCodeClientID     = "device-code"
	oidcDeviceCodeClientSecret = "foobar"
	oidcDeviceCodeScope        = "openid profile email groups"
)

func (s *OIDCScenario) getOIDCEndpoints(t *testing.T, c *http.Client) (endpoints oidcEndpoints) {
	resp, err := c.Get(fmt.Sprintf("%s/.well-known/openid-configuration", LoginBaseURL))
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var metadata struct {
		DeviceAuthorization string `json:"device_authorization_endpoint"`
		Token               string `json:"token_endpoint"`
	}

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&metadata))
	require.NotEmpty(t, metadata.DeviceAuthorization)
	require.NotEmpty(t, metadata.Token)

	return oidcEndpoints{DeviceAuthorization: metadata.DeviceAuthorization, Token: metadata.Token}
}

func (s *OIDCScenario) doRequestDeviceAuthorization(t *testing.T, c *http.Client, endpoints oidcEndpoints) (device oidcDeviceAuthorization) {
	data := url.Values{}
	data.Set("client_id", oidcDeviceCodeClientID)
	data.Set("client_secret", oidcDeviceCodeClientSecret)
	data.Set("scope", oidcDeviceCodeScope)

	resp, err := c.PostForm(endpoints.DeviceAuthorization, data)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&device))

	require.NotEmpty(t, device.DeviceCode)
	require.NotEmpty(t, device.UserCode)
	require.NotEmpty(t, device.VerificationURI)
	require.NotEmpty(t, device.VerificationURIComplete)

	return device
}

func (s *OIDCScenario) doRequestDeviceAuthorizationToken(t *testing.T, c *http.Client, endpoints oidcEndpoints, device oidcDeviceAuthorization) (status int, body map[string]any) {
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", device.DeviceCode)
	data.Set("client_id", oidcDeviceCodeClientID)
	data.Set("client_secret", oidcDeviceCodeClientSecret)

	resp, err := c.PostForm(endpoints.Token, data)
	require.NoError(t, err)

	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(raw, &body))

	return resp.StatusCode, body
}

func (s *OIDCScenario) doPollDeviceAuthorizationToken(t *testing.T, c *http.Client, endpoints oidcEndpoints, device oidcDeviceAuthorization) (token map[string]any) {
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)

		if status, body := s.doRequestDeviceAuthorizationToken(t, c, endpoints, device); status == http.StatusOK {
			return body
		}
	}

	require.Fail(t, "Failed to obtain token after polling device authorization endpoint.")

	return nil
}

func (s *OIDCScenario) doEnterDeviceAuthorizationUserCode(t *testing.T, page *rod.Page, userCode string) {
	s.WaitElementLocatedByID(t, page, "openid-consent-device-auth-stage")
	s.doFillFieldUntilSet(t, s.WaitElementLocatedByID(t, page, "user-code"), userCode)
	s.ClickElementLocatedByID(t, page, "confirm-button")
	s.verifyIsOpenIDConsentDecisionStage(t, page)
}

func (s *OIDCScenario) verifyDeviceAuthorizationBearerToken(t *testing.T, token map[string]any) {
	require.NotEmpty(t, token)
	assert.Equal(t, "bearer", token["token_type"])

	accessToken, ok := token["access_token"].(string)
	require.True(t, ok)

	assert.True(t, strings.HasPrefix(accessToken, "authelia_at_"))
	assert.Equal(t, oidcDeviceCodeScope, token["scope"])
	assert.NotEmpty(t, token["id_token"])
}

func (s *OIDCScenario) verifyOIDCAppAuthorized(t *testing.T, page *rod.Page) (authCodeURL string) {
	welcome, err := s.WaitElementLocatedByID(t, page, "welcome").Text()
	require.NoError(t, err)
	assert.Equal(t, "Logged in as john!", welcome)

	authCodeURL, err = s.WaitElementLocatedByID(t, page, "auth-code-url").Text()
	require.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`/oauth2/callback\?code=authelia_ac_([^&=]+)&`), authCodeURL)

	return authCodeURL
}

func TestRunOIDCScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOIDCSuite())
}
