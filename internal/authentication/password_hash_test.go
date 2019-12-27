package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCannotParseHash(t *testing.T) {
	ok, err := CheckPassword("password", "$6$roSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "Cannot parse the hash $6$roSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")
	assert.False(t, ok)
}

func TestOnlySupportSHA512(t *testing.T) {
	ok, err := CheckPassword("password", "$8$rounds=50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "Authelia only supports salted SHA512 hashing ($6$), not $8$")
	assert.False(t, ok)
}

func TestCannotFindNumberOfRounds(t *testing.T) {
	ok, err := CheckPassword("password", "$6$rounds50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "Cannot match pattern 'rounds=<int>' to find the number of rounds")
	assert.False(t, ok)
}

func TestNumberOfRoundsNotInt(t *testing.T) {
	ok, err := CheckPassword("password", "$6$rounds=abc$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "Cannot find the number of rounds from abc using pattern 'rounds=<int>'. Cause: strconv.ParseInt: parsing \"abc\": invalid syntax")
	assert.False(t, ok)
}

func TestShouldCheckPasswordHashedWithAuthelia(t *testing.T) {
	password := "my;secure*password"
	hash := HashPassword(password, "")
	equal, err := CheckPassword(password, hash)

	require.NoError(t, err)
	assert.True(t, equal)
}
