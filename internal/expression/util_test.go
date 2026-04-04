package expression

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNativeValues(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		have       map[string]any
		expected   any
	}{
		{
			"ShouldHandleBasicCase",
			"true",
			map[string]any{},
			true,
		},
		{
			"ShouldHandleComplexCaseSliceMapOutput",
			`groups.map(i, {"name": i, "oidcID": i})`,
			map[string]any{
				"groups": []string{"abc", "123"},
			},
			[]any{
				map[string]any{"name": "abc", "oidcID": "abc"},
				map[string]any{"name": "123", "oidcID": "123"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := getStandardCELEnvOpts()

			env, err := cel.NewEnv(opts...)
			require.NoError(t, err)

			ast, issues := env.Compile(tc.expression)
			require.NoError(t, issues.Err())

			program, err := env.Program(ast)
			require.NoError(t, err)

			result, _, err := program.Eval(tc.have)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, toNativeValue(result))
		})
	}
}

func TestIsReservedAttribute(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		expected bool
	}{
		{"ShouldReturnTrueForUsername", AttributeUserUsername, true},
		{"ShouldReturnTrueForGroups", AttributeUserGroups, true},
		{"ShouldReturnTrueForDisplayName", AttributeUserDisplayName, true},
		{"ShouldReturnTrueForEmail", AttributeUserEmail, true},
		{"ShouldReturnTrueForEmails", AttributeUserEmails, true},
		{"ShouldReturnTrueForEmailsExtra", AttributeUserEmailsExtra, true},
		{"ShouldReturnTrueForEmailVerified", AttributeUserEmailVerified, true},
		{"ShouldReturnTrueForGivenName", AttributeUserGivenName, true},
		{"ShouldReturnTrueForMiddleName", AttributeUserMiddleName, true},
		{"ShouldReturnTrueForFamilyName", AttributeUserFamilyName, true},
		{"ShouldReturnTrueForNickname", AttributeUserNickname, true},
		{"ShouldReturnTrueForProfile", AttributeUserProfile, true},
		{"ShouldReturnTrueForPicture", AttributeUserPicture, true},
		{"ShouldReturnTrueForWebsite", AttributeUserWebsite, true},
		{"ShouldReturnTrueForGender", AttributeUserGender, true},
		{"ShouldReturnTrueForBirthdate", AttributeUserBirthdate, true},
		{"ShouldReturnTrueForZoneInfo", AttributeUserZoneInfo, true},
		{"ShouldReturnTrueForLocale", AttributeUserLocale, true},
		{"ShouldReturnTrueForPhoneNumber", AttributeUserPhoneNumber, true},
		{"ShouldReturnTrueForPhoneNumberRFC3966", AttributeUserPhoneNumberRFC3966, true},
		{"ShouldReturnTrueForPhoneExtension", AttributeUserPhoneExtension, true},
		{"ShouldReturnTrueForPhoneNumberVerified", AttributeUserPhoneNumberVerified, true},
		{"ShouldReturnTrueForAddress", AttributeUserAddress, true},
		{"ShouldReturnTrueForStreetAddress", AttributeUserStreetAddress, true},
		{"ShouldReturnTrueForLocality", AttributeUserLocality, true},
		{"ShouldReturnTrueForRegion", AttributeUserRegion, true},
		{"ShouldReturnTrueForPostalCode", AttributeUserPostalCode, true},
		{"ShouldReturnTrueForCountry", AttributeUserCountry, true},
		{"ShouldReturnTrueForUpdatedAt", AttributeUserUpdatedAt, true},
		{"ShouldReturnTrueForClaimValue", AttributeOpenIDAuthorizationRequestClaimValue, true},
		{"ShouldReturnTrueForClaimValues", AttributeOpenIDAuthorizationRequestClaimValues, true},
		{"ShouldReturnFalseForCustomAttribute", "custom_attr", false},
		{"ShouldReturnFalseForEmptyString", "", false},
		{"ShouldReturnFalseForUnknownAttribute", "unknown", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsReservedAttribute(tc.key))
		})
	}
}

func TestOptExtraDefaultType(t *testing.T) {
	testCases := []struct {
		name        string
		valueType   string
		multiValued bool
	}{
		{
			"ShouldHandleUnknownType",
			"float",
			false,
		},
		{
			"ShouldHandleEmptyType",
			"",
			false,
		},
		{
			"ShouldHandleUnknownTypeMultiValued",
			"float",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attr := schema.AuthenticationBackendExtraAttribute{
				ValueType:   tc.valueType,
				MultiValued: tc.multiValued,
			}

			opt := optExtra("test_attr", attr)

			assert.NotNil(t, opt)
		})
	}
}
