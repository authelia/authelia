package oidc_test

import (
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	fjwt "github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"gopkg.in/square/go-jose.v2"

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
				{KeyID: abc},
				{KeyID: "123"},
			},
			[]jose.JSONWebKey{
				{KeyID: "123"},
				{KeyID: abc},
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

func TestRFC6750Header(t *testing.T) {
	testCaes := []struct {
		name     string
		have     *fosite.RFC6749Error
		realm    string
		scope    string
		expected string
	}{
		{
			"ShouldEncodeAll",
			&fosite.RFC6749Error{
				ErrorField:       "invalid_example",
				DescriptionField: "A description",
			},
			"abc",
			"openid",
			`realm="abc",error="invalid_example",error_description="A description",scope="openid"`,
		},
		{
			"ShouldEncodeBasic",
			&fosite.RFC6749Error{
				ErrorField:       "invalid_example",
				DescriptionField: "A description",
			},
			"",
			"",
			`error="invalid_example",error_description="A description"`,
		},
	}

	for _, tc := range testCaes {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.RFC6750Header(tc.realm, tc.scope, tc.have))
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
						Session: &oidc.Session{
							DefaultSession: &openid.DefaultSession{
								ExpiresAt: map[fosite.TokenType]time.Time{
									fosite.AccessToken: time.Unix(1000000, 0).UTC(),
								},
								Subject: "asubj",
								Claims: &fjwt.IDTokenClaims{
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
						Session: &oidc.Session{
							DefaultSession: &openid.DefaultSession{
								ExpiresAt: map[fosite.TokenType]time.Time{
									fosite.AccessToken: time.Unix(1000000, 0).UTC(),
								},
								Username: "auser",
								Claims: &fjwt.IDTokenClaims{
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
			[]string{"rclient"},
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

func TestIsJWTProfileAccessToken(t *testing.T) {
	testCases := []struct {
		name     string
		have     *fjwt.Token
		expected bool
	}{
		{
			"ShouldReturnFalseOnNilToken",
			nil,
			false,
		},
		{
			"ShouldReturnFalseOnNilTokenHeader",
			&fjwt.Token{Header: nil},
			false,
		},
		{
			"ShouldReturnFalseOnEmptyHeader",
			&fjwt.Token{Header: map[string]any{}},
			false,
		},
		{
			"ShouldReturnFalseOnInvalidKeyTypeHeaderType",
			&fjwt.Token{Header: map[string]any{oidc.JWTHeaderKeyType: 123}},
			false,
		},
		{
			"ShouldReturnFalseOnInvalidKeyTypeHeaderValue",
			&fjwt.Token{Header: map[string]any{oidc.JWTHeaderKeyType: "JWT"}},
			false,
		},
		{
			"ShouldReturnTrue",
			&fjwt.Token{Header: map[string]any{oidc.JWTHeaderKeyType: oidc.JWTHeaderTypeValueAccessTokenJWT}},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.IsJWTProfileAccessToken(tc.have))
		})
	}
}

func TestGetLangFromRequester(t *testing.T) {
	testCases := []struct {
		name     string
		have     fosite.Requester
		expected language.Tag
	}{
		{
			"ShouldReturnDefault",
			&TestGetLangRequester{},
			language.English,
		},
		{
			"ShouldReturnEmpty",
			&fosite.Request{},
			language.Tag{},
		},
		{
			"ShouldReturnValueFromRequest",
			&fosite.Request{
				Lang: language.French,
			},
			language.French,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.GetLangFromRequester(tc.have))
		})
	}
}

type TestGetLangRequester struct {
}

func (t TestGetLangRequester) SetID(id string) {}

func (t TestGetLangRequester) GetID() string {
	return ""
}

func (t TestGetLangRequester) GetRequestedAt() (requestedAt time.Time) {
	return time.Now().UTC()
}

func (t TestGetLangRequester) GetClient() (client fosite.Client) {
	return nil
}

func (t TestGetLangRequester) GetRequestedScopes() (scopes fosite.Arguments) {
	return nil
}

func (t TestGetLangRequester) GetRequestedAudience() (audience fosite.Arguments) {
	return nil
}

func (t TestGetLangRequester) SetRequestedScopes(scopes fosite.Arguments) {}

func (t TestGetLangRequester) SetRequestedAudience(audience fosite.Arguments) {}

func (t TestGetLangRequester) AppendRequestedScope(scope string) {}

func (t TestGetLangRequester) GetGrantedScopes() (grantedScopes fosite.Arguments) {
	return nil
}

func (t TestGetLangRequester) GetGrantedAudience() (grantedAudience fosite.Arguments) {
	return nil
}

func (t TestGetLangRequester) GrantScope(scope string) {}

func (t TestGetLangRequester) GrantAudience(audience string) {}

func (t TestGetLangRequester) GetSession() (session fosite.Session) {
	return nil
}

func (t TestGetLangRequester) SetSession(session fosite.Session) {}

func (t TestGetLangRequester) GetRequestForm() url.Values {
	return nil
}

func (t TestGetLangRequester) Merge(requester fosite.Requester) {}

func (t TestGetLangRequester) Sanitize(allowedParameters []string) fosite.Requester {
	return nil
}
