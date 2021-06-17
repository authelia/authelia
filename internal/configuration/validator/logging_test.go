package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldSetDefaultLoggingValues(t *testing.T) {
	config := &schema.Configuration{}

	validator := schema.NewStructValidator()

	ValidateLogging(config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	require.NotNil(t, config.Logging.KeepStdout)

	assert.Equal(t, "", config.LogLevel)
	assert.Equal(t, "", config.LogFormat)
	assert.Equal(t, "", config.LogFilePath)

	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "text", config.Logging.Format)
	assert.Equal(t, "", config.Logging.FilePath)
}

func TestShouldRaiseErrorOnInvalidLoggingLevel(t *testing.T) {
	config := &schema.Configuration{
		Logging: schema.LogConfiguration{
			Level: "TRACE",
		},
	}

	validator := schema.NewStructValidator()

	ValidateLogging(config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "the log level 'TRACE' is invalid, must be one of: trace, debug, info, warn, error")
}

// TODO: DEPRECATED TEST. Remove in 4.33.0.
func TestShouldMigrateDeprecatedLoggingConfig(t *testing.T) {
	config := &schema.Configuration{
		LogLevel:    "trace",
		LogFormat:   "json",
		LogFilePath: "/a/b/c",
	}

	validator := schema.NewStructValidator()

	ValidateLogging(config, validator)

	assert.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 3)

	require.NotNil(t, config.Logging.KeepStdout)

	assert.Equal(t, "trace", config.LogLevel)
	assert.Equal(t, "json", config.LogFormat)
	assert.Equal(t, "/a/b/c", config.LogFilePath)

	assert.Equal(t, "trace", config.Logging.Level)
	assert.Equal(t, "json", config.Logging.Format)
	assert.Equal(t, "/a/b/c", config.Logging.FilePath)

	assert.EqualError(t, validator.Warnings()[0], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "log_level", "4.33.0", "log.level"))
	assert.EqualError(t, validator.Warnings()[1], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "log_format", "4.33.0", "log.format"))
	assert.EqualError(t, validator.Warnings()[2], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "log_file_path", "4.33.0", "log.file_path"))
}

func TestShouldRaiseErrorsAndNotOverwriteConfigurationWhenUsingDeprecatedLoggingConfig(t *testing.T) {
	config := &schema.Configuration{
		Logging: schema.LogConfiguration{
			Level:      "info",
			Format:     "text",
			FilePath:   "/x/y/z",
			KeepStdout: true,
		},
		LogLevel:    "debug",
		LogFormat:   "json",
		LogFilePath: "/a/b/c",
	}

	validator := schema.NewStructValidator()

	ValidateLogging(config, validator)

	require.NotNil(t, config.Logging.KeepStdout)

	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "text", config.Logging.Format)
	assert.True(t, config.Logging.KeepStdout)
	assert.Equal(t, "/x/y/z", config.Logging.FilePath)

	assert.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 3)

	assert.EqualError(t, validator.Warnings()[0], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "log_level", "4.33.0", "log.level"))
	assert.EqualError(t, validator.Warnings()[1], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "log_format", "4.33.0", "log.format"))
	assert.EqualError(t, validator.Warnings()[2], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "log_file_path", "4.33.0", "log.file_path"))
}
