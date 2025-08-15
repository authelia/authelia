package oidc_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestClaimsMarshall(t *testing.T) {
	r := &oidc.ClaimsRequests{
		IDToken: map[string]*oidc.ClaimRequest{
			"example": nil,
			"another": nil,
		},
		UserInfo: map[string]*oidc.ClaimRequest{
			"example": nil,
			"another": nil,
		},
	}

	o := r.ToOrdered()

	data, err := json.Marshal(o)
	require.NoError(t, err)

	assert.Equal(t, `{"id_token":{"another":null,"example":null},"userinfo":{"another":null,"example":null}}`, string(data))
}

func TestClaimValidate(t *testing.T) {
	config := schema.IdentityProvidersOpenIDConnect{
		Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
			"example": {
				Claims: []string{oidc.ClaimPreferredUsername, oidc.ClaimFullName, oidc.ClaimEmailAlts, oidc.ClaimPhoneNumber},
			},
		},
		Clients: []schema.IdentityProvidersOpenIDConnectClient{
			{
				ID:           "example-client",
				Scopes:       []string{oidc.ScopeOpenID, "example"},
				ClaimsPolicy: "default",
			},
		},
		ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
			"default": {},
		},
	}

	strategy := oidc.NewCustomClaimsStrategyFromClient(config.Clients[0], config.Scopes, config.ClaimsPolicies)

	client := &oidc.RegisteredClient{
		ID:     config.Clients[0].ID,
		Scopes: config.Clients[0].Scopes,
	}

	requests := &oidc.ClaimsRequests{
		IDToken: map[string]*oidc.ClaimRequest{
			oidc.ClaimFullName: nil,
			oidc.ClaimPreferredUsername: {
				Essential: true,
			},
		},
	}

	ctx := &TestContext{}

	detailer := &testDetailer{
		username: "john",
		groups:   []string{"abc", "123"},
		name:     "John Smith",
		emails:   []string{"john.smith@authelia.com"},
		extra:    map[string]any{},
	}

	extra := map[string]any{}

	err := strategy.HydrateIDTokenClaims(ctx, oauthelia2.ExactScopeStrategy, client, nil, []string{oidc.ClaimPreferredUsername, oidc.ClaimFullName}, requests.IDToken, detailer, time.Now(), time.Now(), nil, extra, false)
	assert.NoError(t, oauthelia2.ErrorToDebugRFC6749Error(err))

	assert.Equal(t, map[string]any{oidc.ClaimAudience: []string{config.Clients[0].ID}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john"}, extra)

	err = strategy.ValidateClaimsRequests(ctx, oauthelia2.ExactScopeStrategy, client, requests)

	assert.NoError(t, oauthelia2.ErrorToDebugRFC6749Error(err))
}

func TestClaimRequest_String(t *testing.T) {
	testCases := []struct {
		name     string
		have     *oidc.ClaimRequest
		expected string
	}{
		{
			"ShouldHandleBlank",
			&oidc.ClaimRequest{},
			"essential 'false'",
		},
		{
			"ShouldHandleNil",
			nil,
			"",
		},
		{
			"ShouldHandleValue",
			&oidc.ClaimRequest{Value: "example"},
			"value 'example', essential 'false'",
		},
		{
			"ShouldHandleValue",
			&oidc.ClaimRequest{Essential: true},
			"essential 'true'",
		},
		{
			"ShouldHandleValues",
			&oidc.ClaimRequest{Essential: true, Values: []any{"abc", "123"}},
			"values ['abc','123'], essential 'true'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.String())
		})
	}
}

func TestClaimRequest_Matches(t *testing.T) {
	testCases := []struct {
		name     string
		have     *oidc.ClaimRequest
		value    any
		expected bool
	}{
		{
			"ShouldMatchUndefined",
			&oidc.ClaimRequest{},
			"abc",
			true,
		},
		{
			"ShouldMatchNil",
			nil,
			"abc",
			true,
		},
		{
			"ShouldMatchInt64ToInt",
			&oidc.ClaimRequest{Value: 123},
			int64(123),
			true,
		},
		{
			"ShouldMatchInt64ToIntArray",
			&oidc.ClaimRequest{Values: []any{123}},
			int64(123),
			true,
		},
		{
			"ShouldMatchInt64ToInt",
			&oidc.ClaimRequest{Value: 123},
			int64(123),
			true,
		},
		{
			"ShouldMatchStringToString",
			&oidc.ClaimRequest{Value: "abc"},
			"abc",
			true,
		},
		{
			"ShouldMatchStringToStringArray",
			&oidc.ClaimRequest{Values: []any{"abc"}},
			"abc",
			true,
		},
		{
			"ShouldMatchBoolToBool",
			&oidc.ClaimRequest{Value: true},
			true,
			true,
		},
		{
			"ShouldNotMatchBoolToBool",
			&oidc.ClaimRequest{Value: true},
			false,
			false,
		},
		{
			"ShouldMatchBoolToBoolArray",
			&oidc.ClaimRequest{Values: []any{true}},
			true,
			true,
		},
		{
			"ShouldNotMatchBoolToBoolArray",
			&oidc.ClaimRequest{Values: []any{false}},
			true,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.Matches(tc.value))
		})
	}
}

