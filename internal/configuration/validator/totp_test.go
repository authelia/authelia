package validator

import (
	"testing"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldSetDefaultTOTPValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := schema.TOTPConfiguration{}

	ValidateTOTP(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "Authelia", config.Issuer)
	assert.Equal(t, defaultTOTPSkew, *config.Skew)
	assert.Equal(t, 30, config.Period)
}

func TestShouldRaiseErrorWhenInvalidTOTPMinimumValues(t *testing.T) {
	var badSkew = -1
	validator := schema.NewStructValidator()
	config := schema.TOTPConfiguration{
		Period: -5,
		Skew:   &badSkew,
	}
	ValidateTOTP(&config, validator)
	assert.Len(t, validator.Errors(), 2)
	assert.EqualError(t, validator.Errors()[0], "TOTP Period must be 1 or more")
	assert.EqualError(t, validator.Errors()[1], "TOTP Skew must be 0 or more")

}
