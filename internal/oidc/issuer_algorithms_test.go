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
