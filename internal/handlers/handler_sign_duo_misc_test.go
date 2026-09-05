package handlers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestDuoPOST(t *testing.T) {
	t.Run("ShouldHandleMalformedBody", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Request.SetBodyString("not json")

		DuoPOST(nil)(mock.Ctx)

		mock.Assert401KO(t, messageMFAValidationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to parse Duo request body", regexpAnyError)
	})
}

func TestSendDuoDevicesResponse(t *testing.T) {
	t.Run("ShouldSendResponse", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		SendDuoDevicesResponse(mock.Ctx, DuoDevicesResponse{
			Result: auth,
			Devices: []DuoDevice{
				{Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device", Capabilities: []string{"push"}},
			},
		})

		body := DuoDevicesResponse{}

		mock.GetResponseData(t, &body)

		assert.Equal(t, auth, body.Result)
		require.Len(t, body.Devices, 1)
		assert.Equal(t, "Test Device", body.Devices[0].DisplayName)
	})
}

func TestHandleAllow(t *testing.T) {
	newSession := func() *session.UserSession {
		return &session.UserSession{
			CookieDomain: exampleDotCom,
			Username:     testUsername,
			DisplayName:  testDisplayName,
			Emails:       []string{testEmail},
		}
	}

	t.Run("ShouldRedirectToTargetURL", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
			AccessControl: schema.AccessControl{DefaultPolicy: "two_factor"},
		})

		userSession := newSession()

		HandleAllow(mock.Ctx, userSession, &bodySignDuoRequest{TargetURL: testRedirectionURLString})

		mock.Assert200OK(t, redirectResponse{Redirect: testRedirectionURLString})

		assert.True(t, userSession.AuthenticationMethodRefs.Duo)
		assert.NotZero(t, userSession.SecondFactorAuthnTimestamp)
	})

	t.Run("ShouldDelegateToFlowHandler", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		userSession := newSession()

		HandleAllow(mock.Ctx, userSession, &bodySignDuoRequest{Flow: "not-a-flow"})

		mock.Assert200KO(t, messageAuthenticationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to find flow handler for the given flow parameters", nil)
	})
}

func TestSetValues(t *testing.T) {
	userSession := session.UserSession{Username: testUsername, DisplayName: testDisplayName}

	testCases := []struct {
		name      string
		method    string
		passcode  string
		targetURL string
		expected  map[string]string
		err       string
	}{
		{
			name:      "ShouldSetPushValues",
			method:    duo.Push,
			targetURL: testRedirectionURLString,
			expected: map[string]string{
				"username":         testUsername,
				"factor":           duo.Push,
				"device":           "ABC",
				"display_username": testDisplayName,
				"pushinfo":         "target%20url=" + testRedirectionURLString,
			},
		},
		{
			name:     "ShouldSetPhoneValues",
			method:   duo.Phone,
			expected: map[string]string{"factor": duo.Phone, "device": "ABC"},
		},
		{
			name:     "ShouldSetSMSValues",
			method:   duo.SMS,
			expected: map[string]string{"factor": duo.SMS, "device": "ABC"},
		},
		{
			name:     "ShouldSetOTPValues",
			method:   duo.OTP,
			passcode: "123456",
			expected: map[string]string{"factor": duo.OTP, "passcode": "123456"},
		},
		{
			name:   "ShouldErrorOnOTPWithoutPasscode",
			method: duo.OTP,
			err:    "no passcode received from user: john",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			values, err := SetValues(userSession, "ABC", tc.method, "127.0.0.1", tc.targetURL, tc.passcode)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, values)

				return
			}

			require.NoError(t, err)

			assert.Equal(t, "127.0.0.1", values.Get("ipaddr"))

			for key, expected := range tc.expected {
				assert.Equal(t, expected, values.Get(key), key)
			}
		})
	}
}

