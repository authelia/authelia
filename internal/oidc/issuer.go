// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"crypto"
	"slices"
	"sort"

	"authelia.com/provider/jose"
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

// GetIssuerJWK returns the JSON Web Key which matches the alg and use values, and the kid value when it is not empty.
// When more than one key matches the first is returned. The identifier named by the alg value is preferred, and when
// no key matches it a key configured with the paired Edwards-curve identifier is returned with the alg value
// requested, as EdDSA from RFC 8037 Section 3.1 and Ed25519 from RFC 9864 Section 2.2 name the same algorithm for an
// Ed25519 key.
func (i *Issuer) GetIssuerJWK(ctx context.Context, kid, alg, use string) (jwk *jose.JSONWebKey, err error) {
	return i.searchJWKS(kid, alg, use, false)
}

// GetIssuerStrictJWK returns the JSON Web Key which matches the alg and use values, and the kid value when it is not
// empty, with the same Edwards-curve identifier resolution as GetIssuerJWK. It differs in that a lookup which more
// than one key matches returns an error, so an alg value reported by GetIssuerSigningAlgValues may still return an
// error here. The paired identifier is only considered when no key matches the alg value, so this ambiguity is never
// resolved through it.
func (i *Issuer) GetIssuerStrictJWK(ctx context.Context, kid, alg, use string) (jwk *jose.JSONWebKey, err error) {
	return i.searchJWKS(kid, alg, use, true)
}

// GetIssuerSigningAlgValues returns the sorted unique algorithms of the keys with a use of sig, excluding none. A key
// configured with either of the EdDSA and Ed25519 identifiers reports both, per RFC 8037 Section 3.1 and RFC 9864
// Section 2.2. Every value reported resolves with GetIssuerJWK given an empty kid and a use of sig, and every alg
// value which resolves that way is reported.
func (i *Issuer) GetIssuerSigningAlgValues(ctx context.Context) (algs []string) {
	return i.keyAlgValues(KeyUseSignature)
}

// GetIssuerDecryptionAlgValues returns the sorted unique algorithms of the keys with a use of enc. No identifier is
// paired for these keys, so the values are exactly those which resolve with GetIssuerJWK given an empty kid and a use
// of enc.
func (i *Issuer) GetIssuerDecryptionAlgValues(ctx context.Context) (algs []string) {
	return i.keyAlgValues(KeyUseEncryption)
}

func (i *Issuer) keyAlgValues(use string) (algs []string) {
	if i == nil || i.jwks == nil {
		return nil
	}

	for _, key := range i.jwks.Keys {
		if key.Use != use || key.Algorithm == "" || key.Algorithm == SigningAlgNone {
			continue
		}

		algs = append(algs, key.Algorithm)

		if use != KeyUseSignature {
			continue
		}

		if paired, ok := SigningAlgEdwardsPair(key.Algorithm); ok {
			algs = append(algs, paired)
		}
	}

	slices.Sort(algs)

	return slices.Compact(algs)
}

func (i *Issuer) searchJWKS(kid, alg, use string, strict bool) (jwk *jose.JSONWebKey, err error) {
	if jwk, err = jwt.SearchJWKS(i.jwks, kid, alg, use, strict); err == nil {
		return jwk, nil
	}

	paired, ok := SigningAlgEdwardsPair(alg)
	if !ok || i.hasKey(kid, alg, use) {
		return nil, err
	}

	pairedJWK, pairedErr := jwt.SearchJWKS(i.jwks, kid, paired, use, strict)
	if pairedErr != nil {
		return nil, err
	}

	edwards := *pairedJWK
	edwards.Algorithm = alg

	return &edwards, nil
}

// hasKey returns true when a key matches the alg and use values, and the kid value when it is not empty.
func (i *Issuer) hasKey(kid, alg, use string) bool {
	return slices.ContainsFunc(i.jwks.Keys, func(key jose.JSONWebKey) bool {
		return key.Algorithm == alg && key.Use == use && (kid == "" || key.KeyID == kid)
	})
}
