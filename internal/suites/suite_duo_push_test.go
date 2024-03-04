package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

type DuoPushWebDriverSuite struct {
	*RodSuite
}

func NewDuoPushWebDriverSuite() *DuoPushWebDriverSuite {
	return &DuoPushWebDriverSuite{
		RodSuite: NewRodSuite(""),
	}
}

func (s *DuoPushWebDriverSuite) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *DuoPushWebDriverSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *DuoPushWebDriverSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *DuoPushWebDriverSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)

		s.collectCoverage(s.Page)
		s.MustClose()
	}()

	// Set default 2FA preference and clean up any Duo device already in DB.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferred2FAMethod(ctx, "john", "totp"))
	require.NoError(s.T(), provider.DeletePreferredDuoDevice(ctx, "john"))
}

func (s *DuoPushWebDriverSuite) TestShouldBypassDeviceSelection() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "allow",
		StatusMessage: "Allowing unknown user",
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (s *DuoPushWebDriverSuite) TestShouldDenyDeviceSelection() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "deny",
		StatusMessage: "We're sorry, access is not allowed.",
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Device selection was denied by Duo policy")
}

func (s *DuoPushWebDriverSuite) TestShouldAskUserToRegister() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:          "enroll",
		EnrollPortalURL: "https://api-example.duosecurity.com/portal?code=1234567890ABCDEF&akey=12345ABCDEFGHIJ67890",
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.WaitElementLocatedByClassName(s.T(), s.Context(ctx), "state-not-registered")
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "No compatible device found")
	enrollPage := s.Page.MustWaitOpen()
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "register-link").MustClick()
	s.Page = enrollPage()

	assert.Contains(s.T(), s.WaitElementLocatedByClassName(s.T(), s.Context(ctx), "description").MustText(), "This enrollment code has expired. Contact your administrator to get a new enrollment code.")
}

func (s *DuoPushWebDriverSuite) TestShouldAutoSelectDevice() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"auto", "push", "sms", "mobile_otp"},
		}},
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Allow)

	// Authenticate.
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	// Switch Method where single Device should be selected automatically.
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.verifyIsHome(s.T(), s.Context(ctx))

	// Re-Login the user.
	s.doLogout(s.T(), s.Context(ctx))
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	// And check the latest method and device is still used.
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "push-notification-method")
	// Meaning the authentication is successful.
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (s *DuoPushWebDriverSuite) TestShouldSelectDevice() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set default 2FA preference to enable Select Device link in frontend.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "ABCDEFGHIJ1234567890", Method: "push"}))

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"auto", "push", "sms", "mobile_otp"},
		}, {
			Device:       "1234567890ABCDEFGHIJ",
			DisplayName:  "Test Device 2",
			Capabilities: []string{"auto", "push", "sms", "mobile_otp"},
		}},
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Allow)

	// Authenticate.
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	// Switch Method where Device Selection should open automatically.
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	// Check for available Device 1.
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "device-12345ABCDEFGHIJ67890")
	// Test Back button.
	s.doClickButton(s.T(), s.Context(ctx), "device-selection-back")
	// then select Device 2 for further use and be redirected.
	s.doChangeDevice(s.T(), s.Context(ctx), "1234567890ABCDEFGHIJ")
	s.verifyIsHome(s.T(), s.Context(ctx))

	// Re-Login the user.
	s.doLogout(s.T(), s.Context(ctx))
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	// And check the latest method and device is still used.
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "push-notification-method")
	// Meaning the authentication is successful.
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (s *DuoPushWebDriverSuite) TestShouldFailInitialSelectionBecauseOfUnsupportedMethod() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"auto", "sms"},
		}},
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.WaitElementLocatedByClassName(s.T(), s.Context(ctx), "state-not-registered")
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "No compatible device found")
}

