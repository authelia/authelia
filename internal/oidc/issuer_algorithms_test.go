package oidc_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"authelia.com/provider/oauth2/token/jose"
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

			parsed, err := jose.ParseSigned(token, []jose.SignatureAlgorithm{jose.SignatureAlgorithm(tc.request)})

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

func assertIssuerSigns(t *testing.T, key schema.CryptographicKey, alg string) {
	t.Helper()

	issuer := oidc.NewIssuer([]schema.JWK{{KeyID: "example", Use: oidc.KeyUseSignature, Algorithm: alg, Key: key}})

	jwk, err := issuer.GetIssuerJWK(context.Background(), "example", alg, oidc.KeyUseSignature)

	require.NoError(t, err)
	assert.Equal(t, alg, jwk.Algorithm)

	token, _, err := jwt.EncodeCompactSigned(context.Background(), jwt.JWTClaims{Subject: "john"}.ToMapClaims(), &jwt.Headers{Extra: map[string]any{}}, jwk)

	require.NoError(t, err)

	parsed, err := jose.ParseSigned(token, []jose.SignatureAlgorithm{jose.SignatureAlgorithm(alg)})

	require.NoError(t, err)

	payload, err := parsed.Verify(jwk.Public())

	require.NoError(t, err)
	assert.Contains(t, string(payload), `"sub":"john"`)
}
