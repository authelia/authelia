package validator

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateLogging validates the logging configuration.
func ValidateLogging(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.Log.Level == "" {
		configuration.Log.Level = schema.DefaultLoggingConfiguration.Level
	}

	if configuration.Log.Format == "" {
		configuration.Log.Format = schema.DefaultLoggingConfiguration.Format
	}

	if !utils.IsStringInSlice(configuration.Log.Level, validLoggingLevels) {
		validator.Push(fmt.Errorf(errFmtLoggingLevelInvalid, configuration.Log.Level, strings.Join(validLoggingLevels, ", ")))
	}
}