func TestNewClaimRequestsMatcher(t *testing.T) {
	attribute := "example-attribute"
	policy := "example-policy"
	scope := "example-scope"
	scopes := []string{"example-scope"}
	claim := "example-claim"

	config := schema.IdentityProvidersOpenIDConnect{
		Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
			scope: {
				Claims: []string{claim},
			},
		},
		ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
			policy: {
				CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
					claim: {
						Name:      claim,
						Attribute: attribute,
					},
				},
			},
		},
		Clients: []schema.IdentityProvidersOpenIDConnectClient{
			{
				ID:           "example-client",
				Scopes:       scopes,
				ClaimsPolicy: policy,
			},
		},
	}

	client := &oidc.RegisteredClient{
		ID:             config.Clients[0].ID,
		ClaimsStrategy: oidc.NewCustomClaimsStrategyFromClient(config.Clients[0], config.Scopes, config.ClaimsPolicies),
		Scopes:         config.Clients[0].Scopes,
	}

	strategy := oauthelia2.ExactScopeStrategy
	start := time.Unix(123123123, 0)
	requested := time.Unix(123123100, 0)
	ctx := &TestContext{}

	var (
		extra map[string]any
	)

	srcNumbers := []any{
		float64(123),
		float32(123),
		int64(123),
		int32(123),
		int16(123),
		int8(123),
		123,
		uint64(123),
		uint32(123),
		uint16(123),
		uint8(123),
		uint(123),
	}

	testCases := []struct {
		name       string
		claims     []string
		requested  any
		detailer   any
		expectedID any
		expectedUI any
		numbers    bool
		err        string
	}{
		{
			"ShouldPassString",
			[]string{"example-claim"},
			"apple",
			"apple",
			"apple",
			"apple",
			false,
			"",
		},
		{
			"ShouldNotPassNonEssentialString",
			[]string{},
			"apple",
			"apple",
			nil,
			"apple",
			false,
			"",
		},
		{
			"ShouldNotPassMismatchedTypes",
			[]string{},
			"apple",
			123,
			nil,
			123,
			false,
			"",
		},
		{
			name:    "ShouldPassNumbers",
			claims:  []string{"example-claim"},
			numbers: true,
		},
	}

	caser := cases.Title(language.English)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.numbers {
				for _, src := range srcNumbers {
					for _, dst := range srcNumbers {
						st := reflect.TypeOf(src)
						dt := reflect.TypeOf(dst)

						t.Run(fmt.Sprintf("%sTo%s", caser.String(st.String()), caser.String(dt.String())), func(t *testing.T) {
							extra = make(map[string]any)

							detailer := &testDetailer{
								username: "john",
								groups:   []string{"abc", "123"},
								name:     "John Smith",
								emails:   []string{"john.smith@authelia.com"},
								extra: map[string]any{
									attribute: dst,
								},
							}

							requests := &oidc.ClaimsRequests{
								IDToken: map[string]*oidc.ClaimRequest{
									claim: {
										Value: src,
									},
								},
								UserInfo: map[string]*oidc.ClaimRequest{
									claim: {
										Value: src,
									},
								},
							}

							assert.NoError(t, oauthelia2.ErrorToDebugRFC6749Error(client.ClaimsStrategy.HydrateIDTokenClaims(ctx, strategy, client, scopes, tc.claims, requests.IDToken, detailer, requested, start, map[string]any{}, extra, false)))

							assert.Equal(t, dst, extra[claim])

							extra = make(map[string]any)

							assert.NoError(t, client.ClaimsStrategy.HydrateUserInfoClaims(ctx, strategy, client, scopes, tc.claims, requests.UserInfo, detailer, requested, start, map[string]any{}, extra))

							assert.Equal(t, dst, extra[claim])

							requests = &oidc.ClaimsRequests{
								IDToken: map[string]*oidc.ClaimRequest{
									claim: {
										Values: []any{src},
									},
								},
								UserInfo: map[string]*oidc.ClaimRequest{
									claim: {
										Values: []any{src},
									},
								},
							}

							assert.NoError(t, client.ClaimsStrategy.HydrateIDTokenClaims(ctx, strategy, client, scopes, tc.claims, requests.IDToken, detailer, requested, start, map[string]any{}, extra, false))

							assert.Equal(t, dst, extra[claim])

							extra = make(map[string]any)

							assert.NoError(t, client.ClaimsStrategy.HydrateUserInfoClaims(ctx, strategy, client, scopes, tc.claims, requests.UserInfo, detailer, requested, start, map[string]any{}, extra))

							assert.Equal(t, dst, extra[claim])
						})
					}
				}

				return
			}

			extra = make(map[string]any)

			detailer := &testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					attribute: tc.detailer,
				},
			}

			requests := &oidc.ClaimsRequests{
				IDToken: map[string]*oidc.ClaimRequest{
					claim: {
						Value: tc.requested,
					},
				},
				UserInfo: map[string]*oidc.ClaimRequest{
					claim: {
						Value: tc.requested,
					},
				},
			}

			err := client.ClaimsStrategy.HydrateIDTokenClaims(ctx, strategy, client, scopes, tc.claims, requests.IDToken, detailer, requested, start, map[string]any{}, extra, false)

			if tc.err == "" {
				require.NoError(t, oauthelia2.ErrorToDebugRFC6749Error(err))
				assert.Equal(t, tc.expectedID, extra[claim])
			} else {
				assert.EqualError(t, oauthelia2.ErrorToDebugRFC6749Error(err), tc.err)
			}

			extra = make(map[string]any)

			err = client.ClaimsStrategy.HydrateUserInfoClaims(ctx, strategy, client, scopes, tc.claims, requests.UserInfo, detailer, requested, start, map[string]any{}, extra)

			if tc.err == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedUI, extra[claim])
			} else {
				assert.EqualError(t, oauthelia2.ErrorToDebugRFC6749Error(err), tc.err)
			}
		})
	}
}

