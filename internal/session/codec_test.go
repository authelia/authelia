// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/random"
)

func TestSecureCodec_Verify(t *testing.T) {
	codec := newTestCodec(t)

	keys := newTestHMACKeys()
	keys.Session = []byte("another-hmac-key")

	other, err := NewCodec(testSecret, keys, random.NewMathematical())
	require.NoError(t, err)

	data := []byte("a-session-identifier")

	testCases := []struct {
		name      string
		data      []byte
		signature string
		expected  bool
	}{
		{"ShouldVerifyValidSignature", data, codec.Sign(data), true},
		{"ShouldNotVerifySignatureOfOtherData", data, codec.Sign([]byte("other-data")), false},
		{"ShouldNotVerifySignatureFromAnotherKey", data, other.Sign(data), false},
		{"ShouldNotVerifyTruncatedSignature", data, codec.Sign(data)[:32], false},
		{"ShouldNotVerifyInvalidHex", data, "not-a-hex-signature", false},
		{"ShouldNotVerifyEmptySignature", data, "", false},
		{"ShouldNotVerifyCSRFSignature", data, codec.SignCSRF(data), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, codec.Verify(tc.data, tc.signature))
		})
	}
}

func TestSecureCodec_VerifyCSRF(t *testing.T) {
	codec := newTestCodec(t)

	keys := newTestHMACKeys()
	keys.SessionCSRF = []byte("another-csrf-hmac-key")

	other, err := NewCodec(testSecret, keys, random.NewMathematical())
	require.NoError(t, err)

	data := []byte("a-csrf-secret")

	testCases := []struct {
		name      string
		data      []byte
		signature string
		expected  bool
	}{
		{"ShouldVerifyValidSignature", data, codec.SignCSRF(data), true},
		{"ShouldNotVerifySignatureOfOtherData", data, codec.SignCSRF([]byte("other-data")), false},
		{"ShouldNotVerifySignatureFromAnotherCSRFKey", data, other.SignCSRF(data), false},
		{"ShouldNotVerifySessionSignature", data, codec.Sign(data), false},
		{"ShouldNotVerifySessionAuthzSignature", data, codec.SignAuthz(data), false},
		{"ShouldNotVerifyTruncatedSignature", data, codec.SignCSRF(data)[:32], false},
		{"ShouldNotVerifyInvalidHex", data, "not-a-hex-signature", false},
		{"ShouldNotVerifyEmptySignature", data, "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, codec.VerifyCSRF(tc.data, tc.signature))
		})
	}
}

func TestSecureCodec_SignCSRFShouldDifferFromSign(t *testing.T) {
	codec := newTestCodec(t)

	data := []byte("the-same-data")

	assert.NotEqual(t, codec.Sign(data), codec.SignCSRF(data))
	assert.Len(t, codec.SignCSRF(data), 64)
}

func TestSecureCodec_GenerateSessionID(t *testing.T) {
	codec := newTestCodec(t)

	id, err := codec.GenerateSessionID()

	require.NoError(t, err)
	assert.Len(t, id, 32)

	other, err := codec.GenerateSessionID()

	require.NoError(t, err)
	assert.NotEqual(t, id, other)
}

func TestSecureCodec_GenerateSessionIDShouldReturnRandomError(t *testing.T) {
	codec, err := NewCodec(testSecret, newTestHMACKeys(), &failingRandom{Provider: random.NewMathematical()})
	require.NoError(t, err)

	id, err := codec.GenerateSessionID()

	assert.EqualError(t, err, "bad stuff")
	assert.Empty(t, id)
}

func TestSecureCodec_GeneratePublicID(t *testing.T) {
	codec := newTestCodec(t)

	id, err := codec.GeneratePublicID()

	require.NoError(t, err)

	pid, err := uuid.Parse(id)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), pid.Version())
}

func TestSecureCodec_SealShouldReturnErrorWhenEncryptionFails(t *testing.T) {
	codec := &SecureCodec{encKey: []byte("short"), hmacKey: []byte(testHMACKey), random: random.NewMathematical()}

	data, err := codec.Seal(testDomain, "id", NewUserSession(testUsername))

	assert.Nil(t, data)
	assert.EqualError(t, err, "unable to encrypt session: crypto/aes: invalid key size 5")
}

func TestSecureCodec_SealShouldReturnErrorWhenMarshalFails(t *testing.T) {
	codec := newTestCodec(t)

	userSession := NewUserSession(testUsername)
	userSession.WebAuthn = &WebAuthn{
		SessionData: &webauthn.SessionData{
			Extensions: protocol.SessionExtensions{
				Extra: map[string]any{"unsupported": func() {}},
			},
		},
	}

	data, err := codec.Seal(testDomain, "id", userSession)

	assert.Nil(t, data)
	assert.ErrorContains(t, err, "unable to marshal session: ")
}

