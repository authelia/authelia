package handler

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestIsAuthorizationMatching(t *testing.T) {
	type args struct {
		levelRequired authorization.Level
		levelCurrent  authentication.Level
	}

	testCases := []struct {
		name     string
		have     args
		expected authorizationMatching
	}{
		{"BypassNotAuthenticatedAuthorized", args{authorization.Bypass, authentication.NotAuthenticated}, Authorized},
		{"BypassOneFactorAuthorized", args{authorization.Bypass, authentication.OneFactor}, Authorized},
		{"BypassTwoFactorAuthorized", args{authorization.Bypass, authentication.TwoFactor}, Authorized},
		{"OneFactorNotAuthenticatedNotAuthorized", args{authorization.OneFactor, authentication.NotAuthenticated}, NotAuthorized},
		{"OneFactorNotAuthenticatedAuthorized", args{authorization.OneFactor, authentication.OneFactor}, Authorized},
		{"OneFactorTwoFactorAuthorized", args{authorization.OneFactor, authentication.TwoFactor}, Authorized},
		{"TwoFactorNotAuthenticatedNotAuthorized", args{authorization.TwoFactor, authentication.NotAuthenticated}, NotAuthorized},
		{"TwoFactorOneFactorNotAuthorized", args{authorization.TwoFactor, authentication.OneFactor}, NotAuthorized},
		{"TwoFactorTwoFactorAuthorized", args{authorization.TwoFactor, authentication.TwoFactor}, Authorized},
		{"DeniedNotAuthenticatedNotAuthorized", args{authorization.Denied, authentication.NotAuthenticated}, NotAuthorized},
		{"DeniedOneFactorNotAuthorizedForbidden", args{authorization.Denied, authentication.OneFactor}, Forbidden},
		{"DeniedTwoFactorNotAuthorizedForbidden", args{authorization.Denied, authentication.TwoFactor}, Forbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isAuthorizationMatching(tc.have.levelRequired, tc.have.levelCurrent)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestHeaderAuthorizationParseBasic(t *testing.T) {
	type result struct {
		username string
		password string
		err      string
	}

	testCases := []struct {
		name     string
		have     string
		expected result
	}{
		{"ShouldParse", "Basic am9objpwYXNzd29yZA==", result{"john", "password", ""}},
		{"ShouldFailToParseWithoutPrefix", "alzefzlfzemjfej==", result{"", "", "header is malformed: does not appear to have a scheme"}},
		{"ShouldFailToParseWithoutSeparator", "Basic am9obiBwYXNzd29yZA==", result{"", "", "header is malformed: format of header must be <user>:<password> but either doesn't have a colon or username"}},
		{"ShouldFailToParseWithBadEncoding", "Basic alzefzlfzemjfej==", result{"", "", "header is malformed: could not decode credentials: illegal base64 data at input byte 16"}},
		{"ShouldFailToParseWithBadScheme", "Digest username=john", result{"", "", "header is malformed: unexpected scheme 'Digest': expected scheme 'Basic'"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualUsername, actualPassword, theError := headerAuthorizationParseBasic([]byte(tc.have))

			assert.Equal(t, tc.expected.username, actualUsername)
			assert.Equal(t, tc.expected.password, actualPassword)

			if tc.expected.err != "" {
				assert.EqualError(t, theError, tc.expected.err)
			} else {
				assert.NoError(t, theError)
			}
		})
	}
}

func TestIsSchemeSecure(t *testing.T) {
	testCases := []struct {
		name     string
		have     *url.URL
		expected bool
	}{
		{"ShouldConsiderHTTPSSecure", mustParseURL("https://example.com"), true},
		{"ShouldConsiderWSSSecure", mustParseURL("wss://example.com"), true},
		{"ShouldNotConsiderHTTPSecure", mustParseURL("http://example.com"), false},
		{"ShouldNotConsiderWSSecure", mustParseURL("ws://example.com"), false},
		{"ShouldNotConsiderTCPSecure", mustParseURL("tcp://example.com"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isSchemeSecure(tc.have)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestIsURLUnderProtectedDomain(t *testing.T) {
	testCases := []struct {
		name     string
		have     *url.URL
		domain   string
		expected bool
	}{
		{"ShouldConsiderSubdomainUnderProtectedDomain", mustParseURL("https://app.example.com"), "example.com", true},
		{"ShouldConsiderWSSSubdomainUnderProtectedDomain", mustParseURL("wss://app.example.com"), "example.com", true},
		{"ShouldConsiderExactDomainUnderProtectedDomain", mustParseURL("https://example.com"), "example.com", true},
		{"ShouldNotConsiderDifferentDomainProtected", mustParseURL("https://app.not.com"), "example.com", false},
		{"ShouldNotConsiderSubSubDomainProtected", mustParseURL("https://app.home.example.com"), "example.com", false},
		{"ShouldNotConsiderBadDomainProtected", mustParseURL("https://com"), "example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isURLUnderProtectedDomain(tc.have, tc.domain)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetProfileRefreshSettings(t *testing.T) {
	testCases := []struct {
		name             string
		have             schema.AuthenticationBackendConfiguration
		expected         bool
		expectedInterval time.Duration
	}{
		{
			"ShouldUseDefaults",
			schema.AuthenticationBackendConfiguration{
				RefreshInterval: schema.RefreshIntervalDefault,
				LDAP:            &schema.LDAPAuthenticationBackendConfiguration{},
			},
			true,
			time.Minute * 5,
		},
		{
			"ShouldHonorAlways",
			schema.AuthenticationBackendConfiguration{
				RefreshInterval: schema.ProfileRefreshAlways,
				LDAP:            &schema.LDAPAuthenticationBackendConfiguration{},
			},
			true,
			time.Duration(0),
		},
		{
			"ShouldHonorDisabled",
			schema.AuthenticationBackendConfiguration{
				RefreshInterval: schema.ProfileRefreshDisabled,
				LDAP:            &schema.LDAPAuthenticationBackendConfiguration{},
			},
			false,
			time.Duration(0),
		},
		{
			"ShouldHonorCustomDuration",
			schema.AuthenticationBackendConfiguration{
				RefreshInterval: "10m",
				LDAP:            &schema.LDAPAuthenticationBackendConfiguration{},
			},
			true,
			time.Minute * 10,
		},
		{
			"ShouldDisableWithoutLDAP",
			schema.AuthenticationBackendConfiguration{
				RefreshInterval: "10m",
				LDAP:            nil,
			},
			false,
			time.Duration(0),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, actualInterval := getProfileRefreshSettings(tc.have)

			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.expectedInterval, actualInterval)
		})
	}
}

func mustParseURL(input string) *url.URL {
	u, err := url.ParseRequestURI(input)

	if err != nil {
		panic(err)
	}

	return u
}
