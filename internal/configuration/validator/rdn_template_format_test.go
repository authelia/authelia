package validator

import (
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/templates"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestRDNTemplateFormats(t *testing.T) {
	_ = schema.AuthenticationBackendLDAPAttributes{
		Username:       "uid",
		DisplayName:    "displayName",
		Mail:           "mail",
		GivenName:      "givenName",
		FamilyName:     "sn",
		MiddleName:     "middleName",
		Nickname:       "nickname",
		Profile:        "profileURL",
		Picture:        "jpegPhoto",
		Website:        "labeledURI",
		Gender:         "gender",
		Birthdate:      "birthdate",
		ZoneInfo:       "timezone",
		Locale:         "preferredLanguage",
		PhoneNumber:    "telephoneNumber",
		PhoneExtension: "extension",
		MemberOf:       "memberOf",
		StreetAddress:  "street",
		Locality:       "l",
		Region:         "st",
		PostalCode:     "postalCode",
		Country:        "c",
	}

	testCases := []struct {
		name           string
		dn             string
		format         string
		templateData   map[string]interface{}
		expectedOutput string
		shouldFail     bool
	}{
		{
			name:   "SimpleGivenAndFamilyName",
			dn:     "cn",
			format: "[[ .given_name ]] [[ .family_name ]]",
			templateData: map[string]interface{}{
				"given_name":  "John",
				"family_name": "Doe",
			},
			expectedOutput: "cn=John Doe",
		},
		{
			name:   "UsernameOnly",
			dn:     "uid",
			format: "[[ .username ]]",
			templateData: map[string]interface{}{
				"username": "jdoe",
			},
			expectedOutput: "uid=jdoe",
		},
		{
			name:   "FullNameComplex",
			dn:     "cn",
			format: "[[ .given_name ]] [[ .middle_name ]] [[ .family_name ]]",
			templateData: map[string]interface{}{
				"given_name":  "John",
				"middle_name": "Q",
				"family_name": "Doe",
			},
			expectedOutput: "cn=John Q Doe",
		},
		{
			name:   "WithEmail",
			dn:     "cn",
			format: "[[ .given_name ]] [[ .family_name ]]",
			templateData: map[string]interface{}{
				"given_name":  "John",
				"family_name": "Doe",
				"emails":      []string{"john.doe@example.com"},
			},
			expectedOutput: "cn=John Doe",
		},
		{
			name:   "DottedFormat",
			dn:     "uid",
			format: "[[ .given_name ]].[[ .family_name ]]",
			templateData: map[string]interface{}{
				"given_name":  "John",
				"family_name": "Doe",
			},
			expectedOutput: "uid=John.Doe",
		},
		{
			name:   "LowercaseTransform",
			dn:     "uid",
			format: "[[ .username | lower ]]",
			templateData: map[string]interface{}{
				"username": "JDoe",
			},
			expectedOutput: "uid=jdoe",
		},
		{
			name:   "WithDisplayName",
			dn:     "cn",
			format: "[[ .display_name ]]",
			templateData: map[string]interface{}{
				"display_name": "John Doe",
			},
			expectedOutput: "cn=John Doe",
		},
		{
			name:   "MultipleAttributes",
			dn:     "cn",
			format: "[[ .full_name ]]",
			templateData: map[string]interface{}{
				"full_name":    "John Doe",
				"phone_number": "+1-555-0123",
			},
			expectedOutput: "cn=John Doe",
		},
		{
			name:   "WithNickname",
			dn:     "cn",
			format: "[[ .nickname ]]",
			templateData: map[string]interface{}{
				"nickname": "Johnny",
			},
			expectedOutput: "cn=Johnny",
		},
		{
			name:   "ConditionalWithField",
			dn:     "cn",
			format: "[[ if .given_name ]][[ .given_name ]][[ else ]]Unknown[[ end ]]",
			templateData: map[string]interface{}{
				"given_name": "John",
			},
			expectedOutput: "cn=John",
		},
		{
			name:   "ConditionalEmpty",
			dn:     "cn",
			format: "[[ if .given_name ]][[ .given_name ]][[ else ]]Unknown[[ end ]]",
			templateData: map[string]interface{}{
				"given_name": "",
			},
			expectedOutput: "cn=Unknown",
		},
		{
			name:   "AddressField",
			dn:     "l",
			format: "[[ .address.locality ]]",
			templateData: map[string]interface{}{
				"address": map[string]interface{}{
					"locality": "San Francisco",
				},
			},
			expectedOutput: "l=San Francisco",
		},
		{
			name:   "MultipleAddressFields",
			dn:     "street",
			format: "[[ .address.street_address ]]",
			templateData: map[string]interface{}{
				"address": map[string]interface{}{
					"street_address": "123 Main St",
					"locality":       "San Francisco",
					"region":         "CA",
				},
			},
			expectedOutput: "street=123 Main St",
		},
		{
			name:   "CompleteRDN",
			dn:     "cn",
			format: "[[ .given_name ]] [[ .family_name ]]",
			templateData: map[string]interface{}{
				"given_name":  "John",
				"family_name": "Doe",
			},
			expectedOutput: "cn=John Doe",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := template.New("rdn").Delims("[[", "]]").Funcs(templates.FuncMap()).Parse(tc.format)

			if tc.shouldFail {
				if err != nil {
					return
				}

				var output strings.Builder

				err = tmpl.Execute(&output, tc.templateData)
				assert.Error(t, err, "Expected template execution to fail")

				return
			}

			require.NoError(t, err, "Template parsing should succeed")

			var rdnValue strings.Builder

			err = tmpl.Execute(&rdnValue, tc.templateData)
			require.NoError(t, err, "Template execution should succeed")

			fullRDN := tc.dn + "=" + rdnValue.String()

			assert.Equal(t, tc.expectedOutput, fullRDN, "Template output should match expected")
		})
	}
}

