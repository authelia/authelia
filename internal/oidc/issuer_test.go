package oidc

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestNewIssuerDefaultKeyID(t *testing.T) {
	testCases := []struct {
		name     string
		keys     []schema.JWK
		expected string
	}{
		{
			"ShouldReturnEmptyForNoKeys",
			nil,
			"",
		},
		{
			"ShouldReturnEmptyForEmptyKeys",
			[]schema.JWK{},
			"",
		},
		{
			"ShouldReturnKeyIDForRS256Sig",
			[]schema.JWK{
				{KeyID: "my-key", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256},
			},
			"my-key",
		},
		{
			"ShouldReturnFirstMatchingRS256Sig",
			[]schema.JWK{
				{KeyID: "ec-key", Use: KeyUseSignature, Algorithm: SigningAlgECDSAUsingP256AndSHA256},
				{KeyID: "rsa-key", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256},
			},
			"rsa-key",
		},
		{
			"ShouldReturnEmptyWhenNoRS256",
			[]schema.JWK{
				{KeyID: "ec-key", Use: KeyUseSignature, Algorithm: SigningAlgECDSAUsingP256AndSHA256},
			},
			"",
		},
		{
			"ShouldReturnEmptyWhenUseIsNotSig",
			[]schema.JWK{
				{KeyID: "enc-key", Use: KeyUseEncryption, Algorithm: SigningAlgRSAUsingSHA256},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, NewIssuerDefaultKeyID(tc.keys))
		})
	}
}

func TestNewJSONWebKeySet(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		keys        []schema.JWK
		expectedNil bool
		expectedLen int
	}{
		{
			"ShouldReturnNilForNilKeys",
			nil,
			true,
			0,
		},
		{
			"ShouldReturnNilForEmptyKeys",
			[]schema.JWK{},
			true,
			0,
		},
		{
			"ShouldReturnSetForSingleKey",
			[]schema.JWK{
				{KeyID: "rsa-1", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256, Key: rsaKey},
			},
			false,
			1,
		},
		{
			"ShouldReturnSetForMultipleKeys",
			[]schema.JWK{
				{KeyID: "rsa-1", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256, Key: rsaKey},
				{KeyID: "rsa-2", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA384, Key: rsaKey},
			},
			false,
			2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewJSONWebKeySet(tc.keys)

			if tc.expectedNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Len(t, result.Keys, tc.expectedLen)
			}
		})
	}
}

func TestNewJSONWebKeySetPublic(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		keys        []schema.JWK
		expectedLen int
	}{
		{
			"ShouldReturnEmptySetForNoKeys",
			nil,
			0,
		},
		{
			"ShouldReturnPublicKeysOnly",
			[]schema.JWK{
				{KeyID: "rsa-1", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256, Key: rsaKey},
			},
			1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewJSONWebKeySetPublic(tc.keys)

			require.NotNil(t, result)
			assert.Len(t, result.Keys, tc.expectedLen)

			for _, key := range result.Keys {
				assert.True(t, key.IsPublic())
			}
		})
	}
}

func TestNewJSONWebKey(t *testing.T) {
	testCases := []struct {
		name      string
		key       func(t *testing.T) schema.JWK
		expectedK string
		expectedA string
		expectedU string
	}{
		{
			"ShouldCreateRSAJWK",
			func(t *testing.T) schema.JWK {
				k, err := rsa.GenerateKey(rand.Reader, 2048)
				require.NoError(t, err)

				return schema.JWK{KeyID: "rsa-1", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256, Key: k}
			},
			"rsa-1",
			SigningAlgRSAUsingSHA256,
			KeyUseSignature,
		},
		{
			"ShouldCreateECDSAJWK",
			func(t *testing.T) schema.JWK {
				k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				require.NoError(t, err)

				return schema.JWK{KeyID: "ec-1", Use: KeyUseSignature, Algorithm: SigningAlgECDSAUsingP256AndSHA256, Key: k}
			},
			"ec-1",
			SigningAlgECDSAUsingP256AndSHA256,
			KeyUseSignature,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jwk := NewJSONWebKey(tc.key(t))

			assert.Equal(t, tc.expectedK, jwk.KeyID)
			assert.Equal(t, tc.expectedA, jwk.Algorithm)
			assert.Equal(t, tc.expectedU, jwk.Use)
			assert.NotNil(t, jwk.Key)
		})
	}
}

