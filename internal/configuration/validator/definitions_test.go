// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
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

	t.Run("ShouldAddDefaultURIParser", func(t *testing.T) {
		config := &schema.Configuration{}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		assert.Len(t, validator.Errors(), 0)
		assert.Equal(t, map[string]schema.URIParser{schema.DefinitionURIParserDefault: {}}, config.Definitions.URIParsers)
	})

	t.Run("ShouldNotReplaceConfiguredDefaultURIParser", func(t *testing.T) {
		config := &schema.Configuration{
			Definitions: schema.Definitions{
				URIParsers: map[string]schema.URIParser{
					schema.DefinitionURIParserDefault: {Allow: []string{"user_info"}},
				},
			},
		}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		assert.Len(t, validator.Errors(), 0)
		assert.Equal(t, map[string]schema.URIParser{schema.DefinitionURIParserDefault: {Allow: []string{"user_info"}}}, config.Definitions.URIParsers)
	})

	t.Run("ShouldErrUnknownURIParserClass", func(t *testing.T) {
		config := &schema.Configuration{
			Definitions: schema.Definitions{
				URIParsers: map[string]schema.URIParser{
					"custom": {Allow: []string{"nope"}, Reject: []string{"nope"}},
				},
			},
		}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		require.Len(t, validator.Errors(), 2)
		assert.EqualError(t, validator.Errors()[0], "definitions: uri_parsers: custom: option 'allow' must only include the values "+utils.StringJoinOr(utils.URLAmbiguityNames())+" but has 'nope'")
		assert.EqualError(t, validator.Errors()[1], "definitions: uri_parsers: custom: option 'reject' must only include the values "+utils.StringJoinOr(utils.URLAmbiguityNames())+" but has 'nope'")
	})

	t.Run("ShouldErrConflictingURIParserClass", func(t *testing.T) {
		config := &schema.Configuration{
			Definitions: schema.Definitions{
				URIParsers: map[string]schema.URIParser{
					"custom": {Allow: []string{"user_info"}, Reject: []string{"user_info"}},
				},
			},
		}
		validator := schema.NewStructValidator()

		ValidateDefinitions(config, validator)

		require.Len(t, validator.Errors(), 1)
		assert.EqualError(t, validator.Errors()[0], "definitions: uri_parsers: custom: the value 'user_info' can't be in both the 'allow' and 'reject' options")
	})
}
