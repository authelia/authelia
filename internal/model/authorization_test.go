package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorization_Parse(t *testing.T) {
	testCases := []struct {
		name       string
		have       string
		expected   string
		expectedsr string
		scheme     AuthorizationScheme
		value      string
		username   string
		password   string
		err        string
	}{
		{
			"ShouldParseBearer",
			"Bearer abc123",
			"Bearer abc123",
			"Bearer",
			AuthorizationSchemeBearer,
			"abc123",
			"",
			"",
			"",
		},
		{
			"ShouldParseBearerAnyCase",
			"BeareR abc123",
			"Bearer abc123",
			"BeareR",
			AuthorizationSchemeBearer,
			"abc123",
			"",
			"",
			"",
		},
		{
			"ShouldFailParseBearerNoScheme",
			"Bearer",
			"",
			"",
			AuthorizationSchemeNone,
			"",
			"",
			"",
			"invalid scheme: the scheme is missing",
		},
		{
			"ShouldFailParseBearerEmpty",
			"Bearer ",
			"",
			"",
			AuthorizationSchemeNone,
			"",
			"",
			"",
			"invalid value: bearer scheme value must not be empty",
		},
		{
			"ShouldFailParseBearerWithBadValues",
			"Bearer !(@)#&!@$&(^T)*@#&^!",
			"",
			"",
			AuthorizationSchemeNone,
			"",
			"",
			"",
			"invalid value: bearer scheme value must only contain characters noted in RFC6750 2.1",
		},
		{
			"ShouldParseBasic",
			"Basic YWJjOjEyMw==",
			"Basic YWJjOjEyMw==",
			"Basic",
			AuthorizationSchemeBasic,
			"YWJjOjEyMw==",
			"abc",
			"123",
			"",
		},
		{
			"ShouldFailParseBasicNoUsername",
			"Basic OjEyMw==",
			"",
			"",
			AuthorizationSchemeNone,
			"",
			"",
			"",
			"invalid value: failed to find the username in the decoded basic value as it was empty",
		},
		{
			"ShouldFailParseBasicNoPassword",
			"Basic YWJjOg==",
			"",
			"",
			AuthorizationSchemeNone,
			"",
			"",
			"",
			"invalid value: failed to find the password in the decoded basic value as it was empty",
		},
		{
			"ShouldFailParseBasicNoSep",
			"Basic YWJjMTIz",
			"",
			"",
			AuthorizationSchemeNone,
			"",
			"",
			"",
			"invalid value: failed to find the username password separator in the decoded basic scheme value",
		},
		{
			"ShouldFailParseBasicBadBase64",
			"Basic ===YWJjMTIz",
			"",
			"",
			AuthorizationSchemeNone,
			"",
			"",
			"",
			"invalid value: failed to parse base64 basic scheme value: illegal base64 data at input byte 0",
		},
		{
			"ShouldFailParseBadScheme",
			"Baser YWJjOjEyMw==",
			"",
			"",
			AuthorizationSchemeNone,
			"",
			"",
			"",
			"invalid scheme: scheme with name 'baser' is unknown",
		},
		{
			"ShouldFailParseEmpty",
			"",
			"",
			"",
			AuthorizationSchemeNone,
			"",
			"",
			"",
			"invalid value: the value provided to be parsed was empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authz := NewAuthorization()

			err := authz.Parse(tc.have)

			if len(tc.err) == 0 {
				assert.NoError(t, err)
				assert.Equal(t, tc.scheme, authz.Scheme())
				assert.Equal(t, tc.expectedsr, authz.SchemeRaw())
				assert.Equal(t, tc.password, authz.password)
				assert.Equal(t, tc.username, authz.username)
				assert.Equal(t, tc.expected, authz.EncodeHeader())

				assert.EqualError(t, authz.Parse(tc.have), "invalid state: this scheme has already performed a parse action")

				bauthz := NewAuthorization()

				assert.NoError(t, bauthz.ParseBytes([]byte(tc.have)))
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestAuthorization_ParsBasic(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
		expected string
		err      string
	}{
		{
			"ShouldParseGoodValues",
			"abc",
			"123",
			"YWJjOjEyMw==",
			"",
		},
		{
			"ShouldFailUsernameWithColon",
			"abc:abc",
			"123",
			"",
			"invalid value: username must not contain the ':' character",
		},
		{
			"ShouldFailUsernameEmpty",
			"",
			"123",
			"",
			"invalid value: username must not be empty",
		},
		{
			"ShouldFailPasswordEmpty",
			"abc",
			"",
			"",
			"invalid value: password must not be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authz := NewAuthorization()

			err := authz.ParseBasic(tc.username, tc.password)

			if len(tc.err) == 0 {
				assert.NoError(t, err)

				assert.Equal(t, AuthorizationSchemeBasic, authz.Scheme())
				assert.Equal(t, fmt.Sprintf("Basic %s", tc.expected), authz.EncodeHeader())
				assert.Equal(t, tc.expected, authz.value)
				assert.Equal(t, tc.expected, authz.Value())
				assert.Equal(t, tc.username, authz.username)
				assert.Equal(t, tc.username, authz.BasicUsername())
				assert.Equal(t, tc.password, authz.password)

				username, password := authz.Basic()

				assert.Equal(t, tc.username, username)
				assert.Equal(t, tc.password, password)

				assert.EqualError(t, authz.ParseBasic(tc.username, tc.password), "invalid state: this scheme has already performed a parse action")
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestAuthorization_ParsBearer(t *testing.T) {
	testCases := []struct {
		name     string
		bearer   string
		expected string
		err      string
	}{
		{
			"ShouldParseGoodValues",
			"abc",
			"abc",
			"",
		},
		{
			"ShouldFailParseBadBearerValue",
			"abc!(*@^&(!@*^$",
			"",
			"invalid value: bearer scheme value must only contain characters noted in RFC6750 2.1",
		},
		{
			"ShouldFailParseEmptyBearerValue",
			"",
			"",
			"invalid value: bearer scheme value must not be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authz := NewAuthorization()

			err := authz.ParseBearer(tc.bearer)

			if len(tc.err) == 0 {
				assert.NoError(t, err)

				assert.Equal(t, AuthorizationSchemeBearer, authz.Scheme())
				assert.Equal(t, fmt.Sprintf("Bearer %s", tc.expected), authz.EncodeHeader())
				assert.Equal(t, tc.expected, authz.value)
				assert.Equal(t, tc.expected, authz.Value())

				username, password := authz.Basic()
				assert.Equal(t, "", authz.BasicUsername())
				assert.Equal(t, "", username)
				assert.Equal(t, "", password)

				assert.EqualError(t, authz.ParseBearer(tc.bearer), "invalid state: this scheme has already performed a parse action")
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestAuthorization_NotParsed(t *testing.T) {
	authz := NewAuthorization()

	assert.Equal(t, "", authz.EncodeHeader())
	assert.Equal(t, "", authz.BasicUsername())

	username, password := authz.Basic()
	assert.Equal(t, "", username)
	assert.Equal(t, "", password)
}

func TestAuthorization_SchemeNone(t *testing.T) {
	authz := NewAuthorization()
	authz.parsed = true

	assert.Equal(t, "", authz.EncodeHeader())
	assert.Equal(t, "", authz.BasicUsername())

	username, password := authz.Basic()
	assert.Equal(t, "", username)
	assert.Equal(t, "", password)
}

func TestAuthorization_Misc(t *testing.T) {
	authz := NewAuthorization()
	authz.parsed = true
	authz.scheme = -1

	assert.Equal(t, "", authz.EncodeHeader())

	assert.Equal(t, "", AuthorizationScheme(-1).String())
}

func TestNewAuthorizationSchemes(t *testing.T) {
	testCases := []struct {
		name      string
		have      []string
		expected  AuthorizationSchemes
		expectedf func(t *testing.T, schemes AuthorizationSchemes)
	}{
		{
			"ShouldParseEmpty",
			nil,
			nil,
			nil,
		},
		{
			"ShouldParseBasic",
			[]string{"BaSiC"},
			AuthorizationSchemes{AuthorizationSchemeBasic},
			func(t *testing.T, schemes AuthorizationSchemes) {
				assert.False(t, schemes.Has(AuthorizationSchemeNone))
				assert.True(t, schemes.Has(AuthorizationSchemeBasic))
				assert.False(t, schemes.Has(AuthorizationSchemeBearer))
			},
		},
		{
			"ShouldParseBearer",
			[]string{"Bearer"},
			AuthorizationSchemes{AuthorizationSchemeBearer},
			func(t *testing.T, schemes AuthorizationSchemes) {
				assert.False(t, schemes.Has(AuthorizationSchemeNone))
				assert.False(t, schemes.Has(AuthorizationSchemeBasic))
				assert.True(t, schemes.Has(AuthorizationSchemeBearer))
			},
		},
		{
			"ShouldParseBoth",
			[]string{"Bearer", "Basic"},
			AuthorizationSchemes{AuthorizationSchemeBearer, AuthorizationSchemeBasic},
			func(t *testing.T, schemes AuthorizationSchemes) {
				assert.False(t, schemes.Has(AuthorizationSchemeNone))
				assert.True(t, schemes.Has(AuthorizationSchemeBasic))
				assert.True(t, schemes.Has(AuthorizationSchemeBearer))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewAuthorizationSchemes(tc.have...)

			assert.Equal(t, tc.expected, actual)

			if tc.expectedf != nil {
				tc.expectedf(t, actual)
			}
		})
	}
}
