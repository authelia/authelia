//go:build go1.27

package oidc_test

import (
	"crypto/mldsa"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestIssuerSignsMLDSA(t *testing.T) {
	testCases := []struct {
		name   string
		params mldsa.Parameters
		alg    string
	}{
		{"ML-DSA-44", mldsa.MLDSA44(), oidc.SigningAlgMLDSA44},
		{"ML-DSA-65", mldsa.MLDSA65(), oidc.SigningAlgMLDSA65},
		{"ML-DSA-87", mldsa.MLDSA87(), oidc.SigningAlgMLDSA87},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := mldsa.GenerateKey(tc.params)

			require.NoError(t, err)

			alg, ok := oidc.SigningAlgFromMLDSAKey(key)

			require.True(t, ok)
			require.Equal(t, tc.alg, alg)

			assertIssuerSigns(t, key, tc.alg)
		})
	}
}
