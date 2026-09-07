package authentication

import (
	"encoding/json"
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
		{
			"ShouldHandleSpecialCharacters",
			&UserDetails{
				Username:    "john.o'connor",
				DisplayName: "John O'Connor-Smith",
				Emails:      []string{"john@example.com", "john.oconnor@example.com"},
				Groups:      []string{"admin", "users"},
			},
			"john.o'connor",
			"John O'Connor-Smith",
			[]string{"admin", "users"},
			[]string{"john@example.com", "john.oconnor@example.com"},
		},
		{
			"ShouldHandleEmptyGroups",
			&UserDetails{
				Username:    "jane",
				DisplayName: "Jane Doe",
				Emails:      []string{"jane@example.com"},
				Groups:      []string{},
			},
			"jane",
			"Jane Doe",
			[]string{},
			[]string{"jane@example.com"},
		},
		{
			"ShouldHandleUnicodeDisplayName",
			&UserDetails{
				Username:    "zhang",
				DisplayName: "张三",
				Emails:      []string{"zhang@example.com"},
				Groups:      []string{"users"},
			},
			"zhang",
			"张三",
			[]string{"users"},
			[]string{"zhang@example.com"},
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

func TestUserDetailsExtended_MarshalJSON(t *testing.T) {
	tag, err := language.Parse("en-US")
	require.NoError(t, err)

	have := &UserDetailsExtended{
		GivenName:  "john",
		FamilyName: "smith",
		Profile:    &url.URL{Scheme: "https", Host: "example.com", Path: "/profile"},
		Picture:    &url.URL{Scheme: "https", Host: "example1.com"},
		Website:    &url.URL{Scheme: "https", Host: "example2.com"},
		Locale:     &tag,
		UserDetails: &UserDetails{
			Username: "john",
			Emails:   []string{"john@example.com"},
		},
		Password: "supersecret",
	}

	data, err := json.Marshal(have)
	require.NoError(t, err)

	assert.NotContains(t, string(data), "supersecret")
	assert.NotContains(t, string(data), `"password"`)

	var raw map[string]any

	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Equal(t, "https://example.com/profile", raw["profile"])
	assert.Equal(t, "https://example1.com", raw["picture"])
	assert.Equal(t, "https://example2.com", raw["website"])
	assert.Equal(t, "en-US", raw["locale"])
	assert.Equal(t, "john", raw["given_name"])
	assert.Equal(t, "smith", raw["family_name"])
	assert.Equal(t, "john", raw["username"])
	assert.Equal(t, []any{"john@example.com"}, raw["mail"])
}

func TestUserDetailsExtended_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name    string
		data    string
		err     string
		asserts func(t *testing.T, d *UserDetailsExtended)
	}{
		{
			"ShouldUnmarshalFullObject",
			`{
				"username": "john",
				"given_name": "john",
				"family_name": "smith",
				"mail": "john@example.com",
				"profile": "https://example.com/profile",
				"picture": "https://example1.com",
				"website": "https://example2.com",
				"locale": "en-US",
				"password": "supersecret"
			}`,
			"",
			func(t *testing.T, d *UserDetailsExtended) {
				assert.Equal(t, "john", d.GivenName)
				assert.Equal(t, "smith", d.FamilyName)
				assert.Equal(t, "supersecret", d.Password)
				assert.Equal(t, []string{"john@example.com"}, d.Emails)
				assert.Equal(t, "https://example.com/profile", stringURL(d.Profile))
				assert.Equal(t, "https://example1.com", stringURL(d.Picture))
				assert.Equal(t, "https://example2.com", stringURL(d.Website))
				require.NotNil(t, d.Locale)
				assert.Equal(t, "en-US", d.Locale.String())
			},
		},
		{
			"ShouldUnmarshalMailArray",
			`{"mail": ["john@example.com", "other@example.com"]}`,
			"",
			func(t *testing.T, d *UserDetailsExtended) {
				assert.Equal(t, []string{"john@example.com", "other@example.com"}, d.Emails)
			},
		},
		{
			"ShouldHandleEmptyMailString",
			`{"mail": ""}`,
			"",
			func(t *testing.T, d *UserDetailsExtended) {
				assert.Nil(t, d.UserDetails)
			},
		},
		{
			"ShouldHandleNoSpecialFields",
			`{"gender": "male"}`,
			"",
			func(t *testing.T, d *UserDetailsExtended) {
				assert.Equal(t, "male", d.Gender)
				assert.Empty(t, d.Password)
				assert.Nil(t, d.Profile)
				assert.Nil(t, d.Picture)
				assert.Nil(t, d.Website)
				assert.Nil(t, d.Locale)
			},
		},
		{
			"ShouldErrorOnInvalidPassword",
			`{"password": 123}`,
			"invalid password:",
			nil,
		},
		{
			"ShouldErrorOnInvalidProfileType",
			`{"profile": 123}`,
			"invalid profile:",
			nil,
		},
		{
			"ShouldErrorOnInvalidPictureType",
			`{"picture": 123}`,
			"invalid picture:",
			nil,
		},
		{
			"ShouldErrorOnInvalidWebsiteType",
			`{"website": 123}`,
			"invalid website:",
			nil,
		},
		{
			"ShouldErrorOnInvalidLocaleType",
			`{"locale": 123}`,
			"invalid locale:",
			nil,
		},
		{
			"ShouldErrorOnInvalidLocaleValue",
			`{"locale": "!!!not-a-locale!!!"}`,
			"invalid locale:",
			nil,
		},
		{
			"ShouldErrorOnInvalidMailType",
			`{"mail": 123}`,
			"mail must be a string or array of strings:",
			nil,
		},
		{
			"ShouldErrorOnMalformedJSON",
			`{`,
			"unexpected end of JSON input",
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := &UserDetailsExtended{}

			err := json.Unmarshal([]byte(tc.data), d)

			if tc.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err)

				return
			}

			require.NoError(t, err)

			if tc.asserts != nil {
				tc.asserts(t, d)
			}
		})
	}
}

func TestUserDetailsExtended_JSONRoundTrip(t *testing.T) {
	tag, err := language.Parse("en-US")
	require.NoError(t, err)

	have := &UserDetailsExtended{
		GivenName:  "john",
		FamilyName: "smith",
		Profile:    &url.URL{Scheme: "https", Host: "example.com", Path: "/profile"},
		Picture:    &url.URL{Scheme: "https", Host: "example1.com"},
		Website:    &url.URL{Scheme: "https", Host: "example2.com"},
		Locale:     &tag,
		UserDetails: &UserDetails{
			Username: "john",
			Emails:   []string{"john@example.com"},
		},
		Extra: map[string]any{"example": "value"},
	}

	data, err := json.Marshal(have)
	require.NoError(t, err)

	result := &UserDetailsExtended{}

	require.NoError(t, json.Unmarshal(data, result))

	assert.Equal(t, have.GivenName, result.GivenName)
	assert.Equal(t, have.FamilyName, result.FamilyName)
	assert.Equal(t, have.GetProfile(), result.GetProfile())
	assert.Equal(t, have.GetPicture(), result.GetPicture())
	assert.Equal(t, have.GetWebsite(), result.GetWebsite())
	assert.Equal(t, have.GetLocale(), result.GetLocale())
	assert.Equal(t, have.Emails, result.Emails)
	assert.Equal(t, have.Extra, result.Extra)
	assert.Empty(t, result.Password)
}
