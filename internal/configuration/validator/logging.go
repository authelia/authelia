package validator

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateLogging validates the logging configuration.
func ValidateLogging(configuration *schema.Configuration, validator *schema.StructValidator) {
	applyDeprecatedLoggingConfiguration(configuration, validator)

	if configuration.Logging.Level == "" {
		configuration.Logging.Level = schema.DefaultLoggingConfiguration.Level
	}

	if configuration.Logging.Format == "" {
		configuration.Logging.Format = schema.DefaultLoggingConfiguration.Format
	}

	if configuration.Logging.FilePath == "" && configuration.Logging.KeepStdout {
		configuration.Logging.KeepStdout = false
	}

	if !utils.IsStringInSlice(configuration.Logging.Level, validLoggingLevels) {
		validator.Push(fmt.Errorf(errFmtLoggingLevelInvalid, configuration.Logging.Level, strings.Join(validLoggingLevels, ", ")))
	}
}

func applyDeprecatedLoggingConfiguration(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.LogLevel != "" {
		validator.PushWarning(fmt.Errorf(errFmtDeprecatedConfigurationKey, "log_level", "4.33.0", "logging.level"))

		if configuration.Logging.Level == "" {
			configuration.Logging.Level = configuration.LogLevel
		}
	}

	if configuration.LogFormat != "" {
		validator.PushWarning(fmt.Errorf(errFmtDeprecatedConfigurationKey, "log_format", "4.33.0", "logging.format"))

		if configuration.Logging.Format == "" {
			configuration.Logging.Format = configuration.LogFormat
		}
	}

	if configuration.LogFilePath != "" {
		validator.PushWarning(fmt.Errorf(errFmtDeprecatedConfigurationKey, "log_file_path", "4.33.0", "logging.file_path"))

		if configuration.Logging.FilePath == "" {
			configuration.Logging.FilePath = configuration.LogFilePath
		}
	}
}
