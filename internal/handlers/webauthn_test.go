package handlers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestWebAuthnFormatError(t *testing.T) {
	testCases := []struct {
		name     string
		have     error
		expected string
	}{
		{
			"ShouldHandleStandardError",
			fmt.Errorf("abc123"),
			"abc123",
		},
		{
			"ShouldHandleProtocolErrorWithDevInfo",
			&protocol.Error{
				Type:    "a_error",
				Details: "a bad thing",
				DevInfo: "example",
			},
			"a bad thing (a_error): example",
		},
		{
			"ShouldHandleProtocolErrorWithDevInfoWithoutType",
			&protocol.Error{
				Details: "a bad thing",
				DevInfo: "example",
			},
			"a bad thing: example",
		},
		{
			"ShouldHandleProtocolErrorWithoutDevInfo",
			&protocol.Error{
				Details: "a bad thing",
			},
			"a bad thing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := formatWebAuthnError(tc.have)

			assert.EqualError(t, actual, tc.expected)
		})
	}
}

func TestWebAuthnGetUserx(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(t *testing.T, mock *mocks.MockAutheliaCtx)
		have        string
		haveDisplay string
		haveRPID    string
		err         string
		expected    func(t *testing.T, mock *mocks.MockAutheliaCtx, user *model.WebAuthnUser)
	}{
		{
			"ShouldTestNormalUseCase",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: exampleDotCom, Username: testUsername, UserID: "john123"}, nil),

					mock.StorageMock.EXPECT().LoadWebAuthnCredentialsByUsername(mock.Ctx, exampleDotCom, testUsername).Return([]model.WebAuthnCredential{
						{
							ID:              1,
							RPID:            exampleDotCom,
							Username:        testUsername,
							Description:     "Primary",
							KID:             model.NewBase64([]byte("abc123")),
							AttestationType: "fido-u2f",
							PublicKey:       []byte("data"),
							SignCount:       0,
							CloneWarning:    false,
						},
						{
							ID:              2,
							RPID:            exampleDotCom,
							Username:        testUsername,
							Description:     "Secondary",
							KID:             model.NewBase64([]byte("123abc")),
							AttestationType: "packed",
							Transport:       "usb,nfc",
							PublicKey:       []byte("data"),
							SignCount:       100,
							CloneWarning:    false,
						},
					}, nil),
				)
			},
			testUsername,
			testDisplayName,
			exampleDotCom,
			"",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, user *model.WebAuthnUser) {
				require.NotNil(t, user)

				assert.Equal(t, []byte("john123"), user.WebAuthnID())
				assert.Equal(t, testUsername, user.WebAuthnName())
				assert.Equal(t, testUsername, user.Username)

				assert.Equal(t, "", user.WebAuthnIcon())

				assert.Equal(t, testDisplayName, user.WebAuthnDisplayName())
				assert.Equal(t, testDisplayName, user.DisplayName)

				require.Len(t, user.Credentials, 2)

				assert.Equal(t, 1, user.Credentials[0].ID)
				assert.Equal(t, exampleDotCom, user.Credentials[0].RPID)
				assert.Equal(t, testUsername, user.Credentials[0].Username)
				assert.Equal(t, "Primary", user.Credentials[0].Description)
				assert.Equal(t, "", user.Credentials[0].Transport)
				assert.Equal(t, "fido-u2f", user.Credentials[0].AttestationType)
				assert.Equal(t, []byte("data"), user.Credentials[0].PublicKey)
				assert.Equal(t, uint32(0), user.Credentials[0].SignCount)
				assert.False(t, user.Credentials[0].CloneWarning)

				descriptors := user.WebAuthnCredentialDescriptors()
				assert.Equal(t, "fido-u2f", descriptors[0].AttestationType)
				assert.Equal(t, "abc123", string(descriptors[0].CredentialID))
				assert.Equal(t, protocol.PublicKeyCredentialType, descriptors[0].Type)

				assert.Len(t, descriptors[0].Transport, 0)

				assert.Equal(t, 2, user.Credentials[1].ID)
				assert.Equal(t, exampleDotCom, user.Credentials[1].RPID)
				assert.Equal(t, testUsername, user.Credentials[1].Username)
				assert.Equal(t, "Secondary", user.Credentials[1].Description)
				assert.Equal(t, "usb,nfc", user.Credentials[1].Transport)
				assert.Equal(t, "packed", user.Credentials[1].AttestationType)
				assert.Equal(t, []byte("data"), user.Credentials[1].PublicKey)
				assert.Equal(t, uint32(100), user.Credentials[1].SignCount)
				assert.False(t, user.Credentials[1].CloneWarning)

				assert.Equal(t, "packed", descriptors[1].AttestationType)
				assert.Equal(t, "123abc", string(descriptors[1].CredentialID))
				assert.Equal(t, protocol.PublicKeyCredentialType, descriptors[1].Type)

				assert.Len(t, descriptors[1].Transport, 2)
				assert.Equal(t, protocol.AuthenticatorTransport("usb"), descriptors[1].Transport[0])
				assert.Equal(t, protocol.AuthenticatorTransport("nfc"), descriptors[1].Transport[1])
			},
		},
		{
			"ShouldGenerateNewUser",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Providers.Random = mock.RandomMock

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(nil, nil),
					mock.RandomMock.EXPECT().
						StringCustom(64, random.CharSetASCII).
						Return("=ckBRe.%fp{w#K[qw4)AWMZrAP)(z3NUt5n3g?;>'^Rp>+eE4z>[^.<3?&n;LM#w"),
					mock.StorageMock.EXPECT().
						SaveWebAuthnUser(mock.Ctx, model.WebAuthnUser{RPID: exampleDotCom, Username: testUsername, DisplayName: testDisplayName, UserID: "=ckBRe.%fp{w#K[qw4)AWMZrAP)(z3NUt5n3g?;>'^Rp>+eE4z>[^.<3?&n;LM#w"}).
						Return(nil),
					mock.StorageMock.EXPECT().LoadWebAuthnCredentialsByUsername(mock.Ctx, exampleDotCom, testUsername).Return([]model.WebAuthnCredential{
						{
							ID:              1,
							RPID:            exampleDotCom,
							Username:        testUsername,
							Description:     "Primary",
							KID:             model.NewBase64([]byte("abc123")),
							AttestationType: "fido-u2f",
							PublicKey:       []byte("data"),
							SignCount:       0,
							CloneWarning:    false,
						},
						{
							ID:              2,
							RPID:            exampleDotCom,
							Username:        testUsername,
							Description:     "Secondary",
							KID:             model.NewBase64([]byte("123abc")),
							AttestationType: "packed",
							Transport:       "usb,nfc",
							PublicKey:       []byte("data"),
							SignCount:       100,
							CloneWarning:    false,
						},
					}, nil),
				)
			},
			testUsername,
			testDisplayName,
			exampleDotCom,
			"",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, user *model.WebAuthnUser) {
				require.NotNil(t, user)

				assert.Equal(t, []byte("=ckBRe.%fp{w#K[qw4)AWMZrAP)(z3NUt5n3g?;>'^Rp>+eE4z>[^.<3?&n;LM#w"), user.WebAuthnID())
				assert.Equal(t, testUsername, user.WebAuthnName())
				assert.Equal(t, testUsername, user.Username)

				assert.Equal(t, "", user.WebAuthnIcon())

				assert.Equal(t, testDisplayName, user.WebAuthnDisplayName())
				assert.Equal(t, testDisplayName, user.DisplayName)

				require.Len(t, user.Credentials, 2)

				assert.Equal(t, 1, user.Credentials[0].ID)
				assert.Equal(t, exampleDotCom, user.Credentials[0].RPID)
				assert.Equal(t, testUsername, user.Credentials[0].Username)
				assert.Equal(t, "Primary", user.Credentials[0].Description)
				assert.Equal(t, "", user.Credentials[0].Transport)
				assert.Equal(t, "fido-u2f", user.Credentials[0].AttestationType)
				assert.Equal(t, []byte("data"), user.Credentials[0].PublicKey)
				assert.Equal(t, uint32(0), user.Credentials[0].SignCount)
				assert.False(t, user.Credentials[0].CloneWarning)

				descriptors := user.WebAuthnCredentialDescriptors()
				assert.Equal(t, "fido-u2f", descriptors[0].AttestationType)
				assert.Equal(t, "abc123", string(descriptors[0].CredentialID))
				assert.Equal(t, protocol.PublicKeyCredentialType, descriptors[0].Type)

				assert.Len(t, descriptors[0].Transport, 0)

				assert.Equal(t, 2, user.Credentials[1].ID)
				assert.Equal(t, exampleDotCom, user.Credentials[1].RPID)
				assert.Equal(t, testUsername, user.Credentials[1].Username)
				assert.Equal(t, "Secondary", user.Credentials[1].Description)
				assert.Equal(t, "usb,nfc", user.Credentials[1].Transport)
				assert.Equal(t, "packed", user.Credentials[1].AttestationType)
				assert.Equal(t, []byte("data"), user.Credentials[1].PublicKey)
				assert.Equal(t, uint32(100), user.Credentials[1].SignCount)
				assert.False(t, user.Credentials[1].CloneWarning)

				assert.Equal(t, "packed", descriptors[1].AttestationType)
				assert.Equal(t, "123abc", string(descriptors[1].CredentialID))
				assert.Equal(t, protocol.PublicKeyCredentialType, descriptors[1].Type)

				assert.Len(t, descriptors[1].Transport, 2)
				assert.Equal(t, protocol.AuthenticatorTransport("usb"), descriptors[1].Transport[0])
				assert.Equal(t, protocol.AuthenticatorTransport("nfc"), descriptors[1].Transport[1])
			},
		},
		{
			"ShouldGenerateNewUser",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Providers.Random = mock.RandomMock

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(nil, nil),
					mock.RandomMock.EXPECT().
						StringCustom(64, random.CharSetASCII).
						Return("=ckBRe.%fp{w#K[qw4)AWMZrAP)(z3NUt5n3g?;>'^Rp>+eE4z>[^.<3?&n;LM#w"),
					mock.StorageMock.EXPECT().
						SaveWebAuthnUser(mock.Ctx, model.WebAuthnUser{RPID: exampleDotCom, Username: testUsername, DisplayName: testDisplayName, UserID: "=ckBRe.%fp{w#K[qw4)AWMZrAP)(z3NUt5n3g?;>'^Rp>+eE4z>[^.<3?&n;LM#w"}).
						Return(fmt.Errorf("broken pipe")),
				)
			},
			testUsername,
			testDisplayName,
			exampleDotCom,
			"broken pipe",
			nil,
		},
		{
			"ShouldHandleEmptyDisplayName",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: exampleDotCom, Username: testUsername, UserID: "john123"}, nil),

					mock.StorageMock.EXPECT().LoadWebAuthnCredentialsByUsername(mock.Ctx, exampleDotCom, testUsername).Return([]model.WebAuthnCredential{
						{
							ID:              1,
							RPID:            exampleDotCom,
							Username:        testUsername,
							Description:     "Primary",
							KID:             model.NewBase64([]byte("abc123")),
							AttestationType: "fido-u2f",
							PublicKey:       []byte("data"),
							SignCount:       0,
							CloneWarning:    false,
						},
					}, nil),
				)
			},
			testUsername,
			testDisplayName,
			exampleDotCom,
			"",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, user *model.WebAuthnUser) {
				require.NotNil(t, user)

				assert.Equal(t, testUsername, user.WebAuthnDisplayName())
				assert.Equal(t, testUsername, user.DisplayName)
			},
		},
		{
			"ShouldHandleEmptyDisplayName",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: exampleDotCom, Username: testUsername, UserID: "john123"}, nil),

					mock.StorageMock.EXPECT().LoadWebAuthnCredentialsByUsername(mock.Ctx, exampleDotCom, testUsername).Return([]model.WebAuthnCredential{
						{
							ID:              1,
							RPID:            exampleDotCom,
							Username:        testUsername,
							Description:     "Primary",
							KID:             model.NewBase64([]byte("abc123")),
							AttestationType: "fido-u2f",
							PublicKey:       []byte("data"),
							SignCount:       0,
							CloneWarning:    false,
						},
					}, nil),
				)
			},
			testUsername,
			"",
			exampleDotCom,
			"",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, user *model.WebAuthnUser) {
				require.NotNil(t, user)

				assert.Equal(t, testUsername, user.WebAuthnDisplayName())
				assert.Equal(t, testUsername, user.DisplayName)
			},
		},
		{
			"ShouldHandleLoadWebAuthnUserErr",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(nil, fmt.Errorf("invalid host")),
				)
			},
			testUsername,
			testDisplayName,
			exampleDotCom,
			"invalid host",
			nil,
		},
		{
			"ShouldHandleLoadWebAuthnCredentialErr",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadWebAuthnUser(mock.Ctx, exampleDotCom, testUsername).
						Return(&model.WebAuthnUser{ID: 1, RPID: exampleDotCom, Username: testUsername, UserID: "john123"}, nil),

					mock.StorageMock.EXPECT().LoadWebAuthnCredentialsByUsername(mock.Ctx, exampleDotCom, testUsername).Return(nil, fmt.Errorf("invalid key")),
				)
			},
			testUsername,
			testDisplayName,
			exampleDotCom,
			"invalid key",
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			tc.setup(t, mock)

			user, err := handleGetWebAuthnUserByRPID(mock.Ctx, tc.have, tc.haveDisplay, tc.haveRPID)
			if tc.err == "" {
				assert.NoError(t, err)

				tc.expected(t, mock, user)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestWebAuthnGetUserWithErr(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	userSession := session.UserSession{
		Username: testUsername,
	}

	ctx.StorageMock.EXPECT().
		LoadWebAuthnUser(ctx.Ctx, exampleDotCom, testUsername).
		Return(&model.WebAuthnUser{ID: 1, RPID: exampleDotCom, Username: testUsername, UserID: "john123"}, nil)

	ctx.StorageMock.EXPECT().
		LoadWebAuthnCredentialsByUsername(ctx.Ctx, exampleDotCom, testUsername).
		Return(nil, errors.New("not found"))

	user, err := handleGetWebAuthnUserByRPID(ctx.Ctx, userSession.Username, userSession.DisplayName, exampleDotCom)

	assert.EqualError(t, err, "not found")
	assert.Nil(t, user)
}

func TestWebAuthnNewWebAuthnShouldReturnErrWhenHeadersNotAvailable(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)
	ctx.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)

	w, err := handleNewWebAuthn(ctx.Ctx)

	assert.Nil(t, w)
	assert.EqualError(t, err, "missing required X-Forwarded-Host header")
}

