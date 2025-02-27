package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func ValidateDefinitions(config *schema.Configuration, validator *schema.StructValidator) {
	for name := range config.Definitions.UserAttributes {
		if !isUserAttributeDefinitionNameValid(name, config) {
			validator.Push(fmt.Errorf(errFmtDefinitionsUserAttributesReservedOrDefined, name, name))
		}
	}
}
