package handlers

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestWebAuthnAssertionGET(t *testing.T) {
	decode := func(in string) []byte {
		value, err := base64.StdEncoding.DecodeString(in)
		if err != nil {
			t.Fatal("Failed to decode base64 string:", err)
		}

		return value
	}

	testCases := []struct {
		name             string
		config           *schema.WebAuthn
		setup            func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected         *regexp.Regexp
		expectedStatus   int
		validateResponse func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldSuccess",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: time.Now(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       4,
					CloneWarning:    false,
					Discoverable:    false,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, "login.example.com", testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: "login.example.com", Username: testUsername, UserID: "ZytlJlVuWzdgN2BxTyI8Uy9uS2xpJSdsT2ZsJUA5UEBve1c2NENCKDNSWWphaGVCJEhlQ3wpYT9HQGBwIi8zQA=="}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, "login.example.com", testUsername).
						Return([]model.WebAuthnCredential{credential}, nil),
				)
			},
			regexp.MustCompile(`^\{"status":"OK","data":\{"publicKey":\{"challenge":"[a-zA-Z0-9/_-]+={0,2}","timeout":60000,"rpId":"login.example.com","allowCredentials":\[\{"type":"public-key","id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","transports":\["usb"]}],"userVerification":"preferred"}}}$`),
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				require.NotNil(t, us.WebAuthn)
				require.NotNil(t, us.WebAuthn.SessionData)

				assert.Equal(t, "", us.WebAuthn.Description)
				assert.Equal(t, []byte{0x5a, 0x79, 0x74, 0x6c, 0x4a, 0x6c, 0x56, 0x75, 0x57, 0x7a, 0x64, 0x67, 0x4e, 0x32, 0x42, 0x78, 0x54, 0x79, 0x49, 0x38, 0x55, 0x79, 0x39, 0x75, 0x53, 0x32, 0x78, 0x70, 0x4a, 0x53, 0x64, 0x73, 0x54, 0x32, 0x5a, 0x73, 0x4a, 0x55, 0x41, 0x35, 0x55, 0x45, 0x42, 0x76, 0x65, 0x31, 0x63, 0x32, 0x4e, 0x45, 0x4e, 0x43, 0x4b, 0x44, 0x4e, 0x53, 0x57, 0x57, 0x70, 0x68, 0x61, 0x47, 0x56, 0x43, 0x4a, 0x45, 0x68, 0x6c, 0x51, 0x33, 0x77, 0x70, 0x59, 0x54, 0x39, 0x48, 0x51, 0x47, 0x42, 0x77, 0x49, 0x69, 0x38, 0x7a, 0x51, 0x41, 0x3d, 0x3d}, us.WebAuthn.UserID)
			},
		},
		{
			"ShouldHandleCredentialError",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, "login.example.com", testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: "login.example.com", Username: testUsername, UserID: "ZytlJlVuWzdgN2BxTyI8Uy9uS2xpJSdsT2ZsJUA5UEBve1c2NENCKDNSWWphaGVCJEhlQ3wpYT9HQGBwIi8zQA=="}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, "login.example.com", testUsername).
						Return(nil, fmt.Errorf("failed")),
				)
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Authentication failed, please retry later."}`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn authentication challenge for user 'john': error occurred retrieving the WebAuthn user configuration from the storage backend", "failed")
			},
		},
		{
			"ShouldHandleAnonymous",
			&schema.DefaultWebAuthnConfiguration,
			nil,
			regexp.MustCompile(`^\{"status":"KO","message":"Authentication failed, please retry later."}`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn authentication challenge", "user is anonymous")
			},
		},
		{
			"ShouldHandleBadCookieDomain",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Authentication failed, please retry later."}`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn authentication challenge: error occurred retrieving the user session data", "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
			},
		},
		{
			"ShouldHandleBadOrigin",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.Ctx.Request.Header.Set("X-Original-URL", "!@NJK#N!@#IKJ!@NJK")
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Authentication failed, please retry later."}`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn authentication challenge for user 'john': error occurred provisioning the configuration", "failed to parse X-Original-URL header: parse \"!@NJK#N!@#IKJ!@NJK\": invalid URI for request")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			if tc.config != nil {
				mock.Ctx.Configuration.WebAuthn = *tc.config
			}

			mock.Ctx.Request.Header.Set("X-Original-URL", "https://login.example.com:8080")

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			WebAuthnAssertionGET(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Regexp(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.validateResponse != nil {
				tc.validateResponse(t, mock)
			}
		})
	}
}

