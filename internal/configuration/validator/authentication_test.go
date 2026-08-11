package validator

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-crypt/crypt/algorithm/scrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldRaiseErrorWhenBothBackendsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	backendConfig := schema.Configuration{}

	backendConfig.AuthenticationBackend.LDAP = &schema.AuthenticationBackendLDAP{}
	backendConfig.AuthenticationBackend.File = &schema.AuthenticationBackendFile{
		Path: "/tmp",
	}

	ValidateAuthenticationBackend(&backendConfig, validator)

	require.Len(t, validator.Errors(), 6)
	assert.EqualError(t, validator.Errors()[0], "authentication_backend: please ensure only one of the 'file' or 'ldap' backend is configured")
	assert.EqualError(t, validator.Errors()[1], "authentication_backend: ldap: option 'address' is required")
	assert.EqualError(t, validator.Errors()[2], "authentication_backend: ldap: option 'user' is required")
	assert.EqualError(t, validator.Errors()[3], "authentication_backend: ldap: option 'password' is required")
	assert.EqualError(t, validator.Errors()[4], "authentication_backend: ldap: option 'users_filter' is required")
	assert.EqualError(t, validator.Errors()[5], "authentication_backend: ldap: option 'groups_filter' is required")
}

func TestShouldRaiseErrorWhenNoBackendProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	backendConfig := schema.Configuration{}

	ValidateAuthenticationBackend(&backendConfig, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "authentication_backend: you must ensure either the 'file' or 'ldap' authentication backend is configured")
}

type FileBasedAuthenticationBackend struct {
	suite.Suite
	config    schema.Configuration
	validator *schema.StructValidator
}

func (suite *FileBasedAuthenticationBackend) SetupTest() {
	password := schema.DefaultPasswordConfig

	suite.validator = schema.NewStructValidator()
	suite.config.AuthenticationBackend = schema.AuthenticationBackend{}
	suite.config.AuthenticationBackend.File = &schema.AuthenticationBackendFile{Path: "/a/path", Password: password}
}

