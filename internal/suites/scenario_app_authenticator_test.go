// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/suites/appdevice"
)

const (
	appAuthenticatorTextValue = `(selector) => {
	const element = document.querySelector(selector);

	if (!element) {
		return '';
	}

	return (element.value || element.textContent || '').trim();
}`

	appAuthenticatorTextMatches = `(selector, pattern) => {
	const element = document.querySelector(selector);

	if (!element) {
		return false;
	}

	return new RegExp(pattern).test((element.value || element.textContent || '').trim());
}`
)

var (
	reAppAuthenticatorQR     = regexp.MustCompile(`^authelia-app://login\.example\.com:8080/api/register/app/[A-Za-z0-9_-]+$`)
	reAppAuthenticatorNumber = regexp.MustCompile(`^[0-9]{2,4}$`)
)

type AppAuthenticatorScenario struct {
	*RodSuite

	nonceLength int
}

func NewAppAuthenticatorScenario(nonceLength int) *AppAuthenticatorScenario {
	return &AppAuthenticatorScenario{
		RodSuite:    NewRodSuite(""),
		nonceLength: nonceLength,
	}
}

func (s *AppAuthenticatorScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *AppAuthenticatorScenario) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *AppAuthenticatorScenario) SetupTest() {
	s.doSetupTest(HomeBaseURL)
}

func (s *AppAuthenticatorScenario) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)

		s.collectCoverage(s.Page)
		s.MustClose()
	}()

	provider, err := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), err)

	devices, err := provider.LoadAppAuthenticatorDevicesByUsername(ctx, "john")
	require.NoError(s.T(), err)

	for _, device := range devices {
		require.NoError(s.T(), provider.DeleteAppAuthenticatorDevice(ctx, device.DeviceID))
	}

	require.NoError(s.T(), provider.SavePreferred2FAMethod(ctx, "john", "totp"))
}

func (s *AppAuthenticatorScenario) TestShouldRegisterAndSignIn() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SecureBaseURL)

	device := s.doRegisterAppDevice(s.T(), s.Context(ctx))

	s.doLogout(s.T(), s.Context(ctx))
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, targetURL)
	s.doChangeMethod(s.T(), s.Context(ctx), "app-authenticator")

	number := s.doWaitAppNumber(s.T(), s.Context(ctx), "#app-authenticator-number")

	challenge, err := device.Check()
	s.Require().NoError(err)
	s.Require().NotNil(challenge)
	s.Equal(s.nonceLength, challenge.NonceLength)

	result, err := device.Approve(number)
	s.Require().NoError(err)
	s.Equal("approved", result)

	s.verifySecretAuthorized(s.T(), s.Context(ctx))

	s.doLogout(s.T(), s.Context(ctx))
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, targetURL)

	number = s.doWaitAppNumber(s.T(), s.Context(ctx), "#app-authenticator-number")

	challenge, err = device.Check()
	s.Require().NoError(err)
	s.Require().NotNil(challenge)

	result, err = device.Approve(number)
	s.Require().NoError(err)
	s.Equal("approved", result)

	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func (s *AppAuthenticatorScenario) TestShouldShowDeniedWhenDeviceDenies() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	device := s.doRegisterAppDevice(s.T(), s.Context(ctx))

	s.doLogout(s.T(), s.Context(ctx))
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "app-authenticator")
	s.doWaitAppNumber(s.T(), s.Context(ctx), "#app-authenticator-number")

	challenge, err := device.Check()
	s.Require().NoError(err)
	s.Require().NotNil(challenge)

	result, err := device.Deny()
	s.Require().NoError(err)
	s.Equal("denied", result)

	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "app-authenticator-denied")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
}

func (s *AppAuthenticatorScenario) TestShouldNotApproveWithWrongNumber() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	device := s.doRegisterAppDevice(s.T(), s.Context(ctx))

	s.doLogout(s.T(), s.Context(ctx))
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "app-authenticator")

	number := s.doWaitAppNumber(s.T(), s.Context(ctx), "#app-authenticator-number")

	challenge, err := device.Check()
	s.Require().NoError(err)
	s.Require().NotNil(challenge)

	result, err := device.Approve(appAuthenticatorWrongNumber(number))
	s.Require().NoError(err)
	s.Equal("rejected", result)

	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "app-authenticator-not-approved")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
}

