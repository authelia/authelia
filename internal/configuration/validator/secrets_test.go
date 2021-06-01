package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldValidateCorrectSecretKeys(t *testing.T) {
	assert.True(t, IsSecretKey("jwt_secret"))
	assert.True(t, IsSecretKey("authelia.jwt_secret.file"))
	assert.False(t, IsSecretKey("totp.issuer"))
}

func TestShouldCreateCorrectSecretEnvNames(t *testing.T) {
	assert.Equal(t, "authelia.jwt_secret.file", SecretNameToEnvName("jwt_secret"))
	assert.Equal(t, "authelia.not_a_real_secret.file", SecretNameToEnvName("not_a_real_secret"))
}
