package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	fjwt "github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestKeyManager(t *testing.T) {
	config := &schema.OpenIDConnectConfiguration{
		Discovery: schema.OpenIDConnectDiscovery{
			DefaultKeyID: "kid-RS256-sig",
		},
		IssuerJWKS: []schema.JWK{
			{
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAUsingSHA256,
				Key:              keyRSA2048,
				CertificateChain: certRSA2048,
			},
			{
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAUsingSHA384,
				Key:              keyRSA2048,
				CertificateChain: certRSA2048,
			},
			{
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAUsingSHA512,
				Key:              keyRSA4096,
				CertificateChain: certRSA4096,
			},
			{
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAPSSUsingSHA256,
				Key:              keyRSA2048,
				CertificateChain: certRSA2048,
			},
			{
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAPSSUsingSHA384,
				Key:              keyRSA2048,
				CertificateChain: certRSA2048,
			},
			{
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAPSSUsingSHA512,
				Key:              keyRSA4096,
				CertificateChain: certRSA4096,
			},
			{
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgECDSAUsingP256AndSHA256,
				Key:              keyECDSAP256,
				CertificateChain: certECDSAP256,
			},
			{
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgECDSAUsingP384AndSHA384,
				Key:              keyECDSAP384,
				CertificateChain: certECDSAP384,
			},
			{
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgECDSAUsingP521AndSHA512,
				Key:              keyECDSAP521,
				CertificateChain: certECDSAP521,
			},
		},
	}

	for i, key := range config.IssuerJWKS {
		config.IssuerJWKS[i].KeyID = fmt.Sprintf("kid-%s-%s", key.Algorithm, key.Use)
	}

	manager := NewKeyManager(config)

	assert.NotNil(t, manager)

	assert.Len(t, manager.kids, len(config.IssuerJWKS))
	assert.Len(t, manager.algs, len(config.IssuerJWKS))

	assert.Equal(t, "kid-RS256-sig", manager.kid)

	ctx := context.Background()

	var (
		jwk *JWK
		err error
	)

	jwk = manager.GetByAlg(ctx, "notalg")
	assert.Nil(t, jwk)

	jwk = manager.GetByKID(ctx, "notalg")
	assert.Nil(t, jwk)

	jwk = manager.GetByKID(ctx, "")
	assert.NotNil(t, jwk)
	assert.Equal(t, config.Discovery.DefaultKeyID, jwk.KeyID())

	jwk, err = manager.GetByHeader(ctx, &fjwt.Headers{Extra: map[string]any{JWTHeaderKeyIdentifier: "notalg"}})
	assert.EqualError(t, err, "jwt header 'kid' with value 'notalg' does not match a managed jwk")
	assert.Nil(t, jwk)

	jwk, err = manager.GetByHeader(ctx, &fjwt.Headers{Extra: map[string]any{}})
	assert.EqualError(t, err, "jwt header did not have a kid")
	assert.Nil(t, jwk)

	jwk, err = manager.GetByHeader(ctx, nil)
	assert.EqualError(t, err, "jwt header was nil")
	assert.Nil(t, jwk)

	kid, err := manager.GetKIDFromAlgStrict(ctx, "notalg")
	assert.EqualError(t, err, "alg not found")
	assert.Equal(t, "", kid)

	kid = manager.GetKIDFromAlg(ctx, "notalg")
	assert.Equal(t, config.Discovery.DefaultKeyID, kid)

	set := manager.Set(ctx)

	assert.NotNil(t, set)
	assert.Len(t, set.Keys, len(config.IssuerJWKS))

	data, err := json.Marshal(&set)
	assert.NoError(t, err)
	assert.NotNil(t, data)

	out := jose.JSONWebKeySet{}
	assert.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, *set, out)

	for _, alg := range []string{SigningAlgRSAUsingSHA256, SigningAlgRSAUsingSHA384, SigningAlgRSAPSSUsingSHA512, SigningAlgRSAPSSUsingSHA256, SigningAlgRSAPSSUsingSHA384, SigningAlgRSAPSSUsingSHA512, SigningAlgECDSAUsingP256AndSHA256, SigningAlgECDSAUsingP384AndSHA384, SigningAlgECDSAUsingP521AndSHA512} {
		t.Run(alg, func(t *testing.T) {
			expectedKID := fmt.Sprintf("kid-%s-%s", alg, KeyUseSignature)

			t.Run("ShouldGetCorrectKey", func(t *testing.T) {

				jwk = manager.GetByKID(ctx, expectedKID)
				assert.NotNil(t, jwk)
				assert.Equal(t, expectedKID, jwk.KeyID())

				jwk = manager.GetByAlg(ctx, alg)
				assert.NotNil(t, jwk)

				assert.Equal(t, alg, jwk.alg.Alg())
				assert.Equal(t, expectedKID, jwk.KeyID())

				kid, err = manager.GetKIDFromAlgStrict(ctx, alg)
				assert.NoError(t, err)
				assert.Equal(t, expectedKID, kid)

				kid = manager.GetKIDFromAlg(ctx, alg)
				assert.Equal(t, expectedKID, kid)

				jwk, err = manager.GetByHeader(ctx, &fjwt.Headers{Extra: map[string]any{JWTHeaderKeyIdentifier: expectedKID}})
				assert.NoError(t, err)
				assert.NotNil(t, jwk)

				assert.Equal(t, expectedKID, jwk.KeyID())
			})

			t.Run("ShouldUseCorrectSigner", func(t *testing.T) {
				var tokenString, sig, sigb string
				var token *fjwt.Token

				tokenString, sig, err = manager.Generate(ctx, fjwt.MapClaims{}, &fjwt.Headers{Extra: map[string]any{JWTHeaderKeyIdentifier: expectedKID}})
				assert.NoError(t, err)

				sigb, err = manager.GetSignature(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, sig, sigb)

				sigb, err = manager.Validate(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, sig, sigb)

				token, err = manager.Decode(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, expectedKID, token.Header[JWTHeaderKeyIdentifier])

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
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAUsingSHA256,
				Key:              keyRSA2048,
				CertificateChain: certRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa2048-rs384",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAUsingSHA384,
				Key:              keyRSA2048,
				CertificateChain: certRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa2048-rs512",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAUsingSHA512,
				Key:              keyRSA2048,
				CertificateChain: certRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-rs256",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAUsingSHA256,
				Key:              keyRSA4096,
				CertificateChain: certRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-rs384",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAUsingSHA384,
				Key:              keyRSA4096,
				CertificateChain: certRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-rs512",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAUsingSHA512,
				Key:              keyRSA4096,
				CertificateChain: certRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa2048-rs256",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAPSSUsingSHA256,
				Key:              keyRSA2048,
				CertificateChain: certRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa2048-ps384",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAPSSUsingSHA384,
				Key:              keyRSA2048,
				CertificateChain: certRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa2048-ps512",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAPSSUsingSHA512,
				Key:              keyRSA2048,
				CertificateChain: certRSA2048,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-ps256",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAPSSUsingSHA256,
				Key:              keyRSA4096,
				CertificateChain: certRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-ps384",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAPSSUsingSHA384,
				Key:              keyRSA4096,
				CertificateChain: certRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "rsa4096-ps512",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgRSAPSSUsingSHA512,
				Key:              keyRSA4096,
				CertificateChain: certRSA4096,
			},
		},
		{
			schema.JWK{
				KeyID:            "ecdsaP256",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgECDSAUsingP256AndSHA256,
				Key:              keyECDSAP256,
				CertificateChain: certECDSAP256,
			},
		},
		{
			schema.JWK{
				KeyID:            "ecdsaP384",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgECDSAUsingP384AndSHA384,
				Key:              keyECDSAP384,
				CertificateChain: certECDSAP384,
			},
		},
		{
			schema.JWK{
				KeyID:            "ecdsaP521",
				Use:              KeyUseSignature,
				Algorithm:        SigningAlgECDSAUsingP521AndSHA512,
				Key:              keyECDSAP521,
				CertificateChain: certECDSAP521,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.have.KeyID, func(t *testing.T) {
			t.Run("Generating", func(t *testing.T) {
				var (
					jwk *JWK
				)

				ctx := context.Background()

				jwk = NewJWK(tc.have)

				signer := jwk.Strategy()

				claims := fjwt.MapClaims{}
				header := &fjwt.Headers{
					Extra: map[string]any{
						"kid": jwk.kid,
					},
				}

				tokenString, sig, err := signer.Generate(ctx, claims, header)

				assert.NoError(t, err)
				assert.NotEqual(t, "", tokenString)
				assert.NotEqual(t, "", sig)

				sigd, err := signer.GetSignature(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, sig, sigd)

				token, err := signer.Decode(ctx, tokenString)
				assert.NoError(t, err)
				assert.NotNil(t, token)
				fmt.Println(tokenString)

				assert.True(t, token.Valid())
				assert.Equal(t, jwk.alg.Alg(), string(token.Method))

				sigv, err := signer.Validate(ctx, tokenString)
				assert.NoError(t, err)
				assert.Equal(t, sig, sigv)
			})

			t.Run("Marshalling", func(t *testing.T) {
				var (
					jwk  *JWK
					out  jose.JSONWebKey
					data []byte
					err  error
				)

				jwk = NewJWK(tc.have)

				strategy := jwk.Strategy()

				assert.NotNil(t, strategy)

				signer, ok := strategy.(*Signer)

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
