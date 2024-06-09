package handlers

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestWebAuthnRegistrationPUT(t *testing.T) {
	testCases := []struct {
		name           string
		config         *schema.WebAuthn
		have           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       *regexp.Regexp
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldSuccess",
			&schema.DefaultWebAuthnConfiguration,
			`{"description":"test"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: exampleDotCom, Username: testUsername, UserID: "ZytlJlVuWzdgN2BxTyI8Uy9uS2xpJSdsT2ZsJUA5UEBve1c2NENCKDNSWWphaGVCJEhlQ3wpYT9HQGBwIi8zQA=="}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, exampleDotCom, testUsername).
						Return(nil, nil),
				)
			},
			regexp.MustCompile(`^\{"status":"OK","data":\{"publicKey":\{"rp":\{"name":"Authelia","id":"example.com"},"user":\{"name":"john","displayName":"john","id":"ZytlJlVuWzdgN2BxTyI8Uy9uS2xpJSdsT2ZsJUA5UEBve1c2NENCKDNSWWphaGVCJEhlQ3wpYT9HQGBwIi8zQA=="},"challenge":"[a-zA-Z0-9/_-]+={0,2}","pubKeyCredParams":\[\{"type":"public-key","alg":-7},\{"type":"public-key","alg":-35},\{"type":"public-key","alg":-36},\{"type":"public-key","alg":-257},\{"type":"public-key","alg":-258},\{"type":"public-key","alg":-259},\{"type":"public-key","alg":-37},\{"type":"public-key","alg":-38},\{"type":"public-key","alg":-39},\{"type":"public-key","alg":-8}],"timeout":60000,"authenticatorSelection":\{"authenticatorAttachment":"cross-platform","requireResidentKey":false,"residentKey":"discouraged","userVerification":"preferred"},"attestation":"indirect","extensions":\{"credProps":true}}}}$`),
			fasthttp.StatusOK,
			nil,
		},
		{
			"ShouldErrorOnInvalidOrigin",
			&schema.DefaultWebAuthnConfiguration,
			`{"description":"test"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.Ctx.Request.Header.Set("X-Original-URL", "haoiu123!J@#*()!@HJ$!@*(OJOIFQJNW()D@JE()_@JK")
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Unable to register your security key."}$`),
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn registration challenge for user 'john': error occurred provisioning the configuration", "failed to parse X-Original-URL header: parse \"haoiu123!J@#*()!@HJ$!@*(OJOIFQJNW()D@JE()_@JK\": invalid URI for request")
			},
		},
		{
			"ShouldErrorOnAnonymous",
			&schema.DefaultWebAuthnConfiguration,
			`{"description":"test"}`,
			nil,
			regexp.MustCompile(`^\{"status":"KO","message":"Unable to register your security key."}$`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn registration challenge", "user is anonymous")
			},
		},
		{
			"ShouldErrorOnBadSessionDomain",
			&schema.DefaultWebAuthnConfiguration,
			`{"description":"test"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Unable to register your security key."}$`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn registration challenge: error occurred retrieving the user session data", "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
			},
		},
		{
			"ShouldErrorOnBadBody",
			&schema.DefaultWebAuthnConfiguration,
			`{"description":test"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Unable to register your security key."}$`),
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn registration challenge for user 'john': error parsing the request body", "invalid character 'e' in literal true (expecting 'r')")
			},
		},
		{
			"ShouldErrorOnBadDescLength",
			&schema.DefaultWebAuthnConfiguration,
			`{"description":"testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Unable to register your security key."}$`),
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn registration challenge for user 'john': error occurred validating the description chosen by the user", "description has a length of 72 but must be between 1 and 64")
			},
		},
		{
			"ShouldHandleStorageErrorCredentials",
			&schema.DefaultWebAuthnConfiguration,
			`{"description":"test"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: exampleDotCom, Username: testUsername, UserID: "ZytlJlVuWzdgN2BxTyI8Uy9uS2xpJSdsT2ZsJUA5UEBve1c2NENCKDNSWWphaGVCJEhlQ3wpYT9HQGBwIi8zQA=="}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, exampleDotCom, testUsername).
						Return(nil, fmt.Errorf("database closed")),
				)
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Unable to register your security key."}$`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn registration challenge for user 'john': error occurred retrieving the WebAuthn user configuration from the storage backend", "database closed")
			},
		},
		{
			"ShouldHandleStorageErrorUser",
			&schema.DefaultWebAuthnConfiguration,
			`{"description":"test"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(nil, fmt.Errorf("no user x")),
				)
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Unable to register your security key."}$`),
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn registration challenge for user 'john': error occurred retrieving the WebAuthn user configuration from the storage backend", "no user x")
			},
		},
		{
			"ShouldHandleConflict",
			&schema.DefaultWebAuthnConfiguration,
			`{"description":"test"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: exampleDotCom, Username: testUsername, UserID: "ZytlJlVuWzdgN2BxTyI8Uy9uS2xpJSdsT2ZsJUA5UEBve1c2NENCKDNSWWphaGVCJEhlQ3wpYT9HQGBwIi8zQA=="}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, exampleDotCom, testUsername).
						Return([]model.WebAuthnCredential{{Description: "test"}}, nil),
				)
			},
			regexp.MustCompile(`^\{"status":"KO","message":"Another one of your security keys is already registered with that display name."}$`),
			fasthttp.StatusConflict,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating a WebAuthn registration challenge for user 'john': error occurred validating the description chosen by the user", "the description 'test' already exists for the user")
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

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			WebAuthnRegistrationPUT(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Regexp(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestWebAuthnRegistrationDELETE(t *testing.T) {
	testCases := []struct {
		name           string
		config         *schema.WebAuthn
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
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
				us.AuthenticationLevel = authentication.OneFactor
				us.WebAuthn = &session.WebAuthn{}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)
			},
		},
		{
			"ShouldErrorAnonymous",
			&schema.DefaultWebAuthnConfiguration,
			nil,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred deleting a WebAuthn registration challenge", "user is anonymous")
			},
		},
		{
			"ShouldErrorBadCookieDomain",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred deleting a WebAuthn registration challenge: error occurred retrieving the user session data", "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
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

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			WebAuthnRegistrationDELETE(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestWebAuthnRegistrationPOST(t *testing.T) {
	decode := func(in string) []byte {
		value, err := base64.StdEncoding.DecodeString(in)
		if err != nil {
			t.Fatal("Failed to decode base64 string:", err)
		}

		return value
	}

	const (
		dataPOSTFmt        = `{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"attestationObject":"o2NmbXRmcGFja2VkZ2F0dFN0bXSjY2FsZyZjc2lnWEcwRQIhAI505i2XKRL3xsFcSNRz6crTg7_AIpJIsVOjuv8MKW6jAiBJCrqGIc9kKSgS1x54lq53SWUpVNXmlakZfp5NIXrJcmN4NWOBWQHeMIIB2jCCAX2gAwIBAgIBATANBgkqhkiG9w0BAQsFADBgMQswCQYDVQQGEwJVUzERMA8GA1UECgwIQ2hyb21pdW0xIjAgBgNVBAsMGUF1dGhlbnRpY2F0b3IgQXR0ZXN0YXRpb24xGjAYBgNVBAMMEUJhdGNoIENlcnRpZmljYXRlMB4XDTE3MDcxNDAyNDAwMFoXDTQzMTEyMTA4MDAzOVowYDELMAkGA1UEBhMCVVMxETAPBgNVBAoMCENocm9taXVtMSIwIAYDVQQLDBlBdXRoZW50aWNhdG9yIEF0dGVzdGF0aW9uMRowGAYDVQQDDBFCYXRjaCBDZXJ0aWZpY2F0ZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABI1hfmXJUI5kvMVnOsgqZ5naPBRGaCwljEY__99Y39L6Pmw3i1PXlcSk3_tBme3Xhi8jq68CA7S4kRugVpmU4QGjJTAjMAwGA1UdEwEB_wQCMAAwEwYLKwYBBAGC5RwCAQEEBAMCBSAwDQYJKoZIhvcNAQELBQADSAAwRQIgI4PXvgxbCt2L3tk_p22e3QmDCw0ZOPJ6dIJcp2LoTRACIQDqhWGzBtSCdnTiGq2CjhApHJxER1tBy9vRbRaioTz-ZGhhdXRoRGF0YVikDGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM","clientDataJSON":"%s","transports":["usb"],"publicKeyAlgorithm":-7,"publicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEuBHGqg5RF-1k2nt6MEw4lBdvrfVD77Dlx83VLu9b3rerjXRKRLRRQlF_USnsrSYHCSKAp2UzGVTvZAp8daWysw","authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM"},"type":"public-key","clientExtensionResults":{"credProps":{"rk":false}},"authenticatorAttachment":"cross-platform"}`
		dataClientDataJSON = `{"type":"webauthn.create","challenge":"aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U","origin":"%s","crossOrigin":false}`
	)

	var (
		dataPOSTGood        = fmt.Sprintf(dataPOSTFmt, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(dataClientDataJSON, "https://login.example.com:8080"))))
		dataPOSTBadRPIDHash = fmt.Sprintf(dataPOSTFmt, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(dataClientDataJSON, "http://example.com"))))
	)

	testCases := []struct {
		name             string
		config           *schema.WebAuthn
		setup            func(t *testing.T, mock *mocks.MockAutheliaCtx)
		have             string
		expected         string
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
				us.AuthenticationLevel = authentication.OneFactor
				us.WebAuthn = &session.WebAuthn{
					Description: "test",
					SessionData: &webauthn.SessionData{
						Challenge:        "aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U",
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
						Return(nil, nil),
					mock.StorageMock.
						EXPECT().
						SaveWebAuthnCredential(mock.Ctx, gomock.Any()).
						Return(nil),
					mock.UserProviderMock.
						EXPECT().
						GetDetails(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{"john@example.com"}}, nil),
					mock.NotifierMock.
						EXPECT().
						Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: "john@example.com"}, "Second Factor Method Added", gomock.Any(), gomock.Any()).
						Return(nil),
				)
			},
			dataPOSTGood,
			`{"status":"OK"}`,
			fasthttp.StatusCreated,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)
			},
		},
		{
			"ShouldHandleErrorNotificationSend",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.WebAuthn = &session.WebAuthn{
					Description: "test",
					SessionData: &webauthn.SessionData{
						Challenge:        "aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U",
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
						Return(nil, nil),
					mock.StorageMock.
						EXPECT().
						SaveWebAuthnCredential(mock.Ctx, gomock.Any()).
						Return(nil),
					mock.UserProviderMock.
						EXPECT().
						GetDetails(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{"john@example.com"}}, nil),
					mock.NotifierMock.
						EXPECT().
						Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: "john@example.com"}, "Second Factor Method Added", gomock.Any(), gomock.Any()).
						Return(fmt.Errorf("invalid server")),
				)
			},
			dataPOSTGood,
			`{"status":"OK"}`,
			fasthttp.StatusCreated,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred sending notification to user 'john' while attempting to alert them of an important event", "invalid server")
			},
		},
		{
			"ShouldHandleErrorNotificationUserDetails",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.WebAuthn = &session.WebAuthn{
					Description: "test",
					SessionData: &webauthn.SessionData{
						Challenge:        "aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U",
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
						Return(nil, nil),
					mock.StorageMock.
						EXPECT().
						SaveWebAuthnCredential(mock.Ctx, gomock.Any()).
						Return(nil),
					mock.UserProviderMock.
						EXPECT().
						GetDetails(testUsername).
						Return(nil, fmt.Errorf("failed conn")),
				)
			},
			`{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"attestationObject":"o2NmbXRmcGFja2VkZ2F0dFN0bXSjY2FsZyZjc2lnWEcwRQIhAI505i2XKRL3xsFcSNRz6crTg7_AIpJIsVOjuv8MKW6jAiBJCrqGIc9kKSgS1x54lq53SWUpVNXmlakZfp5NIXrJcmN4NWOBWQHeMIIB2jCCAX2gAwIBAgIBATANBgkqhkiG9w0BAQsFADBgMQswCQYDVQQGEwJVUzERMA8GA1UECgwIQ2hyb21pdW0xIjAgBgNVBAsMGUF1dGhlbnRpY2F0b3IgQXR0ZXN0YXRpb24xGjAYBgNVBAMMEUJhdGNoIENlcnRpZmljYXRlMB4XDTE3MDcxNDAyNDAwMFoXDTQzMTEyMTA4MDAzOVowYDELMAkGA1UEBhMCVVMxETAPBgNVBAoMCENocm9taXVtMSIwIAYDVQQLDBlBdXRoZW50aWNhdG9yIEF0dGVzdGF0aW9uMRowGAYDVQQDDBFCYXRjaCBDZXJ0aWZpY2F0ZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABI1hfmXJUI5kvMVnOsgqZ5naPBRGaCwljEY__99Y39L6Pmw3i1PXlcSk3_tBme3Xhi8jq68CA7S4kRugVpmU4QGjJTAjMAwGA1UdEwEB_wQCMAAwEwYLKwYBBAGC5RwCAQEEBAMCBSAwDQYJKoZIhvcNAQELBQADSAAwRQIgI4PXvgxbCt2L3tk_p22e3QmDCw0ZOPJ6dIJcp2LoTRACIQDqhWGzBtSCdnTiGq2CjhApHJxER1tBy9vRbRaioTz-ZGhhdXRoRGF0YVikDGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiYXFfQVhkdnNETXNLV18xYVkzMVhRaFUxN1pNZzFpMFRLMDEzRHd1a0IyVSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0","transports":["usb"],"publicKeyAlgorithm":-7,"publicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEuBHGqg5RF-1k2nt6MEw4lBdvrfVD77Dlx83VLu9b3rerjXRKRLRRQlF_USnsrSYHCSKAp2UzGVTvZAp8daWysw","authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM"},"type":"public-key","clientExtensionResults":{"credProps":{"rk":false}},"authenticatorAttachment":"cross-platform"}`,
			`{"status":"OK"}`,
			fasthttp.StatusCreated,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred looking up user details for user 'john' while attempting to alert them of an important event", "failed conn")
			},
		},
		{
			"ShouldHandleSaveCredentialError",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.WebAuthn = &session.WebAuthn{
					Description: "test",
					SessionData: &webauthn.SessionData{
						Challenge:        "aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U",
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
						Return(nil, nil),
					mock.StorageMock.
						EXPECT().
						SaveWebAuthnCredential(mock.Ctx, gomock.Any()).
						Return(fmt.Errorf("disk full")),
				)
			},
			`{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"attestationObject":"o2NmbXRmcGFja2VkZ2F0dFN0bXSjY2FsZyZjc2lnWEcwRQIhAI505i2XKRL3xsFcSNRz6crTg7_AIpJIsVOjuv8MKW6jAiBJCrqGIc9kKSgS1x54lq53SWUpVNXmlakZfp5NIXrJcmN4NWOBWQHeMIIB2jCCAX2gAwIBAgIBATANBgkqhkiG9w0BAQsFADBgMQswCQYDVQQGEwJVUzERMA8GA1UECgwIQ2hyb21pdW0xIjAgBgNVBAsMGUF1dGhlbnRpY2F0b3IgQXR0ZXN0YXRpb24xGjAYBgNVBAMMEUJhdGNoIENlcnRpZmljYXRlMB4XDTE3MDcxNDAyNDAwMFoXDTQzMTEyMTA4MDAzOVowYDELMAkGA1UEBhMCVVMxETAPBgNVBAoMCENocm9taXVtMSIwIAYDVQQLDBlBdXRoZW50aWNhdG9yIEF0dGVzdGF0aW9uMRowGAYDVQQDDBFCYXRjaCBDZXJ0aWZpY2F0ZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABI1hfmXJUI5kvMVnOsgqZ5naPBRGaCwljEY__99Y39L6Pmw3i1PXlcSk3_tBme3Xhi8jq68CA7S4kRugVpmU4QGjJTAjMAwGA1UdEwEB_wQCMAAwEwYLKwYBBAGC5RwCAQEEBAMCBSAwDQYJKoZIhvcNAQELBQADSAAwRQIgI4PXvgxbCt2L3tk_p22e3QmDCw0ZOPJ6dIJcp2LoTRACIQDqhWGzBtSCdnTiGq2CjhApHJxER1tBy9vRbRaioTz-ZGhhdXRoRGF0YVikDGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiYXFfQVhkdnNETXNLV18xYVkzMVhRaFUxN1pNZzFpMFRLMDEzRHd1a0IyVSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0","transports":["usb"],"publicKeyAlgorithm":-7,"publicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEuBHGqg5RF-1k2nt6MEw4lBdvrfVD77Dlx83VLu9b3rerjXRKRLRRQlF_USnsrSYHCSKAp2UzGVTvZAp8daWysw","authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM"},"type":"public-key","clientExtensionResults":{"credProps":{"rk":false}},"authenticatorAttachment":"cross-platform"}`,
			`{"status":"KO","message":"Unable to register your security key."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge for user 'john': error occurred saving the credential to the storage backend", "disk full")
			},
		},
		{
			"ShouldHandleLoadCredentialsError",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.WebAuthn = &session.WebAuthn{
					Description: "test",
					SessionData: &webauthn.SessionData{
						Challenge:        "aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U",
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
						Return(nil, fmt.Errorf("no dice")),
				)
			},
			`{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"attestationObject":"o2NmbXRmcGFja2VkZ2F0dFN0bXSjY2FsZyZjc2lnWEcwRQIhAI505i2XKRL3xsFcSNRz6crTg7_AIpJIsVOjuv8MKW6jAiBJCrqGIc9kKSgS1x54lq53SWUpVNXmlakZfp5NIXrJcmN4NWOBWQHeMIIB2jCCAX2gAwIBAgIBATANBgkqhkiG9w0BAQsFADBgMQswCQYDVQQGEwJVUzERMA8GA1UECgwIQ2hyb21pdW0xIjAgBgNVBAsMGUF1dGhlbnRpY2F0b3IgQXR0ZXN0YXRpb24xGjAYBgNVBAMMEUJhdGNoIENlcnRpZmljYXRlMB4XDTE3MDcxNDAyNDAwMFoXDTQzMTEyMTA4MDAzOVowYDELMAkGA1UEBhMCVVMxETAPBgNVBAoMCENocm9taXVtMSIwIAYDVQQLDBlBdXRoZW50aWNhdG9yIEF0dGVzdGF0aW9uMRowGAYDVQQDDBFCYXRjaCBDZXJ0aWZpY2F0ZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABI1hfmXJUI5kvMVnOsgqZ5naPBRGaCwljEY__99Y39L6Pmw3i1PXlcSk3_tBme3Xhi8jq68CA7S4kRugVpmU4QGjJTAjMAwGA1UdEwEB_wQCMAAwEwYLKwYBBAGC5RwCAQEEBAMCBSAwDQYJKoZIhvcNAQELBQADSAAwRQIgI4PXvgxbCt2L3tk_p22e3QmDCw0ZOPJ6dIJcp2LoTRACIQDqhWGzBtSCdnTiGq2CjhApHJxER1tBy9vRbRaioTz-ZGhhdXRoRGF0YVikDGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiYXFfQVhkdnNETXNLV18xYVkzMVhRaFUxN1pNZzFpMFRLMDEzRHd1a0IyVSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0","transports":["usb"],"publicKeyAlgorithm":-7,"publicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEuBHGqg5RF-1k2nt6MEw4lBdvrfVD77Dlx83VLu9b3rerjXRKRLRRQlF_USnsrSYHCSKAp2UzGVTvZAp8daWysw","authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM"},"type":"public-key","clientExtensionResults":{"credProps":{"rk":false}},"authenticatorAttachment":"cross-platform"}`,
			`{"status":"KO","message":"Unable to register your security key."}`,
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge for user 'john': error occurred retrieving the WebAuthn user configuration from the storage backend", "no dice")
			},
		},
		{
			"ShouldHandleLoadUserError",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.WebAuthn = &session.WebAuthn{
					Description: "test",
					SessionData: &webauthn.SessionData{
						Challenge:        "aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U",
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
						Return(nil, fmt.Errorf("not enough cowbell")),
				)
			},
			`{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"attestationObject":"o2NmbXRmcGFja2VkZ2F0dFN0bXSjY2FsZyZjc2lnWEcwRQIhAI505i2XKRL3xsFcSNRz6crTg7_AIpJIsVOjuv8MKW6jAiBJCrqGIc9kKSgS1x54lq53SWUpVNXmlakZfp5NIXrJcmN4NWOBWQHeMIIB2jCCAX2gAwIBAgIBATANBgkqhkiG9w0BAQsFADBgMQswCQYDVQQGEwJVUzERMA8GA1UECgwIQ2hyb21pdW0xIjAgBgNVBAsMGUF1dGhlbnRpY2F0b3IgQXR0ZXN0YXRpb24xGjAYBgNVBAMMEUJhdGNoIENlcnRpZmljYXRlMB4XDTE3MDcxNDAyNDAwMFoXDTQzMTEyMTA4MDAzOVowYDELMAkGA1UEBhMCVVMxETAPBgNVBAoMCENocm9taXVtMSIwIAYDVQQLDBlBdXRoZW50aWNhdG9yIEF0dGVzdGF0aW9uMRowGAYDVQQDDBFCYXRjaCBDZXJ0aWZpY2F0ZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABI1hfmXJUI5kvMVnOsgqZ5naPBRGaCwljEY__99Y39L6Pmw3i1PXlcSk3_tBme3Xhi8jq68CA7S4kRugVpmU4QGjJTAjMAwGA1UdEwEB_wQCMAAwEwYLKwYBBAGC5RwCAQEEBAMCBSAwDQYJKoZIhvcNAQELBQADSAAwRQIgI4PXvgxbCt2L3tk_p22e3QmDCw0ZOPJ6dIJcp2LoTRACIQDqhWGzBtSCdnTiGq2CjhApHJxER1tBy9vRbRaioTz-ZGhhdXRoRGF0YVikDGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiYXFfQVhkdnNETXNLV18xYVkzMVhRaFUxN1pNZzFpMFRLMDEzRHd1a0IyVSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0","transports":["usb"],"publicKeyAlgorithm":-7,"publicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEuBHGqg5RF-1k2nt6MEw4lBdvrfVD77Dlx83VLu9b3rerjXRKRLRRQlF_USnsrSYHCSKAp2UzGVTvZAp8daWysw","authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM"},"type":"public-key","clientExtensionResults":{"credProps":{"rk":false}},"authenticatorAttachment":"cross-platform"}`,
			`{"status":"KO","message":"Unable to register your security key."}`,
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge for user 'john': error occurred retrieving the WebAuthn user configuration from the storage backend", "not enough cowbell")
			},
		},
		{
			"ShouldHandleNoSession",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			dataPOSTGood,
			`{"status":"KO","message":"Unable to register your security key."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge for user 'john': error occurred retrieving the user session data", "registration challenge session data is not present")
			},
		},
		{
			"ShouldHandleBadBody",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.WebAuthn = &session.WebAuthn{
					Description: "test",
					SessionData: &webauthn.SessionData{
						Challenge:        "aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))
			},
			strings.Replace(dataPOSTGood, `{"id":"rwOw`, `{"id":wOw`, 1),
			`{"status":"KO","message":"Unable to register your security key."}`,
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge for user 'john': error parsing the request body", "Parse error for Registration (invalid_request): invalid character 'w' looking for beginning of value")
			},
		},
		{
			"ShouldHandleBadOrigin",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.WebAuthn = &session.WebAuthn{
					Description: "test",
					SessionData: &webauthn.SessionData{
						Challenge:        "aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				mock.Ctx.Request.Header.Set("X-Original-URL", "---123=123=1")
			},
			dataPOSTGood,
			`{"status":"KO","message":"Unable to register your security key."}`,
			fasthttp.StatusBadRequest,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge for user 'john': error occurred provisioning the configuration", "failed to parse X-Original-URL header: parse \"---123=123=1\": invalid URI for request")
			},
		},
		{
			"ShouldHandleAnonymous",
			&schema.DefaultWebAuthnConfiguration,
			nil,
			`{"id":wOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"attestationObject":"o2NmbXRmcGFja2VkZ2F0dFN0bXSjY2FsZyZjc2lnWEcwRQIhAI505i2XKRL3xsFcSNRz6crTg7_AIpJIsVOjuv8MKW6jAiBJCrqGIc9kKSgS1x54lq53SWUpVNXmlakZfp5NIXrJcmN4NWOBWQHeMIIB2jCCAX2gAwIBAgIBATANBgkqhkiG9w0BAQsFADBgMQswCQYDVQQGEwJVUzERMA8GA1UECgwIQ2hyb21pdW0xIjAgBgNVBAsMGUF1dGhlbnRpY2F0b3IgQXR0ZXN0YXRpb24xGjAYBgNVBAMMEUJhdGNoIENlcnRpZmljYXRlMB4XDTE3MDcxNDAyNDAwMFoXDTQzMTEyMTA4MDAzOVowYDELMAkGA1UEBhMCVVMxETAPBgNVBAoMCENocm9taXVtMSIwIAYDVQQLDBlBdXRoZW50aWNhdG9yIEF0dGVzdGF0aW9uMRowGAYDVQQDDBFCYXRjaCBDZXJ0aWZpY2F0ZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABI1hfmXJUI5kvMVnOsgqZ5naPBRGaCwljEY__99Y39L6Pmw3i1PXlcSk3_tBme3Xhi8jq68CA7S4kRugVpmU4QGjJTAjMAwGA1UdEwEB_wQCMAAwEwYLKwYBBAGC5RwCAQEEBAMCBSAwDQYJKoZIhvcNAQELBQADSAAwRQIgI4PXvgxbCt2L3tk_p22e3QmDCw0ZOPJ6dIJcp2LoTRACIQDqhWGzBtSCdnTiGq2CjhApHJxER1tBy9vRbRaioTz-ZGhhdXRoRGF0YVikDGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiYXFfQVhkdnNETXNLV18xYVkzMVhRaFUxN1pNZzFpMFRLMDEzRHd1a0IyVSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0","transports":["usb"],"publicKeyAlgorithm":-7,"publicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEuBHGqg5RF-1k2nt6MEw4lBdvrfVD77Dlx83VLu9b3rerjXRKRLRRQlF_USnsrSYHCSKAp2UzGVTvZAp8daWysw","authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM"},"type":"public-key","clientExtensionResults":{"credProps":{"rk":false}},"authenticatorAttachment":"cross-platform"}`,
			`{"status":"KO","message":"Unable to register your security key."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge", "user is anonymous")
			},
		},
		{
			"ShouldHandleBadCookieDomain",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
			},
			`{"id":wOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"attestationObject":"o2NmbXRmcGFja2VkZ2F0dFN0bXSjY2FsZyZjc2lnWEcwRQIhAI505i2XKRL3xsFcSNRz6crTg7_AIpJIsVOjuv8MKW6jAiBJCrqGIc9kKSgS1x54lq53SWUpVNXmlakZfp5NIXrJcmN4NWOBWQHeMIIB2jCCAX2gAwIBAgIBATANBgkqhkiG9w0BAQsFADBgMQswCQYDVQQGEwJVUzERMA8GA1UECgwIQ2hyb21pdW0xIjAgBgNVBAsMGUF1dGhlbnRpY2F0b3IgQXR0ZXN0YXRpb24xGjAYBgNVBAMMEUJhdGNoIENlcnRpZmljYXRlMB4XDTE3MDcxNDAyNDAwMFoXDTQzMTEyMTA4MDAzOVowYDELMAkGA1UEBhMCVVMxETAPBgNVBAoMCENocm9taXVtMSIwIAYDVQQLDBlBdXRoZW50aWNhdG9yIEF0dGVzdGF0aW9uMRowGAYDVQQDDBFCYXRjaCBDZXJ0aWZpY2F0ZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABI1hfmXJUI5kvMVnOsgqZ5naPBRGaCwljEY__99Y39L6Pmw3i1PXlcSk3_tBme3Xhi8jq68CA7S4kRugVpmU4QGjJTAjMAwGA1UdEwEB_wQCMAAwEwYLKwYBBAGC5RwCAQEEBAMCBSAwDQYJKoZIhvcNAQELBQADSAAwRQIgI4PXvgxbCt2L3tk_p22e3QmDCw0ZOPJ6dIJcp2LoTRACIQDqhWGzBtSCdnTiGq2CjhApHJxER1tBy9vRbRaioTz-ZGhhdXRoRGF0YVikDGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiYXFfQVhkdnNETXNLV18xYVkzMVhRaFUxN1pNZzFpMFRLMDEzRHd1a0IyVSIsIm9yaWdpbiI6Imh0dHBzOi8vbG9naW4uZXhhbXBsZS5jb206ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0","transports":["usb"],"publicKeyAlgorithm":-7,"publicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEuBHGqg5RF-1k2nt6MEw4lBdvrfVD77Dlx83VLu9b3rerjXRKRLRRQlF_USnsrSYHCSKAp2UzGVTvZAp8daWysw","authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtxFAAAAAQECAwQFBgcIAQIDBAUGBwgAIK8DsFfFgodYaxNDOpr2qERqRh4naiuiJYZAydsREYT1pQECAyYgASFYILgRxqoOURftZNp7ejBMOJQXb631Q--w5cfN1S7vW963Ilggq410SkS0UUJRf1Ep7K0mBwkigKdlMxlU72QKfHWlsrM"},"type":"public-key","clientExtensionResults":{"credProps":{"rk":false}},"authenticatorAttachment":"cross-platform"}`,
			`{"status":"KO","message":"Unable to register your security key."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge: error occurred retrieving the user session data", "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
			},
		},
		{
			"ShouldHandleBadRPIDHash",
			&schema.DefaultWebAuthnConfiguration,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationLevel = authentication.OneFactor
				us.WebAuthn = &session.WebAuthn{
					Description: "test",
					SessionData: &webauthn.SessionData{
						Challenge:        "aq_AXdvsDMsKW_1aY31XQhU17ZMg1i0TK013DwukB2U",
						UserID:           decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="),
						Expires:          time.Now().Add(time.Minute),
						UserVerification: "preferred",
					},
				}

				require.NoError(t, mock.Ctx.SaveSession(us))

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: exampleDotCom, Username: testUsername, UserID: string(decode("OiRQc3wmemUzdHlkVjhVSk5Pe35YMCRCOklLYzVzIkMpaEglNkF5dnVKRSlTPCJbRDZDP102WXpiYXdNekRiTA=="))}, nil),
					mock.StorageMock.
						EXPECT().
						LoadWebAuthnCredentialsByUsername(mock.Ctx, exampleDotCom, testUsername).
						Return(nil, nil),
				)

				mock.Ctx.Request.Header.Del("X-Original-URL")
			},
			dataPOSTBadRPIDHash,
			`{"status":"KO","message":"Unable to register your security key."}`,
			fasthttp.StatusForbidden,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)

				assert.Nil(t, us.WebAuthn)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a WebAuthn registration challenge for user 'john': error comparing the response to the WebAuthn session data", "Error validating the authenticator response (verification_error): RP Hash mismatch. Expected 0c6ca0839c3a5683557833f618a2556665df2a088964787d53850b4ad4d3bedc and Received a379a6f6eeafb9a55e378c118034e2751e682fab9f2d30ab13d2125586ce1947")
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

			if len(tc.have) != 0 {
				mock.Ctx.Request.SetBodyString(tc.have)
			}

			WebAuthnRegistrationPOST(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.validateResponse != nil {
				tc.validateResponse(t, mock)
			}
		})
	}
}
