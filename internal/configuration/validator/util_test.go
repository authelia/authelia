package validator

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestMiscMissingCoverage(t *testing.T) {
	kid, err := jwkCalculateKID(struct{}{}, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "", kid)
}

func TestIsCookieDomainValid(t *testing.T) {
	testCases := []struct {
		domain   string
		expected bool
	}{
		{"example.com", false},
		{".example.com", false},
		{"*.example.com", false},
		{"authelia.com", false},
		{"duckdns.org", true},
		{".duckdns.org", true},
		{"example.duckdns.org", false},
		{"shiftcrypto.dev", false},
		{"192.168.2.1", false},
		{"localhost", true},
		{"com", true},
		{"randomnada", true},
	}

	for _, tc := range testCases {
		name := "ShouldFail"

		if tc.expected {
			name = "ShouldPass"
		}

		t.Run(tc.domain, func(t *testing.T) {
			t.Run(name, func(t *testing.T) {
				assert.Equal(t, tc.expected, isCookieDomainAPublicSuffix(tc.domain))
			})
		})
	}
}

func TestSchemaJWKGetPropertiesMissingTests(t *testing.T) {
	props, err := schemaJWKGetProperties(schema.JWK{Key: keyECDSAP224})

	assert.NoError(t, err)
	assert.Equal(t, oidc.KeyUseSignature, props.Use)
	assert.Equal(t, "", props.Algorithm)
	assert.Equal(t, elliptic.P224(), props.Curve)
	assert.Equal(t, -1, props.Bits)

	props, err = schemaJWKGetProperties(schema.JWK{Key: keyECDSAP224.Public()})

	assert.NoError(t, err)
	assert.Equal(t, oidc.KeyUseSignature, props.Use)
	assert.Equal(t, "", props.Algorithm)
	assert.Equal(t, elliptic.P224(), props.Curve)
	assert.Equal(t, -1, props.Bits)

	rsa := &rsa.PrivateKey{}

	*rsa = *keyRSA2048
	rsa.N = nil

	props, err = schemaJWKGetProperties(schema.JWK{Key: rsa})

	assert.NoError(t, err)
	assert.Equal(t, oidc.KeyUseSignature, props.Use)
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, props.Algorithm)
	assert.Equal(t, nil, props.Curve)
	assert.Equal(t, 0, props.Bits)
}

func TestGetResponseObjectAlgFromKID(t *testing.T) {
	c := &schema.IdentityProvidersOpenIDConnect{
		JSONWebKeys: []schema.JWK{
			{KeyID: "abc", Algorithm: "EX256"},
			{KeyID: "123", Algorithm: "EX512"},
		},
	}

	assert.Equal(t, "EX256", getResponseObjectAlgFromKID(c, "abc", "not"))
	assert.Equal(t, "EX512", getResponseObjectAlgFromKID(c, "123", "not"))
	assert.Equal(t, "not", getResponseObjectAlgFromKID(c, "111111", "not"))
}

