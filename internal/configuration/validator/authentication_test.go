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

	backendConfig.LDAP = &schema.AuthenticationBackendLDAP{}
	backendConfig.File = &schema.AuthenticationBackendFile{
		Path: "/tmp",
	}

	ValidateAuthenticationBackend(&backendConfig, validator)

	require.Len(t, validator.Errors(), 7)
	assert.EqualError(t, validator.Errors()[0], "authentication_backend: please ensure only one of the 'file' or 'ldap' backend is configured")
	assert.EqualError(t, validator.Errors()[1], "authentication_backend: ldap: option 'address' is required")
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
	suite.config.File = &schema.AuthenticationBackendFile{Path: "/a/path", Password: password}
}

func (suite *FileBasedAuthenticationBackend) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *FileBasedAuthenticationBackend) TestShouldValidateWatchDefaultResetInterval() {
	suite.config.File.Watch = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.True(suite.config.RefreshInterval.Valid())
	suite.True(suite.config.RefreshInterval.Always())
	suite.False(suite.config.RefreshInterval.Never())
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenNoPathProvided() {
	suite.config.File.Path = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: option 'path' is required")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenBlank() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}

	suite.Equal("", suite.config.File.Password.Algorithm)
	suite.Equal(0, suite.config.File.Password.KeyLength)   //nolint:staticcheck
	suite.Equal(0, suite.config.File.Password.Iterations)  //nolint:staticcheck
	suite.Equal(0, suite.config.File.Password.SaltLength)  //nolint:staticcheck
	suite.Equal(0, suite.config.File.Password.Memory)      //nolint:staticcheck
	suite.Equal(0, suite.config.File.Password.Parallelism) //nolint:staticcheck

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(schema.DefaultPasswordConfig.Algorithm, suite.config.File.Password.Algorithm)
	suite.Equal(schema.DefaultPasswordConfig.KeyLength, suite.config.File.Password.KeyLength)     //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.Iterations, suite.config.File.Password.Iterations)   //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.SaltLength, suite.config.File.Password.SaltLength)   //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.Memory, suite.config.File.Password.Memory)           //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.Parallelism, suite.config.File.Password.Parallelism) //nolint:staticcheck
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationSHA512() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)

	suite.config.File.Password = schema.AuthenticationBackendFilePassword{
		Algorithm:  digestSHA512,
		Iterations: 1000000,
		SaltLength: 8,
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(hashSHA2Crypt, suite.config.File.Password.Algorithm)
	suite.Equal(digestSHA512, suite.config.File.Password.SHA2Crypt.Variant)
	suite.Equal(1000000, suite.config.File.Password.SHA2Crypt.Iterations)
	suite.Equal(8, suite.config.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationSHA512ButNotOverride() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)

	suite.config.File.Password = schema.AuthenticationBackendFilePassword{
		Algorithm:  digestSHA512,
		Iterations: 1000000,
		SaltLength: 8,
		SHA2Crypt: schema.AuthenticationBackendFilePasswordSHA2Crypt{
			Variant:    digestSHA256,
			Iterations: 50000,
			SaltLength: 12,
		},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(hashSHA2Crypt, suite.config.File.Password.Algorithm)
	suite.Equal(digestSHA256, suite.config.File.Password.SHA2Crypt.Variant)
	suite.Equal(50000, suite.config.File.Password.SHA2Crypt.Iterations)
	suite.Equal(12, suite.config.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationSHA512Alt() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)

	suite.config.File.Password = schema.AuthenticationBackendFilePassword{
		Algorithm:  digestSHA512,
		Iterations: 1000000,
		SaltLength: 64,
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(hashSHA2Crypt, suite.config.File.Password.Algorithm)
	suite.Equal(digestSHA512, suite.config.File.Password.SHA2Crypt.Variant)
	suite.Equal(1000000, suite.config.File.Password.SHA2Crypt.Iterations)
	suite.Equal(16, suite.config.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationArgon2() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)

	suite.config.File.Password = schema.AuthenticationBackendFilePassword{
		Algorithm:   "argon2id",
		Iterations:  4,
		Memory:      1024,
		Parallelism: 4,
		KeyLength:   64,
		SaltLength:  64,
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("argon2", suite.config.File.Password.Algorithm)
	suite.Equal("argon2id", suite.config.File.Password.Argon2.Variant)
	suite.Equal(4, suite.config.File.Password.Argon2.Iterations)
	suite.Equal(1048576, suite.config.File.Password.Argon2.Memory)
	suite.Equal(4, suite.config.File.Password.Argon2.Parallelism)
	suite.Equal(64, suite.config.File.Password.Argon2.KeyLength)
	suite.Equal(64, suite.config.File.Password.Argon2.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationArgon2ButNotOverride() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)

	suite.config.File.Password = schema.AuthenticationBackendFilePassword{
		Algorithm:   "argon2id",
		Iterations:  4,
		Memory:      1024,
		Parallelism: 4,
		KeyLength:   64,
		SaltLength:  64,
		Argon2: schema.AuthenticationBackendFilePasswordArgon2{
			Variant:     "argon2d",
			Iterations:  1,
			Memory:      2048,
			Parallelism: 1,
			KeyLength:   32,
			SaltLength:  32,
		},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("argon2", suite.config.File.Password.Algorithm)
	suite.Equal("argon2d", suite.config.File.Password.Argon2.Variant)
	suite.Equal(1, suite.config.File.Password.Argon2.Iterations)
	suite.Equal(2048, suite.config.File.Password.Argon2.Memory)
	suite.Equal(1, suite.config.File.Password.Argon2.Parallelism)
	suite.Equal(32, suite.config.File.Password.Argon2.KeyLength)
	suite.Equal(32, suite.config.File.Password.Argon2.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationWhenOnlySHA512Set() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = digestSHA512

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(hashSHA2Crypt, suite.config.File.Password.Algorithm)
	suite.Equal(digestSHA512, suite.config.File.Password.SHA2Crypt.Variant)
	suite.Equal(schema.DefaultPasswordConfig.SHA2Crypt.Iterations, suite.config.File.Password.SHA2Crypt.Iterations)
	suite.Equal(schema.DefaultPasswordConfig.SHA2Crypt.SaltLength, suite.config.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidArgon2Variant() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = "argon2"
	suite.config.File.Password.Argon2.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'variant' must be one of 'argon2id', 'id', 'argon2i', 'i', 'argon2d', or 'd' but it's configured as 'invalid'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidSHA2CryptVariant() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = hashSHA2Crypt
	suite.config.File.Password.SHA2Crypt.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'variant' must be one of 'sha256' or 'sha512' but it's configured as 'invalid'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidSHA2CryptSaltLength() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = hashSHA2Crypt
	suite.config.File.Password.SHA2Crypt.SaltLength = 40

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'salt_length' is configured as '40' but must be less than or equal to '16'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidPBKDF2Variant() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = "pbkdf2"
	suite.config.File.Password.PBKDF2.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: pbkdf2: option 'variant' must be one of 'sha1', 'sha224', 'sha256', 'sha384', or 'sha512' but it's configured as 'invalid'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidBCryptVariant() {
	suite.config.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = "bcrypt"
	suite.config.File.Password.BCrypt.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: bcrypt: option 'variant' must be one of 'standard' or 'sha256' but it's configured as 'invalid'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSHA2CryptOptionsTooLow() {
	suite.config.File.Password.SHA2Crypt.Iterations = -1
	suite.config.File.Password.SHA2Crypt.SaltLength = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'iterations' is configured as '-1' but must be greater than or equal to '1000'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: sha2crypt: option 'salt_length' is configured as '-1' but must be greater than or equal to '1'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSHA2CryptOptionsTooHigh() {
	suite.config.File.Password.SHA2Crypt.Iterations = 999999999999
	suite.config.File.Password.SHA2Crypt.SaltLength = 99

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'iterations' is configured as '999999999999' but must be less than or equal to '999999999'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: sha2crypt: option 'salt_length' is configured as '99' but must be less than or equal to '16'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenPBKDF2OptionsTooLow() {
	suite.config.File.Password.PBKDF2.Iterations = -1
	suite.config.File.Password.PBKDF2.SaltLength = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: pbkdf2: option 'iterations' is configured as '-1' but must be greater than or equal to '100000'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: pbkdf2: option 'salt_length' is configured as '-1' but must be greater than or equal to '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenPBKDF2OptionsTooHigh() {
	suite.config.File.Password.PBKDF2.Iterations = 2147483649
	suite.config.File.Password.PBKDF2.SaltLength = 2147483650

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: pbkdf2: option 'iterations' is configured as '2147483649' but must be less than or equal to '2147483647'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: pbkdf2: option 'salt_length' is configured as '2147483650' but must be less than or equal to '2147483647'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBCryptOptionsTooLow() {
	suite.config.File.Password.BCrypt.Cost = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: bcrypt: option 'cost' is configured as '-1' but must be greater than or equal to '10'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBCryptOptionsTooHigh() {
	suite.config.File.Password.BCrypt.Cost = 900

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: bcrypt: option 'cost' is configured as '900' but must be less than or equal to '31'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSCryptOptionsTooLow() {
	suite.config.File.Password.SCrypt.Iterations = -1
	suite.config.File.Password.SCrypt.BlockSize = -21
	suite.config.File.Password.SCrypt.Parallelism = -11
	suite.config.File.Password.SCrypt.KeyLength = -77
	suite.config.File.Password.SCrypt.SaltLength = 7

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 5)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: scrypt: option 'iterations' is configured as '-1' but must be greater than or equal to '1'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: scrypt: option 'block_size' is configured as '-21' but must be greater than or equal to '1'")
	suite.EqualError(suite.validator.Errors()[2], "authentication_backend: file: password: scrypt: option 'parallelism' is configured as '-11' but must be greater than or equal to '1'")
	suite.EqualError(suite.validator.Errors()[3], "authentication_backend: file: password: scrypt: option 'key_length' is configured as '-77' but must be greater than or equal to '1'")
	suite.EqualError(suite.validator.Errors()[4], "authentication_backend: file: password: scrypt: option 'salt_length' is configured as '7' but must be greater than or equal to '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSCryptOptionsTooHigh() {
	suite.config.File.Password.SCrypt.Iterations = 59
	suite.config.File.Password.SCrypt.BlockSize = 360287970189639672
	suite.config.File.Password.SCrypt.Parallelism = 1073741825
	suite.config.File.Password.SCrypt.KeyLength = 1374389534409
	suite.config.File.Password.SCrypt.SaltLength = 2147483647

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 5)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: scrypt: option 'iterations' is configured as '59' but must be less than or equal to '58'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: scrypt: option 'block_size' is configured as '360287970189639672' but must be less than or equal to '36028797018963967'")
	suite.EqualError(suite.validator.Errors()[2], "authentication_backend: file: password: scrypt: option 'parallelism' is configured as '1073741825' but must be less than or equal to '1073741823'")
	suite.EqualError(suite.validator.Errors()[3], "authentication_backend: file: password: scrypt: option 'key_length' is configured as '1374389534409' but must be less than or equal to '137438953440'")
	suite.EqualError(suite.validator.Errors()[4], "authentication_backend: file: password: scrypt: option 'salt_length' is configured as '2147483647' but must be less than or equal to '1024'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenArgon2OptionsTooLow() {
	suite.config.File.Password.Argon2.Iterations = -1
	suite.config.File.Password.Argon2.Memory = -1
	suite.config.File.Password.Argon2.Parallelism = -1
	suite.config.File.Password.Argon2.KeyLength = 1
	suite.config.File.Password.Argon2.SaltLength = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 5)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'iterations' is configured as '-1' but must be greater than or equal to '1'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: argon2: option 'parallelism' is configured as '-1' but must be greater than or equal to '1'")
	suite.EqualError(suite.validator.Errors()[2], "authentication_backend: file: password: argon2: option 'memory' is configured as '-1' but must be greater than or equal to '8'")
	suite.EqualError(suite.validator.Errors()[3], "authentication_backend: file: password: argon2: option 'key_length' is configured as '1' but must be greater than or equal to '4'")
	suite.EqualError(suite.validator.Errors()[4], "authentication_backend: file: password: argon2: option 'salt_length' is configured as '-1' but must be greater than or equal to '1'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenArgon2OptionsTooHigh() {
	suite.config.File.Password.Argon2.Iterations = 9999999999
	suite.config.File.Password.Argon2.Memory = 4294967296
	suite.config.File.Password.Argon2.Parallelism = 16777216
	suite.config.File.Password.Argon2.KeyLength = 9999999998
	suite.config.File.Password.Argon2.SaltLength = 9999999997

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 5)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'iterations' is configured as '9999999999' but must be less than or equal to '2147483647'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: argon2: option 'parallelism' is configured as '16777216' but must be less than or equal to '16777215'")
	suite.EqualError(suite.validator.Errors()[2], "authentication_backend: file: password: argon2: option 'memory' is configured as '4294967296' but must be less than or equal to '4294967295'")
	suite.EqualError(suite.validator.Errors()[3], "authentication_backend: file: password: argon2: option 'key_length' is configured as '9999999998' but must be less than or equal to '2147483647'")
	suite.EqualError(suite.validator.Errors()[4], "authentication_backend: file: password: argon2: option 'salt_length' is configured as '9999999997' but must be less than or equal to '2147483647'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenArgon2MemoryTooLow() {
	suite.config.File.Password.Argon2.Memory = 4
	suite.config.File.Password.Argon2.Parallelism = 4

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'memory' is configured as '4' but must be greater than or equal to '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenArgon2MemoryTooLowMultiplier() {
	suite.config.File.Password.Argon2.Memory = 8
	suite.config.File.Password.Argon2.Parallelism = 4

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'memory' is configured as '8' but must be greater than or equal to '32' or '4' (the value of 'parallelism) multiplied by '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBadAlgorithmDefined() {
	suite.config.File.Password.Algorithm = "bogus"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: option 'algorithm' must be one of 'sha2crypt', 'pbkdf2', 'scrypt', 'bcrypt', or 'argon2' but it's configured as 'bogus'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultValues() {
	suite.config.File.Password.Algorithm = ""
	suite.config.File.Password.Iterations = 0  //nolint:staticcheck
	suite.config.File.Password.SaltLength = 0  //nolint:staticcheck
	suite.config.File.Password.Memory = 0      //nolint:staticcheck
	suite.config.File.Password.Parallelism = 0 //nolint:staticcheck

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(schema.DefaultPasswordConfig.Algorithm, suite.config.File.Password.Algorithm)
	suite.Equal(schema.DefaultPasswordConfig.Iterations, suite.config.File.Password.Iterations)   //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.SaltLength, suite.config.File.Password.SaltLength)   //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.Memory, suite.config.File.Password.Memory)           //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.Parallelism, suite.config.File.Password.Parallelism) //nolint:staticcheck
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenResetURLIsInvalid() {
	suite.config.PasswordReset.CustomURL = url.URL{Scheme: "ldap", Host: "google.com"}
	suite.config.PasswordReset.Disable = true

	suite.True(suite.config.PasswordReset.Disable)

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: password_reset: option 'custom_url' is configured to 'ldap://google.com' which has the scheme 'ldap' but the scheme must be either 'http' or 'https'")

	suite.True(suite.config.PasswordReset.Disable)
}

func (suite *FileBasedAuthenticationBackend) TestShouldNotRaiseErrorWhenResetURLIsValid() {
	suite.config.PasswordReset.CustomURL = url.URL{Scheme: schemeHTTPS, Host: "google.com"}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *FileBasedAuthenticationBackend) TestShouldConfigureDisableResetPasswordWhenCustomURL() {
	suite.config.PasswordReset.CustomURL = url.URL{Scheme: schemeHTTPS, Host: "google.com"}
	suite.config.PasswordReset.Disable = true

	suite.True(suite.config.PasswordReset.Disable)

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.False(suite.config.PasswordReset.Disable)
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
	suite.config.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.LDAP.Implementation = schema.LDAPImplementationCustom
	suite.config.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.LDAP.User = testLDAPUser
	suite.config.LDAP.Password = testLDAPPassword
	suite.config.LDAP.BaseDN = testLDAPBaseDN
	suite.config.LDAP.Attributes.Username = "uid"
	suite.config.LDAP.UsersFilter = "({username_attribute}={input})"
	suite.config.LDAP.GroupsFilter = "(cn={input})"
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateDefaultImplementationAndUsernameAttribute() {
	suite.config.LDAP.Implementation = ""
	suite.config.LDAP.Attributes.Username = ""
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Equal(schema.LDAPImplementationCustom, suite.config.LDAP.Implementation)

	suite.Equal(suite.config.LDAP.Attributes.Username, schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Attributes.Username)
	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenImplementationIsInvalidMSAD() {
	suite.config.LDAP.Implementation = "masd"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'implementation' must be one of 'custom', 'activedirectory', 'rfc2307bis', 'freeipa', 'lldap', or 'glauth' but it's configured as 'masd'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenURLNotProvided() {
	suite.config.LDAP.Address = nil
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'address' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenUserNotProvided() {
	suite.config.LDAP.User = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'user' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenPasswordNotProvided() {
	suite.config.LDAP.Password = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'password' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseErrorWhenPasswordNotProvidedWithPermitUnauthenticatedBind() {
	suite.config.LDAP.Password = ""
	suite.config.LDAP.PermitUnauthenticatedBind = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'permit_unauthenticated_bind' can't be enabled when password reset is enabled")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenPasswordProvidedWithPermitUnauthenticatedBind() {
	suite.config.LDAP.Password = "test"
	suite.config.LDAP.PermitUnauthenticatedBind = true
	suite.config.PasswordReset.Disable = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'permit_unauthenticated_bind' can't be enabled when a password is specified")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultPorts() {
	suite.config.LDAP.Address = &schema.AddressLDAP{Address: MustParseAddress("ldap://abc")}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("ldap://abc:389", suite.config.LDAP.Address.String())

	suite.config.LDAP.Address = &schema.AddressLDAP{Address: MustParseAddress("ldaps://abc")}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("ldaps://abc:636", suite.config.LDAP.Address.String())

	suite.config.LDAP.Address = &schema.AddressLDAP{Address: MustParseAddress("ldapi:///a/path")}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("ldapi:///a/path", suite.config.LDAP.Address.String())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseErrorWhenPermitUnauthenticatedBindConfiguredCorrectly() {
	suite.config.LDAP.Password = ""
	suite.config.LDAP.PermitUnauthenticatedBind = true
	suite.config.PasswordReset.Disable = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenBaseDNNotProvided() {
	suite.config.LDAP.BaseDN = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'base_dn' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnEmptyGroupsFilter() {
	suite.config.LDAP.GroupsFilter = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'groups_filter' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnEmptyUsersFilter() {
	suite.config.LDAP.UsersFilter = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseOnEmptyUsernameAttribute() {
	suite.config.LDAP.Attributes.Username = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultImplementation() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(schema.LDAPImplementationCustom, suite.config.LDAP.Implementation)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorOnBadFilterPlaceholders() {
	suite.config.LDAP.UsersFilter = "(&({username_attribute}={0})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.config.LDAP.GroupsFilter = "(&({username_attribute}={1})(member={0})(objectClass=group)(objectCategory=group))"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.True(suite.validator.HasErrors())

	suite.Require().Len(suite.validator.Errors(), 4)
	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' has an invalid placeholder: '{0}' has been removed, please use '{input}' instead")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: ldap: option 'groups_filter' has an invalid placeholder: '{0}' has been removed, please use '{input}' instead")
	suite.EqualError(suite.validator.Errors()[2], "authentication_backend: ldap: option 'groups_filter' has an invalid placeholder: '{1}' has been removed, please use '{username}' instead")
	suite.EqualError(suite.validator.Errors()[3], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{input}' but it's absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultGroupNameAttribute() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("cn", suite.config.LDAP.Attributes.GroupName)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultMailAttribute() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("mail", suite.config.LDAP.Attributes.Mail)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultDisplayNameAttribute() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("displayName", suite.config.LDAP.Attributes.DisplayName)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultRefreshInterval() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Require().NotNil(suite.config.RefreshInterval)
	suite.False(suite.config.RefreshInterval.Always())
	suite.False(suite.config.RefreshInterval.Never())
	suite.Equal(time.Minute*5, suite.config.RefreshInterval.Value())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainEnclosingParenthesis() {
	suite.config.LDAP.UsersFilter = "{username_attribute}={input}"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain enclosing parenthesis: '{username_attribute}={input}' should probably be '({username_attribute}={input})'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenGroupsFilterDoesNotContainEnclosingParenthesis() {
	suite.config.LDAP.GroupsFilter = "cn={input}"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'groups_filter' must contain enclosing parenthesis: 'cn={input}' should probably be '(cn={input})'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainUsernameAttribute() {
	suite.config.LDAP.UsersFilter = "(&({mail_attribute}={input})(objectClass=person))"
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{username_attribute}' but it's absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldHelpDetectNoInputPlaceholder() {
	suite.config.LDAP.UsersFilter = "(&({username_attribute}={mail_attribute})(objectClass=person))"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{input}' but it's absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultTLSMinimumVersion() {
	suite.config.LDAP.TLS = &schema.TLS{MinimumVersion: schema.TLSVersion{}}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.TLS.MinimumVersion.Value, suite.config.LDAP.TLS.MinimumVersion.MinVersion())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotAllowSSL30() {
	suite.config.LDAP.TLS = &schema.TLS{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldErrorOnBadSearchMode() {
	suite.config.LDAP.GroupSearchMode = "memberOF"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'group_search_mode' must be one of 'filter' or 'memberof' but it's configured as 'memberOF'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNoErrorOnPlaceholderSearchMode() {
	suite.config.LDAP.GroupSearchMode = memberof
	suite.config.LDAP.GroupsFilter = filterMemberOfRDN
	suite.config.LDAP.Attributes.MemberOf = memberOf

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldErrorOnMissingPlaceholderSearchMode() {
	suite.config.LDAP.GroupSearchMode = memberof

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'groups_filter' must contain one of the '{memberof:rdn}' or '{memberof:dn}' placeholders when using a group_search_mode of 'memberof' but they're absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldErrorOnMissingDistinguishedNameDN() {
	suite.config.LDAP.Attributes.DistinguishedName = ""
	suite.config.LDAP.GroupsFilter = "(|({memberof:dn}))"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: attributes: option 'distinguished_name' must be provided when using the '{memberof:dn}' placeholder but it's absent")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: ldap: attributes: option 'member_of' must be provided when using the '{memberof:rdn}' or '{memberof:dn}' placeholder but it's absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldErrorOnMissingMemberOfRDN() {
	suite.config.LDAP.Attributes.DistinguishedName = ""
	suite.config.LDAP.GroupsFilter = filterMemberOfRDN

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: attributes: option 'member_of' must be provided when using the '{memberof:rdn}' or '{memberof:dn}' placeholder but it's absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotAllowTLSVerMinGreaterThanVerMax() {
	suite.config.LDAP.TLS = &schema.TLS{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
		MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS1.3 is greater than the maximum version TLS1.2")
}

func TestLDAPAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(LDAPAuthenticationBackendSuite))
}

type ActiveDirectoryAuthenticationBackendSuite struct {
	LDAPImplementationSuite
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.AuthenticationBackend{}
	suite.config.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.LDAP.Implementation = schema.LDAPImplementationActiveDirectory
	suite.config.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.LDAP.User = testLDAPUser
	suite.config.LDAP.Password = testLDAPPassword
	suite.config.LDAP.BaseDN = testLDAPBaseDN
	suite.config.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.TLS
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldSetActiveDirectoryDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory)
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.LDAP.Timeout = time.Second * 2
	suite.config.LDAP.UsersFilter = "(&({username_attribute}={input})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.config.LDAP.Attributes.Username = "cn"
	suite.config.LDAP.Attributes.Mail = "userPrincipalName"
	suite.config.LDAP.Attributes.DisplayName = "name"
	suite.config.LDAP.GroupsFilter = "(&(member={dn})(objectClass=group)(objectCategory=group))"
	suite.config.LDAP.Attributes.GroupName = "distinguishedName"
	suite.config.LDAP.AdditionalUsersDN = "OU=test"
	suite.config.LDAP.AdditionalGroupsDN = "OU=grps"
	suite.config.LDAP.Attributes.MemberOf = member
	suite.config.LDAP.GroupSearchMode = memberof
	suite.config.LDAP.Attributes.DistinguishedName = "objectGUID"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.NotEqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory)

	suite.Equal(member, suite.config.LDAP.Attributes.MemberOf)
	suite.Equal("objectGUID", suite.config.LDAP.Attributes.DistinguishedName)
	suite.Equal(memberof, suite.config.LDAP.GroupSearchMode)
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldRaiseErrorOnInvalidURLWithHTTP() {
	suite.config.LDAP.Address = &schema.AddressLDAP{Address: MustParseAddress("http://dc1:389")}

	validateLDAPAuthenticationAddress(suite.config.LDAP, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'address' with value 'http://dc1:389' is invalid: scheme must be one of 'ldap', 'ldaps', or 'ldapi' but is configured as 'http'")
}

func TestActiveDirectoryAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(ActiveDirectoryAuthenticationBackendSuite))
}

type RFC2307bisAuthenticationBackendSuite struct {
	LDAPImplementationSuite
}

func (suite *RFC2307bisAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.AuthenticationBackend{}
	suite.config.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.LDAP.Implementation = schema.LDAPImplementationRFC2307bis
	suite.config.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.LDAP.User = testLDAPUser
	suite.config.LDAP.Password = testLDAPPassword
	suite.config.LDAP.BaseDN = testLDAPBaseDN
	suite.config.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis.TLS
}

func (suite *RFC2307bisAuthenticationBackendSuite) TestShouldSetDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis)
}

func (suite *RFC2307bisAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.LDAP.Timeout = time.Second * 2
	suite.config.LDAP.UsersFilter = "(&({username_attribute}={input})(objectClass=Person))"
	suite.config.LDAP.Attributes.Username = "o"
	suite.config.LDAP.Attributes.Mail = "Email"
	suite.config.LDAP.Attributes.DisplayName = "Given"
	suite.config.LDAP.GroupsFilter = "(&(member={dn})(objectClass=posixGroup)(objectClass=top))"
	suite.config.LDAP.Attributes.GroupName = "gid"
	suite.config.LDAP.Attributes.MemberOf = member
	suite.config.LDAP.AdditionalUsersDN = "OU=users,OU=OpenLDAP"
	suite.config.LDAP.AdditionalGroupsDN = "OU=groups,OU=OpenLDAP"
	suite.config.LDAP.GroupSearchMode = memberof

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.NotEqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis)

	suite.Equal(member, suite.config.LDAP.Attributes.MemberOf)
	suite.Equal("", suite.config.LDAP.Attributes.DistinguishedName)
	suite.Equal(schema.LDAPGroupSearchModeMemberOf, suite.config.LDAP.GroupSearchMode)
}

func TestRFC2307bisAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(RFC2307bisAuthenticationBackendSuite))
}

type FreeIPAAuthenticationBackendSuite struct {
	LDAPImplementationSuite
}

func (suite *FreeIPAAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.AuthenticationBackend{}
	suite.config.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.LDAP.Implementation = schema.LDAPImplementationFreeIPA
	suite.config.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.LDAP.User = testLDAPUser
	suite.config.LDAP.Password = testLDAPPassword
	suite.config.LDAP.BaseDN = testLDAPBaseDN
	suite.config.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA.TLS
}

func (suite *FreeIPAAuthenticationBackendSuite) TestShouldSetDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA)
}

func (suite *FreeIPAAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.LDAP.Timeout = time.Second * 2
	suite.config.LDAP.UsersFilter = "(&({username_attribute}={input})(objectClass=person)(!(nsAccountLock=TRUE)))"
	suite.config.LDAP.Attributes.Username = "dn"
	suite.config.LDAP.Attributes.Mail = "email"
	suite.config.LDAP.Attributes.DisplayName = "gecos"
	suite.config.LDAP.GroupsFilter = "(&(member={dn})(objectClass=posixgroup))"
	suite.config.LDAP.GroupSearchMode = schema.LDAPGroupSearchModeMemberOf
	suite.config.LDAP.Attributes.GroupName = "groupName"
	suite.config.LDAP.Attributes.MemberOf = member
	suite.config.LDAP.AdditionalUsersDN = "OU=people"
	suite.config.LDAP.AdditionalGroupsDN = "OU=grp"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.NotEqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA)

	suite.Equal(member, suite.config.LDAP.Attributes.MemberOf)
	suite.Equal("", suite.config.LDAP.Attributes.DistinguishedName)
	suite.Equal(schema.LDAPGroupSearchModeMemberOf, suite.config.LDAP.GroupSearchMode)
}

func TestFreeIPAAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(FreeIPAAuthenticationBackendSuite))
}

type LLDAPAuthenticationBackendSuite struct {
	LDAPImplementationSuite
}

func (suite *LLDAPAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.AuthenticationBackend{}
	suite.config.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.LDAP.Implementation = schema.LDAPImplementationLLDAP
	suite.config.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.LDAP.User = testLDAPUser
	suite.config.LDAP.Password = testLDAPPassword
	suite.config.LDAP.BaseDN = testLDAPBaseDN
	suite.config.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP.TLS
}

func (suite *LLDAPAuthenticationBackendSuite) TestShouldSetDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP)
}

func (suite *LLDAPAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.LDAP.Timeout = time.Second * 2
	suite.config.LDAP.UsersFilter = "(&({username_attribute}={input})(objectClass=Person)(!(nsAccountLock=TRUE)))"
	suite.config.LDAP.Attributes.Username = "username"
	suite.config.LDAP.Attributes.Mail = "m"
	suite.config.LDAP.Attributes.DisplayName = "fn"
	suite.config.LDAP.Attributes.MemberOf = member
	suite.config.LDAP.GroupsFilter = "(&(member={dn})(!(objectClass=posixGroup)))"
	suite.config.LDAP.Attributes.GroupName = "grpz"
	suite.config.LDAP.AdditionalUsersDN = "OU=no"
	suite.config.LDAP.AdditionalGroupsDN = "OU=yes"
	suite.config.LDAP.GroupSearchMode = memberof

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.NotEqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP)

	suite.Equal(member, suite.config.LDAP.Attributes.MemberOf)
	suite.Equal("", suite.config.LDAP.Attributes.DistinguishedName)
	suite.Equal(schema.LDAPGroupSearchModeMemberOf, suite.config.LDAP.GroupSearchMode)
}

func TestLLDAPAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(LLDAPAuthenticationBackendSuite))
}

type GLAuthAuthenticationBackendSuite struct {
	LDAPImplementationSuite
}

func (suite *GLAuthAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.AuthenticationBackend{}
	suite.config.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.LDAP.Implementation = schema.LDAPImplementationGLAuth
	suite.config.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.LDAP.User = testLDAPUser
	suite.config.LDAP.Password = testLDAPPassword
	suite.config.LDAP.BaseDN = testLDAPBaseDN
	suite.config.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth.TLS
}

func (suite *GLAuthAuthenticationBackendSuite) TestShouldSetDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth)
}

func (suite *GLAuthAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.LDAP.Timeout = time.Second * 2
	suite.config.LDAP.UsersFilter = "(&({username_attribute}={input})(objectClass=Person)(!(accountStatus=inactive)))"
	suite.config.LDAP.Attributes.Username = "description"
	suite.config.LDAP.Attributes.Mail = "sender"
	suite.config.LDAP.Attributes.DisplayName = "given"
	suite.config.LDAP.GroupsFilter = "(&(member={dn})(objectClass=posixGroup))"
	suite.config.LDAP.Attributes.GroupName = "grp"
	suite.config.LDAP.AdditionalUsersDN = "OU=users,OU=GlAuth"
	suite.config.LDAP.AdditionalGroupsDN = "OU=groups,OU=GLAuth"
	suite.config.LDAP.Attributes.MemberOf = member
	suite.config.LDAP.GroupSearchMode = memberof

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.NotEqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth)

	suite.Equal(member, suite.config.LDAP.Attributes.MemberOf)
	suite.Equal("", suite.config.LDAP.Attributes.DistinguishedName)
	suite.Equal(schema.LDAPGroupSearchModeMemberOf, suite.config.LDAP.GroupSearchMode)
}

func TestGLAuthAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(GLAuthAuthenticationBackendSuite))
}

type LDAPImplementationSuite struct {
	suite.Suite
	config    schema.AuthenticationBackend
	validator *schema.StructValidator
}

func (suite *LDAPImplementationSuite) EqualImplementationDefaults(expected schema.AuthenticationBackendLDAP) {
	suite.Equal(expected.Timeout, suite.config.LDAP.Timeout)
	suite.Equal(expected.AdditionalUsersDN, suite.config.LDAP.AdditionalUsersDN)
	suite.Equal(expected.AdditionalGroupsDN, suite.config.LDAP.AdditionalGroupsDN)
	suite.Equal(expected.UsersFilter, suite.config.LDAP.UsersFilter)
	suite.Equal(expected.GroupsFilter, suite.config.LDAP.GroupsFilter)
	suite.Equal(expected.GroupSearchMode, suite.config.LDAP.GroupSearchMode)

	suite.Equal(expected.Attributes.DistinguishedName, suite.config.LDAP.Attributes.DistinguishedName)
	suite.Equal(expected.Attributes.Username, suite.config.LDAP.Attributes.Username)
	suite.Equal(expected.Attributes.DisplayName, suite.config.LDAP.Attributes.DisplayName)
	suite.Equal(expected.Attributes.Mail, suite.config.LDAP.Attributes.Mail)
	suite.Equal(expected.Attributes.MemberOf, suite.config.LDAP.Attributes.MemberOf)
	suite.Equal(expected.Attributes.GroupName, suite.config.LDAP.Attributes.GroupName)
}

func (suite *LDAPImplementationSuite) NotEqualImplementationDefaults(expected schema.AuthenticationBackendLDAP) {
	suite.NotEqual(expected.Timeout, suite.config.LDAP.Timeout)
	suite.NotEqual(expected.UsersFilter, suite.config.LDAP.UsersFilter)
	suite.NotEqual(expected.GroupsFilter, suite.config.LDAP.GroupsFilter)
	suite.NotEqual(expected.GroupSearchMode, suite.config.LDAP.GroupSearchMode)
	suite.NotEqual(expected.Attributes.Username, suite.config.LDAP.Attributes.Username)
	suite.NotEqual(expected.Attributes.DisplayName, suite.config.LDAP.Attributes.DisplayName)
	suite.NotEqual(expected.Attributes.Mail, suite.config.LDAP.Attributes.Mail)
	suite.NotEqual(expected.Attributes.GroupName, suite.config.LDAP.Attributes.GroupName)

	if expected.Attributes.DistinguishedName != "" {
		suite.NotEqual(expected.Attributes.DistinguishedName, suite.config.LDAP.Attributes.DistinguishedName)
	}

	if expected.AdditionalUsersDN != "" {
		suite.NotEqual(expected.AdditionalUsersDN, suite.config.LDAP.AdditionalUsersDN)
	}

	if expected.AdditionalGroupsDN != "" {
		suite.NotEqual(expected.AdditionalGroupsDN, suite.config.LDAP.AdditionalGroupsDN)
	}

	if expected.Attributes.MemberOf != "" {
		suite.NotEqual(expected.Attributes.MemberOf, suite.config.LDAP.Attributes.MemberOf)
	}
}
