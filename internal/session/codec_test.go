// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/random"
)

func TestSecureCodec_Verify(t *testing.T) {
	codec := newTestCodec(t)

	other, err := NewCodec(testSecret, []byte("another-hmac-key"), random.NewMathematical())
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, codec.Verify(tc.data, tc.signature))
		})
	}
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
	codec, err := NewCodec(testSecret, []byte(testHMACKey), &failingRandom{Provider: random.NewMathematical()})
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

func TestSecureCodecShouldGeneratePublicIDVersion7(t *testing.T) {
	codec := newTestCodec(t)

	var ids [5][16]byte

	var idStrings [5]string

	now := time.Now().UnixMilli()

	for i := 0; i < 5; i++ {
		idStr, err := codec.GeneratePublicID()
		require.NoError(t, err)

		idStrings[i] = idStr
		parsed, err := uuid.Parse(idStr)
		require.NoError(t, err)
		assert.Equal(t, uuid.Version(7), parsed.Version())

		ids[i] = parsed

		var ts [8]byte
		copy(ts[2:], ids[i][:6])

		//nolint:gosec
		timestamp := int64(binary.BigEndian.Uint64(ts[:]))

		minTime := now - 60000
		maxTime := now + 60000

		assert.GreaterOrEqual(t, timestamp, minTime, "timestamp should not be before 1 minute ago")
		assert.LessOrEqual(t, timestamp, maxTime, "timestamp should not be after 1 minute from now")
	}

	for i := 1; i < 5; i++ {
		cmp := bytes.Compare(ids[i-1][:6], ids[i][:6])
		assert.LessOrEqual(t, cmp, 0, "timestamp prefix should be non-decreasing")
	}

	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			assert.NotEqual(t, idStrings[i], idStrings[j])
		}
	}
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
