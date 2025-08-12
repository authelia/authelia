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

	tests := []struct {
		name string
		got  func() string
		want string
	}{
		{name: "GetUsername", got: u.GetUsername, want: "jdoe"},
		{name: "GetDisplayName", got: u.GetDisplayName, want: "John Doe"},
		{name: "GetGivenName", got: u.GetGivenName, want: "John"},
		{name: "GetFamilyName", got: u.GetFamilyName, want: "Doe"},
		{name: "GetMiddleName", got: u.GetMiddleName, want: "Q"},
		{name: "GetNickname", got: u.GetNickname, want: "Johnny"},
		{name: "GetGender", got: u.GetGender, want: "unspecified"},
		{name: "GetBirthdate", got: u.GetBirthdate, want: "2000-01-02"},
		{name: "GetZoneInfo", got: u.GetZoneInfo, want: "Europe/Berlin"},
		{name: "GetPhoneNumber", got: u.GetPhoneNumber, want: "+123456789"},
		{name: "GetPhoneExtension", got: u.GetPhoneExtension, want: "101"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.got())
		})
	}
}

func TestUser_SliceFieldGetters(t *testing.T) {
	u := &User{
		Emails: []string{"a@example.com", "b@example.com"},
		Groups: []string{"admins", "users"},
	}

	tests := []struct {
		name string
		got  func() []string
		want []string
	}{
		{name: "GetEmails", got: u.GetEmails, want: []string{"a@example.com", "b@example.com"}},
		{name: "GetGroups", got: u.GetGroups, want: []string{"admins", "users"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.got())
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

		tests := []struct {
			name string
			got  func() string
			want string
		}{
			{name: "GetProfile", got: u.GetProfile, want: "https://example.com/profile"},
			{name: "GetPicture", got: u.GetPicture, want: "https://example.com/picture.png"},
			{name: "GetWebsite", got: u.GetWebsite, want: "https://example.com"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.want, tc.got())
			})
		}
	})

	t.Run("ShouldHandleNil", func(t *testing.T) {
		u := &User{}
		tests := []struct {
			name string
			got  func() string
		}{
			{name: "GetProfile", got: u.GetProfile},
			{name: "GetPicture", got: u.GetPicture},
			{name: "GetWebsite", got: u.GetWebsite},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, "", tc.got())
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
	tests := []struct {
		name string
		user *User
		want string
	}{
		{name: "ShouldHandleEmptyNumberAndExtension", user: &User{}, want: ""},
		{name: "ShouldHandleNumberOnly", user: &User{PhoneNumber: "123"}, want: "123"},
		{name: "ShouldHandleNumberWithExtension", user: &User{PhoneNumber: "123", PhoneExtension: "456"}, want: "123;ext=456"},
		{name: "ExtensionOnlyIgnored", user: &User{PhoneExtension: "999"}, want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.user.GetPhoneNumberRFC3966())
		})
	}
}

func TestUser_AddressGetters(t *testing.T) {
	t.Run("ShouldHandleNilValue", func(t *testing.T) {
		u := &User{}
		tests := []struct {
			name string
			got  func() string
		}{
			{name: "GetStreetAddress", got: u.GetStreetAddress},
			{name: "GetLocality", got: u.GetLocality},
			{name: "GetRegion", got: u.GetRegion},
			{name: "GetPostalCode", got: u.GetPostalCode},
			{name: "GetCountry", got: u.GetCountry},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, "", tc.got())
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

		tests := []struct {
			name string
			got  func() string
			want string
		}{
			{name: "GetStreetAddress", got: u.GetStreetAddress, want: "123 Test St"},
			{name: "GetLocality", got: u.GetLocality, want: "Testville"},
			{name: "GetRegion", got: u.GetRegion, want: "TS"},
			{name: "GetPostalCode", got: u.GetPostalCode, want: "12345"},
			{name: "GetCountry", got: u.GetCountry, want: "TC"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.want, tc.got())
			})
		}
	})
}

func TestUser_GetExtra(t *testing.T) {
	tests := []struct {
		name string
		user *User
		want map[string]any
	}{
		{name: "ShouldHandleNil", user: &User{}, want: nil},
		{name: "ShouldHandleNotNil", user: &User{Extra: map[string]any{"k": "v", "n": 1}}, want: map[string]any{"k": "v", "n": 1}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.user.GetExtra())
		})
	}
}
