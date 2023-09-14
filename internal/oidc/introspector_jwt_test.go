package oidc_test

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ory/fosite"
	fjwt "github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestStatelessJWTValidator_IntrospectToken(t *testing.T) {
	signer := &fjwt.DefaultSigner{
		GetPrivateKey: func(ctx context.Context) (any, error) {
			return x509PrivateKeyRSA2048, nil
		},
	}

	maketoken := func(method jwt.SigningMethod, claims jwt.MapClaims, header map[string]any) string {
		j := &jwt.Token{
			Header: header,
			Claims: claims,
			Method: method,
		}

		if _, ok := j.Header[oidc.JWTHeaderKeyAlgorithm]; !ok {
			j.Header[oidc.JWTHeaderKeyAlgorithm] = method.Alg()
		}

		token, err := j.SignedString(x509PrivateKeyRSA2048)

		if err != nil {
			panic(err)
		}

		return token
	}

	handler := oidc.StatelessJWTValidator{
		Signer: signer,
		Config: &ScopeStrategyProvider{
			value: fosite.ExactScopeStrategy,
		},
	}

	testCases := []struct {
		name     string
		have     string
		scopes   []string
		expected fosite.TokenUse
		err      string
	}{
		{
			"ShouldHandleAccessTokenJWT",
			maketoken(jwt.SigningMethodRS256, jwt.MapClaims{}, map[string]any{oidc.JWTHeaderKeyType: oidc.JWTHeaderTypeValueAccessTokenJWT}),
			nil,
			fosite.AccessToken,
			"",
		},
		{
			"ShouldHandleAccessTokenJWTWithScopes",
			maketoken(jwt.SigningMethodRS256, jwt.MapClaims{oidc.ClaimScope: "example"}, map[string]any{oidc.JWTHeaderKeyType: oidc.JWTHeaderTypeValueAccessTokenJWT}),
			[]string{"example"},
			fosite.AccessToken,
			"",
		},
		{
			"ShouldHandleAccessTokenJWTWithScopes",
			maketoken(jwt.SigningMethodRS256, jwt.MapClaims{oidc.ClaimScope: "example2"}, map[string]any{oidc.JWTHeaderKeyType: oidc.JWTHeaderTypeValueAccessTokenJWT}),
			[]string{"example"},
			fosite.AccessToken,
			"The requested scope is invalid, unknown, or malformed. The request scope 'example' has not been granted or is not allowed to be requested.",
		},
		{
			"ShouldRejectStandardJWT",
			maketoken(jwt.SigningMethodRS256, jwt.MapClaims{}, map[string]any{oidc.JWTHeaderKeyType: "JWT"}),
			nil,
			fosite.TokenUse(""),
			"The request could not be authorized. Check that you provided valid credentials in the right format. The provided token is not a valid RFC9068 JWT Profile Access Token as it is missing the header 'typ' value of 'at+jwt'.",
		},
		{
			"ShouldRejectNonJWT",
			"authelia_at_example",
			nil,
			fosite.TokenUse(""),
			"The handler is not responsible for this request. The provided token appears to be an opaque token not a JWT.",
		},
		{
			"ShouldRejectTokenWithOpaquePrefix",
			"authelia_at_example.another.example",
			nil,
			fosite.TokenUse(""),
			"The handler is not responsible for this request. The provided token appears to be an opaque token not a JWT.",
		},
		{
			"ShouldRecjectInvalidToken",
			"example.another.example",
			nil,
			fosite.TokenUse(""),
			"Invalid token format. Check that you provided a valid token in the right format. invalid character '\\x16' looking for beginning of object key string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ar := fosite.NewAccessRequest(oidc.NewSession())

			actual, err := handler.IntrospectToken(context.TODO(), tc.have, fosite.AccessToken, ar, tc.scopes)

			assert.Equal(t, tc.expected, actual)

			if len(tc.err) == 0 {
				assert.NoError(t, oidc.ErrorToDebugRFC6749Error(err))
			} else {
				assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.err)
			}
		})
	}
}

type ScopeStrategyProvider struct {
	value fosite.ScopeStrategy
}

func (p *ScopeStrategyProvider) GetScopeStrategy(ctx context.Context) fosite.ScopeStrategy {
	return p.value
}
