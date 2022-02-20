package authentication

import (
	"fmt"
	"testing"

	"github.com/simia-tech/crypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldHashSHA512Password(t *testing.T) {
	hash, err := HashPassword("password", "aFr56HjK3DrB8t3S", HashingAlgorithmSHA512, 50000, 0, 0, 0, 16)

	assert.NoError(t, err)

	code, parameters, salt, hash, _ := crypt.DecodeSettings(hash)

	assert.Equal(t, "6", code)
	assert.Equal(t, "aFr56HjK3DrB8t3S", salt)
	assert.Equal(t, "zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1", hash)
	assert.Equal(t, schema.DefaultPasswordSHA512Configuration.Iterations, parameters.GetInt("rounds", HashingDefaultSHA512Iterations))
}

func TestShouldHashArgon2idPassword(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		schema.DefaultCIPasswordConfiguration.Iterations, schema.DefaultCIPasswordConfiguration.Memory*1024,
		schema.DefaultCIPasswordConfiguration.Parallelism, schema.DefaultCIPasswordConfiguration.KeyLength,
		schema.DefaultCIPasswordConfiguration.SaltLength)

	assert.NoError(t, err)

	code, parameters, salt, key, err := crypt.DecodeSettings(hash)

	assert.NoError(t, err)
	assert.Equal(t, argon2id, code)
	assert.Equal(t, "BpLnfgDsc2WD8F2q", salt)
	assert.Equal(t, "f+Y+KaS12gkNHN0Llc9kqDZuk1OYvoXj8t+5DcPbgY4", key)
	assert.Equal(t, schema.DefaultCIPasswordConfiguration.Iterations, parameters.GetInt("t", HashingDefaultArgon2idTime))
	assert.Equal(t, schema.DefaultCIPasswordConfiguration.Memory*1024, parameters.GetInt("m", HashingDefaultArgon2idMemory))
	assert.Equal(t, schema.DefaultCIPasswordConfiguration.Parallelism, parameters.GetInt("p", HashingDefaultArgon2idParallelism))
	assert.Equal(t, schema.DefaultCIPasswordConfiguration.KeyLength, parameters.GetInt("k", HashingDefaultArgon2idKeyLength))
}

func TestShouldValidateArgon2idHashWithTEqualOne(t *testing.T) {
	hash := "$argon2id$v=19$m=1024,t=1,p=1,k=16$c2FsdG9uY2U$Sk4UjzxXdCrBcyyMYiPEsQ"
	valid, err := CheckPassword("apple", hash)
	assert.True(t, valid)
	assert.NoError(t, err)
}

// This checks the method of hashing (for argon2id) supports all the characters we allow in Authelia's hash function.
func TestArgon2idHashSaltValidValues(t *testing.T) {
	var err error

	var hash string

	datas := utils.SliceString(HashingPossibleSaltCharacters, 16)

	for _, salt := range datas {
		hash, err = HashPassword("password", salt, HashingAlgorithmArgon2id, 1, 8, 1, 32, 16)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("$argon2id$v=19$m=8,t=1,p=1$%s$", salt), hash[0:44])
	}
}

// This checks the method of hashing (for sha512) supports all the characters we allow in Authelia's hash function.
func TestSHA512HashSaltValidValues(t *testing.T) {
	var err error

	var hash string

	datas := utils.SliceString(HashingPossibleSaltCharacters, 16)

	for _, salt := range datas {
		hash, err = HashPassword("password", salt, HashingAlgorithmSHA512, 1000, 0, 0, 0, 16)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("$6$rounds=1000$%s$", salt), hash[0:32])
	}
}

func TestShouldNotHashPasswordWithNonExistentAlgorithm(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", "bogus",
		schema.DefaultCIPasswordConfiguration.Iterations, schema.DefaultCIPasswordConfiguration.Memory*1024,
		schema.DefaultCIPasswordConfiguration.Parallelism, schema.DefaultCIPasswordConfiguration.KeyLength,
		schema.DefaultCIPasswordConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "hashing algorithm input of 'bogus' is invalid, must be one of 'argon2id' and '6'")
}

