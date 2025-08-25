package model

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

/*
TestShouldOnlyMarshalPeriodAndDigitsAndAbsolutelyNeverSecret.
This test is vital to ensuring the TOTP configuration is marshalled correctly. If encoding/json suddenly changes
upstream and the json tag value of '-' doesn't exclude the field from marshalling then this test will pickup this
issue prior to code being shipped.

For this reason it's essential that the marshalled object contains all values populated, especially the secret.
*/
func TestShouldOnlyMarshalPeriodAndDigitsAndAbsolutelyNeverSecret(t *testing.T) {
	object := &TOTPConfiguration{
		ID:        1,
		Username:  "john",
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,

		// DO NOT CHANGE THIS VALUE UNLESS YOU FULLY UNDERSTAND THE COMMENT AT THE TOP OF THIS TEST.
		Secret: []byte("ABC123"),
	}

	object2 := TOTPConfiguration{
		ID:        1,
		Username:  "john",
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,

		// DO NOT CHANGE THIS VALUE UNLESS YOU FULLY UNDERSTAND THE COMMENT AT THE TOP OF THIS TEST.
		Secret: []byte("ABC123"),
	}

	data, err := json.Marshal(object)
	assert.NoError(t, err)

	data2, err := json.Marshal(object2)
	assert.NoError(t, err)

	assert.Equal(t, "{\"created_at\":\"0001-01-01T00:00:00Z\",\"issuer\":\"Authelia\",\"algorithm\":\"SHA1\",\"digits\":6,\"period\":30}", string(data))
	assert.Equal(t, "{\"created_at\":\"0001-01-01T00:00:00Z\",\"issuer\":\"Authelia\",\"algorithm\":\"SHA1\",\"digits\":6,\"period\":30}", string(data2))

	// DO NOT REMOVE OR CHANGE THESE TESTS UNLESS YOU FULLY UNDERSTAND THE COMMENT AT THE TOP OF THIS TEST.
	require.NotContains(t, string(data), "secret")
	require.NotContains(t, string(data2), "secret")
	require.NotContains(t, string(data), "ABC123")
	require.NotContains(t, string(data2), "ABC123")
	require.NotContains(t, string(data), "QUJDMTIz")
	require.NotContains(t, string(data2), "QUJDMTIz")
}

func TestShouldReturnErrWhenImageTooSmall(t *testing.T) {
	object := &TOTPConfiguration{
		ID:        1,
		Username:  "john",
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Secret:    []byte("ABC123"),
	}

	img, err := object.Image(10, 10)

	assert.EqualError(t, err, "can not scale barcode to an image smaller than 41x41")
	assert.Nil(t, img)
}

func TestShouldReturnImage(t *testing.T) {
	object := &TOTPConfiguration{
		ID:        1,
		Username:  "john",
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Secret:    []byte("ABC123"),
	}

	img, err := object.Image(41, 41)

	assert.NoError(t, err)
	require.NotNil(t, img)

	assert.Equal(t, 41, img.Bounds().Dx())
	assert.Equal(t, 41, img.Bounds().Dy())
}

func TestTOTPConfigurationImportExport(t *testing.T) {
	have := TOTPConfigurationExport{
		TOTPConfigurations: []TOTPConfiguration{
			{
				ID:         0,
				CreatedAt:  time.Now(),
				LastUsedAt: sql.NullTime{Valid: false},
				Username:   "john",
				Issuer:     "example",
				Algorithm:  "SHA1",
				Digits:     6,
				Period:     30,
				Secret:     MustRead(80),
			},
			{
				ID:         1,
				CreatedAt:  time.Now(),
				LastUsedAt: sql.NullTime{Time: time.Now(), Valid: true},
				Username:   "abc",
				Issuer:     "example2",
				Algorithm:  "SHA512",
				Digits:     8,
				Period:     90,
				Secret:     MustRead(120),
			},
		},
	}

	out, err := yaml.Marshal(&have)
	require.NoError(t, err)

	imported := TOTPConfigurationExport{}

	require.NoError(t, yaml.Unmarshal(out, &imported))

	require.Equal(t, len(have.TOTPConfigurations), len(imported.TOTPConfigurations))

	for i, actual := range imported.TOTPConfigurations {
		t.Run(actual.Username, func(t *testing.T) {
			expected := have.TOTPConfigurations[i]

			if expected.ID != 0 {
				assert.NotEqual(t, expected.ID, actual.ID)
			} else {
				assert.Equal(t, expected.ID, actual.ID)
			}

			assert.Equal(t, expected.Username, actual.Username)
			assert.Equal(t, expected.Issuer, actual.Issuer)
			assert.Equal(t, expected.Algorithm, actual.Algorithm)
			assert.Equal(t, expected.Digits, actual.Digits)
			assert.Equal(t, expected.Period, actual.Period)
			assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
			assert.WithinDuration(t, expected.LastUsedAt.Time, actual.LastUsedAt.Time, time.Second)
			assert.Equal(t, expected.LastUsedAt.Valid, actual.LastUsedAt.Valid)
		})
	}
}

func MustRead(n int) []byte {
	data := make([]byte, n)

	if _, err := rand.Read(data); err != nil {
		panic(err)
	}

	return data
}
