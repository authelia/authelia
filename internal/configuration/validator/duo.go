package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateDuo validates and updates the Duo configuration.
func ValidateDuo(config *schema.Configuration, validator *schema.StructValidator) {
	if config.DuoAPI.Disable {
		return
	}

	if config.DuoAPI.Hostname == "" && config.DuoAPI.IntegrationKey == "" && config.DuoAPI.SecretKey == "" {
		config.DuoAPI.Disable = true
	}

	if config.DuoAPI.Disable {
		return
	}

	if config.DuoAPI.Hostname == "" {
		validator.Push(fmt.Errorf(errFmtDuoMissingOption, "hostname"))
	}

	if config.DuoAPI.IntegrationKey == "" {
		validator.Push(fmt.Errorf(errFmtDuoMissingOption, "integration_key"))
	}

	if config.DuoAPI.SecretKey == "" {
		validator.Push(fmt.Errorf(errFmtDuoMissingOption, "secret_key"))
	}
}
