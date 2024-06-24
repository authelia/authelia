package validator

import (
	"fmt"
	"os"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateCustomLocales validates and updates the Password Policy configuration.
func ValidateCustomLocales(config *schema.CustomLocales, validator *schema.StructValidator) {
	if !config.Enabled {
		return
	}

	if config.Path == "" {
		validator.Push(fmt.Errorf(errFmtCustomLocalesPathUndefined))
	}

	switch _, err := os.Stat(config.Path); {
	case os.IsNotExist(err):
		validator.Push(fmt.Errorf(errFmtCustomLocalesPathNotExist, config.Path))
		return
	case err != nil:
		validator.Push(fmt.Errorf(errFmtCustomLocalesPathUnknownError, config.Path, err))
		return
	}
}