func TestDuoDevicesGETMisc(t *testing.T) {
	t.Run("ShouldHandleSessionError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		duoMock := mocks.NewMockDuoProvider(mock.Ctrl)

		duoMock.EXPECT().
			PreAuthCall(mock.Ctx, gomock.Any(), gomock.Any()).
			Return(&duo.PreAuthResponse{Result: enroll, EnrollPortalURL: "https://api-example.duosecurity.com/portal?abcdef"}, nil)

		mock.StorageMock.EXPECT().
			SavePreferredDuoDevice(mock.Ctx, gomock.Any()).
			AnyTimes().
			Return(nil)

		mock.StorageMock.EXPECT().
			LoadPreferredDuoDevice(mock.Ctx, gomock.Any()).
			AnyTimes().
			Return(&model.DuoDevice{Username: testUsername, Device: "ABC", Method: duo.Push}, nil)

		DuoDevicesGET(duoMock)(mock.Ctx)

		body := DuoDevicesResponse{}

		mock.GetResponseData(t, &body)

		assert.Equal(t, enroll, body.Result)
		assert.Equal(t, "https://api-example.duosecurity.com/portal?abcdef", body.EnrollURL)
	})
}

func TestDuoPOSTMisc(t *testing.T) {
	t.Run("ShouldHandleSessionProviderError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
		mock.Ctx.Request.SetBodyString(`{}`)

		DuoPOST(nil)(mock.Ctx)

		mock.Assert200KO(t, messageMFAValidationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), errStrUserSessionData, "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
	})
}

func TestPerformDuoAuthenticationMisc(t *testing.T) {
	t.Run("ShouldHandleValuesError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		userSession := &session.UserSession{Username: testUsername}

		err := PerformDuoAuthentication(mock.Ctx, userSession, nil, "ABC", duo.OTP, "127.0.0.1", &bodySignDuoRequest{})

		assert.EqualError(t, err, "no passcode received from user: john")
	})
}

func TestHandleDuoPreAuthResult(t *testing.T) {
	newSession := func() *session.UserSession {
		return &session.UserSession{CookieDomain: exampleDotCom, Username: testUsername}
	}

	t.Run("ShouldHandleAllow", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
			AccessControl: schema.AccessControl{DefaultPolicy: "two_factor"},
		})

		userSession := newSession()

		device, method, err := HandleDuoPreAuthResult(mock.Ctx, userSession, allow, "", nil, "", &bodySignDuoRequest{TargetURL: testRedirectionURLString})

		assert.NoError(t, err)
		assert.Empty(t, device)
		assert.Empty(t, method)

		mock.Assert200OK(t, redirectResponse{Redirect: testRedirectionURLString})

		assert.True(t, userSession.AuthenticationMethodRefs.Duo)
	})

	t.Run("ShouldHandleUnknownResult", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		device, method, err := HandleDuoPreAuthResult(mock.Ctx, newSession(), "not-a-result", "", nil, "", &bodySignDuoRequest{})

		assert.EqualError(t, err, "unknown result: not-a-result")
		assert.Empty(t, device)
		assert.Empty(t, method)
	})
}

func TestHandleNoDevicesAvailable(t *testing.T) {
	t.Run("ShouldHandleDeleteError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.StorageMock.EXPECT().
			DeletePreferredDuoDevice(mock.Ctx, testUsername).
			Return(errors.New("failed to delete"))

		err := HandleNoDevicesAvailable(mock.Ctx, &session.UserSession{Username: testUsername}, "ABC")

		assert.EqualError(t, err, "unable to delete preferred Duo device and method for user 'john': failed to delete")
	})

	t.Run("ShouldPropagateErrorViaHandleAuthResult", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.StorageMock.EXPECT().
			DeletePreferredDuoDevice(mock.Ctx, testUsername).
			Return(errors.New("failed to delete"))

		device, method, err := HandleAuthResult(mock.Ctx, &session.UserSession{Username: testUsername}, nil, "ABC", duo.Push)

		assert.EqualError(t, err, "unable to delete preferred Duo device and method for user 'john': failed to delete")
		assert.Empty(t, device)
		assert.Empty(t, method)
	})
}