func TestNewClaimRequests(t *testing.T) {
	testCases := []struct {
		name              string
		have              any
		err               string
		expected          bool
		subject           string
		issuer            *url.URL
		badSubjects       []string
		badIssuers        []*url.URL
		claims, essential []string
		idToken, userinfo map[string]*oidc.ClaimRequest
	}{
		{
			"ShouldParse",
			`{"id_token":{"sub":{"value":"aaaa"}}}`,
			"",
			true,
			"aaaa",
			&url.URL{},
			[]string{"nah", "aaa"},
			[]*url.URL{},
			[]string{"sub"},
			nil,
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value: "aaaa",
				},
			},
			nil,
		},
		{
			"ShouldParseUserInfo",
			`{"userinfo":{"sub":{"value":"aaaa"}}}`,
			"",
			true,
			"aaaa",
			&url.URL{},
			[]string{"nah", "aaa"},
			[]*url.URL{},
			[]string{"sub"},
			nil,
			nil,
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value: "aaaa",
				},
			},
		},
		{
			"ShouldParseBoth",
			`{"userinfo":{"sub":{"value":"aaaa"}},"id_token":{"sub":{"value":"aaaa"}}}`,
			"",
			true,
			"aaaa",
			&url.URL{},
			[]string{"nah", "aaa"},
			[]*url.URL{},
			[]string{"sub"},
			nil,
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value: "aaaa",
				},
			},
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value: "aaaa",
				},
			},
		},
		{
			"ShouldParseBothIDTokenEssential",
			`{"userinfo":{"sub":{"value":"aaaa"}},"id_token":{"sub":{"value":"aaaa","essential":true}}}`,
			"",
			true,
			"aaaa",
			&url.URL{},
			[]string{"nah", "aaa"},
			[]*url.URL{},
			nil,
			[]string{"sub"},
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value:     "aaaa",
					Essential: true,
				},
			},
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value: "aaaa",
				},
			},
		},
		{
			"ShouldParseBothUserInfoEssential",
			`{"userinfo":{"sub":{"value":"aaaa","essential":true}},"id_token":{"sub":{"value":"aaaa"}}}`,
			"",
			true,
			"aaaa",
			&url.URL{},
			[]string{"nah", "aaa"},
			[]*url.URL{},
			nil,
			[]string{"sub"},
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value: "aaaa",
				},
			},
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value:     "aaaa",
					Essential: true,
				},
			},
		},
		{
			"ShouldParseUserInfoEssential",
			`{"userinfo":{"sub":{"value":"aaaa","essential":true}}}`,
			"",
			true,
			"aaaa",
			&url.URL{},
			[]string{"nah", "aaa"},
			[]*url.URL{},
			nil,
			[]string{"sub"},
			nil,
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value:     "aaaa",
					Essential: true,
				},
			},
		},
		{
			"ShouldParseBothEssential",
			`{"userinfo":{"sub":{"value":"aaaa","essential":true}},"id_token":{"sub":{"value":"aaaa","essential":true}}}`,
			"",
			true,
			"aaaa",
			&url.URL{},
			[]string{"nah", "aaa"},
			[]*url.URL{},
			nil,
			[]string{"sub"},
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value:     "aaaa",
					Essential: true,
				},
			},
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value:     "aaaa",
					Essential: true,
				},
			},
		},
		{
			"ShouldNotMatchInteger",
			`{"userinfo":{"sub":{"value":1}}}`,
			"",
			false,
			"",
			&url.URL{},
			[]string{},
			[]*url.URL{},
			[]string{"sub"},
			nil,
			nil,
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value: float64(1),
				},
			},
		},
		{
			"ShouldNotMatchIntegerIDToken",
			`{"id_token":{"sub":{"value":1}}}`,
			"",
			false,
			"",
			&url.URL{},
			[]string{},
			[]*url.URL{},
			[]string{"sub"},
			nil,
			map[string]*oidc.ClaimRequest{
				"sub": {
					Value: float64(1),
				},
			},
			nil,
		},
		{
			"ShouldParseRequestOnly",
			`{"userinfo":{"sub":null}}`,
			"",
			true,
			"",
			&url.URL{},
			[]string{},
			[]*url.URL{},
			[]string{"sub"},
			nil,
			nil,
			map[string]*oidc.ClaimRequest{
				"sub": nil,
			},
		},
		{
			"ShouldNotParse",
			`{"iaaa"}}}`,
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The OAuth 2.0 client included a malformed 'claims' parameter in the authorization request. Error occurred attempting to parse the 'claims' parameter: invalid character '}' after object key.",
			false,
			"",
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{}

			if tc.have != nil {
				form.Set(oidc.FormParameterClaims, tc.have.(string))
			}

			var (
				requested         string
				ok                bool
				requests          *oidc.ClaimsRequests
				err               error
				claims, essential []string
			)

			claims, essential = requests.ToSlices()

			assert.Nil(t, claims)
			assert.Nil(t, essential)

			assert.Nil(t, requests.ToSlice())

			assert.Nil(t, requests.GetIDTokenRequests())
			assert.Nil(t, requests.GetUserInfoRequests())

			requested, ok = requests.MatchesIssuer(nil)
			assert.Equal(t, "", requested)
			assert.True(t, ok)

			requested, ok = requests.MatchesSubject("xxxx")
			assert.Equal(t, "", requested)
			assert.True(t, ok)

			requests, err = oidc.NewClaimRequests(form)

			if tc.err != "" {
				assert.Nil(t, requests)
				assert.EqualError(t, oauthelia2.ErrorToDebugRFC6749Error(err), tc.err)

				return
			}

			require.NotNil(t, requests)

			claims, essential = requests.ToSlices()

			assert.Equal(t, tc.claims, claims)
			assert.Equal(t, tc.essential, essential)

			allClaims := requests.ToSlice()

			assert.True(t, utils.IsStringSliceContainsAll(claims, allClaims))
			assert.True(t, utils.IsStringSliceContainsAll(essential, allClaims))

			requested, ok = requests.MatchesSubject(tc.subject)
			assert.Equal(t, tc.subject, requested)
			assert.Equal(t, tc.expected, ok)

			requested, ok = requests.MatchesIssuer(tc.issuer)
			assert.Equal(t, tc.issuer.String(), requested)
			assert.True(t, ok)

			for _, badSubject := range tc.badSubjects {
				requested, ok = requests.MatchesSubject(badSubject)

				assert.Equal(t, tc.subject, requested)
				assert.False(t, ok)
			}

			for _, badIssuer := range tc.badIssuers {
				requested, ok = requests.MatchesIssuer(badIssuer)

				assert.Equal(t, tc.issuer.String(), requested)
				assert.False(t, ok)
			}

			assert.Equal(t, tc.idToken, requests.GetIDTokenRequests())
			assert.Equal(t, tc.userinfo, requests.GetUserInfoRequests())
		})
	}

	form := url.Values{}

	form.Set(oidc.FormParameterClaims, `{"id_token":{"sub":{"value":"aaaa"}}}`)

	requests, err := oidc.NewClaimRequests(form)
	require.NoError(t, err)

	assert.NotNil(t, requests)

	var (
		requested string
		ok        bool
	)

	requested, ok = requests.MatchesSubject("aaaa")
	assert.Equal(t, "aaaa", requested)
	assert.True(t, ok)

	requested, ok = requests.MatchesSubject("aaaaa")
	assert.Equal(t, "aaaa", requested)
	assert.False(t, ok)
}

