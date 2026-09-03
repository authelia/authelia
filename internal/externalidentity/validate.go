// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash"

	"authelia.com/provider/oauth2/token/jose"
)

// ValidateIDToken parses and fully validates an ID Token, returning the claims Authelia consumes.
func ValidateIDToken(ctx context.Context, keys KeySet, raw string, opts ValidateOptions) (claims *IdentityClaims, err error) {
	var jws *jose.JSONWebSignature

	// An ID Token is always in the JWS Compact Serialization (OpenID Connect Core 1.0 Section 2), and only the configured
	// algorithm is accepted, which excludes the 'none' algorithm and any symmetric algorithm keyed with a public key.
	if jws, err = jose.ParseSignedCompact(raw, []jose.SignatureAlgorithm{jose.SignatureAlgorithm(opts.Alg)}); err != nil || len(jws.Signatures) != 1 {
		return nil, fmt.Errorf("error validating the id token: %w", ErrTokenSignatureInvalid)
	}

	var payload []byte

	if payload, err = verifyJWSSignature(ctx, keys, jws, opts); err != nil {
		if errors.Is(err, ErrTokenNoKey) {
			return nil, fmt.Errorf("error validating the id token: %w", ErrTokenNoKey)
		}

		return nil, fmt.Errorf("error validating the id token: %w", ErrTokenSignatureInvalid)
	}

	mapped := map[string]any{}

	if err = json.Unmarshal(payload, &mapped); err != nil {
		return nil, fmt.Errorf("error validating the id token: %w: the payload is not a JSON object", ErrTokenClaimInvalid)
	}

	if claims, err = validateClaims(mapped, opts); err != nil {
		return nil, fmt.Errorf("error validating the id token: %w", err)
	}

	return claims, nil
}

func verifyJWSSignature(ctx context.Context, keys KeySet, jws *jose.JSONWebSignature, opts ValidateOptions) (payload []byte, err error) {
	kid := jws.Signatures[0].Protected.KeyID

	for _, fresh := range []bool{false, true} {
		var jwks *jose.JSONWebKeySet

		if jwks, err = keys.Resolve(ctx, opts.JWKSURI, fresh); err != nil {
			return nil, err
		}

		candidates := selectKeys(jwks, kid, opts.Alg)

		if len(candidates) == 0 {
			if fresh {
				return nil, ErrTokenNoKey
			}

			continue
		}

		for _, candidate := range candidates {
			if payload, err = jws.Verify(candidate.Key); err == nil {
				return payload, nil
			}
		}
	}

	return nil, ErrTokenSignatureInvalid
}

func selectKeys(jwks *jose.JSONWebKeySet, kid, alg string) (candidates []jose.JSONWebKey) {
	if jwks == nil {
		return nil
	}

	for _, jwk := range jwks.Keys {
		if jwk.Use != "" && jwk.Use != "sig" {
			continue
		}

		if jwk.Algorithm != "" && jwk.Algorithm != alg {
			continue
		}

		if kid != "" && jwk.KeyID != kid {
			continue
		}

		candidates = append(candidates, jwk)
	}

	return candidates
}

func validateClaims(mapped map[string]any, opts ValidateOptions) (claims *IdentityClaims, err error) {
	var issuer string

	if issuer, err = validateIssuer(mapped, opts); err != nil {
		return nil, err
	}

	var subject string

	if subject, err = validateSubject(mapped); err != nil {
		return nil, err
	}

	if err = validateAudience(mapped, opts); err != nil {
		return nil, err
	}

	if err = validateTime(mapped, opts); err != nil {
		return nil, err
	}

	if err = validateNonce(mapped, opts); err != nil {
		return nil, err
	}

	if err = validateAccessTokenHash(mapped, opts); err != nil {
		return nil, err
	}

	preferred, _ := mapped[claimPreferredUsername].(string)
	name, _ := mapped[claimName].(string)
	email, _ := mapped[claimEmail].(string)

	return &IdentityClaims{
		Issuer:                         issuer,
		Subject:                        subject,
		PreferredUsername:              preferred,
		Name:                           name,
		Email:                          email,
		AuthenticationMethodsReference: stringsClaim(mapped[claimAuthnMethodRefs]),
	}, nil
}

func validateIssuer(mapped map[string]any, opts ValidateOptions) (issuer string, err error) {
	if opts.Issuer == "" {
		return "", fmt.Errorf("%w: the expected 'iss' value is required but it is absent", ErrValidationOptionsInvalid)
	}

	issuer, _ = mapped[claimIssuer].(string)

	switch {
	case issuer == "":
		return "", fmt.Errorf("%w: the 'iss' claim is required but it is absent", ErrTokenClaimInvalid)
	case issuer != opts.Issuer:
		return "", fmt.Errorf("%w: the 'iss' claim value '%s' does not match the expected value '%s'", ErrTokenClaimInvalid, issuer, opts.Issuer)
	}

	return issuer, nil
}

func validateSubject(mapped map[string]any) (subject string, err error) {
	subject, _ = mapped[claimSubject].(string)

	switch {
	case subject == "":
		return "", fmt.Errorf("%w: the 'sub' claim is required but it is absent", ErrTokenClaimInvalid)
	case len(subject) > 255:
		return "", fmt.Errorf("%w: the 'sub' claim must not exceed 255 characters", ErrTokenClaimInvalid)
	}

	return subject, nil
}

