package validator

import (
	"crypto/tls"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldRaiseErrorWhenBothBackendsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	backendConfig := schema.AuthenticationBackend{}

	backendConfig.LDAP = &schema.LDAPAuthenticationBackend{}
	backendConfig.File = &schema.FileAuthenticationBackend{
		Path: "/tmp",
	}

	ValidateAuthenticationBackend(&backendConfig, validator)

	require.Len(t, validator.Errors(), 7)
	assert.EqualError(t, validator.Errors()[0], "authentication_backend: please ensure only one of the 'file' or 'ldap' backend is configured")
	assert.EqualError(t, validator.Errors()[1], "authentication_backend: ldap: option 'url' is required")
	assert.EqualError(t, validator.Errors()[2], "authentication_backend: ldap: option 'user' is required")
	assert.EqualError(t, validator.Errors()[3], "authentication_backend: ldap: option 'password' is required")
	assert.EqualError(t, validator.Errors()[4], "authentication_backend: ldap: option 'base_dn' is required")
	assert.EqualError(t, validator.Errors()[5], "authentication_backend: ldap: option 'users_filter' is required")
	assert.EqualError(t, validator.Errors()[6], "authentication_backend: ldap: option 'groups_filter' is required")
}

func TestShouldRaiseErrorWhenNoBackendProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	backendConfig := schema.AuthenticationBackend{}

	ValidateAuthenticationBackend(&backendConfig, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "authentication_backend: you must ensure either the 'file' or 'ldap' authentication backend is configured")
}

type FileBasedAuthenticationBackend struct {
	suite.Suite
	config    schema.AuthenticationBackend
	validator *schema.StructValidator
}

func (suite *FileBasedAuthenticationBackend) SetupTest() {
	password := schema.DefaultPasswordConfig

	suite.validator = schema.NewStructValidator()
	suite.config = schema.AuthenticationBackend{}
	suite.config.File = &schema.FileAuthenticationBackend{Path: "/a/path", Password: password}
}

