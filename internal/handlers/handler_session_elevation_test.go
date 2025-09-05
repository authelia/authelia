package handlers

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net"
	"net/mail"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestUserSessionElevationGET(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleOneFactor",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, testUsername).Return(model.UserInfo{
					DisplayName: testDisplayName,
					Method:      "totp",
					HasTOTP:     true,
					HasWebAuthn: true,
					HasDuo:      true,
				}, nil)
			},
			`{"status":"OK","data":{"require_second_factor":false,"skip_second_factor":false,"can_skip_second_factor":false,"factor_knowledge":true,"elevated":false,"expires":0}}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleOneFactorRequireSecondFactor",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.IdentityValidation.ElevatedSession.RequireSecondFactor = true

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, testUsername).Return(model.UserInfo{
					DisplayName: testDisplayName,
					Method:      "totp",
					HasTOTP:     true,
					HasWebAuthn: true,
					HasDuo:      true,
				}, nil)
			},
			`{"status":"OK","data":{"require_second_factor":true,"skip_second_factor":false,"can_skip_second_factor":false,"factor_knowledge":true,"elevated":false,"expires":0}}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleOneFactorRequireSecondFactorWithoutSecondFactor",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.IdentityValidation.ElevatedSession.RequireSecondFactor = true

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, testUsername).Return(model.UserInfo{
					DisplayName: testDisplayName,
					Method:      "totp",
					HasTOTP:     false,
					HasWebAuthn: false,
					HasDuo:      false,
				}, nil)
			},
			`{"status":"OK","data":{"require_second_factor":false,"skip_second_factor":false,"can_skip_second_factor":false,"factor_knowledge":true,"elevated":false,"expires":0}}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleOneFactorSkipSecondFactor",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.IdentityValidation.ElevatedSession.SkipSecondFactor = true

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, testUsername).Return(model.UserInfo{
					DisplayName: testDisplayName,
					Method:      "totp",
					HasTOTP:     true,
					HasWebAuthn: true,
					HasDuo:      true,
				}, nil)
			},
			`{"status":"OK","data":{"require_second_factor":false,"skip_second_factor":false,"can_skip_second_factor":true,"factor_knowledge":true,"elevated":false,"expires":0}}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleTwoFactor",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.AuthenticationMethodRefs.WebAuthn = true

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"OK","data":{"require_second_factor":false,"skip_second_factor":false,"can_skip_second_factor":false,"factor_knowledge":false,"elevated":false,"expires":0}}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleTwoFactorSkip",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.IdentityValidation.ElevatedSession.SkipSecondFactor = true

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.AuthenticationMethodRefs.WebAuthn = true

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"OK","data":{"require_second_factor":false,"skip_second_factor":true,"can_skip_second_factor":false,"factor_knowledge":false,"elevated":false,"expires":0}}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleAnonymous",
			nil,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred retrieving user session elevation state", "user is anonymous")
			},
		},
		{
			"ShouldHandleBadSessionDomain",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred retrieving user session elevation state: error occurred retrieving the user session data", "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
			},
		},
		{
			"ShouldHandleElevated",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.IdentityValidation.ElevatedSession.SkipSecondFactor = true

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.AuthenticationMethodRefs.WebAuthn = true
				us.Elevations.User = &session.Elevation{
					ID:       1,
					RemoteIP: mock.Ctx.RemoteIP(),
					Expires:  mock.Clock.Now().Add(10 * time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"OK","data":{"require_second_factor":false,"skip_second_factor":true,"can_skip_second_factor":false,"factor_knowledge":false,"elevated":true,"expires":600}}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleElevatedCanSkip",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.AuthenticationMethodRefs.WebAuthn = true
				us.Elevations.User = &session.Elevation{
					ID:       1,
					RemoteIP: mock.Ctx.RemoteIP(),
					Expires:  mock.Clock.Now().Add(10 * time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"OK","data":{"require_second_factor":false,"skip_second_factor":false,"can_skip_second_factor":false,"factor_knowledge":false,"elevated":true,"expires":600}}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleElevationBadIP",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.AuthenticationMethodRefs.WebAuthn = true
				us.Elevations.User = &session.Elevation{
					ID:       1,
					RemoteIP: net.ParseIP("1.1.1.1"),
					Expires:  mock.Clock.Now().Add(10 * time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"OK","data":{"require_second_factor":false,"skip_second_factor":false,"can_skip_second_factor":false,"factor_knowledge":false,"elevated":false,"expires":0}}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.Elevations.User)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "The user session elevation was created from a different remote IP so it has been destroyed", "")
			},
		},
		{
			"ShouldHandleElevationExpired",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.AuthenticationMethodRefs.WebAuthn = true
				us.Elevations.User = &session.Elevation{
					ID:       1,
					RemoteIP: mock.Ctx.RemoteIP(),
					Expires:  mock.Clock.Now().Add(-time.Minute),
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"OK","data":{"require_second_factor":false,"skip_second_factor":false,"can_skip_second_factor":false,"factor_knowledge":false,"elevated":false,"expires":-60}}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.Elevations.User)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			UserSessionElevationGET(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestUserSessionElevationPOST(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleOneFactor",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.RandomMock.EXPECT().
						Read(gomock.Any()).
						SetArg(0, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x22, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15}).
						Return(16, nil),
					mock.RandomMock.EXPECT().
						BytesCustomErr(10, []byte(random.CharSetUnambiguousUpper)).
						Return([]byte("ABC123ABC1"), nil),
					mock.StorageMock.EXPECT().
						SaveOneTimeCode(mock.Ctx, model.OneTimeCode{
							PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
							IssuedAt:  mock.Clock.Now(),
							IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
							ExpiresAt: mock.Clock.Now().Add(time.Minute),
							Username:  testUsername,
							Intent:    model.OTCIntentUserSessionElevation,
							Code:      []byte("ABC123ABC1"),
						}).
						Return("abc123", nil),
					mock.NotifierMock.EXPECT().Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: "john@example.com"}, "Confirm your identity", gomock.Any(), templates.EmailIdentityVerificationOTCValues{
						Title:              "Confirm your identity",
						RevocationLinkURL:  "http://example.com/revoke/one-time-code?id=AQIDBAUGRyKJEBESExQVAA",
						RevocationLinkText: "Revoke",
						DisplayName:        testDisplayName,
						Domain:             "example.com",
						RemoteIP:           "0.0.0.0",
						OneTimeCode:        "ABC123ABC1",
					}).
						Return(nil),
				)
			},
			`{"status":"OK","data":{"delete_id":"AQIDBAUGRyKJEBESExQVAA"}}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleOneFactorFailEmail",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.RandomMock.EXPECT().
						Read(gomock.Any()).
						SetArg(0, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x22, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15}).
						Return(16, nil),
					mock.RandomMock.EXPECT().
						BytesCustomErr(10, []byte(random.CharSetUnambiguousUpper)).
						Return([]byte("ABC123ABC1"), nil),
					mock.StorageMock.EXPECT().
						SaveOneTimeCode(mock.Ctx, model.OneTimeCode{
							PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
							IssuedAt:  mock.Clock.Now(),
							IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
							ExpiresAt: mock.Clock.Now().Add(time.Minute),
							Username:  testUsername,
							Intent:    model.OTCIntentUserSessionElevation,
							Code:      []byte("ABC123ABC1"),
						}).
						Return("abc123", nil),
					mock.NotifierMock.EXPECT().Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: "john@example.com"}, "Confirm your identity", gomock.Any(), templates.EmailIdentityVerificationOTCValues{
						Title:              "Confirm your identity",
						RevocationLinkURL:  "http://example.com/revoke/one-time-code?id=AQIDBAUGRyKJEBESExQVAA",
						RevocationLinkText: "Revoke",
						DisplayName:        testDisplayName,
						Domain:             "example.com",
						RemoteIP:           "0.0.0.0",
						OneTimeCode:        "ABC123ABC1",
					}).
						Return(fmt.Errorf("rejected")),
				)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred creating user session elevation One-Time Code challenge for user 'john': error occurred sending the user the notification", "rejected")
			},
		},
		{
			"ShouldHandleOneFactorFailInsert",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.RandomMock.EXPECT().
						Read(gomock.Any()).
						SetArg(0, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x22, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15}).
						Return(16, nil),
					mock.RandomMock.EXPECT().
						BytesCustomErr(10, []byte(random.CharSetUnambiguousUpper)).
						Return([]byte("ABC123ABC1"), nil),
					mock.StorageMock.EXPECT().
						SaveOneTimeCode(mock.Ctx, model.OneTimeCode{
							PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
							IssuedAt:  mock.Clock.Now(),
							IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
							ExpiresAt: mock.Clock.Now().Add(time.Minute),
							Username:  testUsername,
							Intent:    model.OTCIntentUserSessionElevation,
							Code:      []byte("ABC123ABC1"),
						}).
						Return("", fmt.Errorf("failed to insert")),
				)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred creating user session elevation One-Time Code challenge for user 'john': error occurred saving the challenge to the storage backend", "failed to insert")
			},
		},
		{
			"ShouldHandleOneFactorFailGenerateOTC",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.RandomMock.EXPECT().
						Read(gomock.Any()).
						SetArg(0, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x22, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15}).
						Return(16, nil),
					mock.RandomMock.EXPECT().
						BytesCustomErr(10, []byte(random.CharSetUnambiguousUpper)).
						Return(nil, fmt.Errorf("deadlock")),
				)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred creating user session elevation One-Time Code challenge for user 'john': error occurred generating the challenge", "failed to generate random bytes: deadlock")
			},
		},
		{
			"ShouldHandleOneFactorFailGenerateUUID",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.RandomMock.EXPECT().
						Read(gomock.Any()).
						Return(0, fmt.Errorf("random unavailable")),
				)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred creating user session elevation One-Time Code challenge for user 'john': error occurred generating the challenge", "failed to generate public id: random unavailable")
			},
		},
		{
			"ShouldHandleAnonymous",
			nil,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred creating user session elevation One-Time Code challenge", "user is anonymous")
			},
		},
		{
			"ShouldHandleGetSessionError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred creating user session elevation One-Time Code challenge: error occurred retrieving the user session data", "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.Characters = 10
			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.ElevationLifespan = time.Minute
			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.CodeLifespan = time.Minute

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Ctx.Providers.Random = mock.RandomMock

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			UserSessionElevationPOST(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestUserSessionElevationPUT(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		have           string
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleValidCode",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(time.Minute),
					Username:  testUsername,
					Intent:    model.OTCIntentUserSessionElevation,
					Code:      []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCode(mock.Ctx, testUsername, model.OTCIntentUserSessionElevation, "ABC123ABC1").
						Return(code, nil),
					mock.StorageMock.
						EXPECT().
						ConsumeOneTimeCode(mock.Ctx, code).
						Return(nil),
				)
			},
			`{"otc":"ABC123ABC1"}`,
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleValidCodeWithWhiteSpace",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(time.Minute),
					Username:  testUsername,
					Intent:    model.OTCIntentUserSessionElevation,
					Code:      []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCode(mock.Ctx, testUsername, model.OTCIntentUserSessionElevation, "ABC123ABC1").
						Return(code, nil),
					mock.StorageMock.
						EXPECT().
						ConsumeOneTimeCode(mock.Ctx, code).
						Return(nil),
				)
			},
			`{"otc":"ABC123ABC1   "}`,
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleAnonymous",
			nil,
			`{"otc":"ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge", "user is anonymous")
			},
		},
		{
			"ShouldHandleGetSessionError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
			},
			`{"otc":"ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge: error occurred retrieving the user session data", "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
			},
		},
		{
			"ShouldHandleInvalidCode",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(time.Minute),
					Username:  testUsername,
					Intent:    model.OTCIntentUserSessionElevation,
					Code:      []byte("WRONG"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCode(mock.Ctx, testUsername, model.OTCIntentUserSessionElevation, "ABC123ABC1").
						Return(code, nil),
				)
			},
			`{"otc":"ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge for user 'john'", "the code does not match the code stored in the challenge")
			},
		},
		{
			"ShouldHandleBadJSON",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"otc":ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge for user 'john': error parsing the request body", "unable to parse body: invalid character 'A' looking for beginning of value")
			},
		},
		{
			"ShouldHandleLongCodes",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"otc":"ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge for user 'john': expected maximum code length is 20 but the user provided code was 360 characters in length", "")
			},
		},
		{
			"ShouldHandleConsumeError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(time.Minute),
					Username:  testUsername,
					Intent:    model.OTCIntentUserSessionElevation,
					Code:      []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCode(mock.Ctx, testUsername, model.OTCIntentUserSessionElevation, "ABC123ABC1").
						Return(code, nil),
					mock.StorageMock.
						EXPECT().
						ConsumeOneTimeCode(mock.Ctx, code).
						Return(fmt.Errorf("failed to consume")),
				)
			},
			`{"otc":"ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge for user 'john': error occurred saving the consumption of the code to storage", "failed to consume")
			},
		},
		{
			"ShouldHandleAlreadyConsumedChallenge",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:         1,
					PublicID:   uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:   mock.Clock.Now(),
					IssuedIP:   model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt:  mock.Clock.Now().Add(time.Minute),
					Username:   testUsername,
					Intent:     model.OTCIntentUserSessionElevation,
					ConsumedAt: sql.NullTime{Valid: true, Time: mock.Clock.Now().Add(-time.Minute)},
					Code:       []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCode(mock.Ctx, testUsername, model.OTCIntentUserSessionElevation, "ABC123ABC1").
						Return(code, nil),
				)
			},
			`{"otc":"ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge for user 'john'", "the code challenge has already been consumed")
			},
		},
		{
			"ShouldHandleAlreadyRevokedChallenge",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(time.Minute),
					Username:  testUsername,
					Intent:    model.OTCIntentUserSessionElevation,
					RevokedAt: sql.NullTime{Valid: true, Time: mock.Clock.Now().Add(-time.Minute)},
					Code:      []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCode(mock.Ctx, testUsername, model.OTCIntentUserSessionElevation, "ABC123ABC1").
						Return(code, nil),
				)
			},
			`{"otc":"ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge for user 'john'", "the code challenge has been revoked")
			},
		},
		{
			"ShouldHandleAlreadyExpiredChallenge",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(-time.Minute),
					Username:  testUsername,
					Intent:    model.OTCIntentUserSessionElevation,
					Code:      []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCode(mock.Ctx, testUsername, model.OTCIntentUserSessionElevation, "ABC123ABC1").
						Return(code, nil),
				)
			},
			`{"otc":"ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge for user 'john'", "the code challenge has expired")
			},
		},
		{
			"ShouldHandleInvalidIntent",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(time.Minute),
					Username:  testUsername,
					Intent:    "abc",
					Code:      []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCode(mock.Ctx, testUsername, model.OTCIntentUserSessionElevation, "ABC123ABC1").
						Return(code, nil),
				)
			},
			`{"otc":"ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge for user 'john'", "the code challenge has the 'abc' intent but the 'use' intent is required")
			},
		},
		{
			"ShouldHandleNilChallenge",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCode(mock.Ctx, testUsername, model.OTCIntentUserSessionElevation, "ABC123ABC1").
						Return(nil, nil),
				)
			},
			`{"otc":"ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge for user 'john': error occurred retrieving the code challenge from the storage backend", "the code didn't match any recorded code challenges")
			},
		},
		{
			"ShouldHandleStorageError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCode(mock.Ctx, testUsername, model.OTCIntentUserSessionElevation, "ABC123ABC1").
						Return(nil, fmt.Errorf("not found")),
				)
			},
			`{"otc":"ABC123ABC1"}`,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating user session elevation One-Time Code challenge for user 'john': error occurred retrieving the code challenge from the storage backend", "not found")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.Characters = 10
			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.ElevationLifespan = time.Minute
			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.CodeLifespan = time.Minute

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Ctx.Providers.Random = mock.RandomMock

			if len(tc.have) != 0 {
				mock.Ctx.Request.SetBodyString(tc.have)
			}

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			UserSessionElevationPUT(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestUserSessionElevationDELETE(t *testing.T) {
	id := uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500"))

	eid := base64.RawURLEncoding.EncodeToString(id[:])

	testCases := []struct {
		name           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		have           string
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleValidCode",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(-time.Minute),
					Username:  testUsername,
					Intent:    model.OTCIntentUserSessionElevation,
					Code:      []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCodeByPublicID(mock.Ctx, id).
						Return(code, nil),
					mock.StorageMock.
						EXPECT().
						RevokeOneTimeCode(mock.Ctx, id, model.NewIP(net.ParseIP("0.0.0.0"))).
						Return(nil),
				)
			},
			eid,
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldHandleStorageRevokeError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(-time.Minute),
					Username:  testUsername,
					Intent:    model.OTCIntentUserSessionElevation,
					Code:      []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCodeByPublicID(mock.Ctx, id).
						Return(code, nil),
					mock.StorageMock.
						EXPECT().
						RevokeOneTimeCode(mock.Ctx, id, model.NewIP(net.ParseIP("0.0.0.0"))).
						Return(fmt.Errorf("failed to update")),
				)
			},
			eid,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking user session elevation One-Time Code challenge: error occurred saving the revocation to the storage backend", "failed to update")
			},
		},
		{
			"ShouldHandleStorageRevokeErrorBadIntent",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(-time.Minute),
					Username:  testUsername,
					Intent:    "abc",
					Code:      []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCodeByPublicID(mock.Ctx, id).
						Return(code, nil),
				)
			},
			eid,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking user session elevation One-Time Code challenge", "the code challenge has the 'abc' intent but the 'use' intent is required")
			},
		},
		{
			"ShouldHandleStorageRevokeErrorConsumed",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:         1,
					PublicID:   uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:   mock.Clock.Now(),
					IssuedIP:   model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt:  mock.Clock.Now().Add(-time.Minute),
					Username:   testUsername,
					ConsumedAt: sql.NullTime{Valid: true, Time: mock.Clock.Now().Add(-time.Minute)},
					Intent:     model.OTCIntentUserSessionElevation,
					Code:       []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCodeByPublicID(mock.Ctx, id).
						Return(code, nil),
				)
			},
			eid,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking user session elevation One-Time Code challenge", "the code challenge has already been consumed")
			},
		},
		{
			"ShouldHandleStorageRevokeErrorRevoked",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				code := &model.OneTimeCode{
					ID:        1,
					PublicID:  uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500")),
					IssuedAt:  mock.Clock.Now(),
					IssuedIP:  model.NewIP(net.ParseIP("0.0.0.0")),
					ExpiresAt: mock.Clock.Now().Add(-time.Minute),
					Username:  testUsername,
					RevokedAt: sql.NullTime{Valid: true, Time: mock.Clock.Now().Add(-time.Minute)},
					Intent:    model.OTCIntentUserSessionElevation,
					Code:      []byte("ABC123ABC1"),
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCodeByPublicID(mock.Ctx, id).
						Return(code, nil),
				)
			},
			eid,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking user session elevation One-Time Code challenge", "the code challenge has already been revoked")
			},
		},
		{
			"ShouldHandleStorageRevokeErrorStorage",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadOneTimeCodeByPublicID(mock.Ctx, id).
						Return(nil, fmt.Errorf("invalid user")),
				)
			},
			eid,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking user session elevation One-Time Code challenge: error occurred retrieving the code challenge from the storage backend", "invalid user")
			},
		},
		{
			"ShouldHandleBadUUID",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			base64.RawURLEncoding.EncodeToString([]byte("abc")),
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking user session elevation One-Time Code challenge: error occurred parsing the identifier", "invalid UUID (got 3 bytes)")
			},
		},
		{
			"ShouldHandleBadBase64",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.DisplayName = testDisplayName
				us.Emails = []string{"john@example.com"}

				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			"=====123123",
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking user session elevation One-Time Code challenge: error occurred decoding the identifier", "illegal base64 data at input byte 0")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.Characters = 10
			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.ElevationLifespan = time.Minute
			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.CodeLifespan = time.Minute

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Ctx.Providers.Random = mock.RandomMock

			if len(tc.have) != 0 {
				mock.Ctx.SetUserValue("id", tc.have)
			}

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			UserSessionElevateDELETE(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}
