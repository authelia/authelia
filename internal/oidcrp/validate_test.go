package oidcrp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"authelia.com/provider/oauth2/token/jose"
)

func TestValidateIDToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	other, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pkix, err := x509.MarshalPKIXPublicKey(key.Public())
	require.NoError(t, err)

	now := time.Unix(1700000000, 0)

	keys := &stubKeySet{jwks: &jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{KeyID: "kid1", Key: key.Public(), Algorithm: "RS256", Use: "sig"}}}}

	testCases := []struct {
		Name     string
		Sign     any
		KeyID    string
		Method   jwt.SigningMethod
		Claims   jwt.MapClaims
		Nonce    string
		Keys     KeySet
		Options  func(opts *ValidateOptions)
		Expected *IdentityClaims
		Error    string
	}{
		{
			Name:   "ShouldValidateWellFormedToken",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{
				"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"},
				"exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value",
				"preferred_username": "john", "name": "John Smith", "email": "john@example.com",
				"amr": []string{"pwd", "mfa"},
			},
			Nonce: "nonce-value",
			Expected: &IdentityClaims{
				Issuer: "https://op.example.com", Subject: "abc123",
				PreferredUsername: "john", Name: "John Smith", Email: "john@example.com",
				AuthenticationMethodsReference: []string{"pwd", "mfa"},
			},
		},
		{
			Name:   "ShouldValidateTokenWithinLeeway",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{
				"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"},
				"exp": now.Add(-30 * time.Second).Unix(), "iat": now.Add(-time.Hour).Unix(),
				"nbf": now.Add(30 * time.Second).Unix(), "nonce": "nonce-value",
			},
			Nonce:    "nonce-value",
			Expected: &IdentityClaims{Issuer: "https://op.example.com", Subject: "abc123"},
		},
		{
			Name:   "ShouldSelectSigningKeyOverEncryptionKey",
			Sign:   key,
			KeyID:  "",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Keys: &stubKeySet{jwks: &jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
				{KeyID: "enc1", Key: other.Public(), Algorithm: "RS256", Use: "enc"},
				{KeyID: "sig1", Key: key.Public(), Algorithm: "RS256", Use: "sig"},
			}}},
			Expected: &IdentityClaims{Issuer: "https://op.example.com", Subject: "abc123"},
		},
		{
			Name:   "ShouldSelectKeyMatchingConfiguredAlgorithm",
			Sign:   key,
			KeyID:  "",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Keys: &stubKeySet{jwks: &jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
				{KeyID: "big1", Key: other.Public(), Algorithm: "RS512", Use: "sig"},
				{KeyID: "sig1", Key: key.Public(), Algorithm: "RS256", Use: "sig"},
			}}},
			Expected: &IdentityClaims{Issuer: "https://op.example.com", Subject: "abc123"},
		},
		{
			Name:   "ShouldRaiseErrorOnIssuerMismatch",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://evil.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'iss' claim value 'https://evil.example.com' does not match the expected value 'https://op.example.com'",
		},
		{
			Name:   "ShouldRaiseErrorOnAbsentIssuerClaim",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'iss' claim is required but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnNonStringIssuerClaim",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": 12345, "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'iss' claim is required but it is absent",
		},
		{
			Name:    "ShouldRaiseErrorOnEmptyExpectedIssuer",
			Sign:    key,
			KeyID:   "kid1",
			Method:  jwt.SigningMethodRS256,
			Claims:  jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:   "nonce-value",
			Options: func(opts *ValidateOptions) { opts.Issuer = "" },
			Error:   "error validating the id token: the id token validation options are invalid: the expected 'iss' value is required but it is absent",
		},
		{
			Name:    "ShouldRaiseErrorOnAbsentIssuerClaimWithEmptyExpectedIssuer",
			Sign:    key,
			KeyID:   "kid1",
			Method:  jwt.SigningMethodRS256,
			Claims:  jwt.MapClaims{"sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:   "nonce-value",
			Options: func(opts *ValidateOptions) { opts.Issuer = "" },
			Error:   "error validating the id token: the id token validation options are invalid: the expected 'iss' value is required but it is absent",
		},
		{
			Name:    "ShouldRaiseErrorOnEmptyExpectedClientID",
			Sign:    key,
			KeyID:   "kid1",
			Method:  jwt.SigningMethodRS256,
			Claims:  jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:   "nonce-value",
			Options: func(opts *ValidateOptions) { opts.ClientID = "" },
			Error:   "error validating the id token: the id token validation options are invalid: the client id is required but it is absent",
		},
		{
			Name:    "ShouldRaiseErrorOnEmptyAudienceWithEmptyExpectedClientID",
			Sign:    key,
			KeyID:   "kid1",
			Method:  jwt.SigningMethodRS256,
			Claims:  jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{""}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:   "nonce-value",
			Options: func(opts *ValidateOptions) { opts.ClientID = "" },
			Error:   "error validating the id token: the id token validation options are invalid: the client id is required but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnAudienceMismatch",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"other"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'aud' claim does not contain the client id 'client'",
		},
		{
			Name:   "ShouldRaiseErrorOnAbsentAudience",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'aud' claim is required but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnNonStringAudienceElement",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []any{"client", 12345}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'aud' claim must only contain string values",
		},
		{
			Name:   "ShouldRaiseErrorOnNonStringAudience",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": 12345, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'aud' claim must only contain string values",
		},
		{
			Name:   "ShouldDiscardNonStringAuthenticationMethodsReferenceElements",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{
				"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"},
				"exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value",
				"amr": []any{"pwd", 12345, "otp"},
			},
			Nonce: "nonce-value",
			Expected: &IdentityClaims{
				Issuer: "https://op.example.com", Subject: "abc123",
				AuthenticationMethodsReference: []string{"pwd", "otp"},
			},
		},
		{
			Name:   "ShouldRaiseErrorOnMultipleAudienceWithoutAZP",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client", "other"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'azp' claim is required when the 'aud' claim has multiple values but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnAZPMismatch",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client", "other"}, "azp": "other", "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'azp' claim value 'other' does not match the client id 'client'",
		},
		{
			Name:   "ShouldRaiseErrorOnNonceMismatch",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "wrong"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'nonce' claim does not match the expected value",
		},
		{
			Name:   "ShouldRaiseErrorOnAbsentNonceClaim",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix()},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'nonce' claim is required but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnEmptyNonceClaim",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": ""},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'nonce' claim is required but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnEmptyExpectedNonce",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "",
			Error:  "error validating the id token: the id token validation options are invalid: the expected 'nonce' value is required but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnAbsentNonceClaimWithEmptyExpectedNonce",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix()},
			Nonce:  "",
			Error:  "error validating the id token: the id token validation options are invalid: the expected 'nonce' value is required but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnExpiredToken",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(-time.Hour).Unix(), "iat": now.Add(-2 * time.Hour).Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'exp' claim indicates the token is expired",
		},
		{
			Name:   "ShouldRaiseErrorOnAbsentExpiration",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'exp' claim is required but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnAbsentIssuedAt",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'iat' claim is required but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnFutureIssuedAt",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(2 * time.Hour).Unix(), "iat": now.Add(time.Hour).Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'iat' claim indicates the token was issued in the future",
		},
		{
			Name:   "ShouldRaiseErrorOnFutureNotBefore",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nbf": now.Add(time.Hour).Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'nbf' claim indicates the token is not yet valid",
		},
		{
			Name:   "ShouldRaiseErrorOnAbsentSubject",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'sub' claim is required but it is absent",
		},
		{
			Name:   "ShouldRaiseErrorOnOverlongSubject",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": strings.Repeat("a", 256), "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token contains an invalid claim: the 'sub' claim must not exceed 255 characters",
		},
		{
			Name:   "ShouldRaiseErrorOnUnknownKeyID",
			Sign:   key,
			KeyID:  "kid-unknown",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: no key in the json web key set matched the id token",
		},
		{
			Name:   "ShouldRaiseErrorOnForeignSignature",
			Sign:   other,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token signature could not be verified",
		},
		{
			Name:   "ShouldRaiseErrorOnAlgorithmMismatch",
			Sign:   key,
			KeyID:  "kid1",
			Method: jwt.SigningMethodRS512,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token signature could not be verified",
		},
		{
			Name:   "ShouldRaiseErrorOnUnsignedAlgorithmNone",
			Sign:   jwt.UnsafeAllowNoneSignatureType,
			KeyID:  "kid1",
			Method: jwt.SigningMethodNone,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token signature could not be verified",
		},
		{
			Name:   "ShouldRaiseErrorOnHMACSignedWithPublicKey",
			Sign:   pkix,
			KeyID:  "kid1",
			Method: jwt.SigningMethodHS256,
			Claims: jwt.MapClaims{"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"}, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value"},
			Nonce:  "nonce-value",
			Error:  "error validating the id token: the id token signature could not be verified",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			token := jwt.NewWithClaims(tc.Method, tc.Claims)
			token.Header["kid"] = tc.KeyID

			raw, err := token.SignedString(tc.Sign)
			require.NoError(t, err)

			opts := ValidateOptions{
				Issuer: "https://op.example.com", ClientID: "client", Nonce: tc.Nonce,
				Alg: "RS256", JWKSURI: "https://op.example.com/jwks.json", Now: now, Leeway: time.Minute,
			}

			if tc.Options != nil {
				tc.Options(&opts)
			}

			set := KeySet(keys)

			if tc.Keys != nil {
				set = tc.Keys
			}

			claims, err := ValidateIDToken(context.Background(), set, raw, opts)

			if tc.Error != "" {
				assert.Nil(t, claims)
				require.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, claims)
		})
	}
}