func (suite *FileBasedAuthenticationBackend) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenNoPathProvided() {
	suite.config.File.Path = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: option 'path' is required")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenBlank() {
	suite.config.File.Password = schema.Password{}

	suite.Assert().Equal(0, suite.config.File.Password.KeyLength)
	suite.Assert().Equal(0, suite.config.File.Password.Iterations)
	suite.Assert().Equal(0, suite.config.File.Password.SaltLength)
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)
	suite.Assert().Equal(0, suite.config.File.Password.Memory)
	suite.Assert().Equal(0, suite.config.File.Password.Parallelism)

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(schema.DefaultPasswordConfig.KeyLength, suite.config.File.Password.KeyLength)
	suite.Assert().Equal(schema.DefaultPasswordConfig.Iterations, suite.config.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordConfig.SaltLength, suite.config.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordConfig.Algorithm, suite.config.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordConfig.Memory, suite.config.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordConfig.Parallelism, suite.config.File.Password.Parallelism)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationSHA512() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)

	suite.config.File.Password = schema.Password{
		Algorithm:  digestSHA512,
		Iterations: 1000000,
		SaltLength: 8,
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(hashSHA2Crypt, suite.config.File.Password.Algorithm)
	suite.Assert().Equal(digestSHA512, suite.config.File.Password.SHA2Crypt.Variant)
	suite.Assert().Equal(1000000, suite.config.File.Password.SHA2Crypt.Iterations)
	suite.Assert().Equal(8, suite.config.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationSHA512ButNotOverride() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)

	suite.config.File.Password = schema.Password{
		Algorithm:  digestSHA512,
		Iterations: 1000000,
		SaltLength: 8,
		SHA2Crypt: schema.SHA2CryptPassword{
			Variant:    digestSHA256,
			Iterations: 50000,
			SaltLength: 12,
		},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(hashSHA2Crypt, suite.config.File.Password.Algorithm)
	suite.Assert().Equal(digestSHA256, suite.config.File.Password.SHA2Crypt.Variant)
	suite.Assert().Equal(50000, suite.config.File.Password.SHA2Crypt.Iterations)
	suite.Assert().Equal(12, suite.config.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationSHA512Alt() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)

	suite.config.File.Password = schema.Password{
		Algorithm:  digestSHA512,
		Iterations: 1000000,
		SaltLength: 64,
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(hashSHA2Crypt, suite.config.File.Password.Algorithm)
	suite.Assert().Equal(digestSHA512, suite.config.File.Password.SHA2Crypt.Variant)
	suite.Assert().Equal(1000000, suite.config.File.Password.SHA2Crypt.Iterations)
	suite.Assert().Equal(16, suite.config.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationArgon2() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)

	suite.config.File.Password = schema.Password{
		Algorithm:   "argon2id",
		Iterations:  4,
		Memory:      1024,
		Parallelism: 4,
		KeyLength:   64,
		SaltLength:  64,
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("argon2", suite.config.File.Password.Algorithm)
	suite.Assert().Equal("argon2id", suite.config.File.Password.Argon2.Variant)
	suite.Assert().Equal(4, suite.config.File.Password.Argon2.Iterations)
	suite.Assert().Equal(1048576, suite.config.File.Password.Argon2.Memory)
	suite.Assert().Equal(4, suite.config.File.Password.Argon2.Parallelism)
	suite.Assert().Equal(64, suite.config.File.Password.Argon2.KeyLength)
	suite.Assert().Equal(64, suite.config.File.Password.Argon2.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationArgon2ButNotOverride() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)

	suite.config.File.Password = schema.Password{
		Algorithm:   "argon2id",
		Iterations:  4,
		Memory:      1024,
		Parallelism: 4,
		KeyLength:   64,
		SaltLength:  64,
		Argon2: schema.Argon2Password{
			Variant:     "argon2d",
			Iterations:  1,
			Memory:      2048,
			Parallelism: 1,
			KeyLength:   32,
			SaltLength:  32,
		},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("argon2", suite.config.File.Password.Algorithm)
	suite.Assert().Equal("argon2d", suite.config.File.Password.Argon2.Variant)
	suite.Assert().Equal(1, suite.config.File.Password.Argon2.Iterations)
	suite.Assert().Equal(2048, suite.config.File.Password.Argon2.Memory)
	suite.Assert().Equal(1, suite.config.File.Password.Argon2.Parallelism)
	suite.Assert().Equal(32, suite.config.File.Password.Argon2.KeyLength)
	suite.Assert().Equal(32, suite.config.File.Password.Argon2.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationWhenOnlySHA512Set() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = digestSHA512

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(hashSHA2Crypt, suite.config.File.Password.Algorithm)
	suite.Assert().Equal(digestSHA512, suite.config.File.Password.SHA2Crypt.Variant)
	suite.Assert().Equal(schema.DefaultPasswordConfig.SHA2Crypt.Iterations, suite.config.File.Password.SHA2Crypt.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordConfig.SHA2Crypt.SaltLength, suite.config.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidArgon2Variant() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = "argon2"
	suite.config.File.Password.Argon2.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'variant' is configured as 'invalid' but must be one of the following values: 'argon2id', 'id', 'argon2i', 'i', 'argon2d', 'd'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidSHA2CryptVariant() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = hashSHA2Crypt
	suite.config.File.Password.SHA2Crypt.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'variant' is configured as 'invalid' but must be one of the following values: 'sha256', 'sha512'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidSHA2CryptSaltLength() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = hashSHA2Crypt
	suite.config.File.Password.SHA2Crypt.SaltLength = 40

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'salt_length' is configured as '40' but must be less than or equal to '16'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidPBKDF2Variant() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = "pbkdf2"
	suite.config.File.Password.PBKDF2.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: pbkdf2: option 'variant' is configured as 'invalid' but must be one of the following values: 'sha1', 'sha224', 'sha256', 'sha384', 'sha512'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidBCryptVariant() {
	suite.config.File.Password = schema.Password{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = "bcrypt"
	suite.config.File.Password.BCrypt.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: bcrypt: option 'variant' is configured as 'invalid' but must be one of the following values: 'standard', 'sha256'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSHA2CryptOptionsTooLow() {
	suite.config.File.Password.SHA2Crypt.Iterations = -1
	suite.config.File.Password.SHA2Crypt.SaltLength = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'iterations' is configured as '-1' but must be greater than or equal to '1000'")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: sha2crypt: option 'salt_length' is configured as '-1' but must be greater than or equal to '1'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSHA2CryptOptionsTooHigh() {
	suite.config.File.Password.SHA2Crypt.Iterations = 999999999999
	suite.config.File.Password.SHA2Crypt.SaltLength = 99

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'iterations' is configured as '999999999999' but must be less than or equal to '999999999'")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: sha2crypt: option 'salt_length' is configured as '99' but must be less than or equal to '16'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenPBKDF2OptionsTooLow() {
	suite.config.File.Password.PBKDF2.Iterations = -1
	suite.config.File.Password.PBKDF2.SaltLength = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: pbkdf2: option 'iterations' is configured as '-1' but must be greater than or equal to '100000'")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: pbkdf2: option 'salt_length' is configured as '-1' but must be greater than or equal to '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenPBKDF2OptionsTooHigh() {
	suite.config.File.Password.PBKDF2.Iterations = 2147483649
	suite.config.File.Password.PBKDF2.SaltLength = 2147483650

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: pbkdf2: option 'iterations' is configured as '2147483649' but must be less than or equal to '2147483647'")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: pbkdf2: option 'salt_length' is configured as '2147483650' but must be less than or equal to '2147483647'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBCryptOptionsTooLow() {
	suite.config.File.Password.BCrypt.Cost = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: bcrypt: option 'cost' is configured as '-1' but must be greater than or equal to '10'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBCryptOptionsTooHigh() {
	suite.config.File.Password.BCrypt.Cost = 900

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: bcrypt: option 'cost' is configured as '900' but must be less than or equal to '31'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSCryptOptionsTooLow() {
	suite.config.File.Password.SCrypt.Iterations = -1
	suite.config.File.Password.SCrypt.BlockSize = -21
	suite.config.File.Password.SCrypt.Parallelism = -11
	suite.config.File.Password.SCrypt.KeyLength = -77
	suite.config.File.Password.SCrypt.SaltLength = 7

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 5)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: scrypt: option 'iterations' is configured as '-1' but must be greater than or equal to '1'")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: scrypt: option 'block_size' is configured as '-21' but must be greater than or equal to '1'")
	suite.Assert().EqualError(suite.validator.Errors()[2], "authentication_backend: file: password: scrypt: option 'parallelism' is configured as '-11' but must be greater than or equal to '1'")
	suite.Assert().EqualError(suite.validator.Errors()[3], "authentication_backend: file: password: scrypt: option 'key_length' is configured as '-77' but must be greater than or equal to '1'")
	suite.Assert().EqualError(suite.validator.Errors()[4], "authentication_backend: file: password: scrypt: option 'salt_length' is configured as '7' but must be greater than or equal to '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSCryptOptionsTooHigh() {
	suite.config.File.Password.SCrypt.Iterations = 59
	suite.config.File.Password.SCrypt.BlockSize = 360287970189639672
	suite.config.File.Password.SCrypt.Parallelism = 1073741825
	suite.config.File.Password.SCrypt.KeyLength = 1374389534409
	suite.config.File.Password.SCrypt.SaltLength = 2147483647

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 5)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: scrypt: option 'iterations' is configured as '59' but must be less than or equal to '58'")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: scrypt: option 'block_size' is configured as '360287970189639672' but must be less than or equal to '36028797018963967'")
	suite.Assert().EqualError(suite.validator.Errors()[2], "authentication_backend: file: password: scrypt: option 'parallelism' is configured as '1073741825' but must be less than or equal to '1073741823'")
	suite.Assert().EqualError(suite.validator.Errors()[3], "authentication_backend: file: password: scrypt: option 'key_length' is configured as '1374389534409' but must be less than or equal to '137438953440'")
	suite.Assert().EqualError(suite.validator.Errors()[4], "authentication_backend: file: password: scrypt: option 'salt_length' is configured as '2147483647' but must be less than or equal to '1024'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenArgon2OptionsTooLow() {
	suite.config.File.Password.Argon2.Iterations = -1
	suite.config.File.Password.Argon2.Memory = -1
	suite.config.File.Password.Argon2.Parallelism = -1
	suite.config.File.Password.Argon2.KeyLength = 1
	suite.config.File.Password.Argon2.SaltLength = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 5)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'iterations' is configured as '-1' but must be greater than or equal to '1'")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: argon2: option 'parallelism' is configured as '-1' but must be greater than or equal to '1'")
	suite.Assert().EqualError(suite.validator.Errors()[2], "authentication_backend: file: password: argon2: option 'memory' is configured as '-1' but must be greater than or equal to '8'")
	suite.Assert().EqualError(suite.validator.Errors()[3], "authentication_backend: file: password: argon2: option 'key_length' is configured as '1' but must be greater than or equal to '4'")
	suite.Assert().EqualError(suite.validator.Errors()[4], "authentication_backend: file: password: argon2: option 'salt_length' is configured as '-1' but must be greater than or equal to '1'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenArgon2OptionsTooHigh() {
	suite.config.File.Password.Argon2.Iterations = 9999999999
	suite.config.File.Password.Argon2.Memory = 4294967296
	suite.config.File.Password.Argon2.Parallelism = 16777216
	suite.config.File.Password.Argon2.KeyLength = 9999999998
	suite.config.File.Password.Argon2.SaltLength = 9999999997

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 5)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'iterations' is configured as '9999999999' but must be less than or equal to '2147483647'")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: argon2: option 'parallelism' is configured as '16777216' but must be less than or equal to '16777215'")
	suite.Assert().EqualError(suite.validator.Errors()[2], "authentication_backend: file: password: argon2: option 'memory' is configured as '4294967296' but must be less than or equal to '4294967295'")
	suite.Assert().EqualError(suite.validator.Errors()[3], "authentication_backend: file: password: argon2: option 'key_length' is configured as '9999999998' but must be less than or equal to '2147483647'")
	suite.Assert().EqualError(suite.validator.Errors()[4], "authentication_backend: file: password: argon2: option 'salt_length' is configured as '9999999997' but must be less than or equal to '2147483647'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenArgon2MemoryTooLow() {
	suite.config.File.Password.Argon2.Memory = 4
	suite.config.File.Password.Argon2.Parallelism = 4

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'memory' is configured as '4' but must be greater than or equal to '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenArgon2MemoryTooLowMultiplier() {
	suite.config.File.Password.Argon2.Memory = 8
	suite.config.File.Password.Argon2.Parallelism = 4

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'memory' is configured as '8' but must be greater than or equal to '32' or '4' (the value of 'parallelism) multiplied by '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBadAlgorithmDefined() {
	suite.config.File.Password.Algorithm = "bogus"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: option 'algorithm' is configured as 'bogus' but must be one of the following values: 'sha2crypt', 'pbkdf2', 'scrypt', 'bcrypt', 'argon2'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultValues() {
	suite.config.File.Password.Algorithm = ""
	suite.config.File.Password.Iterations = 0
	suite.config.File.Password.SaltLength = 0
	suite.config.File.Password.Memory = 0
	suite.config.File.Password.Parallelism = 0

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(schema.DefaultPasswordConfig.Algorithm, suite.config.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordConfig.Iterations, suite.config.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordConfig.SaltLength, suite.config.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordConfig.Memory, suite.config.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordConfig.Parallelism, suite.config.File.Password.Parallelism)
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenResetURLIsInvalid() {
	suite.config.PasswordReset.CustomURL = url.URL{Scheme: "ldap", Host: "google.com"}
	suite.config.PasswordReset.Disable = true

	suite.Assert().True(suite.config.PasswordReset.Disable)

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: password_reset: option 'custom_url' is configured to 'ldap://google.com' which has the scheme 'ldap' but the scheme must be either 'http' or 'https'")

	suite.Assert().True(suite.config.PasswordReset.Disable)
}

func (suite *FileBasedAuthenticationBackend) TestShouldNotRaiseErrorWhenResetURLIsValid() {
	suite.config.PasswordReset.CustomURL = url.URL{Scheme: "https", Host: "google.com"}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *FileBasedAuthenticationBackend) TestShouldConfigureDisableResetPasswordWhenCustomURL() {
	suite.config.PasswordReset.CustomURL = url.URL{Scheme: "https", Host: "google.com"}
	suite.config.PasswordReset.Disable = true

	suite.Assert().True(suite.config.PasswordReset.Disable)

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().False(suite.config.PasswordReset.Disable)
}

func TestFileBasedAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(FileBasedAuthenticationBackend))
}

type LDAPAuthenticationBackendSuite struct {
	suite.Suite
	config    schema.AuthenticationBackend
	validator *schema.StructValidator
}

func (suite *LDAPAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.AuthenticationBackend{}
	suite.config.LDAP = &schema.LDAPAuthenticationBackend{}
	suite.config.LDAP.Implementation = schema.LDAPImplementationCustom
	suite.config.LDAP.URL = testLDAPURL
	suite.config.LDAP.User = testLDAPUser
	suite.config.LDAP.Password = testLDAPPassword
	suite.config.LDAP.BaseDN = testLDAPBaseDN
	suite.config.LDAP.UsernameAttribute = "uid"
	suite.config.LDAP.UsersFilter = "({username_attribute}={input})"
	suite.config.LDAP.GroupsFilter = "(cn={input})"
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateDefaultImplementationAndUsernameAttribute() {
	suite.config.LDAP.Implementation = ""
	suite.config.LDAP.UsernameAttribute = ""
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Equal(schema.LDAPImplementationCustom, suite.config.LDAP.Implementation)

	suite.Assert().Equal(suite.config.LDAP.UsernameAttribute, schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.UsernameAttribute)
	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenImplementationIsInvalidMSAD() {
	suite.config.LDAP.Implementation = "masd"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'implementation' is configured as 'masd' but must be one of the following values: 'custom', 'activedirectory'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenURLNotProvided() {
	suite.config.LDAP.URL = ""
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'url' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenUserNotProvided() {
	suite.config.LDAP.User = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'user' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenPasswordNotProvided() {
	suite.config.LDAP.Password = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'password' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseErrorWhenPasswordNotProvidedWithPermitUnauthenticatedBind() {
	suite.config.LDAP.Password = ""
	suite.config.LDAP.PermitUnauthenticatedBind = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'permit_unauthenticated_bind' can't be enabled when password reset is enabled")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenPasswordProvidedWithPermitUnauthenticatedBind() {
	suite.config.LDAP.Password = "test"
	suite.config.LDAP.PermitUnauthenticatedBind = true
	suite.config.PasswordReset.Disable = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'permit_unauthenticated_bind' can't be enabled when a password is specified")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseErrorWhenPermitUnauthenticatedBindConfiguredCorrectly() {
	suite.config.LDAP.Password = ""
	suite.config.LDAP.PermitUnauthenticatedBind = true
	suite.config.PasswordReset.Disable = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenBaseDNNotProvided() {
	suite.config.LDAP.BaseDN = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'base_dn' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnEmptyGroupsFilter() {
	suite.config.LDAP.GroupsFilter = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'groups_filter' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnEmptyUsersFilter() {
	suite.config.LDAP.UsersFilter = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseOnEmptyUsernameAttribute() {
	suite.config.LDAP.UsernameAttribute = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnBadRefreshInterval() {
	suite.config.RefreshInterval = "blah"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: option 'refresh_interval' is configured to 'blah' but it must be either a duration notation or one of 'disable', or 'always': could not parse 'blah' as a duration")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultImplementation() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(schema.LDAPImplementationCustom, suite.config.LDAP.Implementation)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorOnBadFilterPlaceholders() {
	suite.config.LDAP.UsersFilter = "(&({username_attribute}={0})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.config.LDAP.GroupsFilter = "(&({username_attribute}={1})(member={0})(objectClass=group)(objectCategory=group))"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().True(suite.validator.HasErrors())

	suite.Require().Len(suite.validator.Errors(), 4)
	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' has an invalid placeholder: '{0}' has been removed, please use '{input}' instead")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication_backend: ldap: option 'groups_filter' has an invalid placeholder: '{0}' has been removed, please use '{input}' instead")
	suite.Assert().EqualError(suite.validator.Errors()[2], "authentication_backend: ldap: option 'groups_filter' has an invalid placeholder: '{1}' has been removed, please use '{username}' instead")
	suite.Assert().EqualError(suite.validator.Errors()[3], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{input}' but it is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultGroupNameAttribute() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("cn", suite.config.LDAP.GroupNameAttribute)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultMailAttribute() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("mail", suite.config.LDAP.MailAttribute)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultDisplayNameAttribute() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("displayName", suite.config.LDAP.DisplayNameAttribute)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultRefreshInterval() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("5m", suite.config.RefreshInterval)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainEnclosingParenthesis() {
	suite.config.LDAP.UsersFilter = "{username_attribute}={input}"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain enclosing parenthesis: '{username_attribute}={input}' should probably be '({username_attribute}={input})'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenGroupsFilterDoesNotContainEnclosingParenthesis() {
	suite.config.LDAP.GroupsFilter = "cn={input}"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'groups_filter' must contain enclosing parenthesis: 'cn={input}' should probably be '(cn={input})'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainUsernameAttribute() {
	suite.config.LDAP.UsersFilter = "(&({mail_attribute}={input})(objectClass=person))"
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{username_attribute}' but it is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldHelpDetectNoInputPlaceholder() {
	suite.config.LDAP.UsersFilter = "(&({username_attribute}={mail_attribute})(objectClass=person))"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{input}' but it is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultTLSMinimumVersion() {
	suite.config.LDAP.TLS = &schema.TLSConfig{MinimumVersion: schema.TLSVersion{}}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.TLS.MinimumVersion.Value, suite.config.LDAP.TLS.MinimumVersion.MinVersion())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotAllowSSL30() {
	suite.config.LDAP.TLS = &schema.TLSConfig{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotAllowTLSVerMinGreaterThanVerMax() {
	suite.config.LDAP.TLS = &schema.TLSConfig{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
		MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS1.3 is greater than the maximum version TLS1.2")
}

func TestLdapAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(LDAPAuthenticationBackendSuite))
}

type ActiveDirectoryAuthenticationBackendSuite struct {
	suite.Suite
	config    schema.AuthenticationBackend
	validator *schema.StructValidator
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.AuthenticationBackend{}
	suite.config.LDAP = &schema.LDAPAuthenticationBackend{}
	suite.config.LDAP.Implementation = schema.LDAPImplementationActiveDirectory
	suite.config.LDAP.URL = testLDAPURL
	suite.config.LDAP.User = testLDAPUser
	suite.config.LDAP.Password = testLDAPPassword
	suite.config.LDAP.BaseDN = testLDAPBaseDN
	suite.config.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.TLS
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldSetActiveDirectoryDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Timeout,
		suite.config.LDAP.Timeout)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.UsersFilter,
		suite.config.LDAP.UsersFilter)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.UsernameAttribute,
		suite.config.LDAP.UsernameAttribute)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.DisplayNameAttribute,
		suite.config.LDAP.DisplayNameAttribute)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.MailAttribute,
		suite.config.LDAP.MailAttribute)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.GroupsFilter,
		suite.config.LDAP.GroupsFilter)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.GroupNameAttribute,
		suite.config.LDAP.GroupNameAttribute)
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.LDAP.Timeout = time.Second * 2
	suite.config.LDAP.UsersFilter = "(&({username_attribute}={input})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.config.LDAP.UsernameAttribute = "cn"
	suite.config.LDAP.MailAttribute = "userPrincipalName"
	suite.config.LDAP.DisplayNameAttribute = "name"
	suite.config.LDAP.GroupsFilter = "(&(member={dn})(objectClass=group)(objectCategory=group))"
	suite.config.LDAP.GroupNameAttribute = "distinguishedName"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Timeout,
		suite.config.LDAP.Timeout)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.UsersFilter,
		suite.config.LDAP.UsersFilter)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.UsernameAttribute,
		suite.config.LDAP.UsernameAttribute)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.DisplayNameAttribute,
		suite.config.LDAP.DisplayNameAttribute)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.MailAttribute,
		suite.config.LDAP.MailAttribute)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.GroupsFilter,
		suite.config.LDAP.GroupsFilter)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.GroupNameAttribute,
		suite.config.LDAP.GroupNameAttribute)
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldRaiseErrorOnInvalidURLWithHTTP() {
	suite.config.LDAP.URL = "http://dc1:389"

	validateLDAPAuthenticationBackendURL(suite.config.LDAP, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'url' must have either the 'ldap' or 'ldaps' scheme but it is configured as 'http'")
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldRaiseErrorOnInvalidURLWithBadCharacters() {
	suite.config.LDAP.URL = "ldap://dc1:abc"

	validateLDAPAuthenticationBackendURL(suite.config.LDAP, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'url' could not be parsed: parse \"ldap://dc1:abc\": invalid port \":abc\" after host")
}

func TestActiveDirectoryAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(ActiveDirectoryAuthenticationBackendSuite))
}
