package validator

import "github.com/authelia/authelia/v4/internal/configuration/schema"

// ValidateAdministration validates and updates the administration configuration.
func ValidateAdministration(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Administration.AdminGroup == "" {
		config.Administration.AdminGroup = schema.DefaultAdministrationConfiguration.AdminGroup
	}
}
