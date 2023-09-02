package oidc_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/openid"
	fjwt "github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestEncodeJWTSecuredResponseParameters(t *testing.T) {
	testCases := []struct {
		name     string
		issuer   string
		signer   *fjwt.DefaultSigner
		client   oidc.Client
		session  any
		in       url.Values
		expected jwt.MapClaims
		err      string
	}{
		{
			"ShouldErrorOnNilSession",
			"",
			&fjwt.DefaultSigner{
				GetPrivateKey: func(ctx context.Context) (any, error) {
					return nil, nil
				},
			},
			&oidc.BaseClient{},
			nil,
			nil,
			jwt.MapClaims{},
			"The JARM response modes require the Authorize Requester session to be set but it wasn't.",
		},
		{
			"ShouldErrorOnBadTypeSession",
			"",
			&fjwt.DefaultSigner{
				GetPrivateKey: func(ctx context.Context) (any, error) {
					return nil, nil
				},
			},
			&oidc.BaseClient{},
			1,
			nil,
			jwt.MapClaims{},
			"The JARM response modes require the Authorize Requester session to implement either the IDTokenSessionContainer or oauth2.JWTSessionContainer interfaces but it doesn't.",
		},
		{
			"ShouldErrorOnNilKey",
			"https://auth.example.com",
			&fjwt.DefaultSigner{
				GetPrivateKey: func(ctx context.Context) (any, error) {
					return nil, nil
				},
			},
			&oidc.BaseClient{},
			nil,
			nil,
			jwt.MapClaims{},
			"unsupported private key type: <nil>",
		},
		{
			"ShouldErrorOnNilKey",
			"https://auth.example.com",
			&fjwt.DefaultSigner{
				GetPrivateKey: func(ctx context.Context) (any, error) {
					return keyRSA2048, nil
				},
			},
			&oidc.BaseClient{
				ID:                               "example",
				AuthorizationSignedResponseAlg:   oidc.SigningAlgRSAUsingSHA256,
				AuthorizationSignedResponseKeyID: "12345",
			},
			nil,
			nil,
			jwt.MapClaims{
				oidc.ClaimAudience: []any{"example"},
				oidc.ClaimIssuer:   "https://auth.example.com",
			},
			"",
		},
		{
			"ShouldErrorOnNilSigner",
			"https://auth.example.com",
			nil,
			&oidc.BaseClient{
				ID:                               "example",
				AuthorizationSignedResponseAlg:   oidc.SigningAlgRSAUsingSHA256,
				AuthorizationSignedResponseKeyID: "12345",
			},
			nil,
			nil,
			jwt.MapClaims{
				oidc.ClaimAudience: []any{"example"},
				oidc.ClaimIssuer:   "https://auth.example.com",
			},
			"The JARM response modes require the JWTSecuredAuthorizeResponseModeSignerProvider to return a jwt.Signer but it didn't.",
		},
		{
			"ShouldEncodeParameters",
			"https://auth.example.com",
			&fjwt.DefaultSigner{
				GetPrivateKey: func(ctx context.Context) (any, error) {
					return keyRSA2048, nil
				},
			},
			&oidc.BaseClient{
				ID:                               "example",
				AuthorizationSignedResponseAlg:   oidc.SigningAlgRSAUsingSHA256,
				AuthorizationSignedResponseKeyID: "12345",
			},
			nil,
			url.Values{oidc.FormParameterAuthorizationCode: []string{"123"}},
			jwt.MapClaims{
				oidc.ClaimAudience:                  []any{"example"},
				oidc.ClaimIssuer:                    "https://auth.example.com",
				oidc.FormParameterAuthorizationCode: "123",
			},
			"",
		},
		{
			"ShouldEncodeParametersAndRestoreIssuerFromSession",
			"",
			&fjwt.DefaultSigner{
				GetPrivateKey: func(ctx context.Context) (any, error) {
					return keyRSA2048, nil
				},
			},
			&oidc.BaseClient{
				ID:                               "example",
				AuthorizationSignedResponseAlg:   oidc.SigningAlgRSAUsingSHA256,
				AuthorizationSignedResponseKeyID: "12345",
			},
			&oauth2.JWTSession{
				JWTClaims: &fjwt.JWTClaims{
					Issuer: "https://original.example.com",
				},
			},
			url.Values{oidc.FormParameterAuthorizationCode: []string{"123"}},
			jwt.MapClaims{
				oidc.ClaimAudience:                  []any{"example"},
				oidc.ClaimIssuer:                    "https://original.example.com",
				oidc.FormParameterAuthorizationCode: "123",
			},
			"",
		},
		{
			"ShouldEncodeParametersAndRestoreIssuerFromSessionOpenID",
			"",
			&fjwt.DefaultSigner{
				GetPrivateKey: func(ctx context.Context) (any, error) {
					return keyRSA2048, nil
				},
			},
			&oidc.BaseClient{
				ID:                               "example",
				AuthorizationSignedResponseAlg:   oidc.SigningAlgRSAUsingSHA256,
				AuthorizationSignedResponseKeyID: "12345",
			},
			&openid.DefaultSession{
				Claims: &fjwt.IDTokenClaims{
					Issuer: "https://original.example.com",
				},
			},
			url.Values{oidc.FormParameterAuthorizationCode: []string{"123"}},
			jwt.MapClaims{
				oidc.ClaimAudience:                  []any{"example"},
				oidc.ClaimIssuer:                    "https://original.example.com",
				oidc.FormParameterAuthorizationCode: "123",
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{
				Issuers: oidc.IssuersConfig{JWTSecuredResponseMode: tc.issuer},
			}

			if tc.signer != nil {
				config.Signer = tc.signer
			}

			actual, err := oidc.EncodeJWTSecuredResponseParameters(oidc.GenerateJWTSecuredResponse(context.TODO(), config, tc.client, tc.session, tc.in))

			if tc.err != "" {
				assert.Nil(t, actual)
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)

				require.NotNil(t, actual)

				jarm := actual.Get(oidc.FormParameterResponse)

				require.NotEmpty(t, jarm)

				token, _, err := jwt.NewParser().ParseUnverified(jarm, jwt.MapClaims{})

				assert.NoError(t, err)
				require.NotNil(t, token)

				claims, ok := token.Claims.(jwt.MapClaims)

				require.True(t, ok)

				assert.NotEmpty(t, claims[oidc.ClaimJWTID])
				assert.NotEmpty(t, claims[oidc.ClaimExpirationTime])
				assert.NotEmpty(t, claims[oidc.ClaimIssuedAt])

				for claim, value := range tc.expected {
					switch claim {
					case oidc.ClaimJWTID, oidc.ClaimExpirationTime, oidc.ClaimIssuedAt:
						continue
					default:
						assert.Equal(t, value, claims[claim])
					}
				}
			}
		})
	}
}

