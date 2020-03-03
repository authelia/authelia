package authentication

import (
	"fmt"
	"testing"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
	"github.com/simia-tech/crypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldHashSHA512Password(t *testing.T) {
	hash, err := HashPassword("password", "aFr56HjK3DrB8t3S", HashingAlgorithmSHA512, 50000, 0, 0, 0, 16)
	assert.NoError(t, err)
	assert.Equal(t, "$6$rounds=50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1", hash)
}

func TestShouldHashArgon2idPassword(t *testing.T) {
	hashString, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		schema.DefaultPasswordOptionsConfiguration.Iterations, schema.DefaultPasswordOptionsConfiguration.Memory,
		schema.DefaultPasswordOptionsConfiguration.Parallelism, schema.DefaultPasswordOptionsConfiguration.KeyLength,
		schema.DefaultPasswordOptionsConfiguration.SaltLength)

	assert.NoError(t, err)

	code, parameters, salt, hash, err := crypt.DecodeSettings(hashString)

	assert.NoError(t, err)
	assert.Equal(t, "argon2id", code)
	assert.Equal(t, "BpLnfgDsc2WD8F2q", salt)
	assert.Equal(t, "2t9X8nNCN2n3/kFYJ3xWNBg5k/rO782Qr7JJoJIK7G4", hash)
	assert.Equal(t, schema.DefaultPasswordOptionsConfiguration.Iterations, parameters.GetInt("t", HashingDefaultArgon2idTime))
	assert.Equal(t, schema.DefaultPasswordOptionsConfiguration.Memory, parameters.GetInt("m", HashingDefaultArgon2idMemory))
	assert.Equal(t, schema.DefaultPasswordOptionsConfiguration.Parallelism, parameters.GetInt("p", HashingDefaultArgon2idParallelism))
	assert.Equal(t, schema.DefaultPasswordOptionsConfiguration.KeyLength, parameters.GetInt("k", HashingDefaultArgon2idKeyLength))
}

// This checks the method of hashing (for argon2id) supports all the characters we allow in Authelia's hash function
func TestArgon2idHashSaltValidValues(t *testing.T) {
	data := string(HashingPossibleSaltCharacters)
	datas := utils.SplitStringToArrayOfStrings(data, 16)
	var hash string
	var err error
	for _, salt := range datas {
		hash, err = HashPassword("password", salt, HashingAlgorithmArgon2id, 1, 8, 1, 32, 16)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("$argon2id$v=19$m=8,p=1$%s$", salt), hash[0:40])
	}
}

// This checks the method of hashing (for sha512) supports all the characters we allow in Authelia's hash function
func TestSHA512HashSaltValidValues(t *testing.T) {
	data := string(HashingPossibleSaltCharacters)
	datas := utils.SplitStringToArrayOfStrings(data, 16)
	var hash string
	var err error
	for _, salt := range datas {
		hash, err = HashPassword("password", salt, HashingAlgorithmSHA512, 1000, 0, 0, 0, 16)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("$6$rounds=1000$%s$", salt), hash[0:32])
	}
}

func TestShouldNotHashPasswordWithNonExistentAlgorithm(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", "bogus",
		schema.DefaultPasswordOptionsConfiguration.Iterations, schema.DefaultPasswordOptionsConfiguration.Memory,
		schema.DefaultPasswordOptionsConfiguration.Parallelism, schema.DefaultPasswordOptionsConfiguration.KeyLength,
		schema.DefaultPasswordOptionsConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "Hashing algorithm input of 'bogus' is invalid, only values of argon2id and 6 are supported.")
}

func TestShouldNotHashArgon2idPasswordDueToMemoryParallelismMismatch(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		schema.DefaultPasswordOptionsConfiguration.Iterations, 8, 2,
		schema.DefaultPasswordOptionsConfiguration.KeyLength, schema.DefaultPasswordOptionsConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "Memory (argon2id) input of 8 is invalid with a paraellelism input of 2, it must be 16 (parallelism * 8) or higher.")
}

