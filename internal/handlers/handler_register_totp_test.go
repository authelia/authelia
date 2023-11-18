package handlers

import (
	"fmt"
	"net/mail"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/totp"
)

func TestShouldReturnTOTPRegisterOptions(t *testing.T) {
	testCases := []struct {
		name           string
		config         schema.TOTP
		expected       string
		expectedStatus int
	}{
		{
			"ShouldHandleDefaults",
			schema.DefaultTOTPConfiguration,
			"{\"status\":\"OK\",\"data\":{\"algorithm\":\"SHA1\",\"algorithms\":[\"SHA1\"],\"length\":6,\"lengths\":[6],\"period\":30,\"periods\":[30]}}",
			fasthttp.StatusOK,
		},
		{
			"ShouldHandleCustom",
			schema.TOTP{DefaultAlgorithm: "SHA256", AllowedAlgorithms: []string{"SHA1", "SHA256"}, DefaultDigits: 6, AllowedDigits: []int{6, 8}, DefaultPeriod: 30, AllowedPeriods: []int{30, 60, 90}},
			"{\"status\":\"OK\",\"data\":{\"algorithm\":\"SHA256\",\"algorithms\":[\"SHA1\",\"SHA256\"],\"length\":6,\"lengths\":[6,8],\"period\":30,\"periods\":[30,60,90]}}",
			fasthttp.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.TOTP = tc.config

			mock.TOTPMock.EXPECT().Options().Return(*totp.NewTOTPOptionsFromSchema(mock.Ctx.Configuration.TOTP))

			TOTPRegisterGET(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))
		})
	}
}

