package expression

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewUserAttributes(t *testing.T) {
	config := &schema.Configuration{}

	attributes := NewUserAttributes(config)

	assert.NotNil(t, attributes)
	assert.NoError(t, attributes.StartupCheck())
}

func TestStartupCheckError(t *testing.T) {
	config := &schema.Configuration{}

	attributes := &UserAttributesExpressions{}

	assert.NotNil(t, attributes)
	assert.EqualError(t, attributes.StartupCheck(), "error reading config: no authentication backend configured")

	attributes.config = config
	assert.EqualError(t, attributes.StartupCheck(), "error reading config: no authentication backend configured")

	attributes.startup = true
	assert.NoError(t, attributes.StartupCheck())
}

func TestNewUserAttributesLDAP(t *testing.T) {
	testCases := []struct {
		name     string
		config   *schema.Configuration
		expected func(t *testing.T, attributes UserAttributeResolver)
	}{
		{
			"ShouldHandleLDAP",
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					LDAP: &schema.AuthenticationBackendLDAP{},
				},
			},
			nil,
		},
		{
			"ShouldHandleLDAPWithDefinitions",
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					LDAP: &schema.AuthenticationBackendLDAP{},
				},
				Definitions: schema.Definitions{
					UserAttributes: map[string]schema.UserAttribute{
						"example":  {Expression: "'abc' in groups"},
						"example2": {Expression: "groups[0]"},
						"example3": {Expression: `'admin' in groups ? 10 : 5`},
					},
				},
			},
			nil,
		},
		{
			"ShouldHandleLDAPWithExtraAttributes",
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					LDAP: &schema.AuthenticationBackendLDAP{
						Attributes: schema.AuthenticationBackendLDAPAttributes{
							DistinguishedName: "dn",
							Username:          "sAMAccountName",
							DisplayName:       "displayName",
							FamilyName:        "sn",
							GivenName:         "givenName",
							MiddleName:        "middle",
							Nickname:          "nickname",
							Gender:            "gender",
							Birthdate:         "birthdate",
							Website:           "website",
							Profile:           "profile",
							Picture:           "photoURL",
							ZoneInfo:          "zone",
							Locale:            "locale",
							PhoneNumber:       "phoneNumber",
							PhoneExtension:    "phoneExtension",
							StreetAddress:     "street",
							Locality:          "locality",
							Region:            "region",
							PostalCode:        "postalCode",
							Country:           "co",
							Mail:              "mail",
							MemberOf:          "memberOf",
							GroupName:         "cn",
							Extra: map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
								"exampleStr": {
									AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
										ValueType: "string",
									},
								},
								"exampleInt": {
									AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
										ValueType: "integer",
									},
								},
								"exampleBool": {
									AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
										ValueType: "boolean",
									},
								},
								"exampleStrMv": {
									AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
										ValueType:   "string",
										MultiValued: true,
									},
								},
								"exampleIntMv": {
									AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
										ValueType:   "integer",
										MultiValued: true,
									},
								},
								"exampleBoolMv": {
									AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
										ValueType:   "boolean",
										MultiValued: true,
									},
								},
							},
						},
					},
				},
				Definitions: schema.Definitions{
					UserAttributes: map[string]schema.UserAttribute{
						"example":  {Expression: "'abc' in groups"},
						"example2": {Expression: "groups[0]"},
						"example3": {Expression: `'admin' in groups ? 10 : 5`},
					},
				},
			},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attributes := NewUserAttributes(tc.config)

			assert.NotNil(t, attributes)
			assert.NoError(t, attributes.StartupCheck())
		})
	}
}

func TestNewUserAttributesFile(t *testing.T) {
	config := &schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			File: &schema.AuthenticationBackendFile{},
		},
	}

	attributes := NewUserAttributes(config)

	assert.NotNil(t, attributes)
	assert.NoError(t, attributes.StartupCheck())

	config.Definitions.UserAttributes = map[string]schema.UserAttribute{
		"example": {Expression: "'abc' in groups"},
	}

	attributes = NewUserAttributes(config)

	assert.NotNil(t, attributes)
	assert.NoError(t, attributes.StartupCheck())

	config.AuthenticationBackend.File.ExtraAttributes = map[string]schema.AuthenticationBackendExtraAttribute{
		"exampleStr": {
			ValueType: "string",
		},
		"exampleInt": {
			ValueType: "integer",
		},
		"exampleBool": {
			ValueType: "boolean",
		},
		"exampleStrMv": {
			ValueType:   "string",
			MultiValued: true,
		},
		"exampleIntMv": {
			ValueType:   "integer",
			MultiValued: true,
		},
		"exampleBoolMv": {
			ValueType:   "boolean",
			MultiValued: true,
		},
	}

	attributes = NewUserAttributes(config)

	assert.NotNil(t, attributes)
	assert.NoError(t, attributes.StartupCheck())
}