func TestJWTSecuredAuthorizationResponseModeClaims_ToFrom(t *testing.T) {
	testCases := []struct {
		name      string
		have      oidc.JWTSecuredAuthorizationResponseModeClaims
		expected  map[string]any
		expectedx fjwt.MapClaims
	}{
		{
			"ShouldReturnMinimal",
			oidc.JWTSecuredAuthorizationResponseModeClaims{
				JTI:      "example",
				Audience: []string{},
				Extra:    map[string]any{},
			},
			map[string]any{
				oidc.ClaimAudience: []string{},
				oidc.ClaimJWTID:    "example",
			},
			fjwt.MapClaims{
				oidc.ClaimAudience: []string{},
				oidc.ClaimJWTID:    "example",
			},
		},
		{
			"ShouldReturnComprehensive",
			oidc.JWTSecuredAuthorizationResponseModeClaims{
				JTI:       "example",
				Issuer:    "https://auth.example.com",
				IssuedAt:  time.Unix(10000, 0).UTC(),
				ExpiresAt: time.Unix(10000+10, 0).UTC(),
				Audience:  []string{"example"},
				Extra: map[string]any{
					"abc": 123,
				},
			},
			map[string]any{
				oidc.ClaimJWTID:          "example",
				oidc.ClaimIssuer:         "https://auth.example.com",
				oidc.ClaimIssuedAt:       int64(10000),
				oidc.ClaimExpirationTime: int64(10010),
				oidc.ClaimAudience:       []string{"example"},
				"abc":                    123,
			},
			fjwt.MapClaims{
				oidc.ClaimJWTID:          "example",
				oidc.ClaimIssuer:         "https://auth.example.com",
				oidc.ClaimIssuedAt:       int64(10000),
				oidc.ClaimExpirationTime: int64(10010),
				oidc.ClaimAudience:       []string{"example"},
				"abc":                    123,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			havemap := tc.have.ToMap()
			havemapclaims := tc.have.ToMapClaims()

			assert.Equal(t, tc.expected, havemap)
			assert.Equal(t, tc.expectedx, havemapclaims)

			frommap := oidc.JWTSecuredAuthorizationResponseModeClaims{}

			frommap.FromMap(havemap)

			frommampclaims := oidc.JWTSecuredAuthorizationResponseModeClaims{}

			frommampclaims.FromMapClaims(havemapclaims)

			assert.Equal(t, tc.have, frommap)
			assert.Equal(t, tc.have, frommampclaims)
		})
	}

	claims := oidc.JWTSecuredAuthorizationResponseModeClaims{}

	mapclaims := claims.ToMap()

	assert.NotEmpty(t, mapclaims[oidc.ClaimJWTID])

	assert.Nil(t, claims.Get(oidc.ClaimUsername))

	claims.Add(oidc.ClaimUsername, "example")

	assert.Equal(t, "example", claims.Get(oidc.ClaimUsername))
}