func TestRDNTemplateValidation(t *testing.T) {
	testCases := []struct {
		name              string
		ldapAttributes    schema.AuthenticationBackendLDAPAttributes
		rdnFormat         string
		shouldHaveErrors  bool
		expectedErrorText string
	}{
		{
			name: "ValidFieldsWithMapping",
			ldapAttributes: schema.AuthenticationBackendLDAPAttributes{
				GivenName:  "givenName",
				FamilyName: "sn",
			},
			rdnFormat:        "cn=[[ .given_name ]] [[ .family_name ]]",
			shouldHaveErrors: false,
		},
		{
			name: "InvalidFieldWithoutMapping",
			ldapAttributes: schema.AuthenticationBackendLDAPAttributes{
				GivenName: "givenName",
			},
			rdnFormat:         "cn=[[ .given_name ]] [[ .phone_number ]]",
			shouldHaveErrors:  true,
			expectedErrorText: "phone_number",
		},
		{
			name: "AllStandardFields",
			ldapAttributes: schema.AuthenticationBackendLDAPAttributes{
				Username:       "uid",
				DisplayName:    "displayName",
				Mail:           "mail",
				GivenName:      "givenName",
				FamilyName:     "sn",
				MiddleName:     "middleName",
				Nickname:       "nickname",
				PhoneNumber:    "telephoneNumber",
				PhoneExtension: "extension",
			},
			rdnFormat:        "cn=[[ .given_name ]] [[ .middle_name ]] [[ .family_name ]],uid=[[ .username ]],mail=[[ index .emails 0 ]]",
			shouldHaveErrors: false,
		},
		{
			name: "AddressFieldsValid",
			ldapAttributes: schema.AuthenticationBackendLDAPAttributes{
				StreetAddress: "street",
				Locality:      "l",
				Region:        "st",
			},
			rdnFormat:        "street=[[ .address.street_address ]],l=[[ .address.locality ]]",
			shouldHaveErrors: false,
		},
		{
			name: "MixedValidAndInvalid",
			ldapAttributes: schema.AuthenticationBackendLDAPAttributes{
				GivenName: "givenName",
			},
			rdnFormat:         "cn=[[ .given_name ]] [[ .undefined_field ]]",
			shouldHaveErrors:  true,
			expectedErrorText: "undefined_field",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Address:      mustParseAddress("ldap://127.0.0.1"),
					User:         "cn=admin,dc=example,dc=com",
					Password:     "password",
					UsersFilter:  "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
					GroupsFilter: "(member={dn})",
					Attributes:   tc.ldapAttributes,
					UserManagement: schema.AuthenticationBackendLDAPUserManagement{
						CreatedUsersRDNFormat: tc.rdnFormat,
					},
				},
			}

			validator := schema.NewStructValidator()
			validateLDAPAuthenticationBackendUserManagementRDNTemplate(config, validator)

			if tc.shouldHaveErrors {
				assert.NotEmpty(t, validator.Errors(), "Expected validation errors")

				if tc.expectedErrorText != "" {
					found := false

					for _, err := range validator.Errors() {
						if strings.Contains(err.Error(), tc.expectedErrorText) {
							found = true
							break
						}
					}

					assert.True(t, found, "Expected error message to contain '%s'", tc.expectedErrorText)
				}
			} else {
				assert.Empty(t, validator.Errors(), "Expected no validation errors, got: %v", validator.Errors())
			}
		})
	}
}
