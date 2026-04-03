package model

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestUser_StringFieldGetters(t *testing.T) {
	u := &User{
		Username:       "jdoe",
		DisplayName:    "John Doe",
		GivenName:      "John",
		FamilyName:     "Doe",
		MiddleName:     "Q",
		Nickname:       "Johnny",
		Gender:         "unspecified",
		Birthdate:      "2000-01-02",
		ZoneInfo:       "Europe/Berlin",
		PhoneNumber:    "+123456789",
		PhoneExtension: "101",
	}

	testCases := []struct {
		name     string
		have     func() string
		expected string
	}{
		{name: "GetUsername", have: u.GetUsername, expected: "jdoe"},
		{name: "GetDisplayName", have: u.GetDisplayName, expected: "John Doe"},
		{name: "GetGivenName", have: u.GetGivenName, expected: "John"},
		{name: "GetFamilyName", have: u.GetFamilyName, expected: "Doe"},
		{name: "GetMiddleName", have: u.GetMiddleName, expected: "Q"},
		{name: "GetNickname", have: u.GetNickname, expected: "Johnny"},
		{name: "GetGender", have: u.GetGender, expected: "unspecified"},
		{name: "GetBirthdate", have: u.GetBirthdate, expected: "2000-01-02"},
		{name: "GetZoneInfo", have: u.GetZoneInfo, expected: "Europe/Berlin"},
		{name: "GetPhoneNumber", have: u.GetPhoneNumber, expected: "+123456789"},
		{name: "GetPhoneExtension", have: u.GetPhoneExtension, expected: "101"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have())
		})
	}
}

func TestUser_SliceFieldGetters(t *testing.T) {
	u := &User{
		Emails: []string{"a@example.com", "b@example.com"},
		Groups: []string{"admins", "users"},
	}

	testCases := []struct {
		name     string
		have     func() []string
		expected []string
	}{
		{name: "GetEmails", have: u.GetEmails, expected: []string{"a@example.com", "b@example.com"}},
		{name: "GetGroups", have: u.GetGroups, expected: []string{"admins", "users"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have())
		})
	}
}

func TestUser_URLGetters(t *testing.T) {
	t.Run("ShouldHandleNotNil", func(t *testing.T) {
		parse := func(raw string) *url.URL {
			u, err := url.Parse(raw)
			require.NoError(t, err)

			return u
		}

		u := &User{
			Profile: parse("https://example.com/profile"),
			Picture: parse("https://example.com/picture.png"),
			Website: parse("https://example.com"),
		}

		testCases := []struct {
			name     string
			have     func() string
			expected string
		}{
			{name: "GetProfile", have: u.GetProfile, expected: "https://example.com/profile"},
			{name: "GetPicture", have: u.GetPicture, expected: "https://example.com/picture.png"},
			{name: "GetWebsite", have: u.GetWebsite, expected: "https://example.com"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.have())
			})
		}
	})

	t.Run("ShouldHandleNil", func(t *testing.T) {
		u := &User{}
		testCases := []struct {
			name string
			have func() string
		}{
			{name: "GetProfile", have: u.GetProfile},
			{name: "GetPicture", have: u.GetPicture},
			{name: "GetWebsite", have: u.GetWebsite},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, "", tc.have())
			})
		}
	})
}

func TestUser_LocaleGetter(t *testing.T) {
	tag, err := language.Parse("en-GB")
	require.NoError(t, err)

	testCases := []struct {
		name string
		user *User
		want string
	}{
		{name: "NilLocale", user: &User{}, want: ""},
		{name: "ShouldHandleStandard", user: &User{Locale: &tag}, want: "en-GB"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.user.GetLocale())
		})
	}
}

func TestUser_PhoneNumberRFC3966(t *testing.T) {
	testCases := []struct {
		name     string
		user     *User
		expected string
	}{
		{name: "ShouldHandleEmptyNumberAndExtension", user: &User{}, expected: ""},
		{name: "ShouldHandleNumberOnly", user: &User{PhoneNumber: "123"}, expected: "123"},
		{name: "ShouldHandleNumberWithExtension", user: &User{PhoneNumber: "123", PhoneExtension: "456"}, expected: "123;ext=456"},
		{name: "ExtensionOnlyIgnored", user: &User{PhoneExtension: "999"}, expected: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.user.GetPhoneNumberRFC3966())
		})
	}
}

func TestUser_AddressGetters(t *testing.T) {
	t.Run("ShouldHandleNilValue", func(t *testing.T) {
		u := &User{}
		testCases := []struct {
			name string
			have func() string
		}{
			{name: "GetStreetAddress", have: u.GetStreetAddress},
			{name: "GetLocality", have: u.GetLocality},
			{name: "GetRegion", have: u.GetRegion},
			{name: "GetPostalCode", have: u.GetPostalCode},
			{name: "GetCountry", have: u.GetCountry},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, "", tc.have())
			})
		}
	})

	t.Run("ShouldHandleValue", func(t *testing.T) {
		u := &User{
			Address: &UserAddress{
				StreetAddress: "123 Test St",
				Locality:      "Testville",
				Region:        "TS",
				PostalCode:    "12345",
				Country:       "TC",
			},
		}

		testCases := []struct {
			name     string
			have     func() string
			expected string
		}{
			{name: "GetStreetAddress", have: u.GetStreetAddress, expected: "123 Test St"},
			{name: "GetLocality", have: u.GetLocality, expected: "Testville"},
			{name: "GetRegion", have: u.GetRegion, expected: "TS"},
			{name: "GetPostalCode", have: u.GetPostalCode, expected: "12345"},
			{name: "GetCountry", have: u.GetCountry, expected: "TC"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.have())
			})
		}
	})
}

func TestUser_GetExtra(t *testing.T) {
	testCases := []struct {
		name     string
		user     *User
		expected map[string]any
	}{
		{name: "ShouldHandleNil", user: &User{}, expected: nil},
		{name: "ShouldHandleNotNil", user: &User{Extra: map[string]any{"k": "v", "n": 1}}, expected: map[string]any{"k": "v", "n": 1}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.user.GetExtra())
		})
	}
}