func TestSchemaJWKGetPropertiesEnc(t *testing.T) {
	testCases := []struct {
		name        string
		key         func(t *testing.T) any
		expectedUse string
		expectedAlg string
		expectedBts int
		err         string
	}{
		{
			"ShouldReturnNilForNilKey",
			func(t *testing.T) any { return nil },
			"",
			"",
			0,
			"",
		},
		{
			"ShouldReturnA256GCMKWForSymmetric256",
			func(t *testing.T) any { return make([]byte, 256) },
			oidc.KeyUseEncryption,
			oidc.EncryptionAlgA256GCMKW,
			256,
			"",
		},
		{
			"ShouldReturnA192GCMKWForSymmetric192",
			func(t *testing.T) any { return make([]byte, 192) },
			oidc.KeyUseEncryption,
			oidc.EncryptionAlgA192GCMKW,
			192,
			"",
		},
		{
			"ShouldReturnA128GCMKWForSymmetric128",
			func(t *testing.T) any { return make([]byte, 128) },
			oidc.KeyUseEncryption,
			oidc.EncryptionAlgA128GCMKW,
			128,
			"",
		},
		{
			"ShouldReturnDirectForSmallSymmetric",
			func(t *testing.T) any { return make([]byte, 32) },
			oidc.KeyUseEncryption,
			oidc.EncryptionAlgDirect,
			32,
			"",
		},
		{
			"ShouldErrForLargeNonStandardSymmetric",
			func(t *testing.T) any { return make([]byte, 64) },
			"",
			"",
			0,
			"invalid symmetric key length of 64 but the minimum is 32",
		},
		{
			"ShouldReturnEmptyForEd25519PrivateKey",
			func(t *testing.T) any {
				_, priv, err := ed25519.GenerateKey(rand.Reader)
				require.NoError(t, err)

				return priv
			},
			"",
			"",
			0,
			"",
		},
		{
			"ShouldReturnEmptyForEd25519PublicKey",
			func(t *testing.T) any {
				pub, _, err := ed25519.GenerateKey(rand.Reader)
				require.NoError(t, err)

				return pub
			},
			"",
			"",
			0,
			"",
		},
		{
			"ShouldReturnRSAOAEP256ForRSAPrivateKey",
			func(t *testing.T) any {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				require.NoError(t, err)

				return key
			},
			oidc.KeyUseEncryption,
			oidc.EncryptionAlgRSAOAEP256,
			256,
			"",
		},
		{
			"ShouldReturnRSAOAEP256ForRSAPublicKey",
			func(t *testing.T) any {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				require.NoError(t, err)

				return &key.PublicKey
			},
			oidc.KeyUseEncryption,
			oidc.EncryptionAlgRSAOAEP256,
			256,
			"",
		},
		{
			"ShouldReturnRSAOAEP256ForRSAPrivateKeyNilN",
			func(t *testing.T) any {
				return &rsa.PrivateKey{}
			},
			oidc.KeyUseEncryption,
			oidc.EncryptionAlgRSAOAEP256,
			0,
			"",
		},
		{
			"ShouldReturnRSAOAEP256ForRSAPublicKeyNilN",
			func(t *testing.T) any {
				return &rsa.PublicKey{}
			},
			oidc.KeyUseEncryption,
			oidc.EncryptionAlgRSAOAEP256,
			0,
			"",
		},
		{
			"ShouldReturnECDHESA256KWForECDSAPrivateKey",
			func(t *testing.T) any {
				key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				require.NoError(t, err)

				return key
			},
			oidc.KeyUseEncryption,
			oidc.EncryptionAlgECDHESA256KW,
			-1,
			"",
		},
		{
			"ShouldReturnECDHESA256KWForECDSAPublicKey",
			func(t *testing.T) any {
				key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				require.NoError(t, err)

				return &key.PublicKey
			},
			oidc.KeyUseEncryption,
			oidc.EncryptionAlgECDHESA256KW,
			-1,
			"",
		},
		{
			"ShouldErrForUnknownKeyType",
			func(t *testing.T) any { return "not a key" },
			"",
			"",
			0,
			"the key type 'string' is unknown or not valid for the configuration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jwk := schema.JWK{Key: tc.key(t)}

			props, err := schemaJWKGetPropertiesEnc(jwk)

			if tc.err != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, props)
			} else {
				assert.NoError(t, err)

				if tc.expectedUse == "" && tc.expectedAlg == "" && tc.expectedBts == 0 {
					if props != nil {
						assert.Equal(t, "", props.Use)
						assert.Equal(t, "", props.Algorithm)
					}
				} else {
					require.NotNil(t, props)
					assert.Equal(t, tc.expectedUse, props.Use)
					assert.Equal(t, tc.expectedAlg, props.Algorithm)
					assert.Equal(t, tc.expectedBts, props.Bits)
				}
			}
		})
	}
}
