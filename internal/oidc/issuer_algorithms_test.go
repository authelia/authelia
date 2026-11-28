// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"authelia.com/provider/jose"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestIssuerSignsEd25519(t *testing.T) {
	_, key, err := ed25519.GenerateKey(rand.Reader)

	require.NoError(t, err)

	assertIssuerSigns(t, key, oidc.SigningAlgEd25519)
}

func TestIssuerSignsEdDSA(t *testing.T) {
	_, key, err := ed25519.GenerateKey(rand.Reader)

	require.NoError(t, err)

	assertIssuerSigns(t, key, oidc.SigningAlgEdDSA)
}

func TestIssuerEdwardsAlgorithmPair(t *testing.T) {
	testCases := []struct {
		name      string
		configure string
		request   string
	}{
		{name: "ShouldResolveEdDSARequestWithEd25519Key", configure: oidc.SigningAlgEd25519, request: oidc.SigningAlgEdDSA},
		{name: "ShouldResolveEd25519RequestWithEdDSAKey", configure: oidc.SigningAlgEdDSA, request: oidc.SigningAlgEd25519},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, key, err := ed25519.GenerateKey(rand.Reader)

			require.NoError(t, err)

			issuer := oidc.NewIssuer([]schema.JWK{{KeyID: "example", Use: oidc.KeyUseSignature, Algorithm: tc.configure, Key: key}})

			assert.Equal(t, "example", issuer.GetKeyID(context.Background(), "", tc.request))

			jwk, err := issuer.GetIssuerJWK(context.Background(), "example", tc.request, oidc.KeyUseSignature)

			require.NoError(t, err)
			assert.Equal(t, tc.request, jwk.Algorithm)
			assert.Equal(t, "example", jwk.KeyID)

			token, _, err := jwt.EncodeCompactSigned(context.Background(), jwt.JWTClaims{Subject: "john"}.ToMapClaims(), &jwt.Headers{Extra: map[string]any{}}, jwk)

			require.NoError(t, err)

			parsed, err := jose.ParseSignedCompact(token, []jose.SignatureAlgorithm{jose.SignatureAlgorithm(tc.request)})

			require.NoError(t, err)
			assert.Equal(t, tc.request, parsed.Signatures[0].Header.Algorithm)

			payload, err := parsed.Verify(jwk.Public())

			require.NoError(t, err)
			assert.Contains(t, string(payload), `"sub":"john"`)
		})
	}
}

func TestIssuerEdwardsAlgorithmPairShouldNotResolveUnrelatedAlgorithms(t *testing.T) {
	_, key, err := ed25519.GenerateKey(rand.Reader)

	require.NoError(t, err)

	issuer := oidc.NewIssuer([]schema.JWK{{KeyID: "example", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgEd25519, Key: key}})

	_, err = issuer.GetIssuerJWK(context.Background(), "example", oidc.SigningAlgRSAUsingSHA256, oidc.KeyUseSignature)

	assert.EqualError(t, err, "Error occurred retrieving the JSON Web Key. Unable to find JSON web key with kid 'example', use 'sig', and alg 'RS256' in JSON Web Key Set")
}

func TestIssuerEdwardsAlgorithmPairLookup(t *testing.T) {
	_, key, err := ed25519.GenerateKey(rand.Reader)

	require.NoError(t, err)

	ambiguous := []schema.JWK{
		{KeyID: "eddsa-1", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgEdDSA, Key: key},
		{KeyID: "eddsa-2", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgEdDSA, Key: key},
		{KeyID: "ed25519", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgEd25519, Key: key},
	}

	paired := []schema.JWK{
		{KeyID: "ed25519", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgEd25519, Key: key},
	}

	testCases := []struct {
		name     string
		keys     []schema.JWK
		kid      string
		alg      string
		strict   bool
		expected string
		err      string
	}{
		{
			name:   "ShouldErrStrictAmbiguousWithPairPresent",
			keys:   ambiguous,
			alg:    oidc.SigningAlgEdDSA,
			strict: true,
			err:    "Error occurred retrieving the JSON Web Key. Unable to find JSON web key with kid '', use 'sig', and alg 'EdDSA' in JSON Web Key Set",
		},
		{
			name:     "ShouldReturnFirstExactKeyWithPairPresent",
			keys:     ambiguous,
			alg:      oidc.SigningAlgEdDSA,
			expected: "eddsa-1",
		},
		{
			name:     "ShouldReturnExactKeyStrictByKeyIDWithPairPresent",
			keys:     ambiguous,
			kid:      "eddsa-2",
			alg:      oidc.SigningAlgEdDSA,
			strict:   true,
			expected: "eddsa-2",
		},
		{
			name:     "ShouldReturnExactPairedKeyStrict",
			keys:     ambiguous,
			alg:      oidc.SigningAlgEd25519,
			strict:   true,
			expected: "ed25519",
		},
		{
			name:     "ShouldReturnPairedKeyByKeyIDStrict",
			keys:     ambiguous,
			kid:      "ed25519",
			alg:      oidc.SigningAlgEdDSA,
			strict:   true,
			expected: "ed25519",
		},
		{
			name:     "ShouldFallBackToPairWithoutExactKey",
			keys:     paired,
			alg:      oidc.SigningAlgEdDSA,
			expected: "ed25519",
		},
		{
			name:     "ShouldFallBackToPairWithoutExactKeyStrict",
			keys:     paired,
			alg:      oidc.SigningAlgEdDSA,
			strict:   true,
			expected: "ed25519",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issuer := oidc.NewIssuer(tc.keys)

			lookup := issuer.GetIssuerJWK
			if tc.strict {
				lookup = issuer.GetIssuerStrictJWK
			}

			jwk, err := lookup(context.Background(), tc.kid, tc.alg, oidc.KeyUseSignature)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, jwk)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, jwk.KeyID)
			assert.Equal(t, tc.alg, jwk.Algorithm)
		})
	}
}

