package oidc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestClientSecretDigestCompare(t *testing.T) {
	testCases := []struct {
		name      string
		digest    func(t *testing.T) *ClientSecretDigest
		secret    []byte
		err       bool
		errTarget error
	}{
		{
			"ShouldErrWhenNilPasswordDigest",
			func(t *testing.T) *ClientSecretDigest {
				return &ClientSecretDigest{PasswordDigest: nil}
			},
			[]byte("secret"),
			true,
			nil,
		},
		{
			"ShouldErrWhenNilDigest",
			func(t *testing.T) *ClientSecretDigest {
				return &ClientSecretDigest{PasswordDigest: &schema.PasswordDigest{}}
			},
			[]byte("secret"),
			true,
			nil,
		},
		{
			"ShouldSucceedWhenMatchingPlainText",
			func(t *testing.T) *ClientSecretDigest {
				pd, err := schema.DecodePasswordDigest("$plaintext$mysecret")
				require.NoError(t, err)

				return &ClientSecretDigest{PasswordDigest: pd}
			},
			[]byte("mysecret"),
			false,
			nil,
		},
		{
			"ShouldErrWhenMismatch",
			func(t *testing.T) *ClientSecretDigest {
				pd, err := schema.DecodePasswordDigest("$plaintext$mysecret")
				require.NoError(t, err)

				return &ClientSecretDigest{PasswordDigest: pd}
			},
			[]byte("wrongsecret"),
			true,
			errClientSecretMismatch,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := tc.digest(t)

			err := d.Compare(context.Background(), tc.secret)

			if tc.err {
				assert.Error(t, err)

				if tc.errTarget != nil {
					assert.Equal(t, tc.errTarget, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
