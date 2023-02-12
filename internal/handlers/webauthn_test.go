package handlers

import (
	"errors"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestWebauthnGetUser(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	userSession := session.UserSession{
		Username:    "john",
		DisplayName: "John Smith",
	}

	ctx.StorageMock.EXPECT().LoadWebauthnDevicesByUsername(ctx.Ctx, "example.com", "john").Return([]model.WebauthnDevice{
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

	user, err := getWebauthnUserByRPID(ctx.Ctx, userSession, "example.com")

	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, []byte("john"), user.WebAuthnID())
	assert.Equal(t, "john", user.WebAuthnName())
	assert.Equal(t, "john", user.Username)

	assert.Equal(t, "", user.WebAuthnIcon())

	assert.Equal(t, "John Smith", user.WebAuthnDisplayName())
	assert.Equal(t, "John Smith", user.DisplayName)

	require.Len(t, user.Devices, 2)

	assert.Equal(t, 1, user.Devices[0].ID)
	assert.Equal(t, "example.com", user.Devices[0].RPID)
	assert.Equal(t, "john", user.Devices[0].Username)
	assert.Equal(t, "Primary", user.Devices[0].Description)
	assert.Equal(t, "", user.Devices[0].Transport)
	assert.Equal(t, "fido-u2f", user.Devices[0].AttestationType)
	assert.Equal(t, []byte("data"), user.Devices[0].PublicKey)
	assert.Equal(t, uint32(0), user.Devices[0].SignCount)
	assert.False(t, user.Devices[0].CloneWarning)

	descriptors := user.WebAuthnCredentialDescriptors()
	assert.Equal(t, "fido-u2f", descriptors[0].AttestationType)
	assert.Equal(t, "abc123", string(descriptors[0].CredentialID))
	assert.Equal(t, protocol.PublicKeyCredentialType, descriptors[0].Type)

	assert.Len(t, descriptors[0].Transport, 0)

	assert.Equal(t, 2, user.Devices[1].ID)
	assert.Equal(t, "example.com", user.Devices[1].RPID)
	assert.Equal(t, "john", user.Devices[1].Username)
	assert.Equal(t, "Secondary", user.Devices[1].Description)
	assert.Equal(t, "usb,nfc", user.Devices[1].Transport)
	assert.Equal(t, "packed", user.Devices[1].AttestationType)
	assert.Equal(t, []byte("data"), user.Devices[1].PublicKey)
	assert.Equal(t, uint32(100), user.Devices[1].SignCount)
	assert.False(t, user.Devices[1].CloneWarning)

	assert.Equal(t, "packed", descriptors[1].AttestationType)
	assert.Equal(t, "123abc", string(descriptors[1].CredentialID))
	assert.Equal(t, protocol.PublicKeyCredentialType, descriptors[1].Type)

	assert.Len(t, descriptors[1].Transport, 2)
	assert.Equal(t, protocol.AuthenticatorTransport("usb"), descriptors[1].Transport[0])
	assert.Equal(t, protocol.AuthenticatorTransport("nfc"), descriptors[1].Transport[1])
}

func TestWebauthnGetUserWithoutDisplayName(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	userSession := session.UserSession{
		Username: "john",
	}

	ctx.StorageMock.EXPECT().LoadWebauthnDevicesByUsername(ctx.Ctx, "example.com", "john").Return([]model.WebauthnDevice{
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

	user, err := getWebauthnUserByRPID(ctx.Ctx, userSession, "example.com")

	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, "john", user.WebAuthnDisplayName())
	assert.Equal(t, "john", user.DisplayName)
}

func TestWebauthnGetUserWithErr(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	userSession := session.UserSession{
		Username: "john",
	}

	ctx.StorageMock.EXPECT().LoadWebauthnDevicesByUsername(ctx.Ctx, "example.com", "john").Return(nil, errors.New("not found"))

	user, err := getWebauthnUserByRPID(ctx.Ctx, userSession, "example.com")

	assert.EqualError(t, err, "not found")
	assert.Nil(t, user)
}

func TestWebauthnNewWebauthnShouldReturnErrWhenHeadersNotAvailable(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)
	ctx.Ctx.Request.Header.Del("X-Forwarded-Host")

	w, err := newWebauthn(ctx.Ctx)

	assert.Nil(t, w)
	assert.EqualError(t, err, "missing required X-Forwarded-Host header")
}

func TestWebauthnNewWebauthnShouldReturnErrWhenWebauthnNotConfigured(t *testing.T) {
	ctx := mocks.NewMockAutheliaCtx(t)

	ctx.Ctx.Request.Header.Set("X-Forwarded-Host", "example.com")
	ctx.Ctx.Request.Header.Set("X-Forwarded-URI", "/")
	ctx.Ctx.Request.Header.Set("X-Forwarded-Proto", "https")

	w, err := newWebauthn(ctx.Ctx)

	assert.Nil(t, w)
	assert.EqualError(t, err, "Configuration error: Missing RPDisplayName")
}
