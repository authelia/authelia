package suites

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/poy/onpar"

	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestDuoPushSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestDuoPushRedirectionURLScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestUserIsRedirectedToDefaultURL", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			is := is.New(t)

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
			is.NoErr(provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
			ConfigureDuoPreAuth(t, PreAuthAPIResponse)
			ConfigureDuo(t, Allow)

			s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
			s.doChangeMethod(t, s.Context(ctx), "push-notification")
			s.verifyIsHome(t, s.Page)

			// Clean up any Duo device already in DB.
			is.NoErr(provider.DeletePreferredDuoDevice(ctx, "john"))
		})
	})

	methods = []string{
		"TIME-BASED ONE-TIME PASSWORD",
		"SECURITY KEY - WEBAUTHN",
		"PUSH NOTIFICATION",
	}

	TestRunAvailableMethodsScenario(t)
	TestRunUserPreferencesScenario(t)
	t.Run("TestShouldBypassDeviceSelection", TestShouldBypassDeviceSelection)
	t.Run("TestShouldDenyDeviceSelection", TestShouldDenyDeviceSelection)
	t.Run("TestShouldAskUserToRegister", TestShouldAskUserToRegister)
	t.Run("TestShouldAutoSelectDevice", TestShouldAutoSelectDevice)
	t.Run("TestShouldSelectDevice", TestShouldSelectDevice)
	t.Run("TestShouldFailInitialSelectionBecauseOfUnsupportedMethod", TestShouldFailInitialSelectionBecauseOfUnsupportedMethod)
	t.Run("TestShouldSelectNewDeviceAfterSavedDeviceMethodIsNoLongerSupported", TestShouldSelectNewDeviceAfterSavedDeviceMethodIsNoLongerSupported)
	t.Run("TestShouldAutoSelectNewDeviceAfterSavedDeviceIsNoLongerAvailable", TestShouldAutoSelectNewDeviceAfterSavedDeviceIsNoLongerAvailable)
	t.Run("TestShouldFailSelectionBecauseOfSelectionBypassed", TestShouldFailSelectionBecauseOfSelectionBypassed)
	t.Run("TestShouldFailSelectionBecauseOfSelectionDenied", TestShouldFailSelectionBecauseOfSelectionDenied)
	t.Run("TestShouldFailAuthenticationBecausePreauthDenied", TestShouldFailAuthenticationBecausePreauthDenied)
	t.Run("TestShouldSucceedAuthentication", TestShouldSucceedAuthentication)
	t.Run("TestShouldFailAuthentication", TestShouldFailAuthentication)
}

func TestShouldBypassDeviceSelection(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "allow",
		StatusMessage: "Allowing unknown user",
	}

	ConfigureDuoPreAuth(t, PreAuthAPIResponse)

	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.verifyIsHome(t, s.Context(ctx))
}

func TestShouldDenyDeviceSelection(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "deny",
		StatusMessage: "We're sorry, access is not allowed.",
	}

	ConfigureDuoPreAuth(t, PreAuthAPIResponse)

	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.verifyNotificationDisplayed(t, s.Context(ctx), "Device selection was denied by Duo policy")
}

func TestShouldAskUserToRegister(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	is := is.New(t)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:          "enroll",
		EnrollPortalURL: "https://api-example.duosecurity.com/portal?code=1234567890ABCDEF&akey=12345ABCDEFGHIJ67890",
	}

	ConfigureDuoPreAuth(t, PreAuthAPIResponse)

	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.WaitElementLocatedByClassName(t, s.Context(ctx), "state-not-registered")
	s.verifyNotificationDisplayed(t, s.Context(ctx), "No compatible device found")
	enrollPage := s.Page.MustWaitOpen()
	s.WaitElementLocatedByID(t, s.Context(ctx), "register-link").MustClick()
	s.Page = enrollPage()

	is.True(strings.Contains(s.WaitElementLocatedByClassName(t, s.Context(ctx), "description").MustText(), "This enrollment code has expired. Contact your administrator to get a new enrollment code."))
}

func TestShouldAutoSelectDevice(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"auto", "push", "sms", "mobile_otp"},
		}},
	}

	ConfigureDuoPreAuth(t, PreAuthAPIResponse)
	ConfigureDuo(t, Allow)

	// Authenticate.
	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	// Switch Method where single Device should be selected automatically.
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.verifyIsHome(t, s.Context(ctx))

	// Re-Login the user.
	s.doLogout(t, s.Context(ctx))
	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	// And check the latest method and device is still used.
	s.WaitElementLocatedByID(t, s.Context(ctx), "push-notification-method")
	// Meaning the authentication is successful.
	s.verifyIsHome(t, s.Context(ctx))
}

