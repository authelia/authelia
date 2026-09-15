// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/random"
)

func TestSecureCodec_Verify(t *testing.T) {
	codec := newTestCodec(t)

	other, err := NewCodec(testSecret, []byte("another-hmac-key"), []byte(testHMACKey+"-csrf"), random.NewMathematical())
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

	other, err := NewCodec(testSecret, []byte(testHMACKey), []byte("another-csrf-hmac-key"), random.NewMathematical())
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

	for _, char := range id {
		assert.True(t, strings.ContainsRune(randomSessionChars, char), "character '%c' is not in the session charset", char)
	}
}

func TestSecureCodec_GenerateSessionIDShouldReturnRandomError(t *testing.T) {
	codec, err := NewCodec(testSecret, []byte(testHMACKey), []byte(testHMACKey+"-csrf"), &failingRandom{Provider: random.NewMathematical()})
	require.NoError(t, err)

	id, err := codec.GenerateSessionID()

	assert.EqualError(t, err, "bad stuff")
	assert.Empty(t, id)
}

func TestSecureCodec_GeneratePublicID(t *testing.T) {
	codec := newTestCodec(t)

	id, err := codec.GeneratePublicID()

	require.NoError(t, err)

	_, err = uuid.Parse(id)
	assert.NoError(t, err)
}

func TestSecureCodec_SealShouldReturnErrorWhenEncryptionFails(t *testing.T) {
	codec := &SecureCodec{encKey: []byte("short"), hmacKey: []byte(testHMACKey), random: random.NewMathematical(), charsetSessionID: randomSessionChars}

	data, err := codec.Seal(testDomain, "id", NewUserSession(testUsername))

	assert.Nil(t, data)
	assert.EqualError(t, err, "unable to encrypt session: crypto/aes: invalid key size 5")
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

func (r *failingRandom) StringCustomErr(_ int, _ string) (data string, err error) {
	return "", errTestFailure
}