func TestIssuerGetKeyID(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	keys := []schema.JWK{
		{KeyID: "rsa-default", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256, Key: rsaKey},
		{KeyID: "ec-key", Use: KeyUseSignature, Algorithm: SigningAlgECDSAUsingP256AndSHA256, Key: ecKey},
	}

	issuer := NewIssuer(keys)

	testCases := []struct {
		name     string
		kid      string
		alg      string
		expected string
	}{
		{
			"ShouldReturnMatchingKID",
			"ec-key",
			SigningAlgECDSAUsingP256AndSHA256,
			"ec-key",
		},
		{
			"ShouldReturnDefaultForUnknownKID",
			"unknown",
			SigningAlgRSAUsingSHA256,
			"rsa-default",
		},
		{
			"ShouldReturnDefaultForEmptyKIDAndAlg",
			"",
			"",
			"rsa-default",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := issuer.GetKeyID(context.Background(), tc.kid, tc.alg)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIssuerGetPublicJSONWebKeys(t *testing.T) {
	rsaKey1, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	rsaKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		keys        []schema.JWK
		expectedLen int
	}{
		{
			"ShouldReturnSinglePublicKey",
			[]schema.JWK{
				{KeyID: "rsa-1", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256, Key: rsaKey1},
			},
			1,
		},
		{
			"ShouldReturnMultiplePublicKeys",
			[]schema.JWK{
				{KeyID: "rsa-1", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256, Key: rsaKey1},
				{KeyID: "rsa-2", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA384, Key: rsaKey2},
			},
			2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issuer := NewIssuer(tc.keys)

			jwks := issuer.GetPublicJSONWebKeys(&issuerTestContext{})

			require.NotNil(t, jwks)
			assert.Len(t, jwks.Keys, tc.expectedLen)

			for _, key := range jwks.Keys {
				assert.True(t, key.IsPublic())
			}
		})
	}
}

func TestIssuerGetIssuerJWK(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keys := []schema.JWK{
		{KeyID: "rsa-1", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256, Key: rsaKey},
	}

	issuer := NewIssuer(keys)

	testCases := []struct {
		name string
		kid  string
		alg  string
		use  string
		err  bool
	}{
		{
			"ShouldFindExistingKey",
			"rsa-1",
			SigningAlgRSAUsingSHA256,
			KeyUseSignature,
			false,
		},
		{
			"ShouldErrForNonExistentKey",
			"unknown",
			SigningAlgRSAUsingSHA256,
			KeyUseSignature,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jwk, err := issuer.GetIssuerJWK(context.Background(), tc.kid, tc.alg, tc.use)

			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, jwk)
			}
		})
	}
}

func TestIssuerGetIssuerStrictJWK(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keys := []schema.JWK{
		{KeyID: "rsa-1", Use: KeyUseSignature, Algorithm: SigningAlgRSAUsingSHA256, Key: rsaKey},
	}

	issuer := NewIssuer(keys)

	testCases := []struct {
		name string
		kid  string
		alg  string
		use  string
		err  bool
	}{
		{
			"ShouldFindExistingKeyStrict",
			"rsa-1",
			SigningAlgRSAUsingSHA256,
			KeyUseSignature,
			false,
		},
		{
			"ShouldErrForNonExistentKeyStrict",
			"unknown",
			"unknown",
			KeyUseSignature,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jwk, err := issuer.GetIssuerStrictJWK(context.Background(), tc.kid, tc.alg, tc.use)

			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, jwk)
			}
		})
	}
}

type issuerTestContext struct {
	context.Context
}

func (c *issuerTestContext) IssuerURL() (*url.URL, error) { return nil, nil }

func (c *issuerTestContext) GetClock() clock.Provider { return nil }

func (c *issuerTestContext) GetRandom() random.Provider { return nil }

func (c *issuerTestContext) GetConfiguration() *schema.Configuration { return nil }

func (c *issuerTestContext) GetProviderStorage() storage.Provider { return nil }

func (c *issuerTestContext) GetUserProvider() authentication.UserProvider { return nil }

func (c *issuerTestContext) GetProviderUserAttributeResolver() expression.UserAttributeResolver {
	return nil
}
