package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

func TestShouldValidateGoodKeys(t *testing.T) {
	configKeys := validKeys
	val := schema.NewStructValidator()
	ValidateKeys(val, configKeys)

	require.Len(t, val.Errors(), 0)
}

func TestShouldNotValidateBadKeys(t *testing.T) {
	configKeys := validKeys
	configKeys = append(configKeys, "bad_key")
	configKeys = append(configKeys, "totp.skewy")
	val := schema.NewStructValidator()
	ValidateKeys(val, configKeys)

	errs := val.Errors()
	require.Len(t, errs, 2)

	assert.EqualError(t, errs[0], "config key not expected: bad_key")
	assert.EqualError(t, errs[1], "config key not expected: totp.skewy")
}

func TestAllSpecificErrorKeys(t *testing.T) {
	var configKeys []string //nolint:prealloc // This is because the test is dynamic based on the keys that exist in the map.

	var uniqueValues []string

	// Setup configKeys and uniqueValues expected.
	for key, value := range specificErrorKeys {
		configKeys = append(configKeys, key)

		if !utils.IsStringInSlice(value, uniqueValues) {
			uniqueValues = append(uniqueValues, value)
		}
	}

	val := schema.NewStructValidator()
	ValidateKeys(val, configKeys)

	errs := val.Errors()

	// Check only unique errors are shown. Require because if we don't the next test panics.
	require.Len(t, errs, len(uniqueValues))

	// Dynamically check all specific errors.
	for i, value := range uniqueValues {
		assert.EqualError(t, errs[i], value)
	}
}

func TestSpecificErrorKeys(t *testing.T) {
	configKeys := []string{
		"logs_level",
		"logs_file_path",
		"authentication_backend.file.password_options.algorithm",
		"authentication_backend.file.password_options.iterations", // This should not show another error since our target for the specific error is password_options.
		"authentication_backend.file.password_hashing.algorithm",
		"authentication_backend.file.hashing.algorithm",
	}

	val := schema.NewStructValidator()
	ValidateKeys(val, configKeys)

	errs := val.Errors()

	require.Len(t, errs, 5)

	assert.EqualError(t, errs[0], specificErrorKeys["logs_level"])
	assert.EqualError(t, errs[1], specificErrorKeys["logs_file_path"])
	assert.EqualError(t, errs[2], specificErrorKeys["authentication_backend.file.password_options.iterations"])
	assert.EqualError(t, errs[3], specificErrorKeys["authentication_backend.file.password_hashing.algorithm"])
	assert.EqualError(t, errs[4], specificErrorKeys["authentication_backend.file.hashing.algorithm"])
}
