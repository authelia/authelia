package authentication

import (
	"net/mail"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestUserDetails_Addresses(t *testing.T) {
	details := &UserDetails{}

	assert.Equal(t, []mail.Address(nil), details.Addresses())

	details = &UserDetails{
		DisplayName: "Example",
		Emails:      []string{"abc@123.com"},
	}

	assert.Equal(t, []mail.Address{{Name: "Example", Address: "abc@123.com"}}, details.Addresses())

	details = &UserDetails{
		DisplayName: "Example",
		Emails:      []string{"abc@123.com", "two@apple.com"},
	}

	assert.Equal(t, []mail.Address{{Name: "Example", Address: "abc@123.com"}, {Name: "Example", Address: "two@apple.com"}}, details.Addresses())

	details = &UserDetails{
		DisplayName: "",
		Emails:      []string{"abc@123.com"},
	}

	assert.Equal(t, []mail.Address{{Address: "abc@123.com"}}, details.Addresses())
}

func TestLevel_String(t *testing.T) {
	assert.Equal(t, "one_factor", OneFactor.String())
	assert.Equal(t, "two_factor", TwoFactor.String())
	assert.Equal(t, "not_authenticated", NotAuthenticated.String())
	assert.Equal(t, "invalid", Level(-1).String())
}

func TestUserDetails(t *testing.T) {
	testCases := []struct {
		name        string
		have        *UserDetails
		username    string
		displayname string
		groups      []string
		emails      []string
	}{
		{
			"ShouldHandleDefaultValues",
			&UserDetails{},
			"",
			"",
			nil,
			nil,
		},
		{
			"ShouldHandleAllValues",
			&UserDetails{
				Username:    "john",
				DisplayName: "john smith",
				Emails:      []string{"john@example.com"},
				Groups:      []string{"jgroup"},
			},
			"john",
			"john smith",
			[]string{"jgroup"},
			[]string{"john@example.com"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.username, tc.have.GetUsername())
			assert.Equal(t, tc.displayname, tc.have.GetDisplayName())
			assert.Equal(t, tc.groups, tc.have.GetGroups())
			assert.Equal(t, tc.emails, tc.have.GetEmails())
		})
	}
}

func TestUserDetailsExtended(t *testing.T) {
	tag, err := language.Parse("en-US")
	require.NoError(t, err)

	testCases := []struct {
		name       string
		have       *UserDetailsExtended
		given      string
		middle     string
		family     string
		nickname   string
		locale     string
		zoneinfo   string
		profile    string
		picture    string
		website    string
		phone      string
		ext        string
		phonerfc   string
		birthdate  string
		gender     string
		street     string
		locality   string
		region     string
		postalcode string
		country    string
		extra      map[string]any
	}{
		{
			"ShouldHandleDefaultValues",
			&UserDetailsExtended{},
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			nil,
		},
		{
			"ShouldHandleAllValues",
			&UserDetailsExtended{
				GivenName:      "john",
				FamilyName:     "smith",
				MiddleName:     "jones",
				Nickname:       "johnny",
				Profile:        &url.URL{Scheme: "https", Host: "example.com", Path: "/profile", RawQuery: "id=123&type=null", Fragment: "section1"},
				Picture:        &url.URL{Scheme: "https", Host: "example1.com"},
				Website:        &url.URL{Scheme: "https", Host: "example2.com"},
				Gender:         "male",
				Birthdate:      "2024-03-15",
				ZoneInfo:       "yes",
				Locale:         &tag,
				PhoneNumber:    "+1-555-0123",
				PhoneExtension: "123",
				Address: &UserDetailsAddress{
					StreetAddress: "123 Example St",
					Locality:      "An Area",
					Region:        "An Region",
					PostalCode:    "12354-1234",
					Country:       "US",
				},
			},
			"john",
			"jones",
			"smith",
			"johnny",
			"en-US",
			"yes",
			"https://example.com/profile?id=123&type=null#section1",
			"https://example1.com",
			"https://example2.com",
			"+1-555-0123",
			"123",
			"+1-555-0123;ext=123",
			"2024-03-15",
			"male",
			"123 Example St",
			"An Area",
			"An Region",
			"12354-1234",
			"US",
			nil,
		},
		{
			"ShouldHandleAllValuesNoExt",
			&UserDetailsExtended{
				GivenName:   "john",
				FamilyName:  "smith",
				MiddleName:  "jones",
				Nickname:    "johnny",
				Profile:     &url.URL{Scheme: "https", Host: "example.com"},
				Picture:     &url.URL{Scheme: "https", Host: "example1.com"},
				Website:     &url.URL{Scheme: "https", Host: "example2.com"},
				Gender:      "male",
				Birthdate:   "2024",
				ZoneInfo:    "yes",
				Locale:      &tag,
				PhoneNumber: "1235",
				Address: &UserDetailsAddress{
					StreetAddress: "123 Example St",
					Locality:      "An Area",
					Region:        "An Region",
					PostalCode:    "12354",
					Country:       "US",
				},
				Extra: map[string]any{
					"example":      1,
					"string_value": "test",
					"int_value":    42,
					"bool_value":   true,
					"nested_map": map[string]string{
						"key": "value",
					},
					"string_slice": []string{"a", "b", "c"},
				},
			},
			"john",
			"jones",
			"smith",
			"johnny",
			"en-US",
			"yes",
			"https://example.com",
			"https://example1.com",
			"https://example2.com",
			"1235",
			"",
			"1235",
			"2024",
			"male",
			"123 Example St",
			"An Area",
			"An Region",
			"12354",
			"US",
			map[string]any{
				"example":      1,
				"string_value": "test",
				"int_value":    42,
				"bool_value":   true,
				"nested_map": map[string]string{
					"key": "value",
				},
				"string_slice": []string{"a", "b", "c"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.given, tc.have.GetGivenName())
			assert.Equal(t, tc.middle, tc.have.GetMiddleName())
			assert.Equal(t, tc.family, tc.have.GetFamilyName())
			assert.Equal(t, tc.nickname, tc.have.GetNickname())
			assert.Equal(t, tc.locale, tc.have.GetLocale())
			assert.Equal(t, tc.zoneinfo, tc.have.GetZoneInfo())
			assert.Equal(t, tc.profile, tc.have.GetProfile())
			assert.Equal(t, tc.picture, tc.have.GetPicture())
			assert.Equal(t, tc.website, tc.have.GetWebsite())
			assert.Equal(t, tc.phone, tc.have.GetPhoneNumber())
			assert.Equal(t, tc.ext, tc.have.GetPhoneExtension())
			assert.Equal(t, tc.phonerfc, tc.have.GetPhoneNumberRFC3966())
			assert.Equal(t, tc.birthdate, tc.have.GetBirthdate())
			assert.Equal(t, tc.gender, tc.have.GetGender())
			assert.Equal(t, tc.street, tc.have.GetStreetAddress())
			assert.Equal(t, tc.locality, tc.have.GetLocality())
			assert.Equal(t, tc.region, tc.have.GetRegion())
			assert.Equal(t, tc.postalcode, tc.have.GetPostalCode())
			assert.Equal(t, tc.country, tc.have.GetCountry())
			assert.Equal(t, tc.extra, tc.have.GetExtra())
		})
	}
}
