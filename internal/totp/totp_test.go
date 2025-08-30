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
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
)

func TestTOTPGenerateCustom(t *testing.T) {
	testCases := []struct {
		name                        string
		username, algorithm, secret string
		digits                      uint32
		period, secretSize          uint
		setup                       func(t *testing.T, totp *TimeBased)
		err                         string
	}{
		{
			name:       "ShouldGenerateSHA1",
			username:   "john",
			algorithm:  "SHA1",
			digits:     6,
			period:     30,
			secretSize: 32,
		},
		{
			name:       "ShouldGenerateLongSecret",
			username:   "john",
			algorithm:  "SHA1",
			digits:     6,
			period:     30,
			secretSize: 42,
		},
		{
			name:       "ShouldGenerateSHA256",
			username:   "john",
			algorithm:  "SHA256",
			digits:     6,
			period:     30,
			secretSize: 32,
		},
		{
			name:       "ShouldGenerateSHA512",
			username:   "john",
			algorithm:  "SHA512",
			digits:     6,
			period:     30,
			secretSize: 32,
		},
		{
			name:       "ShouldGenerateWithSecret",
			username:   "john",
			algorithm:  "SHA512",
			secret:     "ONTGOYLTMZQXGZDBONSGC43EMFZWMZ3BONTWMYLTMRQXGZBSGMYTEMZRMFYXGZDBONSA",
			digits:     6,
			period:     30,
			secretSize: 32,
		},
		{
			name:       "ShouldGenerateWithBadSecretB32Data",
			username:   "john",
			algorithm:  "SHA512",
			secret:     "@#UNH$IK!J@N#EIKJ@U!NIJKUN@#WIK",
			digits:     6,
			period:     30,
			secretSize: 32,
			err:        "totp generate failed: error decoding base32 string: illegal base32 data at input byte 0",
		},
		{
			name:       "ShouldGenerateWithBadSecretLength",
			username:   "john",
			algorithm:  "SHA512",
			secret:     "ONTGOYLTMZQXGZD",
			digits:     6,
			period:     30,
			secretSize: 0,
		},
		{
			name:       "ShouldHandleGenerateError",
			username:   "john",
			algorithm:  "SHA512",
			secret:     "ONTGOYLTMZQXGZD",
			digits:     6,
			period:     30,
			secretSize: 0,
			setup: func(t *testing.T, totp *TimeBased) {
				totp.issuer = ""
			},
			err: "error generating totp: issuer must be set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			totp := NewTimeBasedProvider(schema.TOTP{
				Issuer:           "Authelia",
				DefaultAlgorithm: "SHA1",
				DefaultDigits:    6,
				DefaultPeriod:    30,
				SecretSize:       32,
			})

			ctx := NewContext(context.TODO(), clock.New(), random.New())

			if tc.setup != nil {
				tc.setup(t, totp)
			}

			c, err := totp.GenerateCustom(ctx, tc.username, tc.algorithm, tc.secret, tc.digits, tc.period, tc.secretSize)
			if tc.err == "" {
				assert.NoError(t, err)
				require.NotNil(t, c)
				assert.Equal(t, tc.period, c.Period)
				assert.Equal(t, tc.digits, c.Digits)
				assert.Equal(t, tc.algorithm, c.Algorithm)

				expectedSecretLen := int(tc.secretSize) //nolint:gosec
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

	ctx := NewContext(context.TODO(), clock.New(), random.New())

	config, err := totp.Generate(ctx, "john")
	assert.NoError(t, err)

	assert.Equal(t, "Authelia", config.Issuer)

	assert.Less(t, time.Since(config.CreatedAt), time.Second)
	assert.Greater(t, time.Since(config.CreatedAt), time.Second*-1)

	assert.Equal(t, uint32(8), config.Digits)
	assert.Equal(t, uint(60), config.Period)
	assert.Equal(t, "SHA256", config.Algorithm)

	secret := make([]byte, base32.StdEncoding.WithPadding(base32.NoPadding).DecodedLen(len(config.Secret)))

	_, err = base32.StdEncoding.WithPadding(base32.NoPadding).Decode(secret, config.Secret)
	assert.NoError(t, err)
	assert.Len(t, secret, 32)

	assert.NotNil(t, totp.Options())
}

func TestTimeBased_Validate(t *testing.T) {
	testCases := []struct {
		name   string
		token  string
		config *model.TOTPConfiguration
		valid  bool
		step   uint64
		err    string
	}{
		{
			"ShouldValidate",
			"035781",
			&model.TOTPConfiguration{
				Issuer:    "Authelia",
				Digits:    6,
				Algorithm: "SHA1",
				Period:    30,
				Secret:    []byte("ONTGOYLTMZQXGZDBONSGC43EMFZWMZ3BONTWMYLTMRQXGZBSGMYTEMZRMFYXGZDBONSA"),
			},
			true,
			0x51615,
			"",
		},
		{
			"ShouldNotValidate",
			"035782",
			&model.TOTPConfiguration{
				Issuer:    "Authelia",
				Digits:    6,
				Algorithm: "SHA1",
				Period:    30,
				Secret:    []byte("ONTGOYLTMZQXGZDBONSGC43EMFZWMZ3BONTWMYLTMRQXGZBSGMYTEMZRMFYXGZDBONSA"),
			},
			false,
			0,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewContext(context.TODO(), clock.NewFixed(time.Unix(10000000, 0)), random.New())

			skew := 1

			totp := NewTimeBasedProvider(schema.TOTP{
				Issuer:           "Authelia",
				DefaultAlgorithm: "SHA256",
				DefaultDigits:    8,
				DefaultPeriod:    60,
				Skew:             &skew,
				SecretSize:       32,
			})

			valid, step, err := totp.Validate(ctx, tc.token, tc.config)
			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.valid, valid)
				assert.Equal(t, tc.step, step)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}
