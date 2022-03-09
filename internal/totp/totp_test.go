package totp

import (
	"encoding/base32"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestTOTPGenerateCustom(t *testing.T) {
	totp := NewTimeBasedProvider(schema.TOTPConfiguration{
		Issuer:    "Authelia",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
	})

	assert.Equal(t, uint(1), totp.skew)

	config, err := totp.GenerateCustom("john", "SHA1", 6, 30, 32)
	assert.NoError(t, err)

	assert.Equal(t, uint(6), config.Digits)
	assert.Equal(t, uint(30), config.Period)
	assert.Equal(t, "SHA1", config.Algorithm)

	assert.Less(t, time.Since(config.CreatedAt), time.Second)
	assert.Greater(t, time.Since(config.CreatedAt), time.Second*-1)

	secret := make([]byte, base32.StdEncoding.WithPadding(base32.NoPadding).DecodedLen(len(config.Secret)))

	_, err = base32.StdEncoding.WithPadding(base32.NoPadding).Decode(secret, config.Secret)
	assert.NoError(t, err)
	assert.Len(t, secret, 32)

	config, err = totp.GenerateCustom("john", "SHA1", 6, 30, 42)
	assert.NoError(t, err)

	assert.Equal(t, uint(6), config.Digits)
	assert.Equal(t, uint(30), config.Period)
	assert.Equal(t, "SHA1", config.Algorithm)

	assert.Less(t, time.Since(config.CreatedAt), time.Second)
	assert.Greater(t, time.Since(config.CreatedAt), time.Second*-1)

	secret = make([]byte, base32.StdEncoding.WithPadding(base32.NoPadding).DecodedLen(len(config.Secret)))

	_, err = base32.StdEncoding.WithPadding(base32.NoPadding).Decode(secret, config.Secret)
	assert.NoError(t, err)
	assert.Len(t, secret, 42)

	_, err = totp.GenerateCustom("", "SHA1", 6, 30, 32)
	assert.EqualError(t, err, "AccountName must be set")
}

func TestTOTPGenerate(t *testing.T) {
	skew := uint(2)

	totp := NewTimeBasedProvider(schema.TOTPConfiguration{
		Issuer:    "Authelia",
		Algorithm: "SHA256",
		Digits:    8,
		Period:    60,
		Skew:      &skew,
	})

	assert.Equal(t, uint(2), totp.skew)

	config, err := totp.Generate("john")
	assert.NoError(t, err)

	assert.Equal(t, "Authelia", config.Issuer)

	assert.Less(t, time.Since(config.CreatedAt), time.Second)
	assert.Greater(t, time.Since(config.CreatedAt), time.Second*-1)

	assert.Equal(t, uint(8), config.Digits)
	assert.Equal(t, uint(60), config.Period)
	assert.Equal(t, "SHA256", config.Algorithm)

	secret := make([]byte, base32.StdEncoding.WithPadding(base32.NoPadding).DecodedLen(len(config.Secret)))

	_, err = base32.StdEncoding.WithPadding(base32.NoPadding).Decode(secret, config.Secret)
	assert.NoError(t, err)
	assert.Len(t, secret, 32)
}
