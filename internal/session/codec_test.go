// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
	"time"

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

	for _, char := range id {
		assert.True(t, strings.ContainsRune(randomSessionChars, char), "character '%c' is not in the session charset", char)
	}
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

	_, err = uuid.Parse(id)
	assert.NoError(t, err)
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

func TestSecureCodec_SealAndOpenShouldRoundTripOpenIDConnectLogout(t *testing.T) {
	codec := newTestCodec(t)

	expires := time.Unix(1789807775, 0).UTC()

	testCases := []struct {
		name string
		have *OpenIDConnectLogout
	}{
		{"ShouldRoundTripAbsent", nil},
		{"ShouldRoundTripWithoutRedirect", &OpenIDConnectLogout{ClientID: "app", Expires: expires}},
		{"ShouldRoundTripWithRedirectAndState", &OpenIDConnectLogout{FlowID: "8c6f6e2a-5b3e-4f0a-9d1c-2e7b3a4f5c6d", ClientID: "app", RedirectURI: "https://app.example.com/logged-out", State: "abc123", Expires: expires}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userSession := NewUserSession("")
			userSession.CookieDomain = testDomain
			userSession.OpenIDConnectLogout = tc.have

			data, err := codec.Seal(testDomain, "id", userSession)
			require.NoError(t, err)

			actual := UserSession{}

			require.NoError(t, codec.Open(testDomain, NewRecord("id", data), &actual))

			if tc.have == nil {
				assert.Nil(t, actual.OpenIDConnectLogout)

				return
			}

			require.NotNil(t, actual.OpenIDConnectLogout)
			assert.Equal(t, tc.have.FlowID, actual.OpenIDConnectLogout.FlowID)
			assert.Equal(t, tc.have.ClientID, actual.OpenIDConnectLogout.ClientID)
			assert.Equal(t, tc.have.RedirectURI, actual.OpenIDConnectLogout.RedirectURI)
			assert.Equal(t, tc.have.State, actual.OpenIDConnectLogout.State)
			assert.True(t, tc.have.Expires.Equal(actual.OpenIDConnectLogout.Expires))
		})
	}
}

type failingRandom struct {
	random.Provider
}

func (r *failingRandom) StringCustomErr(_ int, _ string) (data string, err error) {
	return "", errTestFailure
}