func TestUserAttributesWhiteBoxWithParent(t *testing.T) {
	parent := NewMapActivation(nil, map[string]any{
		AttributeUserUsername:                         "notjsmith",
		AttributeUserGroups:                           []string{"notadmin"},
		AttributeUserDisplayName:                      "notJohn Smith",
		AttributeUserEmail:                            "notjsmith@example.com",
		AttributeUserEmails:                           []string{"notjsmith@example.com", "notalt@example.com"},
		AttributeUserEmailsExtra:                      []string{"notalt@example.com"},
		AttributeUserEmailVerified:                    false,
		AttributeUserGivenName:                        "notjohn",
		AttributeUserMiddleName:                       "notjones",
		AttributeUserFamilyName:                       "notsmith",
		AttributeUserNickname:                         "notjohnny",
		AttributeUserProfile:                          "https://notauthellia.com/jsmith",
		AttributeUserPicture:                          "https://notauthellia.com/jsmith.jpg",
		AttributeUserWebsite:                          "https://notauthellia.com",
		AttributeUserGender:                           "notmale",
		AttributeUserBirthdate:                        "not2020",
		AttributeUserZoneInfo:                         "notzoney",
		AttributeUserLocale:                           "not",
		AttributeUserPhoneNumber:                      "not123567",
		AttributeUserPhoneNumberRFC3966:               "not123567;ext=123",
		AttributeUserPhoneExtension:                   "not123",
		AttributeUserPhoneNumberVerified:              true,
		AttributeUserAddress:                          map[string]any{"country": "notUS", "locality": "notGeneral", "postal_code": "not445500", "region": "notRegion", "street_address": "not123 Bay St"},
		AttributeUserStreetAddress:                    "not123 Bay St",
		AttributeUserLocality:                         "notGeneral",
		AttributeUserRegion:                           "notRegion",
		AttributeUserPostalCode:                       "not445500",
		AttributeUserCountry:                          "notUS",
		AttributeUserUpdatedAt:                        int64(100000009),
		AttributeOpenIDAuthorizationRequestClaimValue: int64(1234),
	})

	activation := NewMapActivation(parent, map[string]any{
		AttributeUserUsername:            "jsmith",
		AttributeUserGroups:              []string{"admin"},
		AttributeUserDisplayName:         "John Smith",
		AttributeUserEmail:               "jsmith@example.com",
		AttributeUserEmails:              []string{"jsmith@example.com", "alt@example.com"},
		AttributeUserEmailsExtra:         []string{"alt@example.com"},
		AttributeUserEmailVerified:       true,
		AttributeUserGivenName:           "john",
		AttributeUserMiddleName:          "jones",
		AttributeUserFamilyName:          "smith",
		AttributeUserNickname:            "johnny",
		AttributeUserProfile:             "https://authellia.com/jsmith",
		AttributeUserPicture:             "https://authellia.com/jsmith.jpg",
		AttributeUserWebsite:             "https://authellia.com",
		AttributeUserGender:              "male",
		AttributeUserBirthdate:           "2020",
		AttributeUserZoneInfo:            "zoney",
		AttributeUserLocale:              "",
		AttributeUserPhoneNumber:         "123567",
		AttributeUserPhoneNumberRFC3966:  "123567;ext=123",
		AttributeUserPhoneExtension:      "123",
		AttributeUserPhoneNumberVerified: false,
		AttributeUserAddress:             map[string]any{"country": "US", "locality": "General", "postal_code": "445500", "region": "Region", "street_address": "123 Bay St"},
		AttributeUserStreetAddress:       "123 Bay St",
		AttributeUserLocality:            "General",
		AttributeUserRegion:              "Region",
		AttributeUserPostalCode:          "445500",
		AttributeUserCountry:             "US",
		AttributeUserUpdatedAt:           int64(100000000),
	})

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
		{AttributeOpenIDAuthorizationRequestClaimValue, int64(1234), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := getStandardCELEnvOpts()

			env, err := cel.NewEnv(opts...)
			require.NoError(t, err)

			ast, issues := env.Compile(tc.name)
			require.NoError(t, issues.Err())

			program, err := env.Program(ast)
			require.NoError(t, err)

			actual, _, _ := program.Eval(activation)
			assert.Equal(t, tc.expected, toNativeValue(actual))
		})
	}
}
