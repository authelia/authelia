package oidc

import (
	"context"
	"crypto"
	"sort"

	"github.com/go-jose/go-jose/v4"

	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewIssuer returns a new *Issuer given the provided JSON Web Keys.
func NewIssuer(keys []schema.JWK) (issuer *Issuer) {
	return &Issuer{jwks: NewJSONWebKeySet(keys), kid: NewIssuerDefaultKeyID(keys)}
}

// NewIssuerDefaultKeyID returns the key id of the default signing key in the provided JSON Web Keys.
func NewIssuerDefaultKeyID(keys []schema.JWK) (kid string) {
	for _, key := range keys {
		if key.Use != KeyUseSignature || key.Algorithm != SigningAlgRSAUsingSHA256 {
			continue
		}

		return key.KeyID
	}

	return ""
}

// NewJSONWebKeySet returns a *jose.JSONWebKeySet given the provided JSON Web Keys.
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

// NewJSONWebKeySetPublic returns a *jose.JSONWebKeySet with only the public portion of the provided JSON Web Keys.
func NewJSONWebKeySetPublic(jwks []schema.JWK) (jwkSet *jose.JSONWebKeySet) {
	keys := make([]jose.JSONWebKey, len(jwks))

	for i, jwk := range jwks {
		k := NewJSONWebKey(jwk)

		keys[i] = k.Public()
	}

	sort.Sort(SortedJSONWebKey(keys))

	return &jose.JSONWebKeySet{Keys: keys}
}

// NewJSONWebKey returns a jose.JSONWebKey given the provided JSON Web Key.
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

// Issuer holds the JSON Web Key Set used to issue tokens along with the default key id.
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

// GetPublicJSONWebKeys returns the public portion of the JSON Web Key Set.
func (i *Issuer) GetPublicJSONWebKeys(ctx Context) (jwks *jose.JSONWebKeySet) {
	keys := make([]jose.JSONWebKey, len(i.jwks.Keys))

	for j, jwk := range i.jwks.Keys {
		keys[j] = jwk.Public()
	}

	return &jose.JSONWebKeySet{
		Keys: keys,
	}
}

// GetIssuerJWK returns the JSON Web Key which matches the given kid, alg, and use.
func (i *Issuer) GetIssuerJWK(ctx context.Context, kid, alg, use string) (jwk *jose.JSONWebKey, err error) {
	return jwt.SearchJWKS(i.jwks, kid, alg, use, false)
}

// GetIssuerStrictJWK returns the JSON Web Key which strictly matches the given kid, alg, and use.
func (i *Issuer) GetIssuerStrictJWK(ctx context.Context, kid, alg, use string) (jwk *jose.JSONWebKey, err error) {
	return jwt.SearchJWKS(i.jwks, kid, alg, use, true)
}
