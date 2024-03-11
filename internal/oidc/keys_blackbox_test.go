package oidc_test

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"testing"

	fjwt "authelia.com/provider/oauth2/token/jwt"
	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestKeyManager(t *testing.T) {
	config := &schema.IdentityProvidersOpenIDConnect{
		JSONWebKeys: []schema.JWK{
			{
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAUsingSHA256,
				Key:              x509PrivateKeyRSA2048,
				CertificateChain: x509CertificateChainRSA2048,
			},
			{
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAUsingSHA384,
				Key:              x509PrivateKeyRSA2048,
				CertificateChain: x509CertificateChainRSA2048,
			},
			{
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAUsingSHA512,
				Key:              x509PrivateKeyRSA4096,
				CertificateChain: x509CertificateChainRSA4096,
			},
			{
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAPSSUsingSHA256,
				Key:              x509PrivateKeyRSA2048,
				CertificateChain: x509CertificateChainRSA2048,
			},
			{
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAPSSUsingSHA384,
				Key:              x509PrivateKeyRSA2048,
				CertificateChain: x509CertificateChainRSA2048,
			},
			{
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAPSSUsingSHA512,
				Key:              x509PrivateKeyRSA4096,
				CertificateChain: x509CertificateChainRSA4096,
			},
			{
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgECDSAUsingP256AndSHA256,
				Key:              x509PrivateKeyECDSAP256,
				CertificateChain: x509CertificateChainECDSAP256,
			},
			{
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgECDSAUsingP384AndSHA384,
				Key:              x509PrivateKeyECDSAP384,
				CertificateChain: x509CertificateChainECDSAP384,
			},
			{
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgECDSAUsingP521AndSHA512,
				Key:              x509PrivateKeyECDSAP521,
				CertificateChain: x509CertificateChainECDSAP521,
			},
		},
	}

	config.Discovery.DefaultKeyIDs = map[string]string{}

	for i, key := range config.JSONWebKeys {
		kid := fmt.Sprintf("kid-%s-%s", key.Algorithm, key.Use)

		config.JSONWebKeys[i].KeyID = kid

		if _, ok := config.Discovery.DefaultKeyIDs[key.Algorithm]; !ok {
			config.Discovery.DefaultKeyIDs[key.Algorithm] = kid
		}
	}

	manager := oidc.NewKeyManager(config)

	assert.NotNil(t, manager)

	ctx := context.Background()

	assert.Equal(t, "kid-RS256-sig", manager.GetDefaultKeyID(ctx))

	require.NotNil(t, manager.Get(ctx, "kid-RS256-sig", oidc.SigningAlgRSAUsingSHA256))
	assert.Equal(t, "kid-RS256-sig", manager.Get(ctx, "kid-RS256-sig", oidc.SigningAlgRSAUsingSHA256).KeyID())
	assert.Equal(t, "kid-RS256-sig", manager.Get(ctx, "", oidc.SigningAlgRSAUsingSHA256).KeyID())
	assert.Nil(t, manager.Get(ctx, "", "NOKEY"))

	assert.Equal(t, "kid-RS256-sig", manager.GetKeyID(ctx, "", oidc.SigningAlgRSAUsingSHA256))
	assert.Equal(t, "kid-RS256-sig", manager.GetKeyID(ctx, "kid-RS256-sig", oidc.SigningAlgRSAPSSUsingSHA256))
	assert.Equal(t, "kid-RS256-sig", manager.GetKeyID(ctx, "", ""))
	assert.Equal(t, "kid-PS256-sig", manager.GetKeyID(ctx, "kid-PS256-sig", oidc.SigningAlgRSAPSSUsingSHA256))
	assert.Equal(t, "kid-PS256-sig", manager.GetKeyID(ctx, "", oidc.SigningAlgRSAPSSUsingSHA256))

	var (
		jwk              *oidc.JWK
		tokenString, sig string
		sum              []byte
		token            *fjwt.Token
		err              error
	)

	jwk = manager.GetByAlg(ctx, "notalg")
	assert.Nil(t, jwk)

	jwk = manager.GetByKID(ctx, "notalg")
	assert.Nil(t, jwk)

	jwk = manager.GetByKID(ctx, "")
	assert.NotNil(t, jwk)
	assert.Equal(t, config.Discovery.DefaultKeyIDs[oidc.SigningAlgRSAUsingSHA256], jwk.KeyID())

	jwk, err = manager.GetByHeader(ctx, &fjwt.Headers{Extra: map[string]any{oidc.JWTHeaderKeyIdentifier: "notalg"}})
	assert.EqualError(t, err, "jwt header 'kid' with value 'notalg' does not match a managed jwk")
	assert.Nil(t, jwk)

	jwk, err = manager.GetByHeader(ctx, &fjwt.Headers{Extra: map[string]any{}})
	assert.EqualError(t, err, "jwt header did not match a known jwk")
	assert.Nil(t, jwk)

	jwk, err = manager.GetByHeader(ctx, &fjwt.Headers{Extra: map[string]any{oidc.JWTHeaderKeyIdentifier: "x"}})
	assert.EqualError(t, err, "jwt header 'kid' with value 'x' does not match a managed jwk")
	assert.Nil(t, jwk)

	jwk, err = manager.GetByHeader(ctx, &fjwt.Headers{Extra: map[string]any{oidc.JWTHeaderKeyAlgorithm: "x"}})
	assert.EqualError(t, err, "jwt header 'alg' with value 'x' does not match a managed jwk")
	assert.Nil(t, jwk)

	jwk, err = manager.GetByHeader(ctx, nil)
	assert.EqualError(t, err, "jwt header was nil")
	assert.Nil(t, jwk)

	kid, err := manager.GetKeyIDFromAlgStrict(ctx, "notalg")
	assert.EqualError(t, err, "alg not found")
	assert.Equal(t, "", kid)

	kid = manager.GetKeyIDFromAlg(ctx, "notalg")
	assert.Equal(t, config.Discovery.DefaultKeyIDs[oidc.SigningAlgRSAUsingSHA256], kid)

	set := manager.Set(ctx)

	assert.NotNil(t, set)
	assert.Len(t, set.Keys, len(config.JSONWebKeys))

	data, err := json.Marshal(&set)
	assert.NoError(t, err)
	assert.NotNil(t, data)

	out := jose.JSONWebKeySet{}
	assert.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, *set, out)

	jwk, err = manager.GetByTokenString(ctx, badTokenString)
	assert.EqualError(t, err, "token is malformed: token contains an invalid number of segments")
	assert.Nil(t, jwk)

	tokenString, sig, err = manager.Generate(ctx, nil, nil)
	assert.EqualError(t, err, "error getting jwk from header: jwt header was nil")
	assert.Equal(t, "", tokenString)
	assert.Equal(t, "", sig)

	sig, err = manager.Validate(ctx, badTokenString)
	assert.EqualError(t, err, "error getting jwk from token string: token is malformed: token contains an invalid number of segments")
	assert.Equal(t, "", sig)

	token, err = manager.Decode(ctx, badTokenString)
	assert.EqualError(t, err, "error getting jwk from token string: token is malformed: token contains an invalid number of segments")
	assert.Nil(t, token)

	sum, err = manager.Hash(ctx, []byte(abc))
	assert.NoError(t, err)
	assert.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", fmt.Sprintf("%x", sum))

	assert.Equal(t, crypto.SHA256.Size(), manager.GetSigningMethodLength(ctx))

	for _, alg := range []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAPSSUsingSHA512, oidc.SigningAlgRSAPSSUsingSHA256, oidc.SigningAlgRSAPSSUsingSHA384, oidc.SigningAlgRSAPSSUsingSHA512, oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgECDSAUsingP384AndSHA384, oidc.SigningAlgECDSAUsingP521AndSHA512} {
		t.Run(alg, func(t *testing.T) {
			expectedKID := fmt.Sprintf("kid-%s-%s", alg, oidc.KeyUseSignature)

			t.Run("ShouldGetCorrectKey", func(t *testing.T) {
				jwk = manager.GetByKID(ctx, expectedKID)
				assert.NotNil(t, jwk)
				assert.Equal(t, expectedKID, jwk.KeyID())

				jwk = manager.GetByAlg(ctx, alg)
				assert.NotNil(t, jwk)

				assert.Equal(t, alg, jwk.GetSigningMethod().Alg())
				assert.Equal(t, expectedKID, jwk.KeyID())

				kid, err = manager.GetKeyIDFromAlgStrict(ctx, alg)
				assert.NoError(t, err)
				assert.Equal(t, expectedKID, kid)

				kid = manager.GetKeyIDFromAlg(ctx, alg)
				assert.Equal(t, expectedKID, kid)

				jwk, err = manager.GetByHeader(ctx, &fjwt.Headers{Extra: map[string]any{oidc.JWTHeaderKeyIdentifier: expectedKID}})
				assert.NoError(t, err)
				assert.NotNil(t, jwk)

				assert.Equal(t, expectedKID, jwk.KeyID())
			})

			t.Run("ShouldUseCorrectSigner", func(t *testing.T) {
				var sigb string

				tokenString, sig, err = manager.Generate(ctx, fjwt.MapClaims{}, &fjwt.Headers{Extra: map[string]any{oidc.JWTHeaderKeyIdentifier: expectedKID}})
				assert.NoError(t, err)

				sigb, err = manager.GetSignature(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, sig, sigb)

				sigb, err = manager.Validate(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, sig, sigb)

				token, err = manager.Decode(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, expectedKID, token.Header[oidc.JWTHeaderKeyIdentifier])

				jwk, err = manager.GetByTokenString(ctx, tokenString)
				assert.NoError(t, err)

				sigb, err = jwk.Strategy().Validate(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, sig, sigb)
			})
		})
	}
}

