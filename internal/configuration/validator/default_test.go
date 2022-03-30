package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidateDefault(t *testing.T) {
	validator := schema.NewStructValidator()

	ValidateDefault(schema.DefaultConfiguration{}, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	ValidateDefault(schema.DefaultConfiguration{UserSecondFactorMethod: "totp"}, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	ValidateDefault(schema.DefaultConfiguration{UserSecondFactorMethod: "webauthn"}, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	ValidateDefault(schema.DefaultConfiguration{UserSecondFactorMethod: "mobile_push"}, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)
}

func TestValidateDefault_ShouldRaiseError(t *testing.T) {
	validator := schema.NewStructValidator()

	config := schema.DefaultConfiguration{
		UserSecondFactorMethod: "example",
	}

	ValidateDefault(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.Len(t, validator.Warnings(), 0)
	assert.EqualError(t, validator.Errors()[0], "default: option 'user_second_factor_method' is configured as 'example' but must be one of the following values: 'totp', 'webauthn', 'mobile_push'")
}
