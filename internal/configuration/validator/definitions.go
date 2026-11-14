// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"maps"
	"slices"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateDefinitions validates the definitions configuration.
func ValidateDefinitions(config *schema.Configuration, validator *schema.StructValidator) {
	for name := range config.Definitions.UserAttributes {
		if !isUserAttributeDefinitionNameValid(name, config) {
			validator.Push(fmt.Errorf(errFmtDefinitionsUserAttributesReservedOrDefined, name, name))
		}
	}

	validateDefinitionsURIParsers(config, validator)
}

func validateDefinitionsURIParsers(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Definitions.URIParsers == nil {
		config.Definitions.URIParsers = map[string]schema.URIParser{}
	}

	if _, ok := config.Definitions.URIParsers[schema.DefinitionURIParserDefault]; !ok {
		config.Definitions.URIParsers[schema.DefinitionURIParserDefault] = schema.URIParser{}
	}

	for _, name := range slices.Sorted(maps.Keys(config.Definitions.URIParsers)) {
		parser := config.Definitions.URIParsers[name]

		for _, option := range []struct {
			name   string
			values []string
		}{{"allow", parser.Allow}, {"reject", parser.Reject}} {
			for _, value := range option.values {
				if _, ok := utils.URLAmbiguityFromName(value); !ok {
					validator.Push(fmt.Errorf(errFmtDefinitionsURIParsersName, name, option.name, utils.StringJoinOr(utils.URLAmbiguityNames()), value))
				}
			}
		}

		for _, value := range parser.Allow {
			if _, ok := utils.URLAmbiguityFromName(value); ok && utils.IsStringInSlice(value, parser.Reject) {
				validator.Push(fmt.Errorf(errFmtDefinitionsURIParsersConflict, name, value))
			}
		}
	}
}