func TestShouldNotHashArgon2idPasswordDueToMemoryParallelismMismatch(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		schema.DefaultCIPasswordConfiguration.Iterations, 8, 2,
		schema.DefaultCIPasswordConfiguration.KeyLength, schema.DefaultCIPasswordConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "argon2id memory input of '8' is invalid with a parallelism input of '2', it must be 16 (parallelism * 8) or higher")
}

func TestShouldNotHashArgon2idPasswordDueToMemoryLessThanEight(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		schema.DefaultCIPasswordConfiguration.Iterations, 1, schema.DefaultCIPasswordConfiguration.Parallelism,
		schema.DefaultCIPasswordConfiguration.KeyLength, schema.DefaultCIPasswordConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "argon2id memory input of '1' is invalid, it must be 8 or higher")
}

func TestShouldNotHashArgon2idPasswordDueToKeyLengthLessThanSixteen(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		schema.DefaultCIPasswordConfiguration.Iterations, schema.DefaultCIPasswordConfiguration.Memory*1024,
		schema.DefaultCIPasswordConfiguration.Parallelism, 5, schema.DefaultCIPasswordConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "argon2id key length input of '5' is invalid, it must be 16 or higher")
}

func TestShouldNotHashArgon2idPasswordDueParallelismLessThanOne(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		schema.DefaultCIPasswordConfiguration.Iterations, schema.DefaultCIPasswordConfiguration.Memory*1024, -1,
		schema.DefaultCIPasswordConfiguration.KeyLength, schema.DefaultCIPasswordConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "argon2id parallelism input of '-1' is invalid, it must be 1 or higher")
}

func TestShouldNotHashArgon2idPasswordDueIterationsLessThanOne(t *testing.T) {
	hash, err := HashPassword("password", "BpLnfgDsc2WD8F2q", HashingAlgorithmArgon2id,
		0, schema.DefaultCIPasswordConfiguration.Memory*1024, schema.DefaultCIPasswordConfiguration.Parallelism,
		schema.DefaultCIPasswordConfiguration.KeyLength, schema.DefaultCIPasswordConfiguration.SaltLength)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "argon2id iterations input of '0' is invalid, it must be 1 or higher")
}

func TestShouldNotHashPasswordDueToSaltLength(t *testing.T) {
	hash, err := HashPassword("password", "", HashingAlgorithmArgon2id,
		schema.DefaultCIPasswordConfiguration.Iterations, schema.DefaultCIPasswordConfiguration.Memory*1024,
		schema.DefaultCIPasswordConfiguration.Parallelism, schema.DefaultCIPasswordConfiguration.KeyLength, 0)

	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "salt length input of '0' is invalid, it must be 8 or higher")
}

func TestShouldNotHashPasswordDueToSaltCharLengthTooShort(t *testing.T) {
	// The salt 'YQ' is the base64 value for 'a' which is why the length is 1.
	hash, err := HashPassword("password", "YQ", HashingAlgorithmArgon2id,
		schema.DefaultCIPasswordConfiguration.Iterations, schema.DefaultCIPasswordConfiguration.Memory*1024,
		schema.DefaultCIPasswordConfiguration.Parallelism, schema.DefaultCIPasswordConfiguration.KeyLength,
		schema.DefaultCIPasswordConfiguration.SaltLength)
	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "salt input of 'a' is invalid due to it's length being 1, it must be 8 or more characters")
}

func TestShouldNotHashPasswordWithNonBase64CharsInSalt(t *testing.T) {
	hash, err := HashPassword("password", "abc&123", HashingAlgorithmArgon2id,
		schema.DefaultCIPasswordConfiguration.Iterations, schema.DefaultCIPasswordConfiguration.Memory*1024,
		schema.DefaultCIPasswordConfiguration.Parallelism, schema.DefaultCIPasswordConfiguration.KeyLength,
		schema.DefaultCIPasswordConfiguration.SaltLength)
	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "salt input of 'abc&123' is invalid, only base64 strings are valid for input")
}

