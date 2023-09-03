package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldSetDefaultLoggingValues(t *testing.T) {
	config := &schema.Configuration{}

	validator := schema.NewStructValidator()

	ValidateLog(config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	require.NotNil(t, config.Log.KeepStdout)

	assert.Equal(t, "info", config.Log.Level)
	assert.Equal(t, "text", config.Log.Format)
	assert.Equal(t, "", config.Log.FilePath)
}

func TestShouldRaiseErrorOnInvalidLoggingLevel(t *testing.T) {
	config := &schema.Configuration{
		Log: schema.Log{
			Level: "TRACE",
		},
	}

	validator := schema.NewStructValidator()

	ValidateLog(config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "log: option 'level' must be one of 'trace', 'debug', 'info', 'warn', or 'error' but it's configured as 'TRACE'")
}

func TestShouldRaiseErrorOnInvalidLoggingFormat(t *testing.T) {
	config := &schema.Configuration{
		Log: schema.Log{
			Level:  "trace",
			Format: "FORMAT",
		},
	}

	validator := schema.NewStructValidator()

	ValidateLog(config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "log: option 'format' must be one of 'text' or 'json' but it's configured as 'FORMAT'")
}
