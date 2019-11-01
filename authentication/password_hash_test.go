package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldHashPassword(t *testing.T) {
	hash := HashPassword("password", "$6$rounds=50000$aFr56HjK3DrB8t3S")
	assert.Equal(t, "$6$rounds=50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1", hash)
}

func TestShouldCheckPassword(t *testing.T) {
	ok, err := CheckPassword("password", "$6$rounds=50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.NoError(t, err)
	assert.True(t, ok)
}
