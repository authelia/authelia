package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
TestShouldOnlyMarshalPeriodAndDigitsAndAbsolutelyNeverSecret.
This test is vital to ensuring the TOTP configuration is marshalled correctly. If encoding/json suddenly changes
upstream and the json tag value of '-' doesn't exclude the field from marshalling then this test will pickup this
issue prior to code being shipped.

For this reason it's essential that the marshalled object contains all values populated, especially the secret.
*/
func TestShouldOnlyMarshalPeriodAndDigitsAndAbsolutelyNeverSecret(t *testing.T) {
	object := TOTPConfiguration{
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

	assert.Equal(t, "{\"digits\":6,\"period\":30}", string(data))

	// DO NOT REMOVE OR CHANGE THESE TESTS UNLESS YOU FULLY UNDERSTAND THE COMMENT AT THE TOP OF THIS TEST.
	require.NotContains(t, string(data), "secret")
	require.NotContains(t, string(data), "ABC123")
}

func TestShouldReturnErrWhenImageTooSmall(t *testing.T) {
	object := TOTPConfiguration{
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
	object := TOTPConfiguration{
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