func validateAudience(mapped map[string]any, opts ValidateOptions) (err error) {
	if opts.ClientID == "" {
		return fmt.Errorf("%w: the client id is required but it is absent", ErrValidationOptionsInvalid)
	}

	value, has := mapped[claimAudience]

	if !has || value == nil {
		return fmt.Errorf("%w: the 'aud' claim is required but it is absent", ErrTokenClaimInvalid)
	}

	audience, ok := audienceClaim(value)

	if !ok {
		return fmt.Errorf("%w: the 'aud' claim must only contain string values", ErrTokenClaimInvalid)
	}

	if !containsString(audience, opts.ClientID) {
		return fmt.Errorf("%w: the 'aud' claim does not contain the client id '%s'", ErrTokenClaimInvalid, opts.ClientID)
	}

	if len(audience) > 1 {
		azp, is := mapped[claimAuthorizedParty].(string)

		switch {
		case !is || azp == "":
			return fmt.Errorf("%w: the 'azp' claim is required when the 'aud' claim has multiple values but it is absent", ErrTokenClaimInvalid)
		case azp != opts.ClientID:
			return fmt.Errorf("%w: the 'azp' claim value '%s' does not match the client id '%s'", ErrTokenClaimInvalid, azp, opts.ClientID)
		}
	}

	return nil
}

func validateNonce(mapped map[string]any, opts ValidateOptions) (err error) {
	if opts.Nonce == "" {
		return fmt.Errorf("%w: the expected 'nonce' value is required but it is absent", ErrValidationOptionsInvalid)
	}

	nonce, _ := mapped[claimNonce].(string)

	if nonce == "" {
		return fmt.Errorf("%w: the 'nonce' claim is required but it is absent", ErrTokenClaimInvalid)
	}

	if subtle.ConstantTimeCompare([]byte(nonce), []byte(opts.Nonce)) != 1 {
		return fmt.Errorf("%w: the 'nonce' claim does not match the expected value", ErrTokenClaimInvalid)
	}

	return nil
}

func validateTime(mapped map[string]any, opts ValidateOptions) (err error) {
	exp, ok := numericClaim(mapped, "exp")
	if !ok {
		return fmt.Errorf("%w: the 'exp' claim is required but it is absent", ErrTokenClaimInvalid)
	}

	if opts.Now.Add(-opts.Leeway).Unix() >= exp {
		return fmt.Errorf("%w: the 'exp' claim indicates the token is expired", ErrTokenClaimInvalid)
	}

	iat, ok := numericClaim(mapped, "iat")
	if !ok {
		return fmt.Errorf("%w: the 'iat' claim is required but it is absent", ErrTokenClaimInvalid)
	}

	if iat > opts.Now.Add(opts.Leeway).Unix() {
		return fmt.Errorf("%w: the 'iat' claim indicates the token was issued in the future", ErrTokenClaimInvalid)
	}

	if nbf, has := numericClaim(mapped, "nbf"); has && nbf > opts.Now.Add(opts.Leeway).Unix() {
		return fmt.Errorf("%w: the 'nbf' claim indicates the token is not yet valid", ErrTokenClaimInvalid)
	}

	return nil
}

func numericClaim(mapped map[string]any, name string) (value int64, ok bool) {
	switch v := mapped[name].(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case json.Number:
		n, err := v.Int64()

		return n, err == nil
	default:
		return 0, false
	}
}

func audienceClaim(value any) (values []string, ok bool) {
	switch v := value.(type) {
	case string:
		return []string{v}, true
	case []string:
		return v, true
	case []any:
		values = make([]string, 0, len(v))

		for _, item := range v {
			s, is := item.(string)
			if !is {
				return nil, false
			}

			values = append(values, s)
		}

		return values, true
	default:
		return nil, false
	}
}

func stringsClaim(value any) (values []string) {
	switch v := value.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	case []any:
		values = make([]string, 0, len(v))

		for _, item := range v {
			if s, ok := item.(string); ok {
				values = append(values, s)
			}
		}

		return values
	default:
		return nil
	}
}

func validateAccessTokenHash(mapped map[string]any, opts ValidateOptions) (err error) {
	value, has := mapped[claimAccessTokenHash]

	if !has {
		return nil
	}

	claim, ok := value.(string)

	switch {
	case !ok || claim == "":
		return fmt.Errorf("%w: the 'at_hash' claim must be a non-empty string", ErrTokenClaimInvalid)
	case opts.AccessToken == "":
		return fmt.Errorf("%w: the 'at_hash' claim is present but there is no access token to validate it against", ErrTokenClaimInvalid)
	}

	var expected string

	if expected, err = accessTokenHash(opts.Alg, opts.AccessToken); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationOptionsInvalid, err)
	}

	if subtle.ConstantTimeCompare([]byte(claim), []byte(expected)) != 1 {
		return fmt.Errorf("%w: the 'at_hash' claim does not match the access token", ErrTokenClaimInvalid)
	}

	return nil
}

func accessTokenHash(alg, accessToken string) (value string, err error) {
	var h hash.Hash

	switch {
	case len(alg) != 5:
		return "", fmt.Errorf("the algorithm '%s' has no associated hash algorithm", alg)
	case alg[2:] == "256":
		h = sha256.New()
	case alg[2:] == "384":
		h = sha512.New384()
	case alg[2:] == "512":
		h = sha512.New()
	default:
		return "", fmt.Errorf("the algorithm '%s' has no associated hash algorithm", alg)
	}

	h.Write([]byte(accessToken))

	sum := h.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(sum[:len(sum)/2]), nil
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}

	return false
}
