package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/storage"
)

type DuoPushWebDriverSuite struct {
	*RodSuite
}

func NewDuoPushWebDriverSuite() *DuoPushWebDriverSuite {
	return &DuoPushWebDriverSuite{RodSuite: new(RodSuite)}
}

func (s *DuoPushWebDriverSuite) SetupSuite() {
	browser, err := StartRod()

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
	//TODO: MERGE CONFLICT
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	//TODO: MERGE CONFLICT
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)

		s.collectCoverage(s.Page)
		s.MustClose()
	}()

	s.doLogout(s.T(), s.Context(ctx))
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doChangeMethod(s.T(), s.Context(ctx), "one-time-password")
	s.WaitElementLocatedByCSSSelector(s.T(), s.Context(ctx), "one-time-password-method")
	s.doLogout(ctx, s.T())

	// Set default 2FA preference and clean up any Duo device already in DB.
	provider := storage.NewSQLiteProvider("/tmp/db.sqlite3")
	require.NoError(s.T(), provider.SavePreferred2FAMethod("john", "totp"))
	require.NoError(s.T(), provider.DeletePreferredDuoDevice("john"))
}

func (s *DuoPushWebDriverSuite) TestShouldBypassDeviceSelection() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreauthResponse{
		Result:        "allow",
		StatusMessage: "Allowing unknown user",
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.verifyIsHome(ctx, s.T())
}

func (s *DuoPushWebDriverSuite) TestShouldDenyDeviceSelection() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreauthResponse{
		Result:        "deny",
		StatusMessage: "We're sorry, access is not allowed.",
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.verifyNotificationDisplayed(ctx, s.T(), "Device Selection was denied by Duo Policy")
}

func (s *DuoPushWebDriverSuite) TestShouldAskUserToRegister() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreauthResponse{
		Result:          "enroll",
		EnrollPortalURL: "https://api-example.duosecurity.com/portal?code=1234567890ABCDEF&akey=12345ABCDEFGHIJ67890",
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.WaitElementLocatedByClassName(ctx, s.T(), "state-not-registered")
	s.WaitElementLocatedByID(ctx, s.T(), "register-link")
	s.verifyNotificationDisplayed(ctx, s.T(), "No (compatible) device found")
}

func (s *DuoPushWebDriverSuite) TestUserIsAskedToSelectDevice() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreauthResponse{
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

	// Authenticate
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	// Check for available Devices.
	s.WaitElementLocatedByID(ctx, s.T(), "device-12345ABCDEFGHIJ67890")
	// Test Back button.
	s.doClickButton(ctx, s.T(), "device-selection-back")
	// then select a Device for further use and be redirected.
	s.doChangeDevice(ctx, s.T(), "1234567890ABCDEFGHIJ")
	s.verifyIsHome(ctx, s.T())

	// Logout the user and check if defvice was remembered after logout.
	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), GetLoginBaseURL())
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyIsSecondFactorPage(ctx, s.T())
	// And check the latest method and device is still used.
	s.WaitElementLocatedByID(ctx, s.T(), "push-notification-method")
	// Meaning the authentication is successful
	s.verifyIsHome(ctx, s.T())
}

func (s *DuoPushWebDriverSuite) TestShouldFailInitialSelectionBecauseOfUnsupportedMethod() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreauthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"auto", "sms"},
		}},
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.WaitElementLocatedByClassName(ctx, s.T(), "state-not-registered")
	s.WaitElementLocatedByID(ctx, s.T(), "register-link")
	s.verifyNotificationDisplayed(ctx, s.T(), "No (compatible) device found")
}

func (s *DuoPushWebDriverSuite) TestShouldSelectNewDeviceAfterSavedDeviceMethodIsNoLongerSupported() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreauthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"push", "sms"},
		}},
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Allow)

	// Setup unsupported Duo device in DB.
	provider := storage.NewSQLiteProvider("/tmp/db.sqlite3")
	require.NoError(s.T(), provider.SavePreferredDuoDevice("john", "12345ABCDEFGHIJ67890", "sms"))

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.verifyNotificationDisplayed(ctx, s.T(), "Please select a new compatible device")
	s.WaitElementLocatedByID(ctx, s.T(), "device-selection")
	s.doSelectDevice(ctx, s.T(), "12345ABCDEFGHIJ67890")
	s.verifyIsHome(ctx, s.T())
}