func TestIssuerAlgValues(t *testing.T) {
	_, edwards, err := ed25519.GenerateKey(rand.Reader)

	require.NoError(t, err)

	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	require.NoError(t, err)

	testCases := []struct {
		name       string
		keys       []schema.JWK
		signing    []string
		decryption []string
	}{
		{
			name:    "ShouldReportBothForEdDSAKey",
			keys:    []schema.JWK{{KeyID: "eddsa", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgEdDSA, Key: edwards}},
			signing: []string{oidc.SigningAlgEd25519, oidc.SigningAlgEdDSA},
		},
		{
			name:    "ShouldReportBothForEd25519Key",
			keys:    []schema.JWK{{KeyID: "ed25519", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgEd25519, Key: edwards}},
			signing: []string{oidc.SigningAlgEd25519, oidc.SigningAlgEdDSA},
		},
		{
			name: "ShouldReportEachOnceForBothKeys",
			keys: []schema.JWK{
				{KeyID: "eddsa", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgEdDSA, Key: edwards},
				{KeyID: "ed25519", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgEd25519, Key: edwards},
			},
			signing: []string{oidc.SigningAlgEd25519, oidc.SigningAlgEdDSA},
		},
		{
			name: "ShouldReportSortedWithOtherKeys",
			keys: []schema.JWK{
				{KeyID: "ed25519", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgEd25519, Key: edwards},
				{KeyID: "es256", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgECDSAUsingP256AndSHA256, Key: ec},
				{KeyID: "ecdh", Use: oidc.KeyUseEncryption, Algorithm: oidc.EncryptionAlgECDHES, Key: ec},
			},
			signing:    []string{oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgEd25519, oidc.SigningAlgEdDSA},
			decryption: []string{oidc.EncryptionAlgECDHES},
		},
		{
			name: "ShouldNotReportPairWithoutEdwardsKey",
			keys: []schema.JWK{
				{KeyID: "es256", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgECDSAUsingP256AndSHA256, Key: ec},
				{KeyID: "ecdh", Use: oidc.KeyUseEncryption, Algorithm: oidc.EncryptionAlgECDHES, Key: ec},
			},
			signing:    []string{oidc.SigningAlgECDSAUsingP256AndSHA256},
			decryption: []string{oidc.EncryptionAlgECDHES},
		},
		{
			name:       "ShouldNotReportPairForEncryptionKey",
			keys:       []schema.JWK{{KeyID: "enc", Use: oidc.KeyUseEncryption, Algorithm: oidc.SigningAlgEdDSA, Key: edwards}},
			decryption: []string{oidc.SigningAlgEdDSA},
		},
		{
			name: "ShouldReportNoneWithoutKeys",
		},
	}

	candidates := []string{oidc.SigningAlgEd25519, oidc.SigningAlgEdDSA, oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgRSAUsingSHA256}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issuer := oidc.NewIssuer(tc.keys)

			assert.Equal(t, tc.signing, issuer.GetIssuerSigningAlgValues(context.Background()))
			assert.Equal(t, tc.decryption, issuer.GetIssuerDecryptionAlgValues(context.Background()))

			if len(tc.keys) == 0 {
				return
			}

			// Reported and resolvable must be the same set.
			for _, alg := range candidates {
				_, err := issuer.GetIssuerJWK(context.Background(), "", alg, oidc.KeyUseSignature)

				if slices.Contains(tc.signing, alg) {
					assert.NoError(t, err, alg)
				} else {
					assert.Error(t, err, alg)
				}
			}
		})
	}
}

func assertIssuerSigns(t *testing.T, key schema.CryptographicKey, alg string) {
	t.Helper()

	issuer := oidc.NewIssuer([]schema.JWK{{KeyID: "example", Use: oidc.KeyUseSignature, Algorithm: alg, Key: key}})

	jwk, err := issuer.GetIssuerJWK(context.Background(), "example", alg, oidc.KeyUseSignature)

	require.NoError(t, err)
	assert.Equal(t, alg, jwk.Algorithm)

	token, _, err := jwt.EncodeCompactSigned(context.Background(), jwt.JWTClaims{Subject: "john"}.ToMapClaims(), &jwt.Headers{Extra: map[string]any{}}, jwk)

	require.NoError(t, err)

	parsed, err := jose.ParseSignedCompact(token, []jose.SignatureAlgorithm{jose.SignatureAlgorithm(alg)})

	require.NoError(t, err)

	payload, err := parsed.Verify(jwk.Public())

	require.NoError(t, err)
	assert.Contains(t, string(payload), `"sub":"john"`)
}