func TestNewCustomClaimsStrategy(t *testing.T) {
	testCases := []struct {
		name             string
		config           schema.IdentityProvidersOpenIDConnect
		scopes           oauthelia2.Arguments
		requests         *oidc.ClaimsRequests
		detailer         oidc.UserDetailer
		implicit         bool
		err              string
		claims           []string
		original         map[string]any
		expectedIDToken  map[string]any
		expectedUserInfo map[string]any
		errIDToken       string
		errUserInfo      string
	}{
		{
			"ShouldGrantRequestedClaimsCustom",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			nil,
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			false,
			"",
			[]string{},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldGrantRequestedClaimsIDToken",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
					"alt-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope", "alt-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			&oidc.ClaimsRequests{
				IDToken:  map[string]*oidc.ClaimRequest{"example-claim": nil},
				UserInfo: nil,
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			false,
			"",
			[]string{},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldGrantRequestedClaimsIDTokenRequestedValue",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
					"alt-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope", "alt-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			&oidc.ClaimsRequests{
				IDToken:  map[string]*oidc.ClaimRequest{"example-claim": {Value: float64(123)}},
				UserInfo: nil,
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			false,
			"",
			[]string{"example-claim"},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}, "example-claim": 123},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldGrantRequestedClaimsIDTokenRequestedValueFloat64",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
					"alt-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope", "alt-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			&oidc.ClaimsRequests{
				IDToken:  map[string]*oidc.ClaimRequest{"example-claim": {Value: float64(1234)}},
				UserInfo: nil,
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": float64(1234),
				},
			},
			false,
			"",
			[]string{"example-claim"},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}, "example-claim": float64(1234)},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": float64(1234), oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldGrantRequestedClaimsIDTokenRequestedValue",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
					"alt-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope", "alt-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			&oidc.ClaimsRequests{
				IDToken:  map[string]*oidc.ClaimRequest{"example-claim": {Value: float64(1234)}},
				UserInfo: nil,
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": int64(1234),
				},
			},
			false,
			"",
			[]string{"example-claim"},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}, "example-claim": int64(1234)},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": int64(1234), oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldGrantNoPolicy",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
					"alt-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy2": {
						IDToken:     []string{},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope", "alt-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			&oidc.ClaimsRequests{
				IDToken:  map[string]*oidc.ClaimRequest{"example-claim": nil},
				UserInfo: nil,
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			false,
			"",
			[]string{},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldGrantRequestedClaimsIDTokenButNotPropagateIDTokenClaims",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			&oidc.ClaimsRequests{
				IDToken:  map[string]*oidc.ClaimRequest{"example-claim": nil},
				UserInfo: nil,
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			false,
			"",
			[]string{"example-claim"},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimJWTID: "EXAMPLE", oidc.ClaimFullName: "Not John Smith"},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}, "example-claim": 123},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldNotGrantRequestedClaimsIDTokenButNotPropagateIDTokenClaimsNotGranted",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			&oidc.ClaimsRequests{
				IDToken:  map[string]*oidc.ClaimRequest{"example-claim": nil},
				UserInfo: nil,
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			false,
			"",
			[]string{},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimJWTID: "EXAMPLE", oidc.ClaimFullName: "Not John Smith"},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldGrantRequestedClaimsUserInfo",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim", "example-claim-alt"},
					},
					"alt-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
							"example-claim-alt": {
								Name:      "example-claim-alt",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope", "alt-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope", "alt-scope"},
			&oidc.ClaimsRequests{
				IDToken:  nil,
				UserInfo: map[string]*oidc.ClaimRequest{"example-claim": nil, "example-claim-alt": nil},
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			false,
			"",
			[]string{},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": 123, "example-claim-alt": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldGrantRequestedClaimsUserInfoButNotPropagateIDTokenClaims",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			&oidc.ClaimsRequests{
				IDToken:  nil,
				UserInfo: map[string]*oidc.ClaimRequest{"example-claim": nil},
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			false,
			"",
			[]string{},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimJWTID: "EXAMPLE", oidc.ClaimFullName: "Not John Smith"},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldIncludeCustomClaimsInIDTokenFromPolicy",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{"example-claim"},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			nil,
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			false,
			"",
			[]string{},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}, "example-claim": 123},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldIncludeRequestedAtInIDTokenWhenExpresslyRequested",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{"example-claim"},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			&oidc.ClaimsRequests{
				IDToken:  map[string]*oidc.ClaimRequest{oidc.ClaimRequestedAt: nil},
				UserInfo: nil,
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			false,
			"",
			[]string{},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}, "example-claim": 123, oidc.ClaimRequestedAt: int64(123123100)},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
		{
			"ShouldIncludeScopesForImplicitIDToken",
			schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"example-scope": {
						Claims: []string{"example-claim"},
					},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example-policy": {
						IDToken:     []string{"example-claim"},
						AccessToken: []string{},
						CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
							"example-claim": {
								Name:      "example-claim",
								Attribute: "example-claim",
							},
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "example-client",
						Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
						ClaimsPolicy: "example-policy",
					},
				},
			},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, "example-scope"},
			&oidc.ClaimsRequests{
				IDToken:  nil,
				UserInfo: nil,
			},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
				extra: map[string]any{
					"example-claim": 123,
				},
			},
			true,
			"",
			[]string{},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimAudience: []string{"example-client"}, oidc.ClaimEmail: "john.smith@authelia.com", "email_verified": true, "example-claim": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123)},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, "example-claim": 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: int64(123123123), oidc.ClaimRequestedAt: int64(123123100)},
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Len(t, tc.config.Clients, 1)

			client := &oidc.RegisteredClient{
				ID:             tc.config.Clients[0].ID,
				ClaimsStrategy: oidc.NewCustomClaimsStrategyFromClient(tc.config.Clients[0], tc.config.Scopes, tc.config.ClaimsPolicies),
				Scopes:         tc.config.Clients[0].Scopes,
			}

			strategy := oauthelia2.ExactScopeStrategy
			start := time.Unix(123123123, 0)
			requested := time.Unix(123123100, 0)

			ctx := &TestContext{}

			if tc.err != "" {
				assert.EqualError(t, client.ClaimsStrategy.ValidateClaimsRequests(ctx, strategy, client, tc.requests), tc.err)

				return
			}

			assert.NoError(t, client.ClaimsStrategy.ValidateClaimsRequests(ctx, strategy, client, tc.requests))

			var (
				extra    map[string]any
				requests map[string]*oidc.ClaimRequest
			)

			if tc.requests != nil && tc.requests.IDToken != nil {
				requests = tc.requests.IDToken
			} else {
				requests = make(map[string]*oidc.ClaimRequest)
			}

			extra = make(map[string]any)

			err := client.ClaimsStrategy.HydrateIDTokenClaims(ctx, strategy, client, tc.scopes, tc.claims, requests, tc.detailer, requested, start, tc.original, extra, tc.implicit)

			if tc.errIDToken == "" {
				assert.NoError(t, oauthelia2.ErrorToDebugRFC6749Error(err))
			} else {
				assert.EqualError(t, oauthelia2.ErrorToDebugRFC6749Error(err), tc.errIDToken)
			}

			assert.Equal(t, tc.expectedIDToken, extra)

			if tc.requests != nil && tc.requests.UserInfo != nil {
				requests = tc.requests.UserInfo
			} else {
				requests = make(map[string]*oidc.ClaimRequest)
			}

			extra = make(map[string]any)

			err = client.ClaimsStrategy.HydrateUserInfoClaims(ctx, strategy, client, tc.scopes, tc.claims, requests, tc.detailer, requested, start, tc.original, extra)

			if tc.errUserInfo == "" {
				assert.NoError(t, oauthelia2.ErrorToDebugRFC6749Error(err))
			} else {
				assert.EqualError(t, oauthelia2.ErrorToDebugRFC6749Error(err), tc.errUserInfo)
			}

			assert.Equal(t, tc.expectedUserInfo, extra)
		})
	}
}