func (s *DuoPushWebDriverSuite) TestShouldFailSelectionBecauseOfSelectionBypassed() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreauthResponse{
		Result:        "allow",
		StatusMessage: "Allowing unknown user",
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Deny)

	// Setup unsupported Duo device in DB.
	provider := storage.NewSQLiteProvider("/tmp/db.sqlite3")
	require.NoError(s.T(), provider.SavePreferredDuoDevice("john", "12345ABCDEFGHIJ67890", "push"))

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.doClickButton(ctx, s.T(), "selection-link")
	s.verifyNotificationDisplayed(ctx, s.T(), "Device Selection is being bypassed by Duo Policy")
}

func (s *DuoPushWebDriverSuite) TestShouldFailSelectionBecauseOfSelectionDenied() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var PreAuthAPIResponse = duo.PreauthResponse{
		Result:        "deny",
		StatusMessage: "We're sorry, access is not allowed.",
	}

	ConfigureDuoPreAuth(s.T(), PreAuthAPIResponse)
	ConfigureDuo(s.T(), Deny)

	// Setup unsupported Duo device in DB.
	provider := storage.NewSQLiteProvider("/tmp/db.sqlite3")
	require.NoError(s.T(), provider.SavePreferredDuoDevice("john", "12345ABCDEFGHIJ67890", "push"))

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	err := s.WaitElementLocatedByID(ctx, s.T(), "selection-link").Click()
	require.NoError(s.T(), err)
	s.verifyNotificationDisplayed(ctx, s.T(), "Device Selection was denied by Duo Policy")
}

func (s *DuoPushWebDriverSuite) TestShouldSucceedAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	//TODO: MERGE CONFLICT
	// Setup Duo device in DB.
	provider := storage.NewSQLiteProvider("/tmp/db.sqlite3")
	require.NoError(s.T(), provider.SavePreferredDuoDevice("john", "12345ABCDEFGHIJ67890", "push"))
	//TODO: MERGE CONFLICT
	ConfigureDuo(s.T(), Allow)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (s *DuoPushWebDriverSuite) TestShouldFailAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	//TODO: MERGE CONFLICT
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// Setup Duo device in DB.
	provider := storage.NewSQLiteProvider("/tmp/db.sqlite3")
	require.NoError(s.T(), provider.SavePreferredDuoDevice("john", "12345ABCDEFGHIJ67890", "push"))
	//TODO: MERGE CONFLICT
	ConfigureDuo(s.T(), Deny)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.WaitElementLocatedByClassName(s.T(), s.Context(ctx), "failure-icon")
}

type DuoPushDefaultRedirectionSuite struct {
	*RodSuite
}

func NewDuoPushDefaultRedirectionSuite() *DuoPushDefaultRedirectionSuite {
	return &DuoPushDefaultRedirectionSuite{RodSuite: new(RodSuite)}
}

func (s *DuoPushDefaultRedirectionSuite) SetupSuite() {
	browser, err := StartRod()

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
	//TODO: MERGE CONFLICT
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Setup Duo device in DB.
	provider := storage.NewSQLiteProvider("/tmp/db.sqlite3")
	require.NoError(s.T(), provider.SavePreferredDuoDevice("john", "12345ABCDEFGHIJ67890", "push"))
	ConfigureDuo(s.T(), Allow)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")

	// Clean up any Duo device already in DB.
	require.NoError(s.T(), provider.DeletePreferredDuoDevice("john"))
	//TODO: MERGE CONFLICT
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.verifyIsHome(s.T(), s.Page)
}

type DuoPushSuite struct {
	suite.Suite
}

func NewDuoPushSuite() *DuoPushSuite {
	return &DuoPushSuite{}
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
		"PUSH NOTIFICATION",
	}))
}

func (s *DuoPushSuite) TestUserPreferencesScenario() {
	// Setup Duo device in DB.
	provider := storage.NewSQLiteProvider("/tmp/db.sqlite3")
	require.NoError(s.T(), provider.SavePreferredDuoDevice("john", "12345ABCDEFGHIJ67890", "push"))
	ConfigureDuo(s.T(), Allow)

	suite.Run(s.T(), NewUserPreferencesScenario())

	// Clean up any Duo device already in DB.
	require.NoError(s.T(), provider.DeletePreferredDuoDevice("john"))
}

func TestDuoPushSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewDuoPushSuite())
}
