package oidc_test

import (
	"net/url"
	"testing"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewClaimRequests(t *testing.T) {
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

func TestGrantClaimsUserInfo(t *testing.T) {
	strategy := oauthelia2.ExactScopeStrategy

	testCases := []struct {
		name           string
		client         oidc.Client
		scopes         oauthelia2.Arguments
		requests       map[string]*oidc.ClaimRequest
		detailer       oidc.UserDetailer
		claims         map[string]any
		expected       map[string]any
		expectedClaims map[string]any
		expectedScopes map[string]any
	}{
		{
			"ShouldGrantUserInfoClaims",
			&oidc.RegisteredClient{ID: "example", Scopes: []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail}},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail},
			nil,
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
			},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimAudience: []string{"example"}, oidc.ClaimUpdatedAt: 1234},
			nil,
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimAudience: []string{"example"}, oidc.ClaimUpdatedAt: 1234},
		},
		{
			"ShouldGrantUserInfoClaimsMultipleEmails",
			&oidc.RegisteredClient{ID: "example", Scopes: []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail}},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail},
			nil,
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com", "jsmith@authelia.com"},
			},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailAlts: []string{"jsmith@authelia.com"}, oidc.ClaimEmailVerified: true, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimAudience: []string{"example"}, oidc.ClaimUpdatedAt: 1234},
			nil,
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailAlts: []string{"jsmith@authelia.com"}, oidc.ClaimEmailVerified: true, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimAudience: []string{"example"}, oidc.ClaimUpdatedAt: 1234},
		},
		{
			"ShouldGrantUserInfoClaimsNoEmail",
			&oidc.RegisteredClient{ID: "example", Scopes: []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail}},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail},
			nil,
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{},
			},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimAudience: []string{"example"}, oidc.ClaimUpdatedAt: 1234},
			nil,
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimAudience: []string{"example"}, oidc.ClaimUpdatedAt: 1234},
		},
		{
			"ShouldGrantRequestedClaims",
			&oidc.RegisteredClient{ID: "example", Scopes: []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, oidc.ScopeAddress}},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeEmail, oidc.ScopeAddress},
			map[string]*oidc.ClaimRequest{oidc.ClaimGroups: nil, oidc.ClaimUpdatedAt: nil, oidc.ClaimEmailAlts: nil, oidc.ClaimEmail: nil, oidc.ClaimEmailVerified: nil, oidc.ClaimPreferredUsername: nil, oidc.ClaimFullName: nil, oidc.ClaimGivenName: nil, oidc.ClaimAddress: nil, oidc.ClaimMiddleName: nil, oidc.ClaimFamilyName: nil, oidc.ClaimPhoneNumber: nil, oidc.ClaimPhoneNumberVerified: nil, oidc.ClaimNickname: nil, oidc.ClaimProfile: nil, oidc.ClaimPicture: nil, oidc.ClaimWebsite: nil, oidc.ClaimGender: nil, oidc.ClaimZoneinfo: nil, oidc.ClaimLocale: nil, oidc.ClaimBirthdate: nil},
			&testDetailer{
				username: "john",
				groups:   []string{"abc", "123"},
				name:     "John Smith",
				emails:   []string{"john.smith@authelia.com"},
			},
			map[string]any{oidc.ClaimAuthenticationTime: 123},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimAudience: []string{"example"}, oidc.ClaimUpdatedAt: 1234, oidc.ClaimLocale: "", oidc.ClaimFamilyName: "", oidc.ClaimPicture: "", oidc.ClaimWebsite: "", oidc.ClaimProfile: "", oidc.ClaimZoneinfo: "", oidc.ClaimNickname: "", oidc.ClaimGivenName: "", oidc.ClaimMiddleName: "", oidc.ClaimBirthdate: "", oidc.ClaimGender: ""},
			map[string]any{oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimUpdatedAt: 1234, oidc.ClaimLocale: "", oidc.ClaimFamilyName: "", oidc.ClaimPicture: "", oidc.ClaimWebsite: "", oidc.ClaimProfile: "", oidc.ClaimZoneinfo: "", oidc.ClaimNickname: "", oidc.ClaimGivenName: "", oidc.ClaimMiddleName: "", oidc.ClaimBirthdate: "", oidc.ClaimGender: ""},
			map[string]any{oidc.ClaimAuthenticationTime: 123, oidc.ClaimEmail: "john.smith@authelia.com", oidc.ClaimEmailVerified: true, oidc.ClaimGroups: []string{"abc", "123"}, oidc.ClaimFullName: "John Smith", oidc.ClaimPreferredUsername: "john", oidc.ClaimAudience: []string{"example"}, oidc.ClaimUpdatedAt: 1234},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			extra := map[string]any{}
			extraClaims := map[string]any{}
			extraScopes := map[string]any{}

			oidc.GrantClaimRequests(strategy, tc.client, tc.requests, tc.detailer, extra)
			oidc.GrantScopedClaims(strategy, tc.client, tc.scopes, tc.detailer, tc.claims, extra)

			oidc.GrantClaimRequests(strategy, tc.client, tc.requests, tc.detailer, extraClaims)

			oidc.GrantScopedClaims(strategy, tc.client, tc.scopes, tc.detailer, tc.claims, extraScopes)

			for key, value := range extra {
				if key == oidc.ClaimUpdatedAt {
					assert.Contains(t, tc.expected, key, "Extra does not contain key %s", key)

					continue
				}

				assert.Equal(t, tc.expected[key], value, "Extra key %s has value %s", key, value)
			}

			for key := range tc.expected {
				assert.Contains(t, extra, key, "Extra does not contain key %s", key)
			}

			for key, value := range extraClaims {
				if key == oidc.ClaimUpdatedAt {
					assert.Contains(t, tc.expectedClaims, key)

					continue
				}

				assert.Equal(t, tc.expectedClaims[key], value, "Extra Claims key %s has value %s", key, value)
			}

			for key := range tc.expectedClaims {
				assert.Contains(t, extraClaims, key, "Extra Claims does not contain key %s", key)
			}

			for key, value := range extraScopes {
				if key == oidc.ClaimUpdatedAt {
					assert.Contains(t, tc.expectedScopes, key, "Expected Scopes does not contain key %s", key)

					continue
				}

				assert.Equal(t, tc.expectedScopes[key], value, "Extra Scopes key %s has value %s", key, value)
			}

			for key := range tc.expectedScopes {
				assert.Contains(t, extraScopes, key, "Extra Scopes does not contain key %s", key)
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

func (t testDetailer) GetOpenIDConnectPhoneNumber() (number string) {
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

var (
	_ oidc.UserDetailer = (*testDetailer)(nil)
)