func TestShouldNotHashArgon2idPasswordDueToMemoryLessThanEight(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		schema.DefaultPasswordOptionsConfiguration.Iterations, 1, schema.DefaultPasswordOptionsConfiguration.Parallelism,
		schema.DefaultPasswordOptionsConfiguration.KeyLength, schema.DefaultPasswordOptionsConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "Memory (argon2id) input of 1 is invalid, it must be 8 or higher.")
}

func TestShouldNotHashArgon2idPasswordDueToKeyLengthLessThanSixteen(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		schema.DefaultPasswordOptionsConfiguration.Iterations, schema.DefaultPasswordOptionsConfiguration.Memory,
		schema.DefaultPasswordOptionsConfiguration.Parallelism, 5, schema.DefaultPasswordOptionsConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "Key length (argon2id) input of 5 is invalid, it must be 16 or higher.")
}

func TestShouldNotHashArgon2idPasswordDueParallelismLessThanOne(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		schema.DefaultPasswordOptionsConfiguration.Iterations, schema.DefaultPasswordOptionsConfiguration.Memory, -1,
		schema.DefaultPasswordOptionsConfiguration.KeyLength, schema.DefaultPasswordOptionsConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "Parallelism (argon2id) input of -1 is invalid, it must be 1 or higher.")
}

func TestShouldNotHashPasswordDueToSaltLength(t *testing.T) {
	hash, err := HashPassword("password", "", HashingAlgorithmArgon2id,
		schema.DefaultPasswordOptionsConfiguration.Iterations, schema.DefaultPasswordOptionsConfiguration.Memory,
		schema.DefaultPasswordOptionsConfiguration.Parallelism, schema.DefaultPasswordOptionsConfiguration.KeyLength, 0)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "Salt length input of 0 is invalid, it must be 2 or higher.")

	hash, err = HashPassword("password", "", HashingAlgorithmArgon2id,
		schema.DefaultPasswordOptionsConfiguration.Iterations, schema.DefaultPasswordOptionsConfiguration.Memory,
		schema.DefaultPasswordOptionsConfiguration.Parallelism, schema.DefaultPasswordOptionsConfiguration.KeyLength, 20)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "Salt length input of 20 is invalid, it must be 16 or lower.")
}

func TestShouldNotHashPasswordDueToSaltCharLengthTooLong(t *testing.T) {
	hash, err := HashPassword("password", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", HashingAlgorithmArgon2id,
		schema.DefaultPasswordOptionsConfiguration.Iterations, schema.DefaultPasswordOptionsConfiguration.Memory,
		schema.DefaultPasswordOptionsConfiguration.Parallelism, schema.DefaultPasswordOptionsConfiguration.KeyLength,
		schema.DefaultPasswordOptionsConfiguration.SaltLength)
	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "Salt input of abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 is invalid (62 characters), it must be 16 or fewer characters.")
}

func TestShouldNotHashPasswordDueToSaltCharLengthTooShort(t *testing.T) {
	hash, err := HashPassword("password", "a", HashingAlgorithmArgon2id,
		schema.DefaultPasswordOptionsConfiguration.Iterations, schema.DefaultPasswordOptionsConfiguration.Memory,
		schema.DefaultPasswordOptionsConfiguration.Parallelism, schema.DefaultPasswordOptionsConfiguration.KeyLength,
		schema.DefaultPasswordOptionsConfiguration.SaltLength)
	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "Salt input of a is invalid (1 characters), it must be 2 or more characters.")
}