func TestGrantScopeAudienceConsent(t *testing.T) {
	testCases := []struct {
		name             string
		ar               oauthelia2.Requester
		consent          *model.OAuth2ConsentSession
		expected         bool
		expectedScope    oauthelia2.Arguments
		expectedAudience oauthelia2.Arguments
	}{
		{
			"ShouldGrant",
			&oauthelia2.Request{},
			&model.OAuth2ConsentSession{
				GrantedScopes:   []string{"abc"},
				GrantedAudience: []string{"ad"},
			},
			true,
			[]string{"abc"},
			[]string{"ad"},
		},
		{
			"ShouldNotGrant",
			&oauthelia2.Request{},
			&model.OAuth2ConsentSession{
				GrantedScopes:   []string{},
				GrantedAudience: []string{},
			},
			true,
			nil,
			nil,
		},
		{
			"ShouldNotGrantNilConsent",
			&oauthelia2.Request{},
			nil,
			true,
			nil,
			nil,
		},
		{
			"ShouldNotGrantNilRequest",
			nil,
			&model.OAuth2ConsentSession{
				GrantedScopes:   []string{},
				GrantedAudience: []string{},
			},
			false,
			nil,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oidc.GrantScopeAudienceConsent(tc.ar, tc.consent)

			if tc.expected {
				require.NotNil(t, tc.ar)
				assert.Equal(t, tc.expectedScope, tc.ar.GetGrantedScopes())
				assert.Equal(t, tc.expectedAudience, tc.ar.GetGrantedAudience())
			} else {
				assert.Nil(t, tc.ar)
			}
		})
	}
}