func TestShouldNotParseHashWithNoneBase64CharsInKey(t *testing.T) {
	passwordHash, err := ParseHash("$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$^^vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")
	assert.EqualError(t, err, "hash key '^^vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM' contains invalid base64 characters")
	assert.Nil(t, passwordHash)
}

func TestShouldNotParseHashWithNoneBase64CharsInSalt(t *testing.T) {
	passwordHash, err := ParseHash("$argon2id$v=19$m=65536$^^wTFoFjITudo57a$Z4NH/EKkdv6PJ01Ye1twJ61fsmRJujZZn1IXdUOyrJY")
	assert.EqualError(t, err, "invalid base64 characters in salt '^^wTFoFjITudo57a'")
	assert.Nil(t, passwordHash)
}

func TestShouldNotParseWithMalformedHash(t *testing.T) {
	hashExtraField := "$argon2id$v=19$m=65536,t=3,p=2$abc$BpLnfgDsc2WD8F2q$^^vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
	hashMissingSaltAndParams := "$argon2id$v=1$2t9X8nNCN2n3/kFYJ3xWNBg5k/rO782Qr7JJoJIK7G4"
	hashMissingSalt := "$argon2id$v=1$m=65536,t=3,p=2$2t9X8nNCN2n3/kFYJ3xWNBg5k/rO782Qr7JJoJIK7G4"

	passwordHash, err := ParseHash(hashExtraField)
	assert.EqualError(t, err, fmt.Sprintf("hash key 'BpLnfgDsc2WD8F2q' is not the last parameter in the hash '%s', the hash is likely malformed", hashExtraField))
	assert.Nil(t, passwordHash)

	passwordHash, err = ParseHash(hashMissingSaltAndParams)
	assert.EqualError(t, err, fmt.Sprintf("hash key '' is not the last parameter in the hash '%s', the hash is likely malformed", hashMissingSaltAndParams))
	assert.Nil(t, passwordHash)

	passwordHash, err = ParseHash(hashMissingSalt)
	assert.EqualError(t, err, fmt.Sprintf("hash key '' is not the last parameter in the hash '%s', the hash is likely malformed", hashMissingSalt))
	assert.Nil(t, passwordHash)
}

func TestShouldNotParseHashWithEmptyKey(t *testing.T) {
	hash := "$argon2id$v=19$m=65536$fvwTFoFjITudo57a$"
	passwordHash, err := ParseHash(hash)
	assert.EqualError(t, err, fmt.Sprintf("hash key contains no characters or the field length is invalid in the hash '%s'", hash))
	assert.Nil(t, passwordHash)
}

func TestShouldNotParseArgon2idHashWithEmptyVersion(t *testing.T) {
	hash := "$argon2id$m=65536$fvwTFoFjITudo57a$Z4NH/EKkdv6PJ01Ye1twJ61fsmRJujZZn1IXdUOyrJY"
	passwordHash, err := ParseHash(hash)
	assert.EqualError(t, err, fmt.Sprintf("argon2id version parameter not found in hash '%s'", hash))
	assert.Nil(t, passwordHash)
}

func TestShouldNotParseArgon2idHashWithWrongKeyLength(t *testing.T) {
	hash := "$argon2id$v=19$m=65536,k=50$fvwTFoFjITudo57a$Z4NH/EKkdv6PJ01Ye1twJ61fsmRJujZZn1IXdUOyrJY"
	passwordHash, err := ParseHash(hash)
	assert.EqualError(t, err, "argon2id key length parameter '50' does not match the actual key length '32'")
	assert.Nil(t, passwordHash)
}