func TestShouldSelectDevice(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	is := is.New(t)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

	// Set default 2FA preference to enable Select Device link in frontend.
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	is.NoErr(provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "ABCDEFGHIJ1234567890", Method: "push"}))

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

	ConfigureDuoPreAuth(t, PreAuthAPIResponse)
	ConfigureDuo(t, Allow)

	// Authenticate.
	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	// Switch Method where Device Selection should open automatically.
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	// Check for available Device 1.
	s.WaitElementLocatedByID(t, s.Context(ctx), "device-12345ABCDEFGHIJ67890")
	// Test Back button.
	s.doClickButton(t, s.Context(ctx), "device-selection-back")
	// then select Device 2 for further use and be redirected.
	s.doChangeDevice(t, s.Context(ctx), "1234567890ABCDEFGHIJ")
	s.verifyIsHome(t, s.Context(ctx))

	// Re-Login the user.
	s.doLogout(t, s.Context(ctx))
	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	// And check the latest method and device is still used.
	s.WaitElementLocatedByID(t, s.Context(ctx), "push-notification-method")
	// Meaning the authentication is successful.
	s.verifyIsHome(t, s.Context(ctx))
}

func TestShouldFailInitialSelectionBecauseOfUnsupportedMethod(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result: "auth",
		Devices: []duo.Device{{
			Device:       "12345ABCDEFGHIJ67890",
			DisplayName:  "Test Device 1",
			Capabilities: []string{"auto", "sms"},
		}},
	}

	ConfigureDuoPreAuth(t, PreAuthAPIResponse)

	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.WaitElementLocatedByClassName(t, s.Context(ctx), "state-not-registered")
	s.verifyNotificationDisplayed(t, s.Context(ctx), "No compatible device found")
}

func TestShouldSelectNewDeviceAfterSavedDeviceMethodIsNoLongerSupported(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	is := is.New(t)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

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
	is.NoErr(provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "ABCDEFGHIJ1234567890", Method: "sms"}))
	ConfigureDuoPreAuth(t, PreAuthAPIResponse)
	ConfigureDuo(t, Allow)

	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.WaitElementLocatedByID(t, s.Context(ctx), "device-selection")
	s.doSelectDevice(t, s.Context(ctx), "12345ABCDEFGHIJ67890")
	s.verifyIsHome(t, s.Context(ctx))
}

func TestShouldAutoSelectNewDeviceAfterSavedDeviceIsNoLongerAvailable(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	is := is.New(t)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

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
	is.NoErr(provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "ABCDEFGHIJ1234567890", Method: "push"}))
	ConfigureDuoPreAuth(t, PreAuthAPIResponse)
	ConfigureDuo(t, Allow)

	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.verifyIsHome(t, s.Context(ctx))
}

func TestShouldFailSelectionBecauseOfSelectionBypassed(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	is := is.New(t)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "allow",
		StatusMessage: "Allowing unknown user",
	}

	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	is.NoErr(provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(t, PreAuthAPIResponse)
	ConfigureDuo(t, Deny)

	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.doClickButton(t, s.Context(ctx), "selection-link")
	s.verifyNotificationDisplayed(t, s.Context(ctx), "Device selection was bypassed by Duo policy")
}

func TestShouldFailSelectionBecauseOfSelectionDenied(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	is := is.New(t)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "deny",
		StatusMessage: "We're sorry, access is not allowed.",
	}

	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	is.NoErr(provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(t, PreAuthAPIResponse)
	ConfigureDuo(t, Deny)

	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.WaitElementLocatedByID(t, s.Context(ctx), "selection-link").MustClick()
	s.verifyNotificationDisplayed(t, s.Context(ctx), "Device selection was denied by Duo policy")
}

func TestShouldFailAuthenticationBecausePreauthDenied(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	is := is.New(t)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
	}()

	var PreAuthAPIResponse = duo.PreAuthResponse{
		Result:        "deny",
		StatusMessage: "We're sorry, access is not allowed.",
	}

	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	is.NoErr(provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(t, PreAuthAPIResponse)

	s.doLoginOneFactor(t, s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.WaitElementLocatedByClassName(t, s.Context(ctx), "failure-icon")
	s.verifyNotificationDisplayed(t, s.Context(ctx), "There was an issue completing sign in process")
}

func TestShouldSucceedAuthentication(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	is := is.New(t)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
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
	is.NoErr(provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(t, PreAuthAPIResponse)
	ConfigureDuo(t, Allow)

	s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.verifyIsHome(t, s.Context(ctx))
}

func TestShouldFailAuthentication(t *testing.T) {
	s := setupTest(t, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	is := is.New(t)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
		teardownDuoTest(t, s)
		teardownTest(s)
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
	is.NoErr(provider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}))
	ConfigureDuoPreAuth(t, PreAuthAPIResponse)
	ConfigureDuo(t, Deny)

	s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
	s.doChangeMethod(t, s.Context(ctx), "push-notification")
	s.WaitElementLocatedByClassName(t, s.Context(ctx), "failure-icon")
}
