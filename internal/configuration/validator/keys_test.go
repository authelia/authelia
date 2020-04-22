package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
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