func (s *AppAuthenticatorScenario) TestShouldRemoveDeviceInSettings() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	device := s.doRegisterAppDevice(s.T(), s.Context(ctx))

	s.ClickElementLocatedByID(s.T(), s.Context(ctx), "app-authenticator-device-0-delete")
	s.doMaybeVerifyIdentity(s.T(), s.Context(ctx), "#dialog-delete")
	s.ClickElementLocatedByID(s.T(), s.Context(ctx), "dialog-delete")
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Successfully deleted the Authelia app device")
	s.WaitElementLocatedBySelector(s.T(), s.Context(ctx), `#app-authenticator-panel[data-loading="false"]`)
	s.Require().False(s.CheckElementExistsLocatedByID(s.T(), s.Context(ctx), "app-authenticator-device-0"))

	_, err := device.Check()
	s.Require().Error(err)

	var herr *appdevice.HTTPError

	s.Require().True(errors.As(err, &herr))
	s.Equal(http.StatusUnauthorized, herr.Status)
	s.Equal("bad_signature", herr.Code)
}

func (s *AppAuthenticatorScenario) doRegisterAppDevice(t *testing.T, page *rod.Page) *appdevice.Device {
	device := appdevice.New(appdevice.WithHTTPClient(newAppAuthenticatorHTTPClient(t)))

	s.doLoginOneFactor(t, page, "john", "password", false, BaseDomain, "")
	s.doOpenSettings(t, page)
	s.doOpenSettingsMenuClickTwoFactor(t, page)
	s.WaitElementLocatedBySelector(t, page, `#app-authenticator-panel[data-loading="false"]`)
	s.ClickElementLocatedByID(t, page, "app-authenticator-device-add")
	s.doMaybeVerifyIdentity(t, page, "#app-authenticator-qr")

	qr := s.doWaitAppText(t, page, "#app-authenticator-qr", reAppAuthenticatorQR)

	begin, err := device.RegisterBegin(qr)
	require.NoError(t, err)
	require.Equal(t, s.nonceLength, begin.NonceLength)
	require.Equal(t, "none", begin.Attestation)
	require.Equal(t, LoginBaseURL, begin.BaseURL)

	number := s.doWaitAppNumber(t, page, "#app-authenticator-register-number")

	require.NoError(t, device.RegisterConfirm(number))
	require.Equal(t, "john", device.Username)

	s.verifyNotificationDisplayed(t, page, "Successfully added the Authelia app device")
	s.WaitElementLocatedByID(t, page, "app-authenticator-device-0")

	return device
}

func (s *AppAuthenticatorScenario) doWaitAppNumber(t *testing.T, page *rod.Page, selector string) string {
	number := s.doWaitAppText(t, page, selector, reAppAuthenticatorNumber)

	require.Len(t, number, s.nonceLength)

	return number
}

func (s *AppAuthenticatorScenario) doWaitAppText(t *testing.T, page *rod.Page, selector string, pattern *regexp.Regexp) string {
	s.WaitElementLocatedBySelector(t, page, selector)

	if err := page.Timeout(elementLocateTimeout).Wait(rod.Eval(appAuthenticatorTextMatches, selector, pattern.String())); err != nil {
		observed := "unavailable"

		if result, errEval := page.Eval(appAuthenticatorTextValue, selector); errEval == nil {
			observed = result.Value.Str()
		}

		require.Failf(t, "Element text did not match", "expected '%s' to match '%s', observed '%s': %v", selector, pattern, observed, err)
	}

	result, err := page.Eval(appAuthenticatorTextValue, selector)
	require.NoError(t, err)

	return result.Value.Str()
}

func appAuthenticatorWrongNumber(number string) string {
	wrong := []byte(number)

	wrong[0] = '0' + (wrong[0]-'0'+1)%10

	return string(wrong)
}

func newAppAuthenticatorHTTPClient(t *testing.T) *http.Client {
	data, err := os.ReadFile("common/pki/ca.public.crt")
	require.NoError(t, err)

	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(data))

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    pool,
				MinVersion: tls.VersionTLS12,
			},
			DialContext: DialContext,
		},
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
