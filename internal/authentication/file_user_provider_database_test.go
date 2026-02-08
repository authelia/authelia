package authentication

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-crypt/crypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
)

func TestDatabaseModel_Read(t *testing.T) {
	model := &FileDatabaseModel{}

	dir := t.TempDir()

	_, err := os.Create(filepath.Join(dir, "users_database.yml"))

	assert.NoError(t, err)

	assert.EqualError(t, model.Read(filepath.Join(dir, "users_database.yml")), "no file content")

	assert.NoError(t, os.Mkdir(filepath.Join(dir, "x"), 0000))

	f := filepath.Join(dir, "x", "users_database.yml")

	assert.EqualError(t, model.Read(f), fmt.Sprintf("failed to read the '%s' file: open %s: permission denied", f, f))

	f = filepath.Join(dir, "schema.yml")

	file, err := os.Create(f)
	assert.NoError(t, err)

	_, err = file.WriteString("users:\n\tjohn: {}")

	assert.NoError(t, err)

	assert.EqualError(t, model.Read(f), "could not parse the YAML database: yaml: while scanning for the next token at line 2: found character that cannot start any token")
}

func TestDatabaseModelExtended(t *testing.T) {
	mustParseURI := func(in string) *url.URL {
		if u, err := url.ParseRequestURI(in); err != nil {
			panic(err)
		} else {
			return u
		}
	}

	mustParseTag := func(in string) *language.Tag {
		tag, err := language.Parse(in)
		if err != nil {
			panic(err)
		}

		return &tag
	}

	digest, digestErr := crypt.Decode("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng")
	require.NoError(t, digestErr)

	testCases := []struct {
		name     string
		have     *FileDatabaseUserDetailsModel
		expected *FileUserDatabaseUserDetails
		details  *UserDetailsExtended
		extra    map[string]expression.ExtraAttribute
		err      string
		errExtra string
	}{
		{
			"ShouldHandleEmptyStruct",
			&FileDatabaseUserDetailsModel{},
			nil,
			nil,
			nil,
			"error occurred decoding the password hash for 'example': provided encoded hash has an invalid format: the digest doesn't begin with the delimiter '$' and is not one of the other understood formats",
			"",
		},
		{
			"ShouldHandleBadLocale",
			&FileDatabaseUserDetailsModel{
				Password: "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				Locale:   "Example",
			},
			nil,
			nil,
			nil,
			"error occurred parsing user details for 'example': failed to parse the locale attribute with value 'Example': language: tag is not well-formed",
			"",
		},
		{
			"ShouldHandleMinimal",
			&FileDatabaseUserDetailsModel{
				Password: "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
			},
			&FileUserDatabaseUserDetails{
				Username: "example",
				Password: schema.NewPasswordDigest(digest),
			},
			&UserDetailsExtended{
				UserDetails: &UserDetails{
					Username: "example",
					Emails:   []string(nil),
					Groups:   []string(nil),
				},
			},
			nil,
			"",
			"",
		},
		{
			"ShouldHandleMaximum",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        "https://authelia.com",
				Profile:        "https://authelia.com/jsmith",
				Picture:        "https://authelia.com/jsmith.jpg",
				ZoneInfo:       "unzone",
				Locale:         "en-US",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": byte('1'),
				},
			},
			&FileUserDatabaseUserDetails{
				Username:       "example",
				Password:       schema.NewPasswordDigest(digest),
				DisplayName:    "John Smith",
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        mustParseURI("https://authelia.com"),
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				Address:        &FileUserDatabaseUserDetailsAddressModel{StreetAddress: "123 Baker St", Locality: "Internet", Region: "Online", PostalCode: "98765", Country: "US"},
				Extra:          map[string]any{"example": byte('1')},
			},
			&UserDetailsExtended{
				GivenName:      "john",
				FamilyName:     "smith",
				MiddleName:     "jacob",
				Nickname:       "johnny",
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				Website:        mustParseURI("https://authelia.com"),
				Gender:         "male",
				Birthdate:      "2025",
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &UserDetailsAddress{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": byte('1'),
				},
				UserDetails: &UserDetails{
					Username:    "example",
					DisplayName: "John Smith",
					Emails:      []string{"jsmith@example.com"},
					Groups:      []string{"abc"},
				},
			},
			map[string]expression.ExtraAttribute{
				"example": schema.AuthenticationBackendExtraAttribute{
					MultiValued: false,
					ValueType:   "boolean",
				},
			},
			"",
			"error occurred validating extra attributes for user 'example': attribute 'example' has the unknown type 'uint8'",
		},
		{
			"ShouldHandleBadLocale",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        "https://authelia.com",
				Profile:        "https://authelia.com/jsmith",
				Picture:        "https://authelia.com/jsmith.jpg",
				ZoneInfo:       "unzone",
				Locale:         "en123kn12kj3n123",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": true,
				},
			},
			nil,
			nil,
			map[string]expression.ExtraAttribute{
				"example": schema.AuthenticationBackendExtraAttribute{
					MultiValued: false,
					ValueType:   "boolean",
				},
			},
			"error occurred parsing user details for 'example': failed to parse the locale attribute with value 'en123kn12kj3n123': language: tag is not well-formed",
			"",
		},
		{
			"ShouldHandleBadProfile",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        "https://authelia.com",
				Profile:        "notascheme://authelia.com/jsmith",
				Picture:        "https://authelia.com/jsmith.jpg",
				ZoneInfo:       "unzone",
				Locale:         "en-US",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": "1",
				},
			},
			nil,
			nil,
			map[string]expression.ExtraAttribute{
				"example": schema.AuthenticationBackendExtraAttribute{
					MultiValued: false,
					ValueType:   "string",
				},
			},
			"error occurred parsing user details for 'example': failed to parse the profile attribute with value 'notascheme://authelia.com/jsmith': invalid URL scheme 'notascheme' for profile attribute",
			"",
		},
		{
			"ShouldHandleBadWebsite",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        "https://@@:)!(@*#U!()@#!@.com",
				Profile:        "https://authelia.com/jsmith",
				Picture:        "https://authelia.com/jsmith.jpg",
				ZoneInfo:       "unzone",
				Locale:         "en-US",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": 1,
				},
			},
			nil,
			nil,
			map[string]expression.ExtraAttribute{},
			"error occurred parsing user details for 'example': failed to parse the website attribute with value 'https://@@:)!(@*#U!()@#!@.com': parse \"https://@@:)!(@*#U!()@#!@.com\": net/url: invalid userinfo",
			"error occurred validating extra attributes for user 'example': attribute 'example' is unknown",
		},
		{
			"ShouldHandleBadPicture",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        "https://authelia.com",
				Profile:        "https://authelia.com/jsmith",
				Picture:        "ahttps://authelia.com/jsmith.jpg",
				ZoneInfo:       "unzone",
				Locale:         "en-US",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": 1,
				},
			},
			nil,
			nil,
			map[string]expression.ExtraAttribute{
				"example": schema.AuthenticationBackendExtraAttribute{
					MultiValued: false,
					ValueType:   "integer",
				},
			},
			"error occurred parsing user details for 'example': failed to parse the picture attribute with value 'ahttps://authelia.com/jsmith.jpg': invalid URL scheme 'ahttps' for profile attribute",
			"",
		},
		{
			"ShouldHandleWrongExtraTypeMismatch",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        "https://authelia.com",
				Profile:        "https://authelia.com/jsmith",
				Picture:        "https://authelia.com/jsmith.jpg",
				ZoneInfo:       "unzone",
				Locale:         "en-US",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": 1,
				},
			},
			&FileUserDatabaseUserDetails{
				Username:       "example",
				Password:       schema.NewPasswordDigest(digest),
				DisplayName:    "John Smith",
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        mustParseURI("https://authelia.com"),
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				Address:        &FileUserDatabaseUserDetailsAddressModel{StreetAddress: "123 Baker St", Locality: "Internet", Region: "Online", PostalCode: "98765", Country: "US"},
				Extra: map[string]any{
					"example": 1,
				},
			},
			&UserDetailsExtended{
				GivenName:      "john",
				FamilyName:     "smith",
				MiddleName:     "jacob",
				Nickname:       "johnny",
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				Website:        mustParseURI("https://authelia.com"),
				Gender:         "male",
				Birthdate:      "2025",
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &UserDetailsAddress{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": 1,
				},
				UserDetails: &UserDetails{
					Username:    "example",
					DisplayName: "John Smith",
					Emails:      []string{"jsmith@example.com"},
					Groups:      []string{"abc"},
				},
			},
			map[string]expression.ExtraAttribute{
				"example": schema.AuthenticationBackendExtraAttribute{
					MultiValued: false,
					ValueType:   "boolean",
				},
			},
			"",
			"error occurred validating extra attributes for user 'example': attribute 'example' has the known type 'int' but 'boolean' is the expected type",
		},
		{
			"ShouldHandleArrayString",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        "https://authelia.com",
				Profile:        "https://authelia.com/jsmith",
				Picture:        "https://authelia.com/jsmith.jpg",
				ZoneInfo:       "unzone",
				Locale:         "en-US",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": []any{"abc", "123"},
				},
			},
			&FileUserDatabaseUserDetails{
				Username:       "example",
				Password:       schema.NewPasswordDigest(digest),
				DisplayName:    "John Smith",
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        mustParseURI("https://authelia.com"),
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				Address:        &FileUserDatabaseUserDetailsAddressModel{StreetAddress: "123 Baker St", Locality: "Internet", Region: "Online", PostalCode: "98765", Country: "US"},
				Extra: map[string]any{
					"example": []any{"abc", "123"},
				},
			},
			&UserDetailsExtended{
				GivenName:      "john",
				FamilyName:     "smith",
				MiddleName:     "jacob",
				Nickname:       "johnny",
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				Website:        mustParseURI("https://authelia.com"),
				Gender:         "male",
				Birthdate:      "2025",
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &UserDetailsAddress{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": []any{"abc", "123"},
				},
				UserDetails: &UserDetails{
					Username:    "example",
					DisplayName: "John Smith",
					Emails:      []string{"jsmith@example.com"},
					Groups:      []string{"abc"},
				},
			},
			map[string]expression.ExtraAttribute{
				"example": schema.AuthenticationBackendExtraAttribute{
					MultiValued: true,
					ValueType:   "string",
				},
			},
			"",
			"",
		},
		{
			"ShouldHandleArrayInteger",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        "https://authelia.com",
				Profile:        "https://authelia.com/jsmith",
				Picture:        "https://authelia.com/jsmith.jpg",
				ZoneInfo:       "unzone",
				Locale:         "en-US",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": []any{123, 456},
				},
			},
			&FileUserDatabaseUserDetails{
				Username:       "example",
				Password:       schema.NewPasswordDigest(digest),
				DisplayName:    "John Smith",
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        mustParseURI("https://authelia.com"),
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				Address:        &FileUserDatabaseUserDetailsAddressModel{StreetAddress: "123 Baker St", Locality: "Internet", Region: "Online", PostalCode: "98765", Country: "US"},
				Extra: map[string]any{
					"example": []any{123, 456},
				},
			},
			&UserDetailsExtended{
				GivenName:      "john",
				FamilyName:     "smith",
				MiddleName:     "jacob",
				Nickname:       "johnny",
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				Website:        mustParseURI("https://authelia.com"),
				Gender:         "male",
				Birthdate:      "2025",
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &UserDetailsAddress{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": []any{123, 456},
				},
				UserDetails: &UserDetails{
					Username:    "example",
					DisplayName: "John Smith",
					Emails:      []string{"jsmith@example.com"},
					Groups:      []string{"abc"},
				},
			},
			map[string]expression.ExtraAttribute{
				"example": schema.AuthenticationBackendExtraAttribute{
					MultiValued: true,
					ValueType:   "integer",
				},
			},
			"",
			"",
		},
		{
			"ShouldHandleArrayBoolean",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        "https://authelia.com",
				Profile:        "https://authelia.com/jsmith",
				Picture:        "https://authelia.com/jsmith.jpg",
				ZoneInfo:       "unzone",
				Locale:         "en-US",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": []any{true, false},
				},
			},
			&FileUserDatabaseUserDetails{
				Username:       "example",
				Password:       schema.NewPasswordDigest(digest),
				DisplayName:    "John Smith",
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        mustParseURI("https://authelia.com"),
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				Address:        &FileUserDatabaseUserDetailsAddressModel{StreetAddress: "123 Baker St", Locality: "Internet", Region: "Online", PostalCode: "98765", Country: "US"},
				Extra: map[string]any{
					"example": []any{true, false},
				},
			},
			&UserDetailsExtended{
				GivenName:      "john",
				FamilyName:     "smith",
				MiddleName:     "jacob",
				Nickname:       "johnny",
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				Website:        mustParseURI("https://authelia.com"),
				Gender:         "male",
				Birthdate:      "2025",
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &UserDetailsAddress{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": []any{true, false},
				},
				UserDetails: &UserDetails{
					Username:    "example",
					DisplayName: "John Smith",
					Emails:      []string{"jsmith@example.com"},
					Groups:      []string{"abc"},
				},
			},
			map[string]expression.ExtraAttribute{
				"example": schema.AuthenticationBackendExtraAttribute{
					MultiValued: true,
					ValueType:   "boolean",
				},
			},
			"",
			"",
		},
		{
			"ShouldHandleArrayBooleanNotArray",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        "https://authelia.com",
				Profile:        "https://authelia.com/jsmith",
				Picture:        "https://authelia.com/jsmith.jpg",
				ZoneInfo:       "unzone",
				Locale:         "en-US",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": true,
				},
			},
			&FileUserDatabaseUserDetails{
				Username:       "example",
				Password:       schema.NewPasswordDigest(digest),
				DisplayName:    "John Smith",
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				Website:        mustParseURI("https://authelia.com"),
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				Address:        &FileUserDatabaseUserDetailsAddressModel{StreetAddress: "123 Baker St", Locality: "Internet", Region: "Online", PostalCode: "98765", Country: "US"},
				Extra: map[string]any{
					"example": true,
				},
			},
			&UserDetailsExtended{
				GivenName:      "john",
				FamilyName:     "smith",
				MiddleName:     "jacob",
				Nickname:       "johnny",
				Profile:        mustParseURI("https://authelia.com/jsmith"),
				Picture:        mustParseURI("https://authelia.com/jsmith.jpg"),
				Website:        mustParseURI("https://authelia.com"),
				Gender:         "male",
				Birthdate:      "2025",
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &UserDetailsAddress{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": true,
				},
				UserDetails: &UserDetails{
					Username:    "example",
					DisplayName: "John Smith",
					Emails:      []string{"jsmith@example.com"},
					Groups:      []string{"abc"},
				},
			},
			map[string]expression.ExtraAttribute{
				"example": schema.AuthenticationBackendExtraAttribute{
					MultiValued: true,
					ValueType:   "boolean",
				},
			},
			"",
			"error occurred validating extra attributes for user 'example': attribute 'example' has the type 'bool' but '[]boolean' is the expected type",
		},
		{
			"ShouldHandleArrayByte",
			&FileDatabaseUserDetailsModel{
				Password:       "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng",
				DisplayName:    "John Smith",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				ZoneInfo:       "unzone",
				Locale:         "en-US",
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &FileUserDatabaseUserDetailsAddressModel{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": []any{byte('a'), byte('b')},
				},
			},
			&FileUserDatabaseUserDetails{
				Username:       "example",
				Password:       schema.NewPasswordDigest(digest),
				DisplayName:    "John Smith",
				GivenName:      "john",
				MiddleName:     "jacob",
				FamilyName:     "smith",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Email:          "jsmith@example.com",
				Groups:         []string{"abc"},
				Address:        &FileUserDatabaseUserDetailsAddressModel{StreetAddress: "123 Baker St", Locality: "Internet", Region: "Online", PostalCode: "98765", Country: "US"},
				Extra: map[string]any{
					"example": []any{byte('a'), byte('b')},
				},
			},
			&UserDetailsExtended{
				GivenName:      "john",
				FamilyName:     "smith",
				MiddleName:     "jacob",
				Nickname:       "johnny",
				Gender:         "male",
				Birthdate:      "2025",
				ZoneInfo:       "unzone",
				Locale:         mustParseTag("en-US"),
				PhoneNumber:    "129812",
				PhoneExtension: "123",
				Address: &UserDetailsAddress{
					StreetAddress: "123 Baker St",
					Locality:      "Internet",
					Region:        "Online",
					PostalCode:    "98765",
					Country:       "US",
				},
				Extra: map[string]any{
					"example": []any{byte('a'), byte('b')},
				},
				UserDetails: &UserDetails{
					Username:    "example",
					DisplayName: "John Smith",
					Emails:      []string{"jsmith@example.com"},
					Groups:      []string{"abc"},
				},
			},
			map[string]expression.ExtraAttribute{
				"example": schema.AuthenticationBackendExtraAttribute{
					MultiValued: true,
					ValueType:   "boolean",
				},
			},
			"",
			"error occurred validating extra attributes for user 'example': attribute 'example' has the unknown item type 'uint8'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.have.ToDatabaseUserDetailsModel("example")

			if len(tc.err) == 0 {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)

				model := actual.ToUserDetailsModel()

				assert.Equal(t, tc.have, &model)

				details := actual.ToExtendedUserDetails()

				assert.Equal(t, tc.details, details)
			} else {
				assert.EqualError(t, err, tc.err)
			}

			if tc.extra != nil {
				if tc.errExtra == "" {
					assert.NoError(t, tc.have.ValidateExtra("example", tc.extra))
				} else {
					assert.EqualError(t, tc.have.ValidateExtra("example", tc.extra), tc.errExtra)
				}
			}
		})
	}
}