func TestShouldNotHashWithNonBase64CharsInSalt(t *testing.T) {
	hash, err := HashPassword("password", "abc&123", HashingAlgorithmArgon2id,
		schema.DefaultPasswordOptionsConfiguration.Iterations, schema.DefaultPasswordOptionsConfiguration.Memory,
		schema.DefaultPasswordOptionsConfiguration.Parallelism, schema.DefaultPasswordOptionsConfiguration.KeyLength,
		schema.DefaultPasswordOptionsConfiguration.SaltLength)
	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "Salt input of abc&123 is invalid, only characters [a-zA-Z0-9+/] are valid for input.")
}

func TestShouldNotParseWithNoneBase64CharsInHashKey(t *testing.T) {
	_, err := ParseHash("$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$^^vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")
	assert.EqualError(t, err, "Cannot parse hash key contains invalid base64 characters.")
}

func TestShouldNotParseWithNoneBase64CharsInHashSalt(t *testing.T) {
	_, err := ParseHash("$argon2id$v=19$m=65536,t=3,p=2$^^LnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")
	assert.EqualError(t, err, "Cannot parse hash salt contains invalid base64 characters.")
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

	assert.EqualError(t, err, "Cannot parse hash key is not the last parameter, the hash is probably malformed ($6$roSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1).")
	assert.False(t, ok)
}

func TestCannotParseArgon2idHash(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "Cannot parse hash key is not the last parameter, the hash is probably malformed ($argon2id$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM).")
	assert.False(t, ok)
}

func TestOnlySupportSHA512AndArgon2id(t *testing.T) {
	ok, err := CheckPassword("password", "$8$rounds=50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "Authelia only supports salted SHA512 hashing ($6$) and salted argon2id ($argon2id$), not $8$")
	assert.False(t, ok)
}

func TestCannotFindNumberOfRounds(t *testing.T) {
	hash := "$6$rounds50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1"
	ok, err := CheckPassword("password", hash)
	assert.EqualError(t, err, fmt.Sprintf("Cannot parse hash key is not the last parameter, the hash is probably malformed (%s).", hash))
	assert.False(t, ok)
}

func TestCannotMatchArgon2idParamPattern(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$v=19$m65536,t3,p2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "Cannot parse hash key is not the last parameter, the hash is probably malformed ($argon2id$v=19$m65536,t3,p2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM).")
	assert.False(t, ok)
}

func TestArgon2idVersionLessThanSupported(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$v=18$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "Cannot parse hash argon2id versions less than v19 are not supported (hash is version 18).")
	assert.False(t, ok)
}

func TestArgon2idVersionGreaterThanSupported(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$v=20$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")
	assert.EqualError(t, err, "Cannot parse hash argon2id versions greater than v19 are not supported (hash is version 20).")
	assert.False(t, ok)
}

func TestNumberOfRoundsNotInt(t *testing.T) {
	ok, err := CheckPassword("password", "$6$rounds=abc$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")
	assert.EqualError(t, err, "Cannot parse hash sha512 rounds is not numeric (abc).")
	assert.False(t, ok)
}

func TestShouldCheckPasswordArgon2idHashedWithAuthelia(t *testing.T) {
	password := "my;secure*password"
	hash, err := HashPassword(password, "", HashingAlgorithmArgon2id, HashingDefaultArgon2idTime, HashingDefaultArgon2idMemory, HashingDefaultArgon2idParallelism, HashingDefaultArgon2idKeyLength, schema.DefaultPasswordOptionsConfiguration.SaltLength)

	assert.NoError(t, err)

	equal, err := CheckPassword(password, hash)

	require.NoError(t, err)
	assert.True(t, equal)
}

func TestShouldCheckPasswordSHA512HashedWithAuthelia(t *testing.T) {
	password := "my;secure*password"
	hash, err := HashPassword(password, "", HashingAlgorithmSHA512, schema.DefaultPasswordOptionsSHA512Configuration.Iterations, 0, 0, 0, schema.DefaultPasswordOptionsSHA512Configuration.SaltLength)
	assert.NoError(t, err)
	equal, err := CheckPassword(password, hash)

	require.NoError(t, err)
	assert.True(t, equal)
}
