package validator

import (
	"testing"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldSetDefaultIssuer(t *testing.T) {
	validator := schema.NewStructValidator()
	config := schema.TOTPConfiguration{}

	ValidateTOTP(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "Authelia", config.Issuer)
}
