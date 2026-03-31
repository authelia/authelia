package handlers

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestFirstFactorPasskeyGET(t *testing.T) {
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

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			regexp.MustCompile(`^\{"status":"OK","data":\{"publicKey":\{"challenge":"[a-zA-Z0-9/_-]+={0,2}","timeout":60000,"rpId":"login.example.com"}}}$`),
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				require.NotNil(t, us.WebAuthn)
				require.NotNil(t, us.WebAuthn.SessionData)

				assert.Equal(t, "", us.WebAuthn.Description)
				assert.Equal(t, []byte(nil), us.WebAuthn.UserID)
			},
		},
		{
			"ShouldFailAlreadyLoggedIn",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.AuthenticationMethodRefs.KnowledgeBasedAuthentication = true

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Authentication failed, please retry later."}$`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				assert.Nil(t, us.WebAuthn)
				assert.NoError(t, err)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn passkey authentication challenge: error occurred retrieving the user session data", "user is already authenticated")
			},
		},
		{
			"ShouldFailGetSession",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "123")
				mock.Ctx.Request.Header.Set("X-Forwarded-Host", "____")
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Authentication failed, please retry later."}$`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn passkey authentication challenge: error occurred retrieving the user session data", "unable to retrieve session cookie domain: failed to parse X-Original-URL header: parse \"123\": invalid URI for request")
			},
		},
		{
			"ShouldFailCantGetProvider",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)
				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.Ctx.Request.Header.Set("X-Original-URL", "123")
				mock.Ctx.Request.Header.Set("X-Forwarded-Host", "____")
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Authentication failed, please retry later."}$`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				assert.Nil(t, us.WebAuthn)
				assert.NoError(t, err)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn passkey authentication challenge: error occurred provisioning the configuration", "failed to parse X-Original-URL header: parse \"123\": invalid URI for request")
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

			FirstFactorPasskeyGET(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Regexp(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.validateResponse != nil {
				tc.validateResponse(t, mock)
			}
		})
	}
}

