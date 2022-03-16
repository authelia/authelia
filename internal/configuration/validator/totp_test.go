package validator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldSetDefaultTOTPValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		TOTP: schema.TOTPConfiguration{},
	}

	ValidateTOTP(config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "Authelia", config.TOTP.Issuer)
	assert.Equal(t, schema.DefaultTOTPConfiguration.Algorithm, config.TOTP.Algorithm)
	assert.Equal(t, schema.DefaultTOTPConfiguration.Skew, config.TOTP.Skew)
	assert.Equal(t, schema.DefaultTOTPConfiguration.Period, config.TOTP.Period)
}

func TestShouldNormalizeTOTPAlgorithm(t *testing.T) {
	validator := schema.NewStructValidator()

	config := &schema.Configuration{
		TOTP: schema.TOTPConfiguration{
			Algorithm: "sha1",
		},
	}

	ValidateTOTP(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, "SHA1", config.TOTP.Algorithm)
}

func TestShouldRaiseErrorWhenInvalidTOTPAlgorithm(t *testing.T) {
	validator := schema.NewStructValidator()

	config := &schema.Configuration{
		TOTP: schema.TOTPConfiguration{
			Algorithm: "sha3",
		},
	}

	ValidateTOTP(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtTOTPInvalidAlgorithm, strings.Join(schema.TOTPPossibleAlgorithms, "', '"), "SHA3"))
}

func TestShouldRaiseErrorWhenInvalidTOTPValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		TOTP: schema.TOTPConfiguration{
			Period: 5,
			Digits: 20,
		},
	}

	ValidateTOTP(config, validator)

	require.Len(t, validator.Errors(), 2)
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtTOTPInvalidPeriod, 5))
	assert.EqualError(t, validator.Errors()[1], fmt.Sprintf(errFmtTOTPInvalidDigits, 20))
}