func TestTOTPRegisterPUT(t *testing.T) {
	testCases := []struct {
		name           string
		config         schema.TOTP
		have           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldAllowDefaults",
			schema.DefaultTOTPConfiguration,
			`{"algorithm":"SHA1","length":6,"period":30}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.TOTPMock.EXPECT().Options().Return(*totp.NewTOTPOptionsFromSchema(mock.Ctx.Configuration.TOTP)),
					mock.TOTPMock.EXPECT().
						GenerateCustom(testUsername, mock.Ctx.Configuration.TOTP.DefaultAlgorithm, "", uint(mock.Ctx.Configuration.TOTP.DefaultDigits), uint(mock.Ctx.Configuration.TOTP.DefaultPeriod), uint(0)).
						Return(&model.TOTPConfiguration{Username: testUsername, Algorithm: mock.Ctx.Configuration.TOTP.DefaultAlgorithm, Digits: uint(mock.Ctx.Configuration.TOTP.DefaultDigits), Period: uint(mock.Ctx.Configuration.TOTP.DefaultPeriod)}, nil),
				)
			},
			`{"status":"OK","data":{"base32_secret":"","otpauth_url":"otpauth://totp/:john?algorithm=SHA1\u0026digits=6\u0026issuer=\u0026period=30\u0026secret="}}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldDenyLengthNotPermitted",
			schema.DefaultTOTPConfiguration,
			`{"algorithm":"SHA1","length":20,"period":30}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.TOTPMock.EXPECT().Options().Return(*totp.NewTOTPOptionsFromSchema(mock.Ctx.Configuration.TOTP)),
				)
			},
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Validation failed for TOTP registration because the input options were not permitted by the configuration", "")
			},
		},
		{
			"ShouldPreventAnonymous",
			schema.DefaultTOTPConfiguration,
			`{"algorithm":"SHA1","length":6,"period":30}`,
			nil,
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred handling request: anonymous user attempted TOTP registration", "")
			},
		},
		{
			"ShouldPreventBadJSON",
			schema.DefaultTOTPConfiguration,
			`{"algorithm:"SHA1","length":6,"period":30}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred unmarshaling body TOTP registration", "invalid character 'S' after object key")
			},
		},
		{
			"ShouldPreventBadValueIssuer",
			schema.TOTP{
				DefaultAlgorithm:  schema.TOTPAlgorithmSHA1,
				DefaultDigits:     6,
				DefaultPeriod:     30,
				Skew:              schema.DefaultTOTPConfiguration.Skew,
				SecretSize:        schema.TOTPSecretSizeDefault,
				AllowedAlgorithms: []string{schema.TOTPAlgorithmSHA1},
				AllowedDigits:     []int{6},
				AllowedPeriods:    []int{30},
			},
			`{"algorithm":"SHA1","length":6,"period":30}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.TOTPMock.EXPECT().Options().Return(*totp.NewTOTPOptionsFromSchema(mock.Ctx.Configuration.TOTP)),
					mock.TOTPMock.EXPECT().
						GenerateCustom(testUsername, mock.Ctx.Configuration.TOTP.DefaultAlgorithm, "", uint(mock.Ctx.Configuration.TOTP.DefaultDigits), uint(mock.Ctx.Configuration.TOTP.DefaultPeriod), uint(0)).
						Return(nil, fmt.Errorf("no issuer")),
				)
			},
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating TOTP configuration", "no issuer")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.TOTP = tc.config
			mock.Ctx.Request.SetBodyString(tc.have)

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			TOTPRegisterPUT(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestTOTPRegisterDELETE(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldFailAnonymous",
			nil,
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusForbidden,
			nil,
		},
		{
			"ShouldSkipDelete",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldDelete",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer: "abc",
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Nil(t, us.TOTP)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			TOTPRegisterDELETE(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestTOTPRegisterPOST(t *testing.T) {
	testCases := []struct {
		name           string
		config         schema.TOTP
		have           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldFailNoPending",
			schema.DefaultTOTPConfiguration,
			`{"token":"012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred during TOTP registration: the user did not initiate a registration on their current session", "")
			},
		},
		{
			"ShouldFailBadJSON",
			schema.DefaultTOTPConfiguration,
			`{"token":012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer:    "abc",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    testBASE32TOTPSecret,
					Expires:   mock.Clock.Now().Add(time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred unmarshaling body TOTP registration", "invalid character '1' after object key:value pair")
			},
		},
		{
			"ShouldFailBadExpires",
			schema.DefaultTOTPConfiguration,
			`{"token":"012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer:    "abc",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    testBASE32TOTPSecret,
					Expires:   mock.Clock.Now().Add(-time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred during TOTP registration: the registration is expired", "")
			},
		},
		{
			"ShouldFailAnonymous",
			schema.DefaultTOTPConfiguration,
			`{"token":012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.TOTP = &session.TOTP{
					Issuer:    "abc",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    testBASE32TOTPSecret,
					Expires:   mock.Clock.Now().Add(time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred handling request: anonymous user attempted TOTP registration", "")
			},
		},
		{
			"ShouldFailBadCode",
			schema.DefaultTOTPConfiguration,
			`{"token":"012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer:    "abc",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    testBASE32TOTPSecret,
					Expires:   mock.Clock.Now().Add(time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.TOTPMock.EXPECT().
						Validate(
							"012345",
							&model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)},
						).Return(false, nil),
				)
			},
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating TOTP registration", "")
			},
		},
		{
			"ShouldFailBadCodeErr",
			schema.DefaultTOTPConfiguration,
			`{"token":"012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer:    "abc",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    testBASE32TOTPSecret,
					Expires:   mock.Clock.Now().Add(time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.TOTPMock.EXPECT().
						Validate(
							"012345",
							&model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)},
						).Return(false, fmt.Errorf("pink staple")),
				)
			},
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating TOTP registration", "pink staple")
			},
		},
		{
			"ShouldSaveGoodCode",
			schema.DefaultTOTPConfiguration,
			`{"token":"012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer:    "abc",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    testBASE32TOTPSecret,
					Expires:   mock.Clock.Now().Add(time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.TOTPMock.EXPECT().
						Validate(
							"012345",
							&model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)},
						).Return(true, nil),
					mock.StorageMock.EXPECT().
						SaveTOTPConfiguration(mock.Ctx, model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)}).Return(nil),
					mock.UserProviderMock.EXPECT().GetDetails(testUsername).Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{"john@example.com"}}, nil),
					mock.NotifierMock.EXPECT().Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: "john@example.com"}, "Second Factor Method Added", gomock.Any(), gomock.Any()).Return(nil),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldSaveGoodCodeButFailEventLogWithEmailError",
			schema.DefaultTOTPConfiguration,
			`{"token":"012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer:    "abc",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    testBASE32TOTPSecret,
					Expires:   mock.Clock.Now().Add(time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.TOTPMock.EXPECT().
						Validate(
							"012345",
							&model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)},
						).Return(true, nil),
					mock.StorageMock.EXPECT().
						SaveTOTPConfiguration(mock.Ctx, model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)}).Return(nil),
					mock.UserProviderMock.EXPECT().GetDetails(testUsername).Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{"john@example.com"}}, nil),
					mock.NotifierMock.EXPECT().Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: "john@example.com"}, "Second Factor Method Added", gomock.Any(), gomock.Any()).Return(fmt.Errorf("kittens")),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred sending notification to user 'john' while attempting to notify them of an important event", "kittens")
			},
		},
		{
			"ShouldSaveGoodCodeButFailEventLogNoEmail",
			schema.DefaultTOTPConfiguration,
			`{"token":"012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer:    "abc",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    testBASE32TOTPSecret,
					Expires:   mock.Clock.Now().Add(time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.TOTPMock.EXPECT().
						Validate(
							"012345",
							&model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)},
						).Return(true, nil),
					mock.StorageMock.EXPECT().
						SaveTOTPConfiguration(mock.Ctx, model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)}).Return(nil),
					mock.UserProviderMock.EXPECT().GetDetails(testUsername).Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName}, nil),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred looking up user details for user 'john' while attempting to notify them of an important event", "no email address was found for user")
			},
		},
		{
			"ShouldFailToSaveGoodCode",
			schema.DefaultTOTPConfiguration,
			`{"token":"012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer:    "abc",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    testBASE32TOTPSecret,
					Expires:   mock.Clock.Now().Add(time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.TOTPMock.EXPECT().
						Validate(
							"012345",
							&model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)},
						).Return(true, nil),
					mock.StorageMock.EXPECT().
						SaveTOTPConfiguration(mock.Ctx, model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)}).Return(nil),
					mock.UserProviderMock.EXPECT().GetDetails(testUsername).Return(nil, fmt.Errorf("lookup failure")),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred looking up user details for user 'john' while attempting to notify them of an important event", "lookup failure")
			},
		},
		{
			"ShouldFailToGetUserDetailsGoodCode",
			schema.DefaultTOTPConfiguration,
			`{"token":"012345"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer:    "abc",
					Algorithm: "SHA1",
					Digits:    6,
					Period:    30,
					Secret:    testBASE32TOTPSecret,
					Expires:   mock.Clock.Now().Add(time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.TOTPMock.EXPECT().
						Validate(
							"012345",
							&model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)},
						).Return(true, nil),
					mock.StorageMock.EXPECT().
						SaveTOTPConfiguration(mock.Ctx, model.TOTPConfiguration{CreatedAt: mock.Clock.Now(), Username: testUsername, Issuer: "abc", Algorithm: "SHA1", Period: 30, Digits: 6, Secret: []byte(testBASE32TOTPSecret)}).Return(fmt.Errorf("failed to connect")),
				)
			},
			`{"status":"KO","message":"Unable to set up one-time password."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred saving TOTP registration", "failed to connect")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.TOTP = tc.config
			mock.Ctx.Request.SetBodyString(tc.have)
			mock.Clock.Set(time.Unix(0, 0))
			mock.Ctx.Clock = &mock.Clock

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			TOTPRegisterPOST(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestTOTPConfigurationDELETE(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldFailAnonymous",
			nil,
			`{"status":"KO","message":"Unable to delete one-time password."}`,
			fasthttp.StatusForbidden,
			nil,
		},
		{
			"ShouldDelete",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer: "abc",
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.EXPECT().LoadTOTPConfiguration(mock.Ctx, testUsername).Return(&model.TOTPConfiguration{}, nil),
					mock.StorageMock.EXPECT().DeleteTOTPConfiguration(mock.Ctx, testUsername).Return(nil),
					mock.UserProviderMock.EXPECT().GetDetails(testUsername).Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{"john@example.com"}}, nil),
					mock.NotifierMock.EXPECT().Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: "john@example.com"}, "Second Factor Method Removed", gomock.Any(), gomock.Any()).Return(nil),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldDeleteAndLogErrorNotifier",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer: "abc",
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.EXPECT().LoadTOTPConfiguration(mock.Ctx, testUsername).Return(&model.TOTPConfiguration{}, nil),
					mock.StorageMock.EXPECT().DeleteTOTPConfiguration(mock.Ctx, testUsername).Return(nil),
					mock.UserProviderMock.EXPECT().GetDetails(testUsername).Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{"john@example.com"}}, nil),
					mock.NotifierMock.EXPECT().Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: "john@example.com"}, "Second Factor Method Removed", gomock.Any(), gomock.Any()).Return(fmt.Errorf("bad conn")),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred sending notification to user 'john' while attempting to notify them of an important event", "bad conn")
			},
		},
		{
			"ShouldDeleteAndLogErrorGetDetails",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer: "abc",
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.EXPECT().LoadTOTPConfiguration(mock.Ctx, testUsername).Return(&model.TOTPConfiguration{}, nil),
					mock.StorageMock.EXPECT().DeleteTOTPConfiguration(mock.Ctx, testUsername).Return(nil),
					mock.UserProviderMock.EXPECT().GetDetails(testUsername).Return(nil, fmt.Errorf("lookup err")),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred looking up user details for user 'john' while attempting to notify them of an important event", "lookup err")
			},
		},
		{
			"ShouldFailDelete",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer: "abc",
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.EXPECT().LoadTOTPConfiguration(mock.Ctx, testUsername).Return(&model.TOTPConfiguration{}, nil),
					mock.StorageMock.EXPECT().DeleteTOTPConfiguration(mock.Ctx, testUsername).Return(fmt.Errorf("not a sql")),
				)
			},
			`{"status":"KO","message":"Unable to delete one-time password."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred deleting from storage for TOTP configuration delete operation for user 'john'", "not a sql")
			},
		},
		{
			"ShouldFailDeleteLookup",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.TOTP = &session.TOTP{
					Issuer: "abc",
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.EXPECT().LoadTOTPConfiguration(mock.Ctx, testUsername).Return(nil, fmt.Errorf("not found")),
				)
			},
			`{"status":"KO","message":"Unable to delete one-time password."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred loading from storage for TOTP configuration delete operation for user 'john'", "not found")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			TOTPConfigurationDELETE(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}