func (s *DuoPushWebDriverSuite) TestShouldSelectNewDeviceAfterSavedDeviceMethodIsNoLongerSupported() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"push", "sms"},
		}, {
			Device:       "1234567890ABCDEFGHIJ",
			DisplayName:  "Test Device 2",
			Capabilities: []string{"auto", "push", "sms", "mobile_otp"},
		}},
	}

	// Setup unsupported Duo device in DB.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "ABCDEFGHIJ1234567890", Method: "sms"}))
	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Allow)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "device-selection")
	s.doSelectDevice(s.T(), s.Context(ctx), "12345ABCDEFGHIJ67890")
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (s *DuoPushWebDriverSuite) TestShouldAutoSelectNewDeviceAfterSavedDeviceIsNoLongerAvailable() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"push", "sms"},
		}},
	}

	// Setup unsupported Duo device in DB.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "ABCDEFGHIJ1234567890", Method: "push"}))
	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Allow)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (s *DuoPushWebDriverSuite) TestShouldFailSelectionBecauseOfSelectionBypassed() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "allow",
		StatusMessage: "Allowing unknown user",
	}

	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Deny)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.doClickButton(s.T(), s.Context(ctx), "selection-link")
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Device selection was bypassed by Duo policy")
}

func (s *DuoPushWebDriverSuite) TestShouldFailSelectionBecauseOfSelectionDenied() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "deny",
		StatusMessage: "We're sorry, access is not allowed.",
	}

	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Deny)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	err := s.WaitElementLocatedByID(s.T(), s.Context(ctx), "selection-link").Click("left", 1)
	require.NoError(s.T(), err)
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Device selection was denied by Duo policy")
}

func (s *DuoPushWebDriverSuite) TestShouldFailAuthenticationBecausePreauthDenied() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "deny",
		StatusMessage: "We're sorry, access is not allowed.",
	}

	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.WaitElementLocatedByClassName(s.T(), s.Context(ctx), "failure-icon")
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "There was an issue completing sign in process")
}

func (s *DuoPushWebDriverSuite) TestShouldSucceedAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"auto", "push", "sms", "mobile_otp"},
		}},
	}

	// Setup Duo device in DB.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Allow)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (s *DuoPushWebDriverSuite) TestShouldFailAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"auto", "push", "sms", "mobile_otp"},
		}},
	}

	// Setup Duo device in DB.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Deny)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.WaitElementLocatedByClassName(s.T(), s.Context(ctx), "failure-icon")
}

type DuoPushDefaultRedirectionSuite struct {
	*RodSuite
}

func NewDuoPushDefaultRedirectionSuite() *DuoPushDefaultRedirectionSuite {
	return &DuoPushDefaultRedirectionSuite{RodSuite: NewRodSuite(duoPushSuiteName)}
}

func (s *DuoPushDefaultRedirectionSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *DuoPushDefaultRedirectionSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *DuoPushDefaultRedirectionSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *DuoPushDefaultRedirectionSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *DuoPushDefaultRedirectionSuite) TestUserIsRedirectedToDefaultURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "allow",
		StatusMessage: "Allowing unknown user",
	}

	// Setup Duo device in DB.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Allow)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.verifyIsHome(s.T(), s.Page)

	// Clean up any Duo device already in DB.
	require.NoError(s.T(), provider.DeletePreferredDuoDevice(ctx, "john"))
}

type DuoPushSuite struct {
	*BaseSuite
}

func NewDuoPushSuite() *DuoPushSuite {
	return &DuoPushSuite{
		BaseSuite: &BaseSuite{
			Name: duoPushSuiteName,
		},
	}
}

func (s *DuoPushSuite) TestDuoPushWebDriverSuite() {
	suite.Run(s.T(), NewDuoPushWebDriverSuite())
}

func (s *DuoPushSuite) TestDuoPushRedirectionURLSuite() {
	suite.Run(s.T(), NewDuoPushDefaultRedirectionSuite())
}

func (s *DuoPushSuite) TestAvailableMethodsScenario() {
	suite.Run(s.T(), NewAvailableMethodsScenario([]string{
		"TIME-BASED ONE-TIME PASSWORD",
		"SECURITY KEY - WEBAUTHN",
		"PUSH NOTIFICATION",
	}))
}

func (s *DuoPushSuite) TestUserPreferencesScenario() {
	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "allow",
		StatusMessage: "Allowing unknown user",
	}

	ctx := context.Background()

	// Setup Duo device in DB.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	require.NoError(s.T(), provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Allow)

	suite.Run(s.T(), NewUserPreferencesScenario())

	// Clean up any Duo device already in DB.
	require.NoError(s.T(), provider.DeletePreferredDuoDevice(ctx, "john"))
}

func TestDuoPushSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewDuoPushSuite())
}
