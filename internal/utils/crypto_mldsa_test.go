//go:build go1.27

package utils

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMLDSA(t *testing.T) {
	testCases := []struct {
		name        string
		have        string
		expected    string
		expectedSig x509.SignatureAlgorithm
		err         string
	}{
		{"ShouldHandleFullName44", "ML-DSA-44", KeyMLDSAParameters44, x509.MLDSA44, ""},
		{"ShouldHandleFullName65", "ML-DSA-65", KeyMLDSAParameters65, x509.MLDSA65, ""},
		{"ShouldHandleFullName87", "ML-DSA-87", KeyMLDSAParameters87, x509.MLDSA87, ""},
		{"ShouldHandleBareLevel", "44", KeyMLDSAParameters44, x509.MLDSA44, ""},
		{"ShouldHandleLowerCase", "ml-dsa-65", KeyMLDSAParameters65, x509.MLDSA65, ""},
		{"ShouldErrorOnUnknown", "ML-DSA-99", "", x509.UnknownSignatureAlgorithm, "invalid parameters 'ML-DSA-99' were specified: parameters must be 'ML-DSA-44', 'ML-DSA-65', or 'ML-DSA-87'"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedSig, MLDSASignatureAlgorithmFromString(tc.have))

			key, err := GenerateMLDSAKey(tc.have)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, key)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, key)

			assert.True(t, IsMLDSAPrivateKey(key))
			assert.True(t, IsX509PrivateKey(key))

			parameters, ok := MLDSAParameterSetFromKey(key)

			require.True(t, ok)
			assert.Equal(t, tc.expected, parameters)

			public := PublicKeyFromPrivateKey(key)

			require.NotNil(t, public)
			assert.True(t, IsMLDSAPublicKey(public))

			blockPrivate, err := PEMBlockFromX509Key(key, false)

			require.NoError(t, err)
			assert.Equal(t, BlockTypePKCS8PrivateKey, blockPrivate.Type)

			blockPublic, err := PEMBlockFromX509Key(public, false)

			require.NoError(t, err)
			assert.Equal(t, BlockTypePKIXPublicKey, blockPublic.Type)

			keyAlg, sigAlg := KeySigAlgorithmFromString(KeyAlgorithmMLDSA, tc.have)

			assert.Equal(t, x509.MLDSA, keyAlg)
			assert.Equal(t, tc.expectedSig, sigAlg)
		})
	}
}
