package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldHashPassword(t *testing.T) {
	salt := "$6$rounds=5000$aFr56HjK3DrB8t3S$"
	hash := HashPassword("password", &salt)
	assert.Equal(t, "$6$rounds=5000$aFr56HjK3DrB8t3S$3yTiN5991WnlmhE8qlMmayIiUiT5ppq68CIuHBrGgQHJ4RWSCb0AykB0E6Ij761ZTzLaCZKuXpurcBiqDR1hu.", hash)
}

func TestShouldCheckPassword(t *testing.T) {
	ok, err := CheckPassword("password", "$6$rounds=5000$aFr56HjK3DrB8t3S$3yTiN5991WnlmhE8qlMmayIiUiT5ppq68CIuHBrGgQHJ4RWSCb0AykB0E6Ij761ZTzLaCZKuXpurcBiqDR1hu.")

	assert.NoError(t, err)
	assert.True(t, ok)
}