func TestGetAudienceFromClaims(t *testing.T) {
	testCases := []struct {
		name     string
		have     map[string]any
		expected []string
		ok       bool
	}{
		{
			"ShouldNotPanicOnNil",
			nil,
			nil,
			false,
		},
		{
			"ShouldReturnNothingWithoutClaim",
			map[string]any{},
			nil,
			false,
		},
		{
			"ShouldReturnClaimString",
			map[string]any{oidc.ClaimAudience: "abc"},
			[]string{"abc"},
			true,
		},
		{
			"ShouldReturnClaimStringEmpty",
			map[string]any{oidc.ClaimAudience: ""},
			nil,
			true,
		},
		{
			"ShouldReturnClaimSlice",
			map[string]any{oidc.ClaimAudience: []string{"abc"}},
			[]string{"abc"},
			true,
		},
		{
			"ShouldReturnClaimSliceMulti",
			map[string]any{oidc.ClaimAudience: []string{"abc", "123"}},
			[]string{"abc", "123"},
			true,
		},
		{
			"ShouldReturnClaimSliceAny",
			map[string]any{oidc.ClaimAudience: []any{"abc"}},
			[]string{"abc"},
			true,
		},
		{
			"ShouldReturnClaimSliceAnyMulti",
			map[string]any{oidc.ClaimAudience: []any{"abc", "123"}},
			[]string{"abc", "123"},
			true,
		},
		{
			"ShouldReturnNilInvalidClaim",
			map[string]any{oidc.ClaimAudience: []any{"abc", "123", 11}},
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, ok := oidc.GetAudienceFromClaims(tc.have)

			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.ok, ok)
		})
	}
}