func TestFirstFactorPasskeyPOST(t *testing.T) {
	const (
		dataReqFmt         = `{"response":{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtwFAAAAAw","userHandle":"ZXhhbXBsZQ","clientDataJSON":"%s","signature":"MEQCIBlJ2Fxf6ZwLNTCQglz0AW0pD4HlU8W5Yk696jjfxVxhAiAhAMkLh8iKyhW6zSmzwfQDjMF2nKjVHzEs7jLHRPDZ2A"},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}`
		dataReqFmtKLI      = `{"keepMeLoggedIn":true,"response":{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtwFAAAAAw","userHandle":"ZXhhbXBsZQ","clientDataJSON":"%s","signature":"MEQCIBlJ2Fxf6ZwLNTCQglz0AW0pD4HlU8W5Yk696jjfxVxhAiAhAMkLh8iKyhW6zSmzwfQDjMF2nKjVHzEs7jLHRPDZ2A"},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}`
		dataReqNoHandleFmt = `{"response":{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtwFAAAAAw","clientDataJSON":"%s","signature":"MEQCIBlJ2Fxf6ZwLNTCQglz0AW0pD4HlU8W5Yk696jjfxVxhAiAhAMkLh8iKyhW6zSmzwfQDjMF2nKjVHzEs7jLHRPDZ2A"},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}`
		dataClientJSON     = `{"type":"webauthn.get","challenge":"in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE","origin":"%s","crossOrigin":false,"other_keys_can_be_added_here":"do not compare clientDataJSON against a template. See https://goo.gl/yabPex"}`
	)

	var (
		dataReqGood         = fmt.Sprintf(dataReqFmt, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(dataClientJSON, "https://login.example.com:8080"))))
		dataReqGoodKLI      = fmt.Sprintf(dataReqFmtKLI, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(dataClientJSON, "https://login.example.com:8080"))))
		dataReqBadRPIDHash  = fmt.Sprintf(dataReqFmt, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(dataClientJSON, "http://example.com"))))
		dataReqNoHandleGood = fmt.Sprintf(dataReqNoHandleFmt, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(dataClientJSON, "https://login.example.com:8080"))))
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
			"ShouldFailAlreadyLoggedIn",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				us.AuthenticationMethodRefs.KnowledgeBasedAuthentication = true

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.NotNil(t, us.WebAuthn)
			},
		},
		{
			"ShouldFailGetSession",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "123")
				mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "123")
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn passkey authentication challenge: error occurred retrieving the user session data", "unable to retrieve session cookie domain: failed to parse X-Original-URL header: parse \"123\": invalid URI for request")
			},
		},
		{
			"ShouldFailSessionData",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)
				require.NoError(t, mock.Ctx.SaveSession(us))
				mock.Ctx.Request.Header.Set("X-Original-URL", "123")
				mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "123")
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				assert.Nil(t, us.WebAuthn)
				assert.NoError(t, err)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge: error occurred retrieving the user session data", "challenge session data is not present")
			},
		},
		{
			"ShouldFailGetProvider",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)

				us, err := mock.Ctx.GetSession()

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, err)
				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.Ctx.Request.Header.Set("X-Original-URL", "123")
				mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "123")
			},
			dataReqGood,
			"",
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				assert.Nil(t, us.WebAuthn)
				assert.NoError(t, err)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge: error occurred provisioning the configuration", "failed to parse X-Original-URL header: parse \"123\": invalid URI for request")
			},
		},
		{
			name:   "ShouldSuccess",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC().Add(time.Second * -10), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       2,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				updated := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       credential.CreatedAt,
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       3,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, updated).
						Return(nil),
					mock.UserProviderMock.EXPECT().
						GetDetails(gomock.Eq(testUsername)).
						Return(&authentication.UserDetails{Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadBannedIP(mock.Ctx, gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
						Return(nil, nil),
					mock.StorageMock.EXPECT().
						LoadBannedUser(mock.Ctx, gomock.Eq(testUsername)).
						Return(nil, nil),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: true,
							Banned:     false,
							Username:   testUsername,
							Type:       regulation.AuthTypePasskey,
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
					mock.StorageMock.EXPECT().
						UpdateUserSignInDateByUsername(mock.Ctx, gomock.Eq(testUsername)).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusOK,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)
			},
		},
		{
			name:   "ShouldSuccessUpgrade",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.WebAuthn.EnablePasskeyUpgrade = true

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC().Add(time.Second * -10), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       2,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				updated := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       credential.CreatedAt,
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       3,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, updated).
						Return(nil),
					mock.UserProviderMock.EXPECT().
						GetDetails(gomock.Eq(testUsername)).
						Return(&authentication.UserDetails{Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadBannedIP(mock.Ctx, gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
						Return(nil, nil),
					mock.StorageMock.EXPECT().
						LoadBannedUser(mock.Ctx, gomock.Eq(testUsername)).
						Return(nil, nil),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: true,
							Banned:     false,
							Username:   testUsername,
							Type:       regulation.AuthTypePasskey,
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
					mock.StorageMock.EXPECT().
						UpdateUserSignInDateByUsername(mock.Ctx, gomock.Eq(testUsername)).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusOK,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)
			},
		},
		{
			name:   "ShouldNotAllowBannedUserToUsePasskey",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          mock.Ctx.Clock.Now().UTC().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       mock.Ctx.Clock.Now().UTC().Add(time.Second * -10),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC().Add(time.Second * -10), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       2,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				updated := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       credential.CreatedAt,
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       3,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, updated).
						Return(nil),
					mock.UserProviderMock.EXPECT().
						GetDetails(gomock.Eq(testUsername)).
						Return(&authentication.UserDetails{Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadBannedIP(mock.Ctx, gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
						Return(nil, nil),
					mock.StorageMock.EXPECT().
						LoadBannedUser(mock.Ctx, gomock.Eq(testUsername)).
						Return([]model.BannedUser{{ID: 1, Time: mock.Ctx.Clock.Now().UTC().Add(time.Second - 10), Expires: sql.NullTime{Time: mock.Ctx.Clock.Now().UTC().Add(time.Minute), Valid: true}, Username: testUsername, Source: "Passkey"}}, nil),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     true,
							Username:   testUsername,
							Type:       regulation.AuthTypePasskey,
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusUnauthorized,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexp.MustCompile(`^Unsuccessful Passkey authentication attempt by user 'john' and they are banned until \d+:\d+:\d+(AM|PM) on \w+ \d+ \d+ \(\+\d+:\d+\)$`), nil)
			},
		},
		{
			name:   "ShouldNotAllowBannedIPToUsePasskey",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          mock.Ctx.Clock.Now().UTC().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       mock.Ctx.Clock.Now().UTC().Add(time.Second * -10),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC().Add(time.Second * -10), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       2,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				updated := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       credential.CreatedAt,
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       3,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, updated).
						Return(nil),
					mock.UserProviderMock.EXPECT().
						GetDetails(gomock.Eq(testUsername)).
						Return(&authentication.UserDetails{Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadBannedIP(mock.Ctx, gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
						Return([]model.BannedIP{{ID: 1, Time: mock.Ctx.Clock.Now().UTC().Add(time.Second - 10), Expires: sql.NullTime{Time: mock.Ctx.Clock.Now().UTC().Add(time.Minute), Valid: true}}}, nil),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     true,
							Username:   testUsername,
							Type:       regulation.AuthTypePasskey,
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusUnauthorized,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexp.MustCompile(`^Unsuccessful Passkey authentication attempt by user 'john' and they are banned until \d+:\d+:\d+(AM|PM) on \w+ \d+ \d+ \(\+\d+:\d+\)$`), nil)
			},
		},
		{
			name:   "ShouldHandleBanCheckError",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          mock.Ctx.Clock.Now().UTC().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       mock.Ctx.Clock.Now().UTC().Add(time.Second * -10),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC().Add(time.Second * -10), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       2,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				updated := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       credential.CreatedAt,
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       3,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, updated).
						Return(nil),
					mock.UserProviderMock.EXPECT().
						GetDetails(gomock.Eq(testUsername)).
						Return(&authentication.UserDetails{Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadBannedIP(mock.Ctx, gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
						Return(nil, fmt.Errorf("broken")),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusUnauthorized,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to perform Passkey authentication regulation for user 'john'", "broken")
			},
		},
		{
			name:   "ShouldSuccessKeepLoggedIn",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.Session.RememberMe = time.Hour

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: time.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       2,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				updated := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       credential.CreatedAt,
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       3,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, updated).
						Return(nil),
					mock.UserProviderMock.EXPECT().
						GetDetails(gomock.Eq(testUsername)).
						Return(&authentication.UserDetails{Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadBannedIP(mock.Ctx, gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
						Return(nil, nil),
					mock.StorageMock.EXPECT().
						LoadBannedUser(mock.Ctx, gomock.Eq(testUsername)).
						Return(nil, nil),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: true,
							Banned:     false,
							Username:   testUsername,
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
					mock.StorageMock.EXPECT().
						UpdateUserSignInDateByUsername(mock.Ctx, gomock.Eq(testUsername)).
						Return(nil),
				)
			},
			have:           dataReqGoodKLI,
			expectedStatus: fasthttp.StatusOK,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)
				assert.True(t, us.KeepMeLoggedIn)
			},
		},
		{
			name:   "ShouldFailMark",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC().Add(time.Second * -10), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       2,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				updated := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       credential.CreatedAt,
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       3,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, updated).
						Return(nil),
					mock.UserProviderMock.EXPECT().
						GetDetails(gomock.Eq(testUsername)).
						Return(&authentication.UserDetails{Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadBannedIP(mock.Ctx, gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
						Return(nil, nil),
					mock.StorageMock.EXPECT().
						LoadBannedUser(mock.Ctx, gomock.Eq(testUsername)).
						Return(nil, nil),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: true,
							Banned:     false,
							Username:   testUsername,
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(fmt.Errorf("error marking auth")),
					mock.StorageMock.EXPECT().
						UpdateUserSignInDateByUsername(mock.Ctx, gomock.Eq(testUsername)).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusOK,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to record Passkey authentication attempt", "error marking auth")
			},
		},
		{
			name:   "ShouldFailUserDetails",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC().Add(time.Second * -10), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       2,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				updated := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       credential.CreatedAt,
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       3,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, updated).
						Return(nil),
					mock.UserProviderMock.EXPECT().
						GetDetails(gomock.Eq(testUsername)).
						Return(nil, fmt.Errorf("failed to get details")),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusForbidden,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge for user 'john': error retrieving user details", "failed to get details")
			},
		},
		{
			name:   "ShouldFailCloned",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().Add(-1 * time.Hour), Valid: true},
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
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				updated := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       credential.CreatedAt,
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       10000,
					CloneWarning:    true,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, updated).
						Return(nil),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusForbidden,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge for user 'john': error occurred validating the authenticator response", "authenticator sign count indicates that it is cloned")
			},
		},
		{
			name:   "ShouldFailCredentialNotFound",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().Add(-1 * time.Hour), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64([]byte("wrong-key")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       1,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusForbidden,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge: error performing the login validation", "Unable to find the credential for the returned credential ID (invalid_request)")
			},
		},
		{
			name:   "ShouldFailUpdateCredential",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				credential := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       time.Now(),
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC().Add(time.Second * -10), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       2,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				updated := model.WebAuthnCredential{
					ID:              1,
					CreatedAt:       credential.CreatedAt,
					LastUsedAt:      sql.NullTime{Time: mock.Clock.Now().UTC(), Valid: true},
					RPID:            "login.example.com",
					Username:        testUsername,
					Description:     "test",
					KID:             model.NewBase64(decode("rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
					AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
					AttestationType: "packed",
					Attachment:      "cross-platform",
					Transport:       "usb",
					SignCount:       3,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						UpdateWebAuthnCredentialSignIn(mock.Ctx, updated).
						Return(fmt.Errorf("bad data")),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusForbidden,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge for user 'john': error occurred saving the credential sign-in information to the storage backend", "bad data")
			},
		},
		{
			name:   "ShouldFailGetCredentials",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return(nil, fmt.Errorf("failed to get creds")),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusForbidden,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge: error performing the login validation", "Failed to lookup Client-side Discoverable Credential: failed to get creds (invalid_request)")
			},
		},
		{
			name:   "ShouldFailGetHandle",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(nil, fmt.Errorf("bad handle")),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusForbidden,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge: error performing the login validation", "Failed to lookup Client-side Discoverable Credential: bad handle (invalid_request)")
			},
		},
		{
			name:   "ShouldFailGetHandleBlank",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			have:           dataReqNoHandleGood,
			expectedStatus: fasthttp.StatusForbidden,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge: error performing the login validation", "Client-side Discoverable Assertion was attempted with a blank User Handle (invalid_request)")
			},
		},
		{
			name:   "ShouldFailHeaders",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				mock.Ctx.Request.Header.Set("X-Original-URL", "123")
				mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "123")
				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			have:           dataReqGood,
			expectedStatus: fasthttp.StatusForbidden,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge: error occurred provisioning the configuration", "failed to parse X-Original-URL header: parse \"123\": invalid URI for request")
			},
		},
		{
			name:   "ShouldFailBadRPIDHash",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
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
					SignCount:       2,
					CloneWarning:    false,
					Discoverable:    true,
					Present:         true,
					Verified:        true,
					BackupEligible:  false,
					BackupState:     false,
					PublicKey:       []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
				}

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).
						Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil),
					mock.StorageMock.EXPECT().
						LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).
						Return([]model.WebAuthnCredential{credential}, nil),
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)
			},
			have:           dataReqBadRPIDHash,
			expectedStatus: fasthttp.StatusForbidden,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge: error performing the login validation", "Error validating origin (verification_error): Expected Values: [https://login.example.com:8080], Received: http://example.com")
			},
		},
		{
			name:   "ShouldFailBadJSON",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setUpMockClock(mock)

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			have:           "not json",
			expectedStatus: fasthttp.StatusBadRequest,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge: error parsing the request body", "unable to parse body: invalid character 'o' in literal null (expecting 'u')")
			},
		},
		{
			name:   "ShouldFailBadResponseJSON",
			config: &schema.DefaultWebAuthnConfiguration,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
							Time:       mock.Ctx.Clock.Now(),
							Successful: false,
							Banned:     false,
							Username:   "",
							Type:       "Passkey",
							RemoteIP:   model.NullIP{IP: net.ParseIP("0.0.0.0")},
						})).
						Return(nil),
				)

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.WebAuthn = &session.WebAuthn{
					SessionData: &webauthn.SessionData{
						Challenge:        "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE",
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			have:           `{"response":true}`,
			expectedStatus: fasthttp.StatusBadRequest,
			expectedf: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a WebAuthn passkey authentication challenge: error parsing the request body", "Parse error for Assertion (invalid_request): json: cannot unmarshal bool into Go value of type protocol.CredentialAssertionResponse")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			setUpMockClock(mock)

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

			FirstFactorPasskeyPOST(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Regexp(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}
