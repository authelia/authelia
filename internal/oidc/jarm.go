package oidc

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/token/jwt"
)

// EncodeJWTSecuredResponseParameters takes the result from GenerateJWTSecuredResponse and turns it into parameters in the form of url.Values.
func EncodeJWTSecuredResponseParameters(token, _ string, tErr error) (parameters url.Values, err error) {
	if tErr != nil {
		return nil, tErr
	}

	return url.Values{FormParameterResponse: []string{token}}, nil
}

// GenerateJWTSecuredResponse generates the token and signature for a JARM response.
func GenerateJWTSecuredResponse(ctx context.Context, config JWTSecuredResponseModeProvider, client Client, session any, in url.Values) (token, signature string, err error) {
	headers := map[string]any{}

	if alg := client.GetAuthorizationSignedResponseAlg(); len(alg) > 0 {
		headers[JWTHeaderKeyAlgorithm] = alg
	}

	if kid := client.GetAuthorizationSignedResponseKeyID(); len(kid) > 0 {
		headers[JWTHeaderKeyIdentifier] = kid
	}

	var issuer string

	issuer = config.GetJWTSecuredAuthorizeResponseModeIssuer(ctx)

	if len(issuer) == 0 {
		var (
			src   jwt.MapClaims
			value any
			ok    bool
		)

		switch s := session.(type) {
		case nil:
			return "", "", errors.New("The JARM response modes require the Authorize Requester session to be set but it wasn't.")
		case IDTokenSessionContainer:
			src = s.IDTokenClaims().ToMapClaims()
		case oauth2.JWTSessionContainer:
			src = s.GetJWTClaims().ToMapClaims()
		default:
			return "", "", errors.New("The JARM response modes require the Authorize Requester session to implement either the IDTokenSessionContainer or oauth2.JWTSessionContainer interfaces but it doesn't.")
		}

		if value, ok = src[ClaimIssuer]; ok {
			issuer, _ = value.(string)
		}
	}

	claims := &JWTSecuredAuthorizationResponseModeClaims{
		JTI:       uuid.New().String(),
		Issuer:    issuer,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(config.GetJWTSecuredAuthorizeResponseModeLifespan(ctx)),
		Audience:  []string{client.GetID()},
		Extra:     map[string]any{},
	}

	for param := range in {
		claims.Extra[param] = in.Get(param)
	}

	var signer jwt.Signer

	if signer = config.GetJWTSecuredAuthorizeResponseModeSigner(ctx); signer == nil {
		return "", "", errors.New("The JARM response modes require the JWTSecuredAuthorizeResponseModeSignerProvider to return a jwt.Signer but it didn't.")
	}

	return signer.Generate(ctx, claims.ToMapClaims(), &jwt.Headers{Extra: headers})
}

// JWTSecuredAuthorizationResponseModeClaims represent the JWT claims for JARM.
type JWTSecuredAuthorizationResponseModeClaims struct {
	JTI       string
	Issuer    string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Audience  []string
	Extra     map[string]any
}

// ToMap will transform the headers to a map structure.
func (c *JWTSecuredAuthorizationResponseModeClaims) ToMap() map[string]any {
	var ret = mapCopy(c.Extra)

	if c.Issuer != "" {
		ret[ClaimIssuer] = c.Issuer
	} else {
		delete(ret, ClaimIssuer)
	}

	if c.JTI != "" {
		ret[ClaimJWTID] = c.JTI
	} else {
		ret[ClaimJWTID] = uuid.New().String()
	}

	if len(c.Audience) > 0 {
		ret[ClaimAudience] = c.Audience
	} else {
		ret[ClaimAudience] = []string{}
	}

	if !c.IssuedAt.IsZero() {
		ret[ClaimIssuedAt] = c.IssuedAt.Unix()
	} else {
		delete(ret, ClaimIssuedAt)
	}

	if !c.ExpiresAt.IsZero() {
		ret[ClaimExpirationTime] = c.ExpiresAt.Unix()
	} else {
		delete(ret, ClaimExpirationTime)
	}

	return ret
}

// FromMap will set the claims based on a mapping.
func (c *JWTSecuredAuthorizationResponseModeClaims) FromMap(m map[string]any) {
	c.Extra = make(map[string]any)

	for k, v := range m {
		switch k {
		case ClaimJWTID:
			if s, ok := v.(string); ok {
				c.JTI = s
			}
		case ClaimIssuer:
			if s, ok := v.(string); ok {
				c.Issuer = s
			}
		case ClaimAudience:
			c.Audience = toStringSlice(v)
		case ClaimIssuedAt:
			c.IssuedAt = toTime(v, c.IssuedAt)
		case ClaimExpirationTime:
			c.ExpiresAt = toTime(v, c.ExpiresAt)
		default:
			c.Extra[k] = v
		}
	}
}

// Add will add a key-value pair to the extra field.
func (c *JWTSecuredAuthorizationResponseModeClaims) Add(key string, value any) {
	if c.Extra == nil {
		c.Extra = make(map[string]any)
	}

	c.Extra[key] = value
}

// Get will get a value from the extra field based on a given key.
func (c *JWTSecuredAuthorizationResponseModeClaims) Get(key string) any {
	return c.ToMap()[key]
}

// ToMapClaims will return a jwt-go MapClaims representation.
func (c *JWTSecuredAuthorizationResponseModeClaims) ToMapClaims() jwt.MapClaims {
	return c.ToMap()
}

// FromMapClaims will populate claims from a jwt-go MapClaims representation.
func (c *JWTSecuredAuthorizationResponseModeClaims) FromMapClaims(mc jwt.MapClaims) {
	c.FromMap(mc)
}