func (suite *FileBasedAuthenticationBackend) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *FileBasedAuthenticationBackend) TestShouldValidateWatchDefaultResetInterval() {
	suite.config.AuthenticationBackend.File.Watch = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.True(suite.config.AuthenticationBackend.RefreshInterval.Valid())
	suite.True(suite.config.AuthenticationBackend.RefreshInterval.Always())
	suite.False(suite.config.AuthenticationBackend.RefreshInterval.Never())
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenNoPathProvided() {
	suite.config.AuthenticationBackend.File.Path = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: option 'path' is required")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenBlank() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}

	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.Equal(0, suite.config.AuthenticationBackend.File.Password.KeyLength)   //nolint:staticcheck
	suite.Equal(0, suite.config.AuthenticationBackend.File.Password.Iterations)  //nolint:staticcheck
	suite.Equal(0, suite.config.AuthenticationBackend.File.Password.SaltLength)  //nolint:staticcheck
	suite.Equal(0, suite.config.AuthenticationBackend.File.Password.Memory)      //nolint:staticcheck
	suite.Equal(0, suite.config.AuthenticationBackend.File.Password.Parallelism) //nolint:staticcheck

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(schema.DefaultPasswordConfig.Algorithm, suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.Equal(schema.DefaultPasswordConfig.KeyLength, suite.config.AuthenticationBackend.File.Password.KeyLength)     //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.Iterations, suite.config.AuthenticationBackend.File.Password.Iterations)   //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.SaltLength, suite.config.AuthenticationBackend.File.Password.SaltLength)   //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.Memory, suite.config.AuthenticationBackend.File.Password.Memory)           //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.Parallelism, suite.config.AuthenticationBackend.File.Password.Parallelism) //nolint:staticcheck
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationSHA512() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)

	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{
		Algorithm:  schema.SHA512Lower,
		Iterations: 1000000,
		SaltLength: 8,
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(hashSHA2Crypt, suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.Equal(schema.SHA512Lower, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Variant)
	suite.Equal(1000000, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Iterations)
	suite.Equal(8, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationSHA512ButNotOverride() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)

	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{
		Algorithm:  schema.SHA512Lower,
		Iterations: 1000000,
		SaltLength: 8,
		SHA2Crypt: schema.AuthenticationBackendFilePasswordSHA2Crypt{
			Variant:    schema.SHA256Lower,
			Iterations: 50000,
			SaltLength: 12,
		},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(hashSHA2Crypt, suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.Equal(schema.SHA256Lower, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Variant)
	suite.Equal(50000, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Iterations)
	suite.Equal(12, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationSHA512Alt() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)

	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{
		Algorithm:  schema.SHA512Lower,
		Iterations: 1000000,
		SaltLength: 64,
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(hashSHA2Crypt, suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.Equal(schema.SHA512Lower, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Variant)
	suite.Equal(1000000, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Iterations)
	suite.Equal(16, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationArgon2() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)

	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{
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

	suite.Equal("argon2", suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.Equal("argon2id", suite.config.AuthenticationBackend.File.Password.Argon2.Variant)
	suite.Equal(4, suite.config.AuthenticationBackend.File.Password.Argon2.Iterations)
	suite.Equal(1048576, suite.config.AuthenticationBackend.File.Password.Argon2.Memory)
	suite.Equal(4, suite.config.AuthenticationBackend.File.Password.Argon2.Parallelism)
	suite.Equal(64, suite.config.AuthenticationBackend.File.Password.Argon2.KeyLength)
	suite.Equal(64, suite.config.AuthenticationBackend.File.Password.Argon2.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationArgon2ButNotOverride() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)

	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{
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

	suite.Equal("argon2", suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.Equal("argon2d", suite.config.AuthenticationBackend.File.Password.Argon2.Variant)
	suite.Equal(1, suite.config.AuthenticationBackend.File.Password.Argon2.Iterations)
	suite.Equal(2048, suite.config.AuthenticationBackend.File.Password.Argon2.Memory)
	suite.Equal(1, suite.config.AuthenticationBackend.File.Password.Argon2.Parallelism)
	suite.Equal(32, suite.config.AuthenticationBackend.File.Password.Argon2.KeyLength)
	suite.Equal(32, suite.config.AuthenticationBackend.File.Password.Argon2.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldMigrateLegacyConfigurationWhenOnlySHA512Set() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.config.AuthenticationBackend.File.Password.Algorithm = schema.SHA512Lower

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(hashSHA2Crypt, suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.Equal(schema.SHA512Lower, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Variant)
	suite.Equal(schema.DefaultPasswordConfig.SHA2Crypt.Iterations, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Iterations)
	suite.Equal(schema.DefaultPasswordConfig.SHA2Crypt.SaltLength, suite.config.AuthenticationBackend.File.Password.SHA2Crypt.SaltLength)
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidArgon2Variant() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.config.AuthenticationBackend.File.Password.Algorithm = "argon2"
	suite.config.AuthenticationBackend.File.Password.Argon2.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'variant' must be one of 'argon2id', 'id', 'argon2i', 'i', 'argon2d', or 'd' but it's configured as 'invalid'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidSHA2CryptVariant() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.config.AuthenticationBackend.File.Password.Algorithm = hashSHA2Crypt
	suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'variant' must be one of 'sha256' or 'sha512' but it's configured as 'invalid'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidSHA2CryptSaltLength() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.config.AuthenticationBackend.File.Password.Algorithm = hashSHA2Crypt
	suite.config.AuthenticationBackend.File.Password.SHA2Crypt.SaltLength = 40

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'salt_length' is configured as '40' but must be less than or equal to '16'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidPBKDF2Variant() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.config.AuthenticationBackend.File.Password.Algorithm = "pbkdf2"
	suite.config.AuthenticationBackend.File.Password.PBKDF2.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: pbkdf2: option 'variant' must be one of 'sha1', 'sha224', 'sha256', 'sha384', or 'sha512' but it's configured as 'invalid'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorOnInvalidBcryptVariant() {
	suite.config.AuthenticationBackend.File.Password = schema.AuthenticationBackendFilePassword{}
	suite.Equal("", suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.config.AuthenticationBackend.File.Password.Algorithm = "bcrypt"
	suite.config.AuthenticationBackend.File.Password.Bcrypt.Variant = testInvalid

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: bcrypt: option 'variant' must be one of 'standard' or 'sha256' but it's configured as 'invalid'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSHA2CryptOptionsTooLow() {
	suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Iterations = -1
	suite.config.AuthenticationBackend.File.Password.SHA2Crypt.SaltLength = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'iterations' is configured as '-1' but must be greater than or equal to '1000'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: sha2crypt: option 'salt_length' is configured as '-1' but must be greater than or equal to '1'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSHA2CryptOptionsTooHigh() {
	suite.config.AuthenticationBackend.File.Password.SHA2Crypt.Iterations = 999999999999
	suite.config.AuthenticationBackend.File.Password.SHA2Crypt.SaltLength = 99

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: sha2crypt: option 'iterations' is configured as '999999999999' but must be less than or equal to '999999999'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: sha2crypt: option 'salt_length' is configured as '99' but must be less than or equal to '16'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenPBKDF2OptionsTooLow() {
	suite.config.AuthenticationBackend.File.Password.PBKDF2.Iterations = -1
	suite.config.AuthenticationBackend.File.Password.PBKDF2.SaltLength = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: pbkdf2: option 'iterations' is configured as '-1' but must be greater than or equal to '100000'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: pbkdf2: option 'salt_length' is configured as '-1' but must be greater than or equal to '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenPBKDF2OptionsTooHigh() {
	suite.config.AuthenticationBackend.File.Password.PBKDF2.Iterations = 2147483649
	suite.config.AuthenticationBackend.File.Password.PBKDF2.SaltLength = 2147483650

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: pbkdf2: option 'iterations' is configured as '2147483649' but must be less than or equal to '2147483647'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: pbkdf2: option 'salt_length' is configured as '2147483650' but must be less than or equal to '2147483647'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBcryptOptionsTooLow() {
	suite.config.AuthenticationBackend.File.Password.Bcrypt.Cost = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: bcrypt: option 'cost' is configured as '-1' but must be greater than or equal to '10'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBcryptOptionsTooHigh() {
	suite.config.AuthenticationBackend.File.Password.Bcrypt.Cost = 900

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: bcrypt: option 'cost' is configured as '900' but must be less than or equal to '31'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenScryptOptionsTooLow() {
	suite.config.AuthenticationBackend.File.Password.Scrypt.Iterations = -1
	suite.config.AuthenticationBackend.File.Password.Scrypt.BlockSize = -21
	suite.config.AuthenticationBackend.File.Password.Scrypt.Parallelism = -11
	suite.config.AuthenticationBackend.File.Password.Scrypt.KeyLength = -77
	suite.config.AuthenticationBackend.File.Password.Scrypt.SaltLength = 7

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 5)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: scrypt: option 'iterations' is configured as '-1' but must be greater than or equal to '1'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: scrypt: option 'block_size' is configured as '-21' but must be greater than or equal to '1'")
	suite.EqualError(suite.validator.Errors()[2], "authentication_backend: file: password: scrypt: option 'parallelism' is configured as '-11' but must be greater than or equal to '1'")
	suite.EqualError(suite.validator.Errors()[3], "authentication_backend: file: password: scrypt: option 'key_length' is configured as '-77' but must be greater than or equal to '1'")
	suite.EqualError(suite.validator.Errors()[4], "authentication_backend: file: password: scrypt: option 'salt_length' is configured as '7' but must be greater than or equal to '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenScryptOptionsTooHigh() {
	suite.config.AuthenticationBackend.File.Password.Scrypt.Iterations = 59
	suite.config.AuthenticationBackend.File.Password.Scrypt.BlockSize = 360287970189639672
	suite.config.AuthenticationBackend.File.Password.Scrypt.Parallelism = 1073741825
	suite.config.AuthenticationBackend.File.Password.Scrypt.KeyLength = 1374389534409
	suite.config.AuthenticationBackend.File.Password.Scrypt.SaltLength = 2147483647

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 5)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: scrypt: option 'iterations' is configured as '59' but must be less than or equal to '58'")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: file: password: scrypt: option 'block_size' is configured as '360287970189639672' but must be less than or equal to '36028797018963967'")
	suite.EqualError(suite.validator.Errors()[2], "authentication_backend: file: password: scrypt: option 'parallelism' is configured as '1073741825' but must be less than or equal to '1073741823'")
	suite.EqualError(suite.validator.Errors()[3], fmt.Sprintf("authentication_backend: file: password: scrypt: option 'key_length' is configured as '1374389534409' but must be less than or equal to '%d'", scrypt.KeyLengthMax))
	suite.EqualError(suite.validator.Errors()[4], "authentication_backend: file: password: scrypt: option 'salt_length' is configured as '2147483647' but must be less than or equal to '1024'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenArgon2OptionsTooLow() {
	suite.config.AuthenticationBackend.File.Password.Argon2.Iterations = -1
	suite.config.AuthenticationBackend.File.Password.Argon2.Memory = -1
	suite.config.AuthenticationBackend.File.Password.Argon2.Parallelism = -1
	suite.config.AuthenticationBackend.File.Password.Argon2.KeyLength = 1
	suite.config.AuthenticationBackend.File.Password.Argon2.SaltLength = -1

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
	suite.config.AuthenticationBackend.File.Password.Argon2.Iterations = 9999999999
	suite.config.AuthenticationBackend.File.Password.Argon2.Memory = 4294967296
	suite.config.AuthenticationBackend.File.Password.Argon2.Parallelism = 16777216
	suite.config.AuthenticationBackend.File.Password.Argon2.KeyLength = 9999999998
	suite.config.AuthenticationBackend.File.Password.Argon2.SaltLength = 9999999997

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
	suite.config.AuthenticationBackend.File.Password.Argon2.Memory = 4
	suite.config.AuthenticationBackend.File.Password.Argon2.Parallelism = 4

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'memory' is configured as '4' but must be greater than or equal to '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenArgon2MemoryTooLowMultiplier() {
	suite.config.AuthenticationBackend.File.Password.Argon2.Memory = 8
	suite.config.AuthenticationBackend.File.Password.Argon2.Parallelism = 4

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: argon2: option 'memory' is configured as '8' but must be greater than or equal to '32' or '4' (the value of 'parallelism) multiplied by '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBadAlgorithmDefined() {
	suite.config.AuthenticationBackend.File.Password.Algorithm = "bogus"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: option 'algorithm' must be one of 'sha2crypt', 'pbkdf2', 'scrypt', 'bcrypt', or 'argon2' but it's configured as 'bogus'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultValues() {
	suite.config.AuthenticationBackend.File.Password.Algorithm = ""
	suite.config.AuthenticationBackend.File.Password.Iterations = 0  //nolint:staticcheck
	suite.config.AuthenticationBackend.File.Password.SaltLength = 0  //nolint:staticcheck
	suite.config.AuthenticationBackend.File.Password.Memory = 0      //nolint:staticcheck
	suite.config.AuthenticationBackend.File.Password.Parallelism = 0 //nolint:staticcheck

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(schema.DefaultPasswordConfig.Algorithm, suite.config.AuthenticationBackend.File.Password.Algorithm)
	suite.Equal(schema.DefaultPasswordConfig.Iterations, suite.config.AuthenticationBackend.File.Password.Iterations)   //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.SaltLength, suite.config.AuthenticationBackend.File.Password.SaltLength)   //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.Memory, suite.config.AuthenticationBackend.File.Password.Memory)           //nolint:staticcheck
	suite.Equal(schema.DefaultPasswordConfig.Parallelism, suite.config.AuthenticationBackend.File.Password.Parallelism) //nolint:staticcheck
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenResetURLIsInvalid() {
	suite.config.AuthenticationBackend.PasswordReset.CustomURL = url.URL{Scheme: "ldap", Host: "google.com"}
	suite.config.AuthenticationBackend.PasswordReset.Disable = true

	suite.True(suite.config.AuthenticationBackend.PasswordReset.Disable)

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: password_reset: option 'custom_url' is configured to 'ldap://google.com' which has the scheme 'ldap' but the scheme must be either 'http' or 'https'")

	suite.True(suite.config.AuthenticationBackend.PasswordReset.Disable)
}

func (suite *FileBasedAuthenticationBackend) TestShouldNotRaiseErrorWhenResetURLIsValid() {
	suite.config.AuthenticationBackend.PasswordReset.CustomURL = url.URL{Scheme: schemeHTTPS, Host: "google.com"}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *FileBasedAuthenticationBackend) TestShouldConfigureDisableResetPasswordWhenCustomURL() {
	suite.config.AuthenticationBackend.PasswordReset.CustomURL = url.URL{Scheme: schemeHTTPS, Host: "google.com"}
	suite.config.AuthenticationBackend.PasswordReset.Disable = true

	suite.True(suite.config.AuthenticationBackend.PasswordReset.Disable)

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.False(suite.config.AuthenticationBackend.PasswordReset.Disable)
}

func (suite *FileBasedAuthenticationBackend) TestShouldValidateExtraAttributeString() {
	suite.config.AuthenticationBackend.File.ExtraAttributes = map[string]schema.AuthenticationBackendExtraAttribute{
		"custom_attr": {ValueType: "string"},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *FileBasedAuthenticationBackend) TestShouldValidateExtraAttributeInteger() {
	suite.config.AuthenticationBackend.File.ExtraAttributes = map[string]schema.AuthenticationBackendExtraAttribute{
		"custom_int": {ValueType: "integer"},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *FileBasedAuthenticationBackend) TestShouldValidateExtraAttributeBoolean() {
	suite.config.AuthenticationBackend.File.ExtraAttributes = map[string]schema.AuthenticationBackendExtraAttribute{
		"custom_bool": {ValueType: "boolean"},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenExtraAttributeValueTypeMissing() {
	suite.config.AuthenticationBackend.File.ExtraAttributes = map[string]schema.AuthenticationBackendExtraAttribute{
		"custom_attr": {ValueType: ""},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: extra_attributes: custom_attr: option 'value_type' is required")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenExtraAttributeValueTypeInvalid() {
	suite.config.AuthenticationBackend.File.ExtraAttributes = map[string]schema.AuthenticationBackendExtraAttribute{
		"custom_attr": {ValueType: "invalid"},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: extra_attributes: custom_attr: option 'value_type' must be one of 'string', 'integer', or 'boolean' but it's configured as 'invalid'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenExtraAttributeNameReserved() {
	suite.config.AuthenticationBackend.File.ExtraAttributes = map[string]schema.AuthenticationBackendExtraAttribute{
		"username": {ValueType: "string"},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: file: extra_attributes: username: attribute name 'username' is reserved")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseMultipleErrorsForExtraAttributes() {
	suite.config.AuthenticationBackend.File.ExtraAttributes = map[string]schema.AuthenticationBackendExtraAttribute{
		"email": {ValueType: "string"},
		"bad":   {ValueType: ""},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.GreaterOrEqual(len(suite.validator.Errors()), 2)
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenScryptVariantInvalid() {
	suite.config.AuthenticationBackend.File.Password.Algorithm = "scrypt"
	suite.config.AuthenticationBackend.File.Password.Scrypt.Variant = "invalid"
	suite.config.AuthenticationBackend.File.Password.Scrypt.Iterations = 1
	suite.config.AuthenticationBackend.File.Password.Scrypt.BlockSize = 1
	suite.config.AuthenticationBackend.File.Password.Scrypt.Parallelism = 1
	suite.config.AuthenticationBackend.File.Password.Scrypt.KeyLength = 16
	suite.config.AuthenticationBackend.File.Password.Scrypt.SaltLength = 8

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().GreaterOrEqual(len(suite.validator.Errors()), 1)

	found := false

	for _, err := range suite.validator.Errors() {
		if strings.Contains(err.Error(), "variant") {
			found = true

			break
		}
	}

	suite.True(found)
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenScryptYescryptParallelismNotOne() {
	suite.config.AuthenticationBackend.File.Password.Algorithm = hashScrypt
	suite.config.AuthenticationBackend.File.Password.Scrypt.Variant = hashScryptVariantYesCrypt
	suite.config.AuthenticationBackend.File.Password.Scrypt.Iterations = 1
	suite.config.AuthenticationBackend.File.Password.Scrypt.BlockSize = 1
	suite.config.AuthenticationBackend.File.Password.Scrypt.Parallelism = 2
	suite.config.AuthenticationBackend.File.Password.Scrypt.KeyLength = 16
	suite.config.AuthenticationBackend.File.Password.Scrypt.SaltLength = 8

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().GreaterOrEqual(len(suite.validator.Errors()), 1)

	found := false

	for _, err := range suite.validator.Errors() {
		if strings.Contains(err.Error(), "parallelism") && strings.Contains(err.Error(), "yescrypt") {
			found = true

			break
		}
	}

	suite.True(found)
}

func TestFileBasedAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(FileBasedAuthenticationBackend))
}

type LDAPAuthenticationBackendSuite struct {
	suite.Suite
	config    schema.Configuration
	validator *schema.StructValidator
}

func (suite *LDAPAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config.AuthenticationBackend = schema.AuthenticationBackend{}
	suite.config.AuthenticationBackend.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.AuthenticationBackend.LDAP.Implementation = schema.LDAPImplementationCustom
	suite.config.AuthenticationBackend.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.AuthenticationBackend.LDAP.User = testLDAPUser
	suite.config.AuthenticationBackend.LDAP.Password = testLDAPPassword
	suite.config.AuthenticationBackend.LDAP.BaseDN = testLDAPBaseDN
	suite.config.AuthenticationBackend.LDAP.Attributes.Username = "uid"
	suite.config.AuthenticationBackend.LDAP.UsersFilter = "({username_attribute}={input})"
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = "(cn={input})"
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateDefaultImplementationAndUsernameAttribute() {
	suite.config.AuthenticationBackend.LDAP.Implementation = ""
	suite.config.AuthenticationBackend.LDAP.Attributes.Username = ""
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Equal(schema.LDAPImplementationCustom, suite.config.AuthenticationBackend.LDAP.Implementation)

	suite.Equal(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Attributes.Username, suite.config.AuthenticationBackend.LDAP.Attributes.Username)
	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(0, suite.config.AuthenticationBackend.LDAP.Pooling.Retries)
	suite.Equal(0, suite.config.AuthenticationBackend.LDAP.Pooling.Count)
	suite.Equal(time.Duration(0), suite.config.AuthenticationBackend.LDAP.Pooling.Timeout)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateDefaultPooling() {
	suite.config.AuthenticationBackend.LDAP.Pooling.Enable = true
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Equal(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Retries, suite.config.AuthenticationBackend.LDAP.Pooling.Retries)
	suite.Equal(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Count, suite.config.AuthenticationBackend.LDAP.Pooling.Count)
	suite.Equal(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Timeout, suite.config.AuthenticationBackend.LDAP.Pooling.Timeout)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenImplementationIsInvalidMSAD() {
	suite.config.AuthenticationBackend.LDAP.Implementation = "masd"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'implementation' must be one of 'custom', 'activedirectory', 'rfc2307bis', 'freeipa', 'lldap', or 'glauth' but it's configured as 'masd'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenURLNotProvided() {
	suite.config.AuthenticationBackend.LDAP.Address = nil
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'address' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenUserNotProvided() {
	suite.config.AuthenticationBackend.LDAP.User = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'user' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenPasswordNotProvided() {
	suite.config.AuthenticationBackend.LDAP.Password = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'password' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseErrorWhenPasswordNotProvidedWithPermitUnauthenticatedBind() {
	suite.config.AuthenticationBackend.LDAP.Password = ""
	suite.config.AuthenticationBackend.LDAP.PermitUnauthenticatedBind = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'permit_unauthenticated_bind' can't be enabled when password reset is enabled")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenPasswordProvidedWithPermitUnauthenticatedBind() {
	suite.config.AuthenticationBackend.LDAP.Password = "test"
	suite.config.AuthenticationBackend.LDAP.PermitUnauthenticatedBind = true
	suite.config.AuthenticationBackend.PasswordReset.Disable = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'permit_unauthenticated_bind' can't be enabled when a password is specified")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultPorts() {
	suite.config.AuthenticationBackend.LDAP.Address = &schema.AddressLDAP{Address: MustParseAddress("ldap://abc")}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("ldap://abc:389", suite.config.AuthenticationBackend.LDAP.Address.String())

	suite.config.AuthenticationBackend.LDAP.Address = &schema.AddressLDAP{Address: MustParseAddress("ldaps://abc")}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("ldaps://abc:636", suite.config.AuthenticationBackend.LDAP.Address.String())

	suite.config.AuthenticationBackend.LDAP.Address = &schema.AddressLDAP{Address: MustParseAddress("ldapi:///a/path")}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("ldapi:///a/path", suite.config.AuthenticationBackend.LDAP.Address.String())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseErrorWhenPermitUnauthenticatedBindConfiguredCorrectly() {
	suite.config.AuthenticationBackend.LDAP.Password = ""
	suite.config.AuthenticationBackend.LDAP.PermitUnauthenticatedBind = true
	suite.config.AuthenticationBackend.PasswordReset.Disable = true

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnEmptyGroupsFilter() {
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'groups_filter' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnEmptyUsersFilter() {
	suite.config.AuthenticationBackend.LDAP.UsersFilter = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseOnEmptyUsernameAttribute() {
	suite.config.AuthenticationBackend.LDAP.Attributes.Username = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultImplementation() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(schema.LDAPImplementationCustom, suite.config.AuthenticationBackend.LDAP.Implementation)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorOnBadFilterPlaceholders() {
	suite.config.AuthenticationBackend.LDAP.UsersFilter = "(&({username_attribute}={0})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = "(&({username_attribute}={1})(member={0})(objectClass=group)(objectCategory=group))"

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

	suite.Equal("cn", suite.config.AuthenticationBackend.LDAP.Attributes.GroupName)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultMailAttribute() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("mail", suite.config.AuthenticationBackend.LDAP.Attributes.Mail)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultDisplayNameAttribute() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal("displayName", suite.config.AuthenticationBackend.LDAP.Attributes.DisplayName)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultRefreshInterval() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Require().NotNil(suite.config.AuthenticationBackend.RefreshInterval)
	suite.False(suite.config.AuthenticationBackend.RefreshInterval.Always())
	suite.False(suite.config.AuthenticationBackend.RefreshInterval.Never())
	suite.Equal(time.Minute*5, suite.config.AuthenticationBackend.RefreshInterval.Value())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainEnclosingParenthesis() {
	suite.config.AuthenticationBackend.LDAP.UsersFilter = "{username_attribute}={input}"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain enclosing parenthesis: '{username_attribute}={input}' should probably be '({username_attribute}={input})'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenGroupsFilterDoesNotContainEnclosingParenthesis() {
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = "cn={input}"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'groups_filter' must contain enclosing parenthesis: 'cn={input}' should probably be '(cn={input})'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainUsernameAttribute() {
	suite.config.AuthenticationBackend.LDAP.UsersFilter = "(&({mail_attribute}={input})(objectClass=person))"
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{username_attribute}' but it's absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldHelpDetectNoInputPlaceholder() {
	suite.config.AuthenticationBackend.LDAP.UsersFilter = "(&({username_attribute}={mail_attribute})(objectClass=person))"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{input}' but it's absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultTLSMinimumVersion() {
	suite.config.AuthenticationBackend.LDAP.TLS = &schema.TLS{MinimumVersion: schema.TLSVersion{}}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.Equal(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.TLS.MinimumVersion.Value, suite.config.AuthenticationBackend.LDAP.TLS.MinimumVersion.MinVersion())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotAllowSSL30() {
	suite.config.AuthenticationBackend.LDAP.TLS = &schema.TLS{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldErrorOnBadSearchMode() {
	suite.config.AuthenticationBackend.LDAP.GroupSearchMode = "memberOF"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'group_search_mode' must be one of 'filter' or 'memberof' but it's configured as 'memberOF'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNoErrorOnPlaceholderSearchMode() {
	suite.config.AuthenticationBackend.LDAP.GroupSearchMode = memberof
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = filterMemberOfRDN
	suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf = memberOf

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldErrorOnMissingPlaceholderSearchMode() {
	suite.config.AuthenticationBackend.LDAP.GroupSearchMode = memberof

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'groups_filter' must contain one of the '{memberof:rdn}' or '{memberof:dn}' placeholders when using a group_search_mode of 'memberof' but they're absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldErrorOnMissingDistinguishedNameDN() {
	suite.config.AuthenticationBackend.LDAP.Attributes.DistinguishedName = ""
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = "(|({memberof:dn}))"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: attributes: option 'distinguished_name' must be provided when using the '{memberof:dn}' placeholder but it's absent")
	suite.EqualError(suite.validator.Errors()[1], "authentication_backend: ldap: attributes: option 'member_of' must be provided when using the '{memberof:rdn}' or '{memberof:dn}' placeholder but it's absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldErrorOnMissingMemberOfRDN() {
	suite.config.AuthenticationBackend.LDAP.Attributes.DistinguishedName = ""
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = filterMemberOfRDN

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: attributes: option 'member_of' must be provided when using the '{memberof:rdn}' or '{memberof:dn}' placeholder but it's absent")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotAllowTLSVerMinGreaterThanVerMax() {
	suite.config.AuthenticationBackend.LDAP.TLS = &schema.TLS{
		MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
		MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS 1.3 is greater than the maximum version TLS 1.2")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateExtraAttributeString() {
	suite.config.AuthenticationBackend.LDAP.Attributes.Extra = map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
		"custom_attr": {AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "string"}},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateExtraAttributeInteger() {
	suite.config.AuthenticationBackend.LDAP.Attributes.Extra = map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
		"custom_int": {AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "integer"}},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateExtraAttributeBoolean() {
	suite.config.AuthenticationBackend.LDAP.Attributes.Extra = map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
		"custom_bool": {AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "boolean"}},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenExtraAttributeValueTypeMissing() {
	suite.config.AuthenticationBackend.LDAP.Attributes.Extra = map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
		"custom_attr": {AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: ""}},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: attributes: extra: custom_attr: option 'value_type' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenExtraAttributeValueTypeInvalid() {
	suite.config.AuthenticationBackend.LDAP.Attributes.Extra = map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
		"custom_attr": {AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "float"}},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: attributes: extra: custom_attr: option 'value_type' must be one of 'string', 'integer', or 'boolean' but it's configured as 'float'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenExtraAttributeNameReserved() {
	suite.config.AuthenticationBackend.LDAP.Attributes.Extra = map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
		"username": {AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "string"}},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: attributes: extra: username: attribute name 'username' is reserved")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenExtraAttributeCustomNameReserved() {
	suite.config.AuthenticationBackend.LDAP.Attributes.Extra = map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
		"custom": {
			Name: "email",

			AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "string"},
		},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: attributes: extra: custom: attribute name 'email' is reserved")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateExtraAttributeWithCustomName() {
	suite.config.AuthenticationBackend.LDAP.Attributes.Extra = map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
		"custom": {
			Name:                                "myattr",
			AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "string"},
		},
	}

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)
}

func TestLDAPAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(LDAPAuthenticationBackendSuite))
}

type ActiveDirectoryAuthenticationBackendSuite struct {
	LDAPImplementationSuite
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.Configuration{}
	suite.config.AuthenticationBackend.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.AuthenticationBackend.LDAP.Implementation = schema.LDAPImplementationActiveDirectory
	suite.config.AuthenticationBackend.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.AuthenticationBackend.LDAP.User = testLDAPUser
	suite.config.AuthenticationBackend.LDAP.Password = testLDAPPassword
	suite.config.AuthenticationBackend.LDAP.BaseDN = testLDAPBaseDN
	suite.config.AuthenticationBackend.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory.TLS
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldSetActiveDirectoryDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory)
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.AuthenticationBackend.LDAP.Timeout = time.Second * 2
	suite.config.AuthenticationBackend.LDAP.UsersFilter = "(&({username_attribute}={input})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.config.AuthenticationBackend.LDAP.Attributes.Username = "cn"
	suite.config.AuthenticationBackend.LDAP.Attributes.Mail = "userPrincipalName"
	suite.config.AuthenticationBackend.LDAP.Attributes.DisplayName = "name"
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = "(&(member={dn})(objectClass=group)(objectCategory=group))"
	suite.config.AuthenticationBackend.LDAP.Attributes.GroupName = "distinguishedName"
	suite.config.AuthenticationBackend.LDAP.AdditionalUsersDN = "OU=test"
	suite.config.AuthenticationBackend.LDAP.AdditionalGroupsDN = "OU=grps"
	suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf = member
	suite.config.AuthenticationBackend.LDAP.GroupSearchMode = memberof
	suite.config.AuthenticationBackend.LDAP.Attributes.DistinguishedName = "objectGUID"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.NotEqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory)

	suite.Equal(member, suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf)
	suite.Equal("objectGUID", suite.config.AuthenticationBackend.LDAP.Attributes.DistinguishedName)
	suite.Equal(memberof, suite.config.AuthenticationBackend.LDAP.GroupSearchMode)
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldRaiseErrorOnInvalidURLWithHTTP() {
	suite.config.AuthenticationBackend.LDAP.Address = &schema.AddressLDAP{Address: MustParseAddress("http://dc1:389")}

	validateLDAPAuthenticationAddress(suite.config.AuthenticationBackend.LDAP, suite.validator)

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
	suite.config = schema.Configuration{}
	suite.config.AuthenticationBackend.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.AuthenticationBackend.LDAP.Implementation = schema.LDAPImplementationRFC2307bis
	suite.config.AuthenticationBackend.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.AuthenticationBackend.LDAP.User = testLDAPUser
	suite.config.AuthenticationBackend.LDAP.Password = testLDAPPassword
	suite.config.AuthenticationBackend.LDAP.BaseDN = testLDAPBaseDN
	suite.config.AuthenticationBackend.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis.TLS
}

func (suite *RFC2307bisAuthenticationBackendSuite) TestShouldSetDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis)
}

func (suite *RFC2307bisAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.AuthenticationBackend.LDAP.Timeout = time.Second * 2
	suite.config.AuthenticationBackend.LDAP.UsersFilter = "(&({username_attribute}={input})(objectClass=Person))"
	suite.config.AuthenticationBackend.LDAP.Attributes.Username = "o"
	suite.config.AuthenticationBackend.LDAP.Attributes.Mail = "Email"
	suite.config.AuthenticationBackend.LDAP.Attributes.DisplayName = "Given"
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = "(&(member={dn})(objectClass=posixGroup)(objectClass=top))"
	suite.config.AuthenticationBackend.LDAP.Attributes.GroupName = "gid"
	suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf = member
	suite.config.AuthenticationBackend.LDAP.AdditionalUsersDN = "OU=users,OU=OpenLDAP"
	suite.config.AuthenticationBackend.LDAP.AdditionalGroupsDN = "OU=groups,OU=OpenLDAP"
	suite.config.AuthenticationBackend.LDAP.GroupSearchMode = memberof

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.NotEqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis)

	suite.Equal(member, suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf)
	suite.Equal("", suite.config.AuthenticationBackend.LDAP.Attributes.DistinguishedName)
	suite.Equal(schema.LDAPGroupSearchModeMemberOf, suite.config.AuthenticationBackend.LDAP.GroupSearchMode)
}

func TestRFC2307bisAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(RFC2307bisAuthenticationBackendSuite))
}

type FreeIPAAuthenticationBackendSuite struct {
	LDAPImplementationSuite
}

func (suite *FreeIPAAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.Configuration{}
	suite.config.AuthenticationBackend.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.AuthenticationBackend.LDAP.Implementation = schema.LDAPImplementationFreeIPA
	suite.config.AuthenticationBackend.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.AuthenticationBackend.LDAP.User = testLDAPUser
	suite.config.AuthenticationBackend.LDAP.Password = testLDAPPassword
	suite.config.AuthenticationBackend.LDAP.BaseDN = testLDAPBaseDN
	suite.config.AuthenticationBackend.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA.TLS
}

func (suite *FreeIPAAuthenticationBackendSuite) TestShouldSetDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA)
}

func (suite *FreeIPAAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.AuthenticationBackend.LDAP.Timeout = time.Second * 2
	suite.config.AuthenticationBackend.LDAP.UsersFilter = "(&({username_attribute}={input})(objectClass=person)(!(nsAccountLock=TRUE)))"
	suite.config.AuthenticationBackend.LDAP.Attributes.Username = "dn"
	suite.config.AuthenticationBackend.LDAP.Attributes.Mail = "email"
	suite.config.AuthenticationBackend.LDAP.Attributes.DisplayName = "gecos"
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = "(&(member={dn})(objectClass=posixgroup))"
	suite.config.AuthenticationBackend.LDAP.GroupSearchMode = schema.LDAPGroupSearchModeMemberOf
	suite.config.AuthenticationBackend.LDAP.Attributes.GroupName = "groupName"
	suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf = member
	suite.config.AuthenticationBackend.LDAP.AdditionalUsersDN = "OU=people"
	suite.config.AuthenticationBackend.LDAP.AdditionalGroupsDN = "OU=grp"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.NotEqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA)

	suite.Equal(member, suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf)
	suite.Equal("", suite.config.AuthenticationBackend.LDAP.Attributes.DistinguishedName)
	suite.Equal(schema.LDAPGroupSearchModeMemberOf, suite.config.AuthenticationBackend.LDAP.GroupSearchMode)
}

func TestFreeIPAAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(FreeIPAAuthenticationBackendSuite))
}

type LLDAPAuthenticationBackendSuite struct {
	LDAPImplementationSuite
}

func (suite *LLDAPAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.Configuration{}
	suite.config.AuthenticationBackend.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.AuthenticationBackend.LDAP.Implementation = schema.LDAPImplementationLLDAP
	suite.config.AuthenticationBackend.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.AuthenticationBackend.LDAP.User = testLDAPUser
	suite.config.AuthenticationBackend.LDAP.Password = testLDAPPassword
	suite.config.AuthenticationBackend.LDAP.BaseDN = testLDAPBaseDN
	suite.config.AuthenticationBackend.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP.TLS
}

func (suite *LLDAPAuthenticationBackendSuite) TestShouldSetDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP)
}

func (suite *LLDAPAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.AuthenticationBackend.LDAP.Timeout = time.Second * 2
	suite.config.AuthenticationBackend.LDAP.UsersFilter = "(&({username_attribute}={input})(objectClass=Person)(!(nsAccountLock=TRUE)))"
	suite.config.AuthenticationBackend.LDAP.Attributes.Username = "username"
	suite.config.AuthenticationBackend.LDAP.Attributes.Mail = "m"
	suite.config.AuthenticationBackend.LDAP.Attributes.DisplayName = "fn"
	suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf = member
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = "(&(member={dn})(!(objectClass=posixGroup)))"
	suite.config.AuthenticationBackend.LDAP.Attributes.GroupName = "grpz"
	suite.config.AuthenticationBackend.LDAP.AdditionalUsersDN = "OU=no"
	suite.config.AuthenticationBackend.LDAP.AdditionalGroupsDN = "OU=yes"
	suite.config.AuthenticationBackend.LDAP.GroupSearchMode = memberof

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.NotEqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP)

	suite.Equal(member, suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf)
	suite.Equal("", suite.config.AuthenticationBackend.LDAP.Attributes.DistinguishedName)
	suite.Equal(schema.LDAPGroupSearchModeMemberOf, suite.config.AuthenticationBackend.LDAP.GroupSearchMode)
}

func TestLLDAPAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(LLDAPAuthenticationBackendSuite))
}

type GLAuthAuthenticationBackendSuite struct {
	LDAPImplementationSuite
}

func (suite *GLAuthAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.Configuration{}
	suite.config.AuthenticationBackend.LDAP = &schema.AuthenticationBackendLDAP{}
	suite.config.AuthenticationBackend.LDAP.Implementation = schema.LDAPImplementationGLAuth
	suite.config.AuthenticationBackend.LDAP.Address = &schema.AddressLDAP{Address: *testLDAPAddress}
	suite.config.AuthenticationBackend.LDAP.User = testLDAPUser
	suite.config.AuthenticationBackend.LDAP.Password = testLDAPPassword
	suite.config.AuthenticationBackend.LDAP.BaseDN = testLDAPBaseDN
	suite.config.AuthenticationBackend.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth.TLS
}

func (suite *GLAuthAuthenticationBackendSuite) TestShouldSetDefaults() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Len(suite.validator.Warnings(), 0)
	suite.Len(suite.validator.Errors(), 0)

	suite.EqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth)
}

func (suite *GLAuthAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.config.AuthenticationBackend.LDAP.Timeout = time.Second * 2
	suite.config.AuthenticationBackend.LDAP.UsersFilter = "(&({username_attribute}={input})(objectClass=Person)(!(accountStatus=inactive)))"
	suite.config.AuthenticationBackend.LDAP.Attributes.Username = "description"
	suite.config.AuthenticationBackend.LDAP.Attributes.Mail = "sender"
	suite.config.AuthenticationBackend.LDAP.Attributes.DisplayName = "given"
	suite.config.AuthenticationBackend.LDAP.GroupsFilter = "(&(member={dn})(objectClass=posixGroup))"
	suite.config.AuthenticationBackend.LDAP.Attributes.GroupName = "grp"
	suite.config.AuthenticationBackend.LDAP.AdditionalUsersDN = "OU=users,OU=GlAuth"
	suite.config.AuthenticationBackend.LDAP.AdditionalGroupsDN = "OU=groups,OU=GLAuth"
	suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf = member
	suite.config.AuthenticationBackend.LDAP.GroupSearchMode = memberof

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.NotEqualImplementationDefaults(schema.DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth)

	suite.Equal(member, suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf)
	suite.Equal("", suite.config.AuthenticationBackend.LDAP.Attributes.DistinguishedName)
	suite.Equal(schema.LDAPGroupSearchModeMemberOf, suite.config.AuthenticationBackend.LDAP.GroupSearchMode)
}

func TestGLAuthAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(GLAuthAuthenticationBackendSuite))
}

type LDAPImplementationSuite struct {
	suite.Suite
	config    schema.Configuration
	validator *schema.StructValidator
}

func (suite *LDAPImplementationSuite) EqualImplementationDefaults(expected schema.AuthenticationBackendLDAP) {
	suite.Equal(expected.Timeout, suite.config.AuthenticationBackend.LDAP.Timeout)
	suite.Equal(expected.AdditionalUsersDN, suite.config.AuthenticationBackend.LDAP.AdditionalUsersDN)
	suite.Equal(expected.AdditionalGroupsDN, suite.config.AuthenticationBackend.LDAP.AdditionalGroupsDN)
	suite.Equal(expected.UsersFilter, suite.config.AuthenticationBackend.LDAP.UsersFilter)
	suite.Equal(expected.GroupsFilter, suite.config.AuthenticationBackend.LDAP.GroupsFilter)
	suite.Equal(expected.GroupSearchMode, suite.config.AuthenticationBackend.LDAP.GroupSearchMode)

	suite.Equal(expected.Attributes.DistinguishedName, suite.config.AuthenticationBackend.LDAP.Attributes.DistinguishedName)
	suite.Equal(expected.Attributes.Username, suite.config.AuthenticationBackend.LDAP.Attributes.Username)
	suite.Equal(expected.Attributes.DisplayName, suite.config.AuthenticationBackend.LDAP.Attributes.DisplayName)
	suite.Equal(expected.Attributes.Mail, suite.config.AuthenticationBackend.LDAP.Attributes.Mail)
	suite.Equal(expected.Attributes.MemberOf, suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf)
	suite.Equal(expected.Attributes.GroupName, suite.config.AuthenticationBackend.LDAP.Attributes.GroupName)
}

func (suite *LDAPImplementationSuite) NotEqualImplementationDefaults(expected schema.AuthenticationBackendLDAP) {
	suite.NotEqual(expected.Timeout, suite.config.AuthenticationBackend.LDAP.Timeout)
	suite.NotEqual(expected.UsersFilter, suite.config.AuthenticationBackend.LDAP.UsersFilter)
	suite.NotEqual(expected.GroupsFilter, suite.config.AuthenticationBackend.LDAP.GroupsFilter)
	suite.NotEqual(expected.GroupSearchMode, suite.config.AuthenticationBackend.LDAP.GroupSearchMode)
	suite.NotEqual(expected.Attributes.Username, suite.config.AuthenticationBackend.LDAP.Attributes.Username)
	suite.NotEqual(expected.Attributes.DisplayName, suite.config.AuthenticationBackend.LDAP.Attributes.DisplayName)
	suite.NotEqual(expected.Attributes.Mail, suite.config.AuthenticationBackend.LDAP.Attributes.Mail)
	suite.NotEqual(expected.Attributes.GroupName, suite.config.AuthenticationBackend.LDAP.Attributes.GroupName)

	if expected.Attributes.DistinguishedName != "" {
		suite.NotEqual(expected.Attributes.DistinguishedName, suite.config.AuthenticationBackend.LDAP.Attributes.DistinguishedName)
	}

	if expected.AdditionalUsersDN != "" {
		suite.NotEqual(expected.AdditionalUsersDN, suite.config.AuthenticationBackend.LDAP.AdditionalUsersDN)
	}

	if expected.AdditionalGroupsDN != "" {
		suite.NotEqual(expected.AdditionalGroupsDN, suite.config.AuthenticationBackend.LDAP.AdditionalGroupsDN)
	}

	if expected.Attributes.MemberOf != "" {
		suite.NotEqual(expected.Attributes.MemberOf, suite.config.AuthenticationBackend.LDAP.Attributes.MemberOf)
	}
}

func TestLDAPUserManagementRequiredAttributesValidation(t *testing.T) {
	testCases := []struct {
		name                        string
		authenticationBackendConfig *schema.AuthenticationBackend
		expectedErrors              []string
	}{
		{
			name: "ShouldPassWithValidRequiredAttributes",
			authenticationBackendConfig: &schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Address:      mustParseAddress("ldap://127.0.0.1"),
					User:         "cn=admin,dc=example,dc=com",
					Password:     "password",
					UsersFilter:  "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
					GroupsFilter: "(member={dn})",
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						Mail:       "mail",
						GivenName:  "givenName",
						FamilyName: "sn",
					},
					UserManagement: schema.AuthenticationBackendLDAPUserManagement{
						RequiredAttributes: []string{"mail", "given_name", "family_name"},
					},
				},
			},
			expectedErrors: nil,
		},
		{
			name: "ShouldFailWithUnsupportedRequiredAttribute",
			authenticationBackendConfig: &schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Address:      mustParseAddress("ldap://127.0.0.1"),
					User:         "cn=admin,dc=example,dc=com",
					Password:     "password",
					UsersFilter:  "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
					GroupsFilter: "(member={dn})",
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						Mail: "mail",
					},
					UserManagement: schema.AuthenticationBackendLDAPUserManagement{
						RequiredAttributes: []string{"mail", "phone_number"},
					},
				},
			},
			expectedErrors: []string{
				"authentication_backend: ldap: user_management: option 'required_attributes' contains the attribute 'phone_number' which is not a supported attribute: supported attributes are determined by the LDAP attribute mappings and extra attributes configured",
			},
		},
		{
			name: "ShouldPassWithExtraAttributes",
			authenticationBackendConfig: &schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Address:      mustParseAddress("ldap://127.0.0.1"),
					User:         "cn=admin,dc=example,dc=com",
					Password:     "password",
					UsersFilter:  "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
					GroupsFilter: "(member={dn})",
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						Mail: "mail",
						Extra: map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
							"employee_id": {
								Name: "employeeNumber",
								AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
									ValueType: "string",
								},
							},
						},
					},
					UserManagement: schema.AuthenticationBackendLDAPUserManagement{
						RequiredAttributes: []string{"mail", "employee_id"},
					},
				},
			},
			expectedErrors: nil,
		},
		{
			name: "ShouldPassWithAddressAttributes",
			authenticationBackendConfig: &schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Address:      mustParseAddress("ldap://127.0.0.1"),
					User:         "cn=admin,dc=example,dc=com",
					Password:     "password",
					UsersFilter:  "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
					GroupsFilter: "(member={dn})",
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						Mail:          "mail",
						StreetAddress: "streetAddress",
						Locality:      "l",
					},
					UserManagement: schema.AuthenticationBackendLDAPUserManagement{
						RequiredAttributes: []string{"address", "address.street_address", "address.locality"},
					},
				},
			},
			expectedErrors: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := schema.Configuration{
				Administration: schema.Administration{
					Enabled:              true,
					EnableUserManagement: true,
				},
				AuthenticationBackend: *tc.authenticationBackendConfig,
			}

			ValidateAuthenticationBackend(&config, validator)

			if tc.expectedErrors == nil {
				for _, err := range validator.Errors() {
					assert.NotContains(t, err.Error(), "user_management")
				}
			} else {
				var userMgmtErrors []string

				for _, err := range validator.Errors() {
					if errStr := err.Error(); len(errStr) > 0 && len(tc.expectedErrors) > 0 {
						for _, expectedErr := range tc.expectedErrors {
							if errStr == expectedErr {
								userMgmtErrors = append(userMgmtErrors, errStr)
							}
						}
					}
				}

				assert.Equal(t, tc.expectedErrors, userMgmtErrors)
			}
		})
	}
}

//nolint:unparam
func mustParseAddress(uri string) *schema.AddressLDAP {
	addr, err := schema.NewAddress(uri)
	if err != nil {
		panic(err)
	}

	return &schema.AddressLDAP{Address: *addr}
}

func TestLDAPUserManagementRDNTemplateValidation(t *testing.T) {
	testCases := []struct {
		name                        string
		authenticationBackendConfig *schema.AuthenticationBackend
		expectedErrors              []string
	}{
		{
			name: "ShouldPassWithValidRDNTemplate",
			authenticationBackendConfig: &schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Address:      mustParseAddress("ldap://127.0.0.1"),
					User:         "cn=admin,dc=example,dc=com",
					Password:     "password",
					UsersFilter:  "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
					GroupsFilter: "(member={dn})",
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						GivenName:  "givenName",
						FamilyName: "sn",
					},
					UserManagement: schema.AuthenticationBackendLDAPUserManagement{
						CreatedUsersRDNFormat:    "[[ .given_name ]] [[ .family_name ]]",
						CreatedUsersRDNAttribute: "cn",
						RequiredAttributes:       []string{"given_name", "family_name"},
					},
				},
			},
			expectedErrors: nil,
		},
		{
			name: "ShouldFailWithInvalidTemplSyntax",
			authenticationBackendConfig: &schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Address:      mustParseAddress("ldap://127.0.0.1"),
					User:         "cn=admin,dc=example,dc=com",
					Password:     "password",
					UsersFilter:  "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
					GroupsFilter: "(member={dn})",
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						GivenName: "givenName",
					},
					UserManagement: schema.AuthenticationBackendLDAPUserManagement{
						CreatedUsersRDNFormat: "[[ .given_name ",
					},
				},
			},
			expectedErrors: []string{
				"authentication_backend: ldap: user_management: option 'created_users_rdn_format' is invalid:",
			},
		},
		{
			name: "ShouldFailWithUnsupportedField",
			authenticationBackendConfig: &schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Address:      mustParseAddress("ldap://127.0.0.1"),
					User:         "cn=admin,dc=example,dc=com",
					Password:     "password",
					UsersFilter:  "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
					GroupsFilter: "(member={dn})",
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						GivenName: "givenName",
					},
					UserManagement: schema.AuthenticationBackendLDAPUserManagement{
						CreatedUsersRDNFormat:    "[[ .given_name ]] [[ .phone_number ]]",
						CreatedUsersRDNAttribute: "cn",
					},
				},
			},
			expectedErrors: []string{
				"authentication_backend: ldap: user_management: option 'created_users_rdn_format' references field 'phone_number' which is not a supported attribute: ensure the attribute is mapped in the LDAP configuration",
			},
		},
		{
			name: "ShouldPassWithEmptyTemplate",
			authenticationBackendConfig: &schema.AuthenticationBackend{
				LDAP: &schema.AuthenticationBackendLDAP{
					Address:      mustParseAddress("ldap://127.0.0.1"),
					User:         "cn=admin,dc=example,dc=com",
					Password:     "password",
					UsersFilter:  "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
					GroupsFilter: "(member={dn})",
					Attributes: schema.AuthenticationBackendLDAPAttributes{
						GivenName: "givenName",
					},
					UserManagement: schema.AuthenticationBackendLDAPUserManagement{
						CreatedUsersRDNFormat: "",
					},
				},
			},
			expectedErrors: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := schema.Configuration{
				Administration: schema.Administration{
					Enabled:              true,
					EnableUserManagement: true,
				},
				AuthenticationBackend: *tc.authenticationBackendConfig,
			}
			ValidateAuthenticationBackend(&config, validator)

			if tc.expectedErrors == nil {
				for _, err := range validator.Errors() {
					assert.NotContains(t, err.Error(), "created_users_rdn_format")
				}
			} else {
				var rdnErrors []string

				for _, err := range validator.Errors() {
					errStr := err.Error()
					for _, expectedErr := range tc.expectedErrors {
						if strings.Contains(errStr, expectedErr) || strings.HasPrefix(errStr, expectedErr) {
							rdnErrors = append(rdnErrors, errStr)
							break
						}
					}
				}

				assert.Len(t, rdnErrors, len(tc.expectedErrors), "Expected %d RDN template errors but got %d: %v", len(tc.expectedErrors), len(rdnErrors), rdnErrors)
			}
		})
	}
}
