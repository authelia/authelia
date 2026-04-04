package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidateDefinitions(t *testing.T) {
	t.Run("ShouldSucceedWithNoAttributes", func(t *testing.T) {
		config := &schema.Configuration{}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		assert.Len(t, validator.Warnings(), 0)
		assert.Len(t, validator.Errors(), 0)
	})

	t.Run("ShouldSucceedWithValidAttribute", func(t *testing.T) {
		config := &schema.Configuration{
			Definitions: schema.Definitions{
				UserAttributes: map[string]schema.UserAttribute{
					"custom_attr": {Expression: "'value'"},
				},
			},
		}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		assert.Len(t, validator.Warnings(), 0)
		assert.Len(t, validator.Errors(), 0)
	})

	t.Run("ShouldSucceedWithMultipleValidAttributes", func(t *testing.T) {
		config := &schema.Configuration{
			Definitions: schema.Definitions{
				UserAttributes: map[string]schema.UserAttribute{
					"attr_one": {Expression: "'one'"},
					"attr_two": {Expression: "'two'"},
				},
			},
		}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		assert.Len(t, validator.Warnings(), 0)
		assert.Len(t, validator.Errors(), 0)
	})

	t.Run("ShouldErrReservedAttributeName", func(t *testing.T) {
		config := &schema.Configuration{
			Definitions: schema.Definitions{
				UserAttributes: map[string]schema.UserAttribute{
					"username": {Expression: "'value'"},
				},
			},
		}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		assert.Len(t, validator.Warnings(), 0)
		require.Len(t, validator.Errors(), 1)
		assert.EqualError(t, validator.Errors()[0], "definitions: user_attributes: username: attribute name 'username' is either reserved or already defined in the authentication backend")
	})

	t.Run("ShouldErrAttributeConflictsWithFileExtraAttribute", func(t *testing.T) {
		config := &schema.Configuration{
			AuthenticationBackend: schema.AuthenticationBackend{
				File: &schema.AuthenticationBackendFile{
					ExtraAttributes: map[string]schema.AuthenticationBackendExtraAttribute{
						"custom_attr": {ValueType: "string"},
					},
				},
			},
			Definitions: schema.Definitions{
				UserAttributes: map[string]schema.UserAttribute{
					"custom_attr": {Expression: "'value'"},
				},
			},
		}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		assert.Len(t, validator.Warnings(), 0)
		require.Len(t, validator.Errors(), 1)
		assert.EqualError(t, validator.Errors()[0], "definitions: user_attributes: custom_attr: attribute name 'custom_attr' is either reserved or already defined in the authentication backend")
	})

	t.Run("ShouldErrAttributeConflictsWithLDAPExtraAttributeByKey", func(t *testing.T) {
		config := &schema.Configuration{
			AuthenticationBackend: schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						Extra: map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
							"custom_attr": {AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "string"}},
						},
					},
				},
			},
			Definitions: schema.Definitions{
				UserAttributes: map[string]schema.UserAttribute{
					"custom_attr": {Expression: "'value'"},
				},
			},
		}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		assert.Len(t, validator.Warnings(), 0)
		require.Len(t, validator.Errors(), 1)
		assert.EqualError(t, validator.Errors()[0], "definitions: user_attributes: custom_attr: attribute name 'custom_attr' is either reserved or already defined in the authentication backend")
	})

	t.Run("ShouldErrAttributeConflictsWithLDAPExtraAttributeByName", func(t *testing.T) {
		config := &schema.Configuration{
			AuthenticationBackend: schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						Extra: map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
							"ldap_key": {
								Name: "custom_attr",

								AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "string"},
							},
						},
					},
				},
			},
			Definitions: schema.Definitions{
				UserAttributes: map[string]schema.UserAttribute{
					"custom_attr": {Expression: "'value'"},
				},
			},
		}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		assert.Len(t, validator.Warnings(), 0)
		require.Len(t, validator.Errors(), 1)
		assert.EqualError(t, validator.Errors()[0], "definitions: user_attributes: custom_attr: attribute name 'custom_attr' is either reserved or already defined in the authentication backend")
	})

	t.Run("ShouldNotConflictWithDifferentLDAPExtraAttributeName", func(t *testing.T) {
		config := &schema.Configuration{
			AuthenticationBackend: schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						Extra: map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
							"ldap_key": {
								Name:                                "other_name",
								AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "string"},
							},
						},
					},
				},
			},
			Definitions: schema.Definitions{
				UserAttributes: map[string]schema.UserAttribute{
					"custom_attr": {Expression: "'value'"},
				},
			},
		}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		assert.Len(t, validator.Warnings(), 0)
		assert.Len(t, validator.Errors(), 0)
	})
}