func TestWebAuthnAssertionPOST(t *testing.T) {
	const (
		dataReqFmt     = `{"response":{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtwFAAAAAw","clientDataJSON":"%s","signature":"MEQCIBlJ2Fxf6ZwLNTCQglz0AW0pD4HlU8W5Yk696jjfxVxhAiAhAMkLh8iKyhW6zSmzwfQDjMF2nKjVHzEs7jLHRPDZ2A"},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}`
		dataClientJSON = `{"type":"webauthn.get","challenge":"in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE","origin":"%s","crossOrigin":false,"other_keys_can_be_added_here":"do not compare clientDataJSON against a template. See https://goo.gl/yabPex"}`
	)

	var (
		dataReqGood        = fmt.Sprintf(dataReqFmt, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(dataClientJSON, "https://login.example.com:8080"))))
		dataReqBadRPIDHash = fmt.Sprintf(dataReqFmt, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(dataClientJSON, "http://example.com"))))
	)

	decode := func(in string) []byte {
		value, err := base64.StdEncoding.DecodeString(in)
		if err != nil {
			t.Fatal("Failed to decode base64 string:", err)
		}

		return value
	}

	testCases := []struct {
		name           string
		config         *schema.WebAuthn
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		have           string
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldSuccess",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: time.Now(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       0,
					CloneWarning:    false,
					Discoverable:    false,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, "login.example.com", testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: "login.example.com", Username: testUsername, UserID: string(decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="))}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, "login.example.com", testUsername).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.
						EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, gomock.Any()).
						Return(nil),
					mock.StorageMock.
						EXPECT().
						AppendAuthenticationLog(mock.Ctx, gomock.Any()).
						Return(nil),
				)
			},
			dataReqGood,
			"",
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)
			},
		},
		{
			"ShouldFailClone",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: time.Now(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       10000,
					CloneWarning:    false,
					Discoverable:    false,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, "login.example.com", testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: "login.example.com", Username: testUsername, UserID: string(decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="))}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, "login.example.com", testUsername).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.
						EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, gomock.Any()).
						Return(nil),
				)
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn authentication challenge for user 'john': error occurred validating the authenticator response", "authenticator sign count indicates that it is cloned")
			},
		},
		{
			"ShouldFailUpdateSignIn",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: time.Now(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       0,
					CloneWarning:    false,
					Discoverable:    false,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, "login.example.com", testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: "login.example.com", Username: testUsername, UserID: string(decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="))}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, "login.example.com", testUsername).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.
						EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, gomock.Any()).
						Return(fmt.Errorf("failed to update")),
				)
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn authentication challenge for user 'john': error occurred saving the credential sign-in information to the storage backend", "failed to update")
			},
		},
		{
			"ShouldFailAuthLogFailure",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       mock.Clock.Now().UTC(),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().Add(time.Minute).UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       0,
					CloneWarning:    false,
					Discoverable:    false,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, "login.example.com", testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: "login.example.com", Username: testUsername, UserID: string(decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="))}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, "login.example.com", testUsername).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.
						EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, gomock.Any()).
						Return(nil),
					mock.StorageMock.
						EXPECT().
						AppendAuthenticationLog(mock.Ctx, gomock.Any()).
						Return(fmt.Errorf("bad record")),
				)

				mock.Clock.Set(mock.Clock.Now().Add(2 * time.Minute))
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Unable to mark WebAuthn authentication attempt by user 'john'", "bad record")
			},
		},
		{
			"ShouldFailBadRPIDHash",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       mock.Clock.Now().UTC(),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().Add(time.Minute).UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       0,
					CloneWarning:    false,
					Discoverable:    false,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, "login.example.com", testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: "login.example.com", Username: testUsername, UserID: string(decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="))}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, "login.example.com", testUsername).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.
						EXPECT().
						AppendAuthenticationLog(mock.Ctx, gomock.Any()).
						Return(nil),
				)

				mock.Clock.Set(mock.Clock.Now().Add(2 * time.Minute))
			},
			dataReqBadRPIDHash,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Unsuccessful WebAuthn authentication attempt by user 'john'", "Error validating origin (verification_error): Expected Values: [https://login.example.com:8080], Received: http://example.com")
			},
		},
		{
			"ShouldFailBadResponse",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.Clock.Set(mock.Clock.Now().Add(2 * time.Minute))
			},
			`{"response":{"id":true,"rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtwFAAAAAw","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiaW4xY0wtb1dmU2pTZDd1dXdVdnYybmRPQW1SWGIwY09BYlVvVHRBcXZHRSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZSwib3RoZXJfa2V5c19jYW5fYmVfYWRkZWRfaGVyZSI6ImRvIG5vdCBjb21wYXJlIGNsaWVudERhdGFKU09OIGFnYWluc3QgYSB0ZW1wbGF0ZS4gU2VlIGh0dHBzOi8vZ29vLmdsL3lhYlBleCJ9","signature":"MEQCIBlJ2Fxf6ZwLNTCQglz0AW0pD4HlU8W5Yk696jjfxVxhAiAhAMkLh8iKyhW6zSmzwfQDjMF2nKjVHzEs7jLHRPDZ2A"},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}`,
			"",
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn authentication challenge for user 'john': error parsing the request body", "Parse error for Assertion (invalid_request): json: cannot unmarshal bool into Go struct field CredentialAssertionResponse.id of type string")
			},
		},
		{
			"ShouldFailBadJSON",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.Clock.Set(mock.Clock.Now().Add(2 * time.Minute))
			},
			`{"response:{"id":true,"rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtwFAAAAAw","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiaW4xY0wtb1dmU2pTZDd1dXdVdnYybmRPQW1SWGIwY09BYlVvVHRBcXZHRSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZSwib3RoZXJfa2V5c19jYW5fYmVfYWRkZWRfaGVyZSI6ImRvIG5vdCBjb21wYXJlIGNsaWVudERhdGFKU09OIGFnYWluc3QgYSB0ZW1wbGF0ZS4gU2VlIGh0dHBzOi8vZ29vLmdsL3lhYlBleCJ9","signature":"MEQCIBlJ2Fxf6ZwLNTCQglz0AW0pD4HlU8W5Yk696jjfxVxhAiAhAMkLh8iKyhW6zSmzwfQDjMF2nKjVHzEs7jLHRPDZ2A"},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}`,
			"",
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn authentication challenge for user 'john': error parsing the request body", "unable to parse body: invalid character 'i' after object key")
			},
		},
		{
			"ShouldFailOrigin",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.Clock.Set(mock.Clock.Now().Add(2 * time.Minute))

				// This malformed URL is chosen to invoke the url.Parse errors.
				mock.Ctx.Request.Header.Set("X-Original-URL", "!@#*(&jklqnwdkjqwe")
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn authentication challenge for user 'john': error occurred provisioning the configuration", "failed to parse X-Original-URL header: parse \"!@#*(&jklqnwdkjqwe\": invalid URI for request")
			},
		},
		{
			"ShouldFailNilSession",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn authentication challenge for user 'john': error occurred retrieving the user session data", "challenge session data is not present")
			},
		},
		{
			"ShouldFailAnonymous",
			&schema.DefaultWebAuthnConfiguration,
			nil,
			`{"response:{"id":true,"rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtwFAAAAAw","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiaW4xY0wtb1dmU2pTZDd1dXdVdnYybmRPQW1SWGIwY09BYlVvVHRBcXZHRSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZSwib3RoZXJfa2V5c19jYW5fYmVfYWRkZWRfaGVyZSI6ImRvIG5vdCBjb21wYXJlIGNsaWVudERhdGFKU09OIGFnYWluc3QgYSB0ZW1wbGF0ZS4gU2VlIGh0dHBzOi8vZ29vLmdsL3lhYlBleCJ9","signature":"MEQCIBlJ2Fxf6ZwLNTCQglz0AW0pD4HlU8W5Yk696jjfxVxhAiAhAMkLh8iKyhW6zSmzwfQDjMF2nKjVHzEs7jLHRPDZ2A"},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}`,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn authentication challenge", "user is anonymous")
			},
		},
		{
			"ShouldFailBadSessionDomain",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
			},
			`{"response:{"id":true,"rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtwFAAAAAw","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiaW4xY0wtb1dmU2pTZDd1dXdVdnYybmRPQW1SWGIwY09BYlVvVHRBcXZHRSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZSwib3RoZXJfa2V5c19jYW5fYmVfYWRkZWRfaGVyZSI6ImRvIG5vdCBjb21wYXJlIGNsaWVudERhdGFKU09OIGFnYWluc3QgYSB0ZW1wbGF0ZS4gU2VlIGh0dHBzOi8vZ29vLmdsL3lhYlBleCJ9","signature":"MEQCIBlJ2Fxf6ZwLNTCQglz0AW0pD4HlU8W5Yk696jjfxVxhAiAhAMkLh8iKyhW6zSmzwfQDjMF2nKjVHzEs7jLHRPDZ2A"},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}`,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn authentication challenge: error occurred retrieving the user session data", "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
			},
		},
		{
			"ShouldFailLoadWebAuthnUser",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, "login.example.com", testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: "login.example.com", Username: testUsername, UserID: string(decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="))}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, "login.example.com", testUsername).
						Return(nil, fmt.Errorf("failed to load credentials")),
				)
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn authentication challenge for user 'john': error occurred retrieving the WebAuthn user configuration from the storage backend", "failed to load credentials")
			},
		},
		{
			"ShouldFailUpdateAuthLog",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, "login.example.com", testUsername).
						Return(nil, fmt.Errorf("failed load user")),
				)
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn authentication challenge for user 'john': error occurred retrieving the WebAuthn user configuration from the storage backend", "failed load user")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			if tc.config != nil {
				mock.Ctx.Configuration.WebAuthn = *tc.config
			}

			if len(tc.have) != 0 {
				mock.Ctx.Request.SetBodyString(tc.have)
			}

			mock.Ctx.Request.Header.Set("X-Original-URL", "https://login.example.com:8080")

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			WebAuthnAssertionPOST(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Regexp(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

//nolint:godot
/*
[00] sign challenge response post body {"response":{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtwFAAAAAw","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiaW4xY0wtb1dmU2pTZDd1dXdVdnYybmRPQW1SWGIwY09BYlVvVHRBcXZHRSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZSwib3RoZXJfa2V5c19jYW5fYmVfYWRkZWRfaGVyZSI6ImRvIG5vdCBjb21wYXJlIGNsaWVudERhdGFKU09OIGFnYWluc3QgYSB0ZW1wbGF0ZS4gU2VlIGh0dHBzOi8vZ29vLmdsL3lhYlBleCJ9","signature":"MEQCIBlJ2Fxf6ZwLNTCQglz0AW0pD4HlU8W5Yk696jjfxVxhAiAhAMkLh8iKyhW6zSmzwfQDjMF2nKjVHzEs7jLHRPDZ2A"},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}
[00] sign challenge session data {"challenge":"in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE","user_id":"OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA==","allowed_credentials":["rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU="],"expires":"2023-11-26T08:02:44.624158134Z","userVerification":"preferred","description":""}

[00] registration challenge response post body {"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"attestationObject":"o2NmbXRmcGFja2VkZ2F0dFN0bXSjY2FsZyZjc2lnWEcwRQIhAI505i2XKRL3xsFcSNRz6crTg7_AIpJIsVOjuv8MKW6jAiBJCrqGIc9kKSgS1x54lq53SWUpVNXmlakZfp5NIXrJcmN4NWOBWQHeMIIB2jCCAX2gAwIBAgIBATANBgkqhkiG9w0BAQsFADBgMQswCQYDVQQGEwJVUzERMA8GA1UECgwIQ2hyb21pdW0xIjAgBgNVBAsMGUF1dGhlbnRpY2F0b3IgQXR0ZXN0YXRpb24xGjAYBgNVBAMMEUJhdGNoIENlcnRpZmljYXRlMB4XDTE3MDcxNDAyNDAwMFoXDTQzMTEyMTA4MDAzOVowYDELMAkGA1UEBhMCVVMxETAPBgNVBAoMCENocm9taXVtMSIwIAYDVQQLDBlBdXRoZW50aWNhdG9yIEF0dGVzdGF0aW9uMRowGAYDVQQDDBFCYXRjaCBDZXJ0aWZpY2F0ZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABI1hfmXJUI5kvMVnOsgqZ5naPBRGaCwljEY__99Y39L6Pmw3i1PXlcSk3_tBme3Xhi8jq68CA7S4kRugVpmU4QGjJTAjMAwGA1UdEwEB_wQCMAAwEwYLKwYBBAGC5RwCAQEEBAMCBSAwDQYJKoZIhvcNAQELBQADSAAwRQIgI4PXvgxbCt2L3tk_p22e3QmDCw0ZOPJ6dIJcp2LoTRACIQDqhWGzBtSCdnTiGq2CjhApHJxER1tBy9vRbRaioTz-ZGhhdXRoRGF0YVikDGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiYXFfQVhkdnNETXNLV18xYVkzMVhRaFUxN1pNZzFpMFRLMDEzRHd1a0IyVSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0","transports":["usb"],"publicKeyAlgorithm":-7,"publicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEuBHGqg5RF-1k2nt6MEw4lBdvrfVD77Dlx83VLu9b3rerjXRKRLRRQlF_USnsrSYHCSKAp2UzGVTvZAp8daWysw","authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM"},"type":"public-key","clientExtensionResults":{"credProps":{"rk":false}},"authenticatorAttachment":"cross-platform"}
[00] registration challenge session data {"challenge":"aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U","user_id":"OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA==","expires":"2023-11-26T08:01:39.891837286Z","userVerification":"preferred","description":"test"}
*/
