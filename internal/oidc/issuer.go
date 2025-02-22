package oidc

import (
	"context"
	"crypto"
	"sort"

	"authelia.com/provider/oauth2/token/jwt"
	"github.com/go-jose/go-jose/v4"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func NewIssuer(keys []schema.JWK) (issuer *Issuer) {
	return &Issuer{jwks: NewJSONWebKeySet(keys), kid: NewIssuerDefaultKeyID(keys)}
}

func NewIssuerDefaultKeyID(keys []schema.JWK) (kid string) {
	for _, key := range keys {
		if key.Use != KeyUseSignature || key.Algorithm != SigningAlgRSAUsingSHA256 {
			continue
		}

		return key.KeyID
	}

	return ""
}

func NewJSONWebKeySet(jwks []schema.JWK) (jwkSet *jose.JSONWebKeySet) {
	if len(jwks) == 0 {
		return nil
	}

	keys := make([]jose.JSONWebKey, len(jwks))

	for i, jwk := range jwks {
		keys[i] = NewJSONWebKey(jwk)
	}

	sort.Sort(SortedJSONWebKey(keys))

	return &jose.JSONWebKeySet{Keys: keys}
}

func NewJSONWebKeySetPublic(jwks []schema.JWK) (jwkSet *jose.JSONWebKeySet) {
	keys := make([]jose.JSONWebKey, len(jwks))

	for i, jwk := range jwks {
		k := NewJSONWebKey(jwk)

		keys[i] = k.Public()
	}

	sort.Sort(SortedJSONWebKey(keys))

	return &jose.JSONWebKeySet{Keys: keys}
}

func NewJSONWebKey(key schema.JWK) (jwk jose.JSONWebKey) {
	jwk = jose.JSONWebKey{
		Key:                         key.Key,
		KeyID:                       key.KeyID,
		Algorithm:                   key.Algorithm,
		Use:                         key.Use,
		Certificates:                key.CertificateChain.Certificates(),
		CertificateThumbprintSHA256: key.CertificateChain.Thumbprint(crypto.SHA256),
		CertificateThumbprintSHA1:   key.CertificateChain.Thumbprint(crypto.SHA1),
	}

	return jwk
}

type Issuer struct {
	kid  string
	jwks *jose.JSONWebKeySet
}

// GetKeyID returns the JWK Key ID given an kid/alg or the default if it doesn't exist.
func (i *Issuer) GetKeyID(ctx context.Context, kid, alg string) string {
	if jwk, err := i.GetIssuerStrictJWK(ctx, kid, alg, KeyUseSignature); err == nil {
		return jwk.KeyID
	}

	return i.kid
}

func (i *Issuer) GetPublicJSONWebKeys(ctx Context) (jwks *jose.JSONWebKeySet) {
	keys := make([]jose.JSONWebKey, len(i.jwks.Keys))

	for j, jwk := range i.jwks.Keys {
		keys[j] = jwk.Public()
	}

	return &jose.JSONWebKeySet{
		Keys: keys,
	}
}

func (i *Issuer) GetIssuerJWK(ctx context.Context, kid, alg, use string) (jwk *jose.JSONWebKey, err error) {
	return jwt.SearchJWKS(i.jwks, kid, alg, use, false)
}

func (i *Issuer) GetIssuerStrictJWK(ctx context.Context, kid, alg, use string) (jwk *jose.JSONWebKey, err error) {
	return jwt.SearchJWKS(i.jwks, kid, alg, use, true)
}
