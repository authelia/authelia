package oidc_test

import (
	"sort"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestSortedJSONWebKey(t *testing.T) {
	testCases := []struct {
		name     string
		have     []jose.JSONWebKey
		expected []jose.JSONWebKey
	}{
		{
			"ShouldOrderByKID",
			[]jose.JSONWebKey{
				{KeyID: "abc"},
				{KeyID: "123"},
			},
			[]jose.JSONWebKey{
				{KeyID: "123"},
				{KeyID: "abc"},
			},
		},
		{
			"ShouldOrderByAlg",
			[]jose.JSONWebKey{
				{Algorithm: "RS256"},
				{Algorithm: "HS256"},
			},
			[]jose.JSONWebKey{
				{Algorithm: "HS256"},
				{Algorithm: "RS256"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(oidc.SortedJSONWebKey(tc.have))

			assert.Equal(t, tc.expected, tc.have)
		})
	}
}

func TestIntrospectionResponseToMap(t *testing.T) {
	testCases := []struct {
		name        string
		have        fosite.IntrospectionResponder
		expectedaud []string
		expected    map[string]any
	}{
		{
			"ShouldDecodeInactive",
			&oidc.IntrospectionResponse{},
			nil,
			map[string]any{oidc.ClaimActive: false},
		},
		{
			"ShouldReturnInactiveWhenNil",
			nil,
			nil,
			map[string]any{oidc.ClaimActive: false},
		},
		{
			"ShouldReturnActiveWithoutAccessRequester",
			&oidc.IntrospectionResponse{
				Active: true,
			},
			nil,
			map[string]any{oidc.ClaimActive: true},
		},
		{
			"ShouldReturnActiveWithAccessRequester",
			&oidc.IntrospectionResponse{
				Active: true,
				AccessRequester: &fosite.AccessRequest{
					Request: fosite.Request{
						RequestedAt:     time.Unix(100000, 0).UTC(),
						GrantedScope:    fosite.Arguments{oidc.ScopeOpenID, oidc.ScopeProfile},
						GrantedAudience: fosite.Arguments{"https://example.com", "aclient"},
						Client:          &oidc.BaseClient{ID: "aclient"},
					},
				},
			},
			nil,
			map[string]any{
				oidc.ClaimActive:           true,
				oidc.ClaimScope:            "openid profile",
				oidc.ClaimAudience:         []string{"https://example.com", "aclient"},
				oidc.ClaimIssuedAt:         int64(100000),
				oidc.ClaimClientIdentifier: "aclient",
			},
		},
		{
			"ShouldReturnActiveWithAccessRequesterAndSession",
			&oidc.IntrospectionResponse{
				Active: true,
				AccessRequester: &fosite.AccessRequest{
					Request: fosite.Request{
						RequestedAt:     time.Unix(100000, 0).UTC(),
						GrantedScope:    fosite.Arguments{oidc.ScopeOpenID, oidc.ScopeProfile},
						GrantedAudience: fosite.Arguments{"https://example.com", "aclient"},
						Client:          &oidc.BaseClient{ID: "aclient"},
						Session: &model.OpenIDSession{
							DefaultSession: &openid.DefaultSession{
								ExpiresAt: map[fosite.TokenType]time.Time{
									fosite.AccessToken: time.Unix(1000000, 0).UTC(),
								},
								Subject: "asubj",
								Claims: &jwt.IDTokenClaims{
									Extra: map[string]any{
										"aclaim":                 1,
										oidc.ClaimExpirationTime: 0,
									},
								},
							},
						},
					},
				},
			},
			nil,
			map[string]any{
				oidc.ClaimActive:           true,
				oidc.ClaimScope:            "openid profile",
				oidc.ClaimAudience:         []string{"https://example.com", "aclient"},
				oidc.ClaimIssuedAt:         int64(100000),
				oidc.ClaimClientIdentifier: "aclient",
				"aclaim":                   1,
				oidc.ClaimSubject:          "asubj",
				oidc.ClaimExpirationTime:   int64(1000000),
			},
		},
		{
			"ShouldReturnActiveWithAccessRequesterAndSessionWithIDTokenClaimsAndUsername",
			&oidc.IntrospectionResponse{
				Client: &oidc.BaseClient{
					ID:       "rclient",
					Audience: []string{"https://rs.example.com"},
				},
				Active: true,
				AccessRequester: &fosite.AccessRequest{
					Request: fosite.Request{
						RequestedAt:     time.Unix(100000, 0).UTC(),
						GrantedScope:    fosite.Arguments{oidc.ScopeOpenID, oidc.ScopeProfile},
						GrantedAudience: fosite.Arguments{"https://example.com", "aclient"},
						Client:          &oidc.BaseClient{ID: "aclient"},
						Session: &model.OpenIDSession{
							DefaultSession: &openid.DefaultSession{
								ExpiresAt: map[fosite.TokenType]time.Time{
									fosite.AccessToken: time.Unix(1000000, 0).UTC(),
								},
								Username: "auser",
								Claims: &jwt.IDTokenClaims{
									Subject: "asubj",
									Extra: map[string]any{
										"aclaim":                 1,
										oidc.ClaimExpirationTime: 0,
									},
								},
							},
						},
					},
				},
			},
			[]string{"https://rs.example.com", "rclient"},
			map[string]any{
				oidc.ClaimActive:           true,
				oidc.ClaimScope:            "openid profile",
				oidc.ClaimAudience:         []string{"https://example.com", "aclient"},
				oidc.ClaimIssuedAt:         int64(100000),
				oidc.ClaimClientIdentifier: "aclient",
				"aclaim":                   1,
				oidc.ClaimSubject:          "asubj",
				oidc.ClaimExpirationTime:   int64(1000000),
				oidc.ClaimUsername:         "auser",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			aud, introspection := oidc.IntrospectionResponseToMap(tc.have)

			assert.Equal(t, tc.expectedaud, aud)
			assert.Equal(t, tc.expected, introspection)
		})
	}
}