func TestShouldParseArgon2idHash(t *testing.T) {
	passwordHash, err := ParseHash("$argon2id$v=19$m=65536,t=1,p=8$NEwwcVNuQWlQMFpkMndxdg$LlHjiLxPB94pdmOiNwr7Bgy+uy3huSv6y9phCQ+mLls")
	assert.NoError(t, err)
	assert.Equal(t, schema.DefaultCIPasswordConfiguration.Iterations, passwordHash.Iterations)
	assert.Equal(t, schema.DefaultCIPasswordConfiguration.Parallelism, passwordHash.Parallelism)
	assert.Equal(t, schema.DefaultCIPasswordConfiguration.KeyLength, passwordHash.KeyLength)
	assert.Equal(t, schema.DefaultCIPasswordConfiguration.Memory*1024, passwordHash.Memory)
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

	assert.EqualError(t, err, "hash key '' is not the last parameter in the hash '$6$roSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1', the hash is likely malformed")
	assert.False(t, ok)
}

func TestCannotParseArgon2idHash(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "hash key '' is not the last parameter in the hash '$argon2id$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM', the hash is likely malformed")
	assert.False(t, ok)
}

func TestOnlySupportSHA512AndArgon2id(t *testing.T) {
	ok, err := CheckPassword("password", "$8$rounds=50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "invalid hash type prefix '8', must be one of salted sha512 '$6$' or salted argon2id '$argon2id$'")
	assert.False(t, ok)
}

func TestCannotFindNumberOfRounds(t *testing.T) {
	hash := "$6$rounds50000$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1"
	ok, err := CheckPassword("password", hash)

	assert.EqualError(t, err, fmt.Sprintf("hash key 'aFr56HjK3DrB8t3S' is not the last parameter in the hash '%s', the hash is likely malformed", hash))
	assert.False(t, ok)
}

func TestCannotMatchArgon2idParamPattern(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$v=19$m65536,t3,p2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "hash key 'BpLnfgDsc2WD8F2q' is not the last parameter in the hash '$argon2id$v=19$m65536,t3,p2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM', the hash is likely malformed")
	assert.False(t, ok)
}

func TestArgon2idVersionLessThanSupported(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$v=18$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "argon2id version '18' is invalid as versions less than v19 are not supported")
	assert.False(t, ok)
}

func TestArgon2idVersionGreaterThanSupported(t *testing.T) {
	ok, err := CheckPassword("password", "$argon2id$v=20$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM")

	assert.EqualError(t, err, "argon2id version '20' is invalid as versions greater than v19 are not supported")
	assert.False(t, ok)
}

func TestNumberOfRoundsNotInt(t *testing.T) {
	ok, err := CheckPassword("password", "$6$rounds=abc$aFr56HjK3DrB8t3S$zhPQiS85cgBlNhUKKE6n/AHMlpqrvYSnSL3fEVkK0yHFQ.oFFAd8D4OhPAy18K5U61Z2eBhxQXExGU/eknXlY1")

	assert.EqualError(t, err, "sha512 iterations 'abc' is not numeric")
	assert.False(t, ok)
}

func TestShouldCheckPasswordArgon2idHashedWithAuthelia(t *testing.T) {
	password := testPassword
	hash, err := HashPassword(password, "", HashingAlgorithmArgon2id, schema.DefaultCIPasswordConfiguration.Iterations,
		schema.DefaultCIPasswordConfiguration.Memory*1024, schema.DefaultCIPasswordConfiguration.Parallelism,
		schema.DefaultCIPasswordConfiguration.KeyLength, schema.DefaultCIPasswordConfiguration.SaltLength)

	assert.NoError(t, err)

	equal, err := CheckPassword(password, hash)

	require.NoError(t, err)
	assert.True(t, equal)
}

func TestShouldCheckPasswordSHA512HashedWithAuthelia(t *testing.T) {
	password := testPassword
	hash, err := HashPassword(password, "", HashingAlgorithmSHA512, schema.DefaultPasswordSHA512Configuration.Iterations,
		0, 0, 0, schema.DefaultPasswordSHA512Configuration.SaltLength)

	assert.NoError(t, err)

	equal, err := CheckPassword(password, hash)

	require.NoError(t, err)
	assert.True(t, equal)
}
