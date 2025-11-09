package expression_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/authentication"
	. "github.com/authelia/authelia/v4/internal/expression"
)

func TestUserDetailerActivationBlackBox(t *testing.T) {
	activation := &UserDetailerActivation{}

	assert.Nil(t, activation.Parent())

	mustParseURI := func(uri string) *url.URL {
		if uri, err := url.Parse(uri); err != nil {
			panic(err)
		} else {
			return uri
		}
	}

	activation = NewUserDetailerActivation(
		nil,
		&authentication.UserDetailsExtended{
			GivenName:      "john",
			FamilyName:     "smith",
			MiddleName:     "jones",
			Nickname:       "johnny",
			Profile:        mustParseURI("https://authellia.com/jsmith"),
			Picture:        mustParseURI("https://authellia.com/jsmith.jpg"),
			Website:        mustParseURI("https://authellia.com"),
			Gender:         "male",
			Birthdate:      "2020",
			ZoneInfo:       "zoney",
			Locale:         nil,
			PhoneNumber:    "123567",
			PhoneExtension: "123",
			Address: &authentication.UserDetailsAddress{
				StreetAddress: "123 Bay St",
				Locality:      "General",
				Region:        "Region",
				PostalCode:    "445500",
				Country:       "US",
			},
			Extra: map[string]any{
				"example": 1,
			},
			UserDetails: &authentication.UserDetails{
				Username:    "jsmith",
				DisplayName: "John Smith",
				Emails:      []string{"jsmith@example.com", "alt@example.com"},
				Groups:      []string{"admin"},
			},
		}, time.Unix(100000000, 0))

	testCases := []struct {
		name     string
		expected any
		found    bool
	}{
		{AttributeUserUsername, "jsmith", true},
		{AttributeUserGroups, []string{"admin"}, true},
		{AttributeUserDisplayName, "John Smith", true},
		{AttributeUserEmail, "jsmith@example.com", true},
		{AttributeUserEmails, []string{"jsmith@example.com", "alt@example.com"}, true},
		{AttributeUserEmailsExtra, []string{"alt@example.com"}, true},
		{AttributeUserEmailVerified, true, true},
		{AttributeUserGivenName, "john", true},
		{AttributeUserMiddleName, "jones", true},
		{AttributeUserFamilyName, "smith", true},
		{AttributeUserNickname, "johnny", true},
		{AttributeUserProfile, "https://authellia.com/jsmith", true},
		{AttributeUserPicture, "https://authellia.com/jsmith.jpg", true},
		{AttributeUserWebsite, "https://authellia.com", true},
		{AttributeUserGender, "male", true},
		{AttributeUserBirthdate, "2020", true},
		{AttributeUserZoneInfo, "zoney", true},
		{AttributeUserLocale, "", true},
		{AttributeUserPhoneNumber, "123567", true},
		{AttributeUserPhoneNumberRFC3966, "123567;ext=123", true},
		{AttributeUserPhoneExtension, "123", true},
		{AttributeUserPhoneNumberVerified, false, true},
		{AttributeUserAddress, map[string]any{"country": "US", "locality": "General", "postal_code": "445500", "region": "Region", "street_address": "123 Bay St"}, true},
		{AttributeUserStreetAddress, "123 Bay St", true},
		{AttributeUserLocality, "General", true},
		{AttributeUserRegion, "Region", true},
		{AttributeUserPostalCode, "445500", true},
		{AttributeUserCountry, "US", true},
		{AttributeUserUpdatedAt, int64(100000000), true},
		{"example", 1, true},
		{"example2", nil, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, ok := activation.ResolveName(tc.name)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.found, ok)
		})
	}
}

func TestUserDetailerActivationBlackBoxNoValues(t *testing.T) {
	activation := NewUserDetailerActivation(
		nil,
		&authentication.UserDetailsExtended{
			UserDetails: &authentication.UserDetails{
				Username: "jsmith",
			},
		}, time.Unix(100000000, 0))

	testCases := []struct {
		name     string
		expected any
		found    bool
	}{
		{AttributeUserUsername, "jsmith", true},
		{AttributeUserGroups, []string(nil), true},
		{AttributeUserDisplayName, "", true},
		{AttributeUserEmail, "", true},
		{AttributeUserEmails, []string(nil), true},
		{AttributeUserEmailsExtra, nil, true},
		{AttributeUserEmailVerified, true, true},
		{AttributeUserGivenName, "", true},
		{AttributeUserMiddleName, "", true},
		{AttributeUserFamilyName, "", true},
		{AttributeUserNickname, "", true},
		{AttributeUserProfile, "", true},
		{AttributeUserPicture, "", true},
		{AttributeUserWebsite, "", true},
		{AttributeUserGender, "", true},
		{AttributeUserBirthdate, "", true},
		{AttributeUserZoneInfo, "", true},
		{AttributeUserLocale, "", true},
		{AttributeUserPhoneNumber, "", true},
		{AttributeUserPhoneNumberRFC3966, "", true},
		{AttributeUserPhoneExtension, "", true},
		{AttributeUserPhoneNumberVerified, nil, true},
		{AttributeUserAddress, map[string]any(nil), true},
		{AttributeUserStreetAddress, "", true},
		{AttributeUserLocality, "", true},
		{AttributeUserRegion, "", true},
		{AttributeUserPostalCode, "", true},
		{AttributeUserCountry, "", true},
		{AttributeUserUpdatedAt, int64(100000000), true},
		{"example", nil, false},
		{"example2", nil, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, ok := activation.ResolveName(tc.name)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.found, ok)
		})
	}
}