func TestJWKFunctionality(t *testing.T) {
	testCases := []struct {
		have schema.JWK
	}{
		{
			schema.JWK{
				KeyID:            "rsa2048-rs256",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAUsingSHA256,
				Key:              x509PrivateKeyRSA2048,
				CertificateChain: x509CertificateChainRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa2048-rs384",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAUsingSHA384,
				Key:              x509PrivateKeyRSA2048,
				CertificateChain: x509CertificateChainRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa2048-rs512",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAUsingSHA512,
				Key:              x509PrivateKeyRSA2048,
				CertificateChain: x509CertificateChainRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-rs256",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAUsingSHA256,
				Key:              x509PrivateKeyRSA4096,
				CertificateChain: x509CertificateChainRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-rs384",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAUsingSHA384,
				Key:              x509PrivateKeyRSA4096,
				CertificateChain: x509CertificateChainRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-rs512",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAUsingSHA512,
				Key:              x509PrivateKeyRSA4096,
				CertificateChain: x509CertificateChainRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa2048-rs256",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAPSSUsingSHA256,
				Key:              x509PrivateKeyRSA2048,
				CertificateChain: x509CertificateChainRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa2048-ps384",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAPSSUsingSHA384,
				Key:              x509PrivateKeyRSA2048,
				CertificateChain: x509CertificateChainRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa2048-ps512",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAPSSUsingSHA512,
				Key:              x509PrivateKeyRSA2048,
				CertificateChain: x509CertificateChainRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-ps256",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAPSSUsingSHA256,
				Key:              x509PrivateKeyRSA4096,
				CertificateChain: x509CertificateChainRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-ps384",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAPSSUsingSHA384,
				Key:              x509PrivateKeyRSA4096,
				CertificateChain: x509CertificateChainRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-ps512",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAPSSUsingSHA512,
				Key:              x509PrivateKeyRSA4096,
				CertificateChain: x509CertificateChainRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "ecdsaP256",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgECDSAUsingP256AndSHA256,
				Key:              x509PrivateKeyECDSAP256,
				CertificateChain: x509CertificateChainECDSAP256,
			},
		},
		{
			schema.JWK{
				KeyID:            "ecdsaP384",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgECDSAUsingP384AndSHA384,
				Key:              x509PrivateKeyECDSAP384,
				CertificateChain: x509CertificateChainECDSAP384,
			},
		},
		{
			schema.JWK{
				KeyID:            "ecdsaP521",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgECDSAUsingP521AndSHA512,
				Key:              x509PrivateKeyECDSAP521,
				CertificateChain: x509CertificateChainECDSAP521,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.have.KeyID, func(t *testing.T) {
			t.Run("Generating", func(t *testing.T) {
				var (
					jwk *oidc.JWK
				)

				ctx := context.Background()

				jwk = oidc.NewJWK(tc.have)

				signer := jwk.Strategy()

				claims := fjwt.MapClaims{}
				header := &fjwt.Headers{
					Extra: map[string]any{
						oidc.JWTHeaderKeyIdentifier: jwk.KeyID(),
					},
				}

				tokenString, sig, err := signer.Generate(ctx, nil, nil)
				assert.EqualError(t, err, "either claims or header is nil")
				assert.Equal(t, "", tokenString)
				assert.Equal(t, "", sig)

				tokenString, sig, err = signer.Generate(ctx, claims, header)
				assert.NoError(t, err)
				assert.NotEqual(t, "", tokenString)
				assert.NotEqual(t, "", sig)

				sigd, err := signer.GetSignature(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, sig, sigd)

				token, err := signer.Decode(ctx, tokenString)
				assert.NoError(t, err)
				assert.NotNil(t, token)

				assert.True(t, token.Valid())
				assert.Equal(t, jwk.GetSigningMethod().Alg(), string(token.Method))

				sigv, err := signer.Validate(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, sig, sigv)
			})

			t.Run("Marshalling", func(t *testing.T) {
				var (
					jwk  *oidc.JWK
					out  jose.JSONWebKey
					data []byte
					err  error
				)

				jwk = oidc.NewJWK(tc.have)

				strategy := jwk.Strategy()

				assert.NotNil(t, strategy)

				signer, ok := strategy.(*oidc.Signer)

				require.True(t, ok)

				assert.NotNil(t, signer)

				key, err := signer.GetPublicKey(context.Background())
				assert.NoError(t, err)
				assert.NotNil(t, key)

				key, err = jwk.GetPrivateKey(context.Background())
				assert.NoError(t, err)
				assert.NotNil(t, key)

				data, err = json.Marshal(jwk.JWK())

				assert.NoError(t, err)
				require.NotNil(t, data)

				assert.NoError(t, json.Unmarshal(data, &out))

				assert.True(t, out.IsPublic())
				assert.Equal(t, tc.have.KeyID, out.KeyID)
				assert.Equal(t, tc.have.KeyID, jwk.KeyID())
				assert.Equal(t, tc.have.Use, out.Use)
				assert.Equal(t, tc.have.Algorithm, out.Algorithm)
				assert.NotNil(t, out.Key)
				assert.NotNil(t, out.Certificates)
				assert.NotNil(t, out.CertificateThumbprintSHA1)
				assert.NotNil(t, out.CertificateThumbprintSHA256)
				assert.True(t, out.Valid())

				data, err = json.Marshal(jwk.PrivateJWK())

				assert.NoError(t, err)
				require.NotNil(t, data)
				assert.NoError(t, json.Unmarshal(data, &out))

				assert.False(t, out.IsPublic())
				assert.Equal(t, tc.have.KeyID, out.KeyID)
				assert.Equal(t, tc.have.Use, out.Use)
				assert.Equal(t, tc.have.Algorithm, out.Algorithm)
				assert.NotNil(t, out.Key)
				assert.NotNil(t, out.Certificates)
				assert.NotNil(t, out.CertificateThumbprintSHA1)
				assert.NotNil(t, out.CertificateThumbprintSHA256)
				assert.True(t, out.Valid())
			})
		})
	}
}
