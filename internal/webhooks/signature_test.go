// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/events"
)

func TestSignShouldProduceTheDocumentedFormat(t *testing.T) {
	signature := Sign("secret", events.SignatureAlgorithmSHA256, 1789200306, []byte(`{"a":1}`))

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte("1789200306.{\"a\":1}"))

	assert.Equal(t, "t=1789200306,v1="+hex.EncodeToString(mac.Sum(nil)), signature)
}

func TestSignShouldBeStableForKnownVectors(t *testing.T) {
	testCases := []struct {
		name      string
		secret    string
		algorithm string
		timestamp int64
		body      string
		length    int
	}{
		{
			name:      "ShouldSignSHA256",
			secret:    "secret",
			algorithm: events.SignatureAlgorithmSHA256,
			timestamp: 1789200306,
			body:      `{"specversion":"1.0"}`,
			length:    64,
		},
		{
			name:      "ShouldSignSHA512",
			secret:    "secret",
			algorithm: events.SignatureAlgorithmSHA512,
			timestamp: 1789200306,
			body:      `{"specversion":"1.0"}`,
			length:    128,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Sign(tc.secret, tc.algorithm, tc.timestamp, []byte(tc.body))

			prefix := "t=" + strconv.FormatInt(tc.timestamp, 10) + ",v1="

			require.True(t, len(actual) > len(prefix))
			assert.Equal(t, prefix, actual[:len(prefix)])
			assert.Len(t, actual[len(prefix):], tc.length)
		})
	}
}

func TestSignShouldReturnEmptyWithoutASecret(t *testing.T) {
	assert.Equal(t, "", Sign("", events.SignatureAlgorithmSHA256, 1789200306, []byte(`{}`)))
}

func TestSignShouldDefaultToSHA256ForAnUnknownAlgorithm(t *testing.T) {
	assert.Equal(t,
		Sign("secret", events.SignatureAlgorithmSHA256, 1789200306, []byte(`{}`)),
		Sign("secret", "unknown", 1789200306, []byte(`{}`)),
	)
}
