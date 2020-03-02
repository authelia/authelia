package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldHashSHA512Password(t *testing.T) {
	hash := HashPassword("password", "$6$rounds=50000$aFr56HjK3DrB8t3S")
	assert.Equal(t, "$6$rounds=50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1", hash)
}

func TestShouldHashArgon2idPassword(t *testing.T) {
	hash := HashPassword("password", "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q")
	assert.Equal(t, "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM", hash)
}

func TestShouldCheckSHA512Password(t *testing.T) {
	ok, err := CheckPassword("password", "$6$rounds=50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestShouldCheckArgon2idPassword(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestCannotParseSHA512Hash(t *testing.T) {
	ok, err := CheckPassword("password", "$6$roSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "Cannot parse the SHA512 hash $6$roSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")
	assert.False(t, ok)
}

func TestCannotParseArgon2idHash(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "Cannot parse the Argon2id hash $argon2id$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")
	assert.False(t, ok)
}

func TestOnlySupportSHA512AndArgon2id(t *testing.T) {
	ok, err := CheckPassword("password", "$8$rounds=50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "Authelia only supports salted SHA512 hashing ($6$) and salted argon2id ($argon2id$), not $8$")
	assert.False(t, ok)
}

func TestCannotFindNumberOfRounds(t *testing.T) {
	ok, err := CheckPassword("password", "$6$rounds50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "Cannot match pattern 'rounds=<int>' to find the number of rounds. Cause: input does not match format")
	assert.False(t, ok)
}

func TestCannotMatchArgon2idParamPattern(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$v=19$m65536,t3,p2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "Cannot match pattern 'm=<int>,t=<int>,p=<int>' to find the argon2id params. Cause: input does not match format")
	assert.False(t, ok)
}

func TestArgon2idVersionLessThanSupported(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$v=18$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "Argon2 versions less than v19 are not supported (hash is version 18)")
	assert.False(t, ok)
}

func TestArgon2idVersionGreaterThanSupported(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$v=20$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "Argon2 versions greater than v19 are not supported (hash is version 20)")
	assert.False(t, ok)
}

func TestNumberOfRoundsNotInt(t *testing.T) {
	ok, err := CheckPassword("password", "$6$rounds=abc$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "Cannot match pattern 'rounds=<int>' to find the number of rounds. Cause: expected integer")
	assert.False(t, ok)
}

func TestShouldCheckPasswordHashedWithAuthelia(t *testing.T) {
	password := "my;secure*password"
	hash := HashPassword(password, "")
	equal, err := CheckPassword(password, hash)

	require.NoError(t, err)
	assert.True(t, equal)
}
