package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldSetDefaultConfig(t *testing.T) {
	validator := schema.NewStructValidator()
	config := schema.ServerConfiguration{}

	ValidateServer(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, defaultReadBufferSize, config.ReadBufferSize)
	assert.Equal(t, defaultWriteBufferSize, config.WriteBufferSize)
}

func TestShouldRaiseOnNegativeValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := schema.ServerConfiguration{
		ReadBufferSize:  -1,
		WriteBufferSize: -1,
	}

	ValidateServer(&config, validator)

	require.Len(t, validator.Errors(), 2)
	assert.EqualError(t, validator.Errors()[0], "server read buffer size must be above 0")
	assert.EqualError(t, validator.Errors()[1], "server write buffer size must be above 0")
}
