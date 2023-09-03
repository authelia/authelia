package handlers

import (
	"errors"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestWebAuthnGetUser(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	userSession := session.UserSession{
		Username:    "john",
		DisplayName: "John Smith",
	}

	ctx.StorageMock.EXPECT().
		LoadWebAuthnUser(ctx.Ctx, "example.com", "john").
		Return(&model.WebAuthnUser{ID: 1, RPID: "example.com", Username: "john", UserID: "john123"}, nil)

	ctx.StorageMock.EXPECT().LoadWebAuthnCredentialsByUsername(ctx.Ctx, "example.com", "john").Return([]model.WebAuthnCredential{
		{
			ID:              1,
			RPID:            "example.com",
			Username:        "john",
			Description:     "Primary",
			KID:             model.NewBase64([]byte("abc123")),
			AttestationType: "fido-u2f",
			PublicKey:       []byte("data"),
			SignCount:       0,
			CloneWarning:    false,
		},
		{
			ID:              2,
			RPID:            "example.com",
			Username:        "john",
			Description:     "Secondary",
			KID:             model.NewBase64([]byte("123abc")),
			AttestationType: "packed",
			Transport:       "usb,nfc",
			PublicKey:       []byte("data"),
			SignCount:       100,
			CloneWarning:    false,
		},
	}, nil)

	user, err := getWebAuthnUserByRPID(ctx.Ctx, userSession.Username, userSession.DisplayName, "example.com")

	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, []byte("john123"), user.WebAuthnID())
	assert.Equal(t, "john", user.WebAuthnName())
	assert.Equal(t, "john", user.Username)

	assert.Equal(t, "", user.WebAuthnIcon())

	assert.Equal(t, "John Smith", user.WebAuthnDisplayName())
	assert.Equal(t, "John Smith", user.DisplayName)

	require.Len(t, user.Credentials, 2)

	assert.Equal(t, 1, user.Credentials[0].ID)
	assert.Equal(t, "example.com", user.Credentials[0].RPID)
	assert.Equal(t, "john", user.Credentials[0].Username)
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
	assert.Equal(t, "example.com", user.Credentials[1].RPID)
	assert.Equal(t, "john", user.Credentials[1].Username)
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
}

func TestWebAuthnGetNewUser(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	// Use the random mock.
	ctx.Ctx.Providers.Random = ctx.RandomMock

	userSession := session.UserSession{
		Username:    "john",
		DisplayName: "John Smith",
	}

	gomock.InOrder(
		ctx.StorageMock.EXPECT().
			LoadWebAuthnUser(ctx.Ctx, "example.com", "john").
			Return(nil, nil),
		ctx.RandomMock.EXPECT().
			StringCustom(64, random.CharSetASCII).
			Return("=ckBRe.%fp{w#K[qw4)AWMZrAP)(z3NUt5n3g?;>'^Rp>+eE4z>[^.<3?&n;LM#w"),
		ctx.StorageMock.EXPECT().
			SaveWebAuthnUser(ctx.Ctx, model.WebAuthnUser{RPID: "example.com", Username: "john", DisplayName: "John Smith", UserID: "=ckBRe.%fp{w#K[qw4)AWMZrAP)(z3NUt5n3g?;>'^Rp>+eE4z>[^.<3?&n;LM#w"}).
			Return(nil),
		ctx.StorageMock.EXPECT().LoadWebAuthnCredentialsByUsername(ctx.Ctx, "example.com", "john").Return([]model.WebAuthnCredential{
			{
				ID:              1,
				RPID:            "example.com",
				Username:        "john",
				Description:     "Primary",
				KID:             model.NewBase64([]byte("abc123")),
				AttestationType: "fido-u2f",
				PublicKey:       []byte("data"),
				SignCount:       0,
				CloneWarning:    false,
			},
			{
				ID:              2,
				RPID:            "example.com",
				Username:        "john",
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

	user, err := getWebAuthnUserByRPID(ctx.Ctx, userSession.Username, userSession.DisplayName, "example.com")

	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, []byte("=ckBRe.%fp{w#K[qw4)AWMZrAP)(z3NUt5n3g?;>'^Rp>+eE4z>[^.<3?&n;LM#w"), user.WebAuthnID())
	assert.Equal(t, "john", user.WebAuthnName())
	assert.Equal(t, "john", user.Username)

	assert.Equal(t, "", user.WebAuthnIcon())

	assert.Equal(t, "John Smith", user.WebAuthnDisplayName())
	assert.Equal(t, "John Smith", user.DisplayName)

	require.Len(t, user.Credentials, 2)

	assert.Equal(t, 1, user.Credentials[0].ID)
	assert.Equal(t, "example.com", user.Credentials[0].RPID)
	assert.Equal(t, "john", user.Credentials[0].Username)
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
	assert.Equal(t, "example.com", user.Credentials[1].RPID)
	assert.Equal(t, "john", user.Credentials[1].Username)
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
}

func TestWebAuthnGetUserWithoutDisplayName(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	userSession := session.UserSession{
		Username: "john",
	}

	ctx.StorageMock.EXPECT().
		LoadWebAuthnUser(ctx.Ctx, "example.com", "john").
		Return(&model.WebAuthnUser{ID: 1, RPID: "example.com", Username: "john", UserID: "john123"}, nil)

	ctx.StorageMock.EXPECT().LoadWebAuthnCredentialsByUsername(ctx.Ctx, "example.com", "john").Return([]model.WebAuthnCredential{
		{
			ID:              1,
			RPID:            "example.com",
			Username:        "john",
			Description:     "Primary",
			KID:             model.NewBase64([]byte("abc123")),
			AttestationType: "fido-u2f",
			PublicKey:       []byte("data"),
			SignCount:       0,
			CloneWarning:    false,
		},
	}, nil)

	user, err := getWebAuthnUserByRPID(ctx.Ctx, userSession.Username, userSession.DisplayName, "example.com")

	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, "john", user.WebAuthnDisplayName())
	assert.Equal(t, "john", user.DisplayName)
}

func TestWebAuthnGetUserWithErr(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	userSession := session.UserSession{
		Username: "john",
	}

	ctx.StorageMock.EXPECT().
		LoadWebAuthnUser(ctx.Ctx, "example.com", "john").
		Return(&model.WebAuthnUser{ID: 1, RPID: "example.com", Username: "john", UserID: "john123"}, nil)

	ctx.StorageMock.EXPECT().
		LoadWebAuthnCredentialsByUsername(ctx.Ctx, "example.com", "john").
		Return(nil, errors.New("not found"))

	user, err := getWebAuthnUserByRPID(ctx.Ctx, userSession.Username, userSession.DisplayName, "example.com")

	assert.EqualError(t, err, "not found")
	assert.Nil(t, user)
}

func TestWebAuthnNewWebAuthnShouldReturnErrWhenHeadersNotAvailable(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)
	ctx.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)

	w, err := newWebAuthn(ctx.Ctx)

	assert.Nil(t, w)
	assert.EqualError(t, err, "missing required X-Forwarded-Host header")
}

func TestWebAuthnNewWebAuthnShouldReturnErrWhenWebAuthnNotConfigured(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	ctx.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "example.com")
	ctx.Ctx.Request.Header.Set("X-Forwarded-URI", "/")
	ctx.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")

	w, err := newWebAuthn(ctx.Ctx)

	assert.Nil(t, w)
	assert.EqualError(t, err, "error occurred validating the configuration: the field 'RPDisplayName' must be configured but it is empty")
}