func TestWebAuthnNewWebAuthnShouldReturnErrWhenWebAuthnNotConfigured(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	ctx.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, exampleDotCom)
	ctx.Ctx.Request.Header.Set("X-Forwarded-URI", "/")
	ctx.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")

	w, err := handleNewWebAuthn(ctx.Ctx)

	assert.Nil(t, w)
	assert.EqualError(t, err, "error occurred validating the configuration: the field 'RPDisplayName' must be configured but it is empty")
}

func TestWebauthnCredentialCreationIsDiscoverable(t *testing.T) {
	testCases := []struct {
		name      string
		have      *protocol.ParsedCredentialCreationData
		expected  bool
		expectedf func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldBeDiscoverable",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{
						WebAuthnExtensionCredProps: map[string]any{
							WebAuthnExtensionCredPropsResidentKey: true,
						},
					},
				},
			},
			true,
			nil,
		},
		{
			"ShouldNotBeDiscoverableExplicit",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{
						WebAuthnExtensionCredProps: map[string]any{
							WebAuthnExtensionCredPropsResidentKey: false,
						},
					},
				},
			},
			false,
			nil,
		},
		{
			"ShouldNotBeDiscoverableImplicitType",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{
						WebAuthnExtensionCredProps: map[string]any{
							WebAuthnExtensionCredPropsResidentKey: 1,
						},
					},
				},
			},
			false,
			nil,
		},
		{
			"ShouldNotBeDiscoverableImplicitNoRK",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{
						WebAuthnExtensionCredProps: map[string]any{},
					},
				},
			},
			false,
			nil,
		},
		{
			"ShouldNotBeDiscoverableImplicitNoCredPropsType",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{
						WebAuthnExtensionCredProps: 1,
					},
				},
			},
			false,
			nil,
		},
		{
			"ShouldNotBeDiscoverableImplicitNoCredProps",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{},
				},
			},
			false,
			nil,
		},
		{
			"ShouldNotBeDiscoverableImplicitNoCredPropsNil",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{},
			},
			false,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			assert.Equal(t, tc.expected, handleWebAuthnCredentialCreationIsDiscoverable(mock.Ctx, tc.have))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}