func TestHandlePreferredDeviceCheckMisc(t *testing.T) {
	newSession := func() *session.UserSession {
		return &session.UserSession{CookieDomain: exampleDotCom, Username: testUsername}
	}

	t.Run("ShouldHandleEnrollDeleteError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		duoMock := mocks.NewMockDuoProvider(mock.Ctrl)

		duoMock.EXPECT().
			PreAuthCall(mock.Ctx, gomock.Any(), gomock.Any()).
			Return(&duo.PreAuthResponse{Result: enroll}, nil)

		mock.StorageMock.EXPECT().
			DeletePreferredDuoDevice(mock.Ctx, testUsername).
			Return(errors.New("failed to delete"))

		device, method, err := HandlePreferredDeviceCheck(mock.Ctx, newSession(), duoMock, "ABC", duo.Push, &bodySignDuoRequest{})

		assert.EqualError(t, err, "unable to delete preferred Duo device and method for user 'john': failed to delete")
		assert.Empty(t, device)
		assert.Empty(t, method)
	})

	t.Run("ShouldHandleUnknownResult", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		duoMock := mocks.NewMockDuoProvider(mock.Ctrl)

		duoMock.EXPECT().
			PreAuthCall(mock.Ctx, gomock.Any(), gomock.Any()).
			Return(&duo.PreAuthResponse{Result: "not-a-result"}, nil)

		device, method, err := HandlePreferredDeviceCheck(mock.Ctx, newSession(), duoMock, "ABC", duo.Push, &bodySignDuoRequest{})

		assert.EqualError(t, err, "unknown result: not-a-result")
		assert.Empty(t, device)
		assert.Empty(t, method)
	})
}

func TestHandleAutoSelection(t *testing.T) {
	t.Run("ShouldEnrollWhenNoDevices", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		device, method, err := HandleAutoSelection(mock.Ctx, nil, testUsername)

		assert.NoError(t, err)
		assert.Empty(t, device)
		assert.Empty(t, method)

		body := DuoSignResponse{}

		mock.GetResponseData(t, &body)

		assert.Equal(t, enroll, body.Result)
	})

	t.Run("ShouldRequireSelectionWithMultipleMethods", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		devices := []DuoDevice{{Device: "ABC", DisplayName: "Test Device", Capabilities: []string{duo.Push, duo.SMS}}}

		device, method, err := HandleAutoSelection(mock.Ctx, devices, testUsername)

		assert.NoError(t, err)
		assert.Empty(t, device)
		assert.Empty(t, method)

		body := DuoSignResponse{}

		mock.GetResponseData(t, &body)

		assert.Equal(t, auth, body.Result)
		require.Len(t, body.Devices, 1)
		assert.Equal(t, []string{duo.Push, duo.SMS}, body.Devices[0].Capabilities)
	})

	t.Run("ShouldHandleSaveError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.StorageMock.EXPECT().
			SavePreferredDuoDevice(mock.Ctx, model.DuoDevice{Username: testUsername, Device: "ABC", Method: duo.Push}).
			Return(errors.New("failed to save"))

		devices := []DuoDevice{{Device: "ABC", DisplayName: "Test Device", Capabilities: []string{duo.Push}}}

		device, method, err := HandleAutoSelection(mock.Ctx, devices, testUsername)

		assert.EqualError(t, err, "unable to save new preferred Duo device and method for user 'john': failed to save")
		assert.Empty(t, device)
		assert.Empty(t, method)
	})
}

func TestHandleAllowMisc(t *testing.T) {
	t.Run("ShouldHandleSessionRegenerateError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")

		HandleAllow(mock.Ctx, &session.UserSession{Username: testUsername}, &bodySignDuoRequest{})

		mock.Assert401KO(t, messageMFAValidationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Could not regenerate session during Duo authentication for user 'john'", "unable to regenerate user session: unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
	})
}