func TestSecureCodec_GeneratePublicIDShouldReturnRandomError(t *testing.T) {
	codec := newTestCodec(t)

	uuid.SetRand(&failingReader{})

	t.Cleanup(func() {
		uuid.SetRand(nil)
	})

	id, err := codec.GeneratePublicID()

	assert.EqualError(t, err, "bad stuff")
	assert.Empty(t, id)
}

func TestSecureCodec_OpenShouldIgnoreRecordsWithoutData(t *testing.T) {
	codec := newTestCodec(t)

	testCases := []struct {
		name   string
		record Record
	}{
		{"ShouldIgnoreNilRecord", nil},
		{"ShouldIgnoreRecordWithNoData", NewRecord("id", nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userSession := NewUserSession(testUsername)

			require.NoError(t, codec.Open(testDomain, tc.record, &userSession))

			assert.Equal(t, testUsername, userSession.Username)
		})
	}
}

func TestSecureCodec_SealAndOpenShouldRoundTrip(t *testing.T) {
	codec := newTestCodec(t)

	userSession := NewUserSession(testUsername)
	userSession.CookieDomain = testDomain

	data, err := codec.Seal(testDomain, "id", userSession)
	require.NoError(t, err)

	actual := UserSession{}

	require.NoError(t, codec.Open(testDomain, NewRecord("id", data), &actual))

	assert.Equal(t, testUsername, actual.Username)
	assert.Equal(t, testDomain, actual.CookieDomain)
}

func TestNewCodecShouldRequireEveryHMACKey(t *testing.T) {
	testCases := []struct {
		name     string
		have     func(keys *HMACKeys)
		expected string
	}{
		{"ShouldRequireSession", func(keys *HMACKeys) { keys.Session = nil }, "error occurred creating the session codec: every hmac key is required"},
		{"ShouldRequireSessionAuthz", func(keys *HMACKeys) { keys.SessionAuthz = nil }, "error occurred creating the session codec: every hmac key is required"},
		{"ShouldRequireSessionCSRF", func(keys *HMACKeys) { keys.SessionCSRF = []byte{} }, "error occurred creating the session codec: every hmac key is required"},
		{"ShouldRequireDistinctSessionAndAuthz", func(keys *HMACKeys) { keys.SessionAuthz = keys.Session }, "error occurred creating the session codec: every hmac key must be distinct"},
		{"ShouldRequireDistinctSessionAndCSRF", func(keys *HMACKeys) { keys.SessionCSRF = keys.Session }, "error occurred creating the session codec: every hmac key must be distinct"},
		{"ShouldRequireDistinctAuthzAndCSRF", func(keys *HMACKeys) { keys.SessionCSRF = keys.SessionAuthz }, "error occurred creating the session codec: every hmac key must be distinct"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys := newTestHMACKeys()

			tc.have(&keys)

			codec, err := NewCodec(testSecret, keys, random.NewMathematical())

			assert.Nil(t, codec)
			assert.EqualError(t, err, tc.expected)
		})
	}
}

func TestSecureCodecShouldSignWithTheKeyForEachPurpose(t *testing.T) {
	codec := newTestCodec(t)

	data := []byte("a-value")

	signatures := []string{codec.Sign(data), codec.SignAuthz(data), codec.SignCSRF(data)}

	assert.NotEqual(t, signatures[0], signatures[1])
	assert.NotEqual(t, signatures[0], signatures[2])
	assert.NotEqual(t, signatures[1], signatures[2])

	assert.True(t, codec.Verify(data, signatures[0]))
	assert.False(t, codec.Verify(data, signatures[1]))
	assert.False(t, codec.Verify(data, signatures[2]))

	keys := newTestHMACKeys()
	keys.SessionAuthz = []byte("a-rotated-authz-hmac-key")

	rotated, err := NewCodec(testSecret, keys, random.NewMathematical())
	require.NoError(t, err)

	assert.Equal(t, signatures[0], rotated.Sign(data))
	assert.NotEqual(t, signatures[1], rotated.SignAuthz(data))
	assert.Equal(t, signatures[2], rotated.SignCSRF(data))
}

type failingRandom struct {
	random.Provider
}

func (r *failingRandom) Read(_ []byte) (n int, err error) {
	return 0, errTestFailure
}

type failingReader struct{}

func (r *failingReader) Read(_ []byte) (n int, err error) {
	return 0, errTestFailure
}
