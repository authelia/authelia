package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldSetDefaultTOTPValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := schema.TOTPConfiguration{}

	ValidateTOTP(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "Authelia", config.Issuer)
	assert.Equal(t, *schema.DefaultTOTPConfiguration.Skew, *config.Skew)
	assert.Equal(t, schema.DefaultTOTPConfiguration.Period, config.Period)
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

func TestShouldNotRaiseErrorOnValidAlgorithmsMD5(t *testing.T) {
	validator := schema.NewStructValidator()

	config := schema.TOTPConfiguration{
		Algorithm: schema.MD5,
	}
	ValidateTOTP(&config, validator)
	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, config.Algorithm, schema.MD5)
}

func TestShouldNotRaiseErrorOnValidAlgorithmsSHA1(t *testing.T) {
	validator := schema.NewStructValidator()

	config := schema.TOTPConfiguration{
		Algorithm: schema.SHA1,
	}
	ValidateTOTP(&config, validator)
	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, config.Algorithm, schema.SHA1)
}

func TestShouldNotRaiseErrorOnValidAlgorithmsSHA256(t *testing.T) {
	validator := schema.NewStructValidator()

	config := schema.TOTPConfiguration{
		Algorithm: schema.SHA256,
	}
	ValidateTOTP(&config, validator)
	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, config.Algorithm, schema.SHA256)
}

func TestShouldNotRaiseErrorOnValidAlgorithmsSHA512(t *testing.T) {
	validator := schema.NewStructValidator()

	config := schema.TOTPConfiguration{
		Algorithm: schema.SHA512,
	}
	ValidateTOTP(&config, validator)
	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, config.Algorithm, schema.SHA512)
}

func TestShouldRaiseErrorOnInvalidAlgorithms(t *testing.T) {
	validator := schema.NewStructValidator()

	config := schema.TOTPConfiguration{
		Algorithm: "aes",
	}
	ValidateTOTP(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "TOTP Algorithm must be one of md5, sha1, sha256, or sha512")
}

func TestShouldCorrectTOTPAlgorithmCapitalization(t *testing.T) {
	validator := schema.NewStructValidator()

	config := schema.TOTPConfiguration{
		Algorithm: "SHA1",
	}
	assert.Equal(t, config.Algorithm, "SHA1")
	ValidateTOTP(&config, validator)
	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, config.Algorithm, schema.SHA1)
}