func TestOrderedClaimsRequestsSerialized(t *testing.T) {
	hash := func(s string) string {
		sum := sha256.Sum256([]byte(s))
		return fmt.Sprintf("%x", sum[:])
	}

	testCases := []struct {
		name             string
		ocr              *oidc.OrderedClaimsRequests
		ocr2             *oidc.OrderedClaimsRequests
		expectSerialized string
		expectError      bool
	}{
		{
			name:             "ShouldSerializeEmpty",
			ocr:              &oidc.OrderedClaimsRequests{},
			expectSerialized: "{}",
		},
		{
			name: "ShouldSerializeDeterministicallyForIDToken",
			ocr: &oidc.OrderedClaimsRequests{
				IDToken: oidc.OrderedClaimRequests{
					{Claim: "z", Request: &oidc.ClaimRequest{Essential: true}},
					{Claim: "a", Request: &oidc.ClaimRequest{Essential: false}},
					{Claim: "m", Request: &oidc.ClaimRequest{Essential: true}},
				},
			},
			ocr2: &oidc.OrderedClaimsRequests{
				IDToken: oidc.OrderedClaimRequests{
					{Claim: "m", Request: &oidc.ClaimRequest{Essential: true}},
					{Claim: "z", Request: &oidc.ClaimRequest{Essential: true}},
					{Claim: "a", Request: &oidc.ClaimRequest{Essential: false}},
				},
			},
		},
		{
			name: "ShouldSerializeDeterministicallyForUserInfo",
			ocr: &oidc.OrderedClaimsRequests{
				UserInfo: oidc.OrderedClaimRequests{
					{Claim: "email", Request: &oidc.ClaimRequest{Essential: true}},
					{Claim: "name", Request: &oidc.ClaimRequest{Essential: false}},
				},
			},
			ocr2: &oidc.OrderedClaimsRequests{
				UserInfo: oidc.OrderedClaimRequests{
					{Claim: "name", Request: &oidc.ClaimRequest{Essential: false}},
					{Claim: "email", Request: &oidc.ClaimRequest{Essential: true}},
				},
			},
		},
		{
			name: "ShouldSerializeIDTokenAndUserInfoSorted",
			ocr: &oidc.OrderedClaimsRequests{
				IDToken: oidc.OrderedClaimRequests{
					{Claim: "b", Request: &oidc.ClaimRequest{Essential: true}},
					{Claim: "a", Request: &oidc.ClaimRequest{Essential: false}},
				},
				UserInfo: oidc.OrderedClaimRequests{
					{Claim: "c", Request: &oidc.ClaimRequest{Essential: false}},
					{Claim: "a", Request: &oidc.ClaimRequest{Essential: true}},
				},
			},
			expectSerialized: `{"id_token":{"a":{"essential":false},"b":{"essential":true}},"userinfo":{"a":{"essential":true},"c":{"essential":false}}}`,
		},
		{
			name: "ShouldReturnErrorOnUnsupportedValue",
			ocr: &oidc.OrderedClaimsRequests{
				IDToken: oidc.OrderedClaimRequests{
					{Claim: "a", Request: &oidc.ClaimRequest{Value: make(chan int)}},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serialized, signature, err := tc.ocr.Serialized()

			if tc.expectError {
				require.Error(t, err)
				assert.Empty(t, serialized)
				assert.Empty(t, signature)

				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, serialized)
			assert.NotEmpty(t, signature)

			if tc.expectSerialized != "" {
				assert.Equal(t, tc.expectSerialized, serialized)
			}

			assert.Equal(t, hash(serialized), signature)

			if tc.ocr2 != nil {
				serialized2, signature2, err2 := tc.ocr2.Serialized()
				require.NoError(t, err2)
				assert.Equal(t, serialized, serialized2)
				assert.Equal(t, signature, signature2)
				assert.Equal(t, hash(serialized2), signature2)
			}
		})
	}
}

type testDetailer struct {
	username  string
	groups    []string
	name      string
	emails    []string
	given     string
	family    string
	middle    string
	nickname  string
	profile   string
	picture   string
	website   string
	gender    string
	birthdate string
	info      string
	locale    string
	number    string
	extension string
	address   string
	locality  string
	region    string
	postcode  string
	country   string
	extra     map[string]any
}

func (t testDetailer) GetGivenName() (given string) {
	return t.given
}

func (t testDetailer) GetFamilyName() (family string) {
	return t.family
}

func (t testDetailer) GetMiddleName() (middle string) {
	return t.middle
}

func (t testDetailer) GetNickname() (nickname string) {
	return t.nickname
}

func (t testDetailer) GetProfile() (profile string) {
	return t.profile
}

func (t testDetailer) GetPicture() (picture string) {
	return t.picture
}

func (t testDetailer) GetWebsite() (website string) {
	return t.website
}

func (t testDetailer) GetGender() (gender string) {
	return t.gender
}

func (t testDetailer) GetBirthdate() (birthdate string) {
	return t.birthdate
}

func (t testDetailer) GetZoneInfo() (info string) {
	return t.info
}

func (t testDetailer) GetLocale() (locale string) {
	return t.locale
}

func (t testDetailer) GetPhoneNumber() (number string) {
	return t.number
}

func (t testDetailer) GetPhoneExtension() (extension string) {
	return t.extension
}

func (t testDetailer) GetPhoneNumberRFC3966() (number string) {
	return t.number
}

func (t testDetailer) GetStreetAddress() (address string) {
	return t.address
}

func (t testDetailer) GetLocality() (locality string) {
	return t.locality
}

func (t testDetailer) GetRegion() (region string) {
	return t.region
}

func (t testDetailer) GetPostalCode() (postcode string) {
	return t.postcode
}

func (t testDetailer) GetCountry() (country string) {
	return t.country
}

func (t testDetailer) GetUsername() (username string) {
	return t.username
}

func (t testDetailer) GetGroups() (groups []string) {
	return t.groups
}

func (t testDetailer) GetDisplayName() (name string) {
	return t.name
}

func (t testDetailer) GetEmails() (emails []string) {
	return t.emails
}

func (t testDetailer) GetExtra() (extra map[string]any) {
	return t.extra
}

var (
	_ oidc.UserDetailer = (*testDetailer)(nil)
)
