package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