func TestValidateIDTokenRefetchesOnKeyIDMiss(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	now := time.Unix(1700000000, 0)

	keys := &stubKeySet{jwks: &jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{KeyID: "kid1", Key: key.Public(), Algorithm: "RS256", Use: "sig"}}}}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"},
		"exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value",
	})
	token.Header["kid"] = "kid-absent"

	raw, err := token.SignedString(key)
	require.NoError(t, err)

	_, err = ValidateIDToken(context.Background(), keys, raw, ValidateOptions{
		Issuer: "https://op.example.com", ClientID: "client", Nonce: "nonce-value",
		Alg: "RS256", JWKSURI: "https://op.example.com/jwks.json", Now: now, Leeway: time.Minute,
	})

	require.EqualError(t, err, "error validating the id token: no key in the json web key set matched the id token")
	assert.Equal(t, 2, keys.resolved)
}

func TestValidateIDTokenResolvesRotatedKey(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	rotated, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	now := time.Unix(1700000000, 0)

	keys := &rotatingKeySet{
		cached:    &jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{KeyID: "kid1", Key: key.Public(), Algorithm: "RS256", Use: "sig"}}},
		refreshed: &jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{KeyID: "kid2", Key: rotated.Public(), Algorithm: "RS256", Use: "sig"}}},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"},
		"exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "nonce-value",
	})
	token.Header["kid"] = "kid2"

	raw, err := token.SignedString(rotated)
	require.NoError(t, err)

	claims, err := ValidateIDToken(context.Background(), keys, raw, ValidateOptions{
		Issuer: "https://op.example.com", ClientID: "client", Nonce: "nonce-value",
		Alg: "RS256", JWKSURI: "https://op.example.com/jwks.json", Now: now, Leeway: time.Minute,
	})

	require.NoError(t, err)
	assert.Equal(t, &IdentityClaims{Issuer: "https://op.example.com", Subject: "abc123"}, claims)
	assert.Equal(t, 2, keys.resolved)
}

type stubKeySet struct {
	jwks     *jose.JSONWebKeySet
	resolved int
}

func (s *stubKeySet) Resolve(_ context.Context, _ string, _ bool) (jwks *jose.JSONWebKeySet, err error) {
	s.resolved++

	return s.jwks, nil
}

type rotatingKeySet struct {
	cached    *jose.JSONWebKeySet
	refreshed *jose.JSONWebKeySet
	resolved  int
}

func (s *rotatingKeySet) Resolve(_ context.Context, _ string, ignoreCache bool) (jwks *jose.JSONWebKeySet, err error) {
	s.resolved++

	if ignoreCache {
		return s.refreshed, nil
	}

	return s.cached, nil
}
