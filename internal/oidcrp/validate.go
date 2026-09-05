package oidcrp

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"authelia.com/provider/oauth2/token/jose"
)

// ValidateIDToken parses and fully validates an ID Token, returning the claims Authelia consumes.
func ValidateIDToken(ctx context.Context, keys KeySet, raw string, opts ValidateOptions) (claims *IdentityClaims, err error) {
	mapped := jwt.MapClaims{}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{opts.Alg}), jwt.WithoutClaimsValidation())

	if _, err = parser.ParseWithClaims(raw, mapped, keyFunc(ctx, keys, opts)); err != nil {
		if errors.Is(err, ErrTokenNoKey) {
			return nil, fmt.Errorf("error validating the id token: %w", ErrTokenNoKey)
		}

		return nil, fmt.Errorf("error validating the id token: %w", ErrTokenSignatureInvalid)
	}

	if claims, err = validateClaims(mapped, opts); err != nil {
		return nil, fmt.Errorf("error validating the id token: %w", err)
	}

	return claims, nil
}

func keyFunc(ctx context.Context, keys KeySet, opts ValidateOptions) jwt.Keyfunc {
	return func(token *jwt.Token) (key any, err error) {
		kid, _ := token.Header["kid"].(string)

		var jwks *jose.JSONWebKeySet

		if jwks, err = keys.Resolve(ctx, opts.JWKSURI, false); err != nil {
			return nil, err
		}

		if key = selectKey(jwks, kid, opts.Alg); key != nil {
			return key, nil
		}

		if jwks, err = keys.Resolve(ctx, opts.JWKSURI, true); err != nil {
			return nil, err
		}

		if key = selectKey(jwks, kid, opts.Alg); key != nil {
			return key, nil
		}

		return nil, ErrTokenNoKey
	}
}

func selectKey(jwks *jose.JSONWebKeySet, kid, alg string) (key any) {
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

		return jwk.Key
	}

	return nil
}

func validateClaims(mapped jwt.MapClaims, opts ValidateOptions) (claims *IdentityClaims, err error) {
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

func validateIssuer(mapped jwt.MapClaims, opts ValidateOptions) (issuer string, err error) {
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

func validateSubject(mapped jwt.MapClaims) (subject string, err error) {
	subject, _ = mapped[claimSubject].(string)

	switch {
	case subject == "":
		return "", fmt.Errorf("%w: the 'sub' claim is required but it is absent", ErrTokenClaimInvalid)
	case len(subject) > 255:
		return "", fmt.Errorf("%w: the 'sub' claim must not exceed 255 characters", ErrTokenClaimInvalid)
	}

	return subject, nil
}

func validateAudience(mapped jwt.MapClaims, opts ValidateOptions) (err error) {
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

func validateNonce(mapped jwt.MapClaims, opts ValidateOptions) (err error) {
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

func validateTime(mapped jwt.MapClaims, opts ValidateOptions) (err error) {
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

func numericClaim(mapped jwt.MapClaims, name string) (value int64, ok bool) {
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

// audienceClaim returns the values of the 'aud' claim. Every element must be a string as RFC 7519 defines the claim
// in terms of the StringOrURI type. A non string element is not discarded as it would under count the audience, which
// switches off the 'azp' requirement that is the only control distinguishing a token issued to this client from one
// which merely lists this client in its audience.
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

// stringsClaim returns the string values of a claim, discarding any element which is not a string. It is used for
// advisory claims such as 'amr' where rejecting the entire claim over a single malformed element is harsher than the
// value of the claim warrants. It must not be used for a claim a security decision depends on; see audienceClaim.
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

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}

	return false
}
