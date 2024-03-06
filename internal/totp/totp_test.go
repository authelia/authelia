package totp

import (
	"context"
	"encoding/base32"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

func TestTOTPGenerateCustom(t *testing.T) {
	testCases := []struct {
		desc                        string
		username, algorithm, secret string
		digits, period, secretSize  uint
		err                         string
	}{
		{
			desc:       "ShouldGenerateSHA1",
			username:   "john",
			algorithm:  "SHA1",
			digits:     6,
			period:     30,
			secretSize: 32,
		},
		{
			desc:       "ShouldGenerateLongSecret",
			username:   "john",
			algorithm:  "SHA1",
			digits:     6,
			period:     30,
			secretSize: 42,
		},
		{
			desc:       "ShouldGenerateSHA256",
			username:   "john",
			algorithm:  "SHA256",
			digits:     6,
			period:     30,
			secretSize: 32,
		},
		{
			desc:       "ShouldGenerateSHA512",
			username:   "john",
			algorithm:  "SHA512",
			digits:     6,
			period:     30,
			secretSize: 32,
		},
		{
			desc:       "ShouldGenerateWithSecret",
			username:   "john",
			algorithm:  "SHA512",
			secret:     "ONTGOYLTMZQXGZDBONSGC43EMFZWMZ3BONTWMYLTMRQXGZBSGMYTEMZRMFYXGZDBONSA",
			digits:     6,
			period:     30,
			secretSize: 32,
		},
		{
			desc:       "ShouldGenerateWithBadSecretB32Data",
			username:   "john",
			algorithm:  "SHA512",
			secret:     "@#UNH$IK!J@N#EIKJ@U!NIJKUN@#WIK",
			digits:     6,
			period:     30,
			secretSize: 32,
			err:        "totp generate failed: error decoding base32 string: illegal base32 data at input byte 0",
		},
		{
			desc:       "ShouldGenerateWithBadSecretLength",
			username:   "john",
			algorithm:  "SHA512",
			secret:     "ONTGOYLTMZQXGZD",
			digits:     6,
			period:     30,
			secretSize: 0,
		},
	}

	totp := NewTimeBasedProvider(schema.TOTP{
		Issuer:           "Authelia",
		DefaultAlgorithm: "SHA1",
		DefaultDigits:    6,
		DefaultPeriod:    30,
		SecretSize:       32,
	})

	ctx := NewContext(context.TODO(), &clock.Real{}, &random.Cryptographical{})

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := totp.GenerateCustom(ctx, tc.username, tc.algorithm, tc.secret, tc.digits, tc.period, tc.secretSize)
			if tc.err == "" {
				assert.NoError(t, err)
				require.NotNil(t, c)
				assert.Equal(t, tc.period, c.Period)
				assert.Equal(t, tc.digits, c.Digits)
				assert.Equal(t, tc.algorithm, c.Algorithm)

				expectedSecretLen := int(tc.secretSize)
				if tc.secret != "" {
					expectedSecretLen = base32.StdEncoding.WithPadding(base32.NoPadding).DecodedLen(len(tc.secret))
				}

				secret := make([]byte, expectedSecretLen)

				n, err := base32.StdEncoding.WithPadding(base32.NoPadding).Decode(secret, c.Secret)
				assert.NoError(t, err)
				assert.Len(t, secret, expectedSecretLen)
				assert.Equal(t, expectedSecretLen, n)
			} else {
				assert.Nil(t, c)
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestTOTPGenerate(t *testing.T) {
	skew := 2

	totp := NewTimeBasedProvider(schema.TOTP{
		Issuer:           "Authelia",
		DefaultAlgorithm: "SHA256",
		DefaultDigits:    8,
		DefaultPeriod:    60,
		Skew:             &skew,
		SecretSize:       32,
	})

	assert.Equal(t, uint(2), totp.skew)

	ctx := NewContext(context.TODO(), &clock.Real{}, &random.Cryptographical{})

	config, err := totp.Generate(ctx, "john")
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
