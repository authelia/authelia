package validator

import (
	"testing"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestShouldRaiseErrorsWhenNoBackendProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	backendConfig := schema.AuthenticationBackendConfiguration{}

	ValidateAuthenticationBackend(&backendConfig, validator)

	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Please provide `ldap` or `file` object in `authentication_backend`")
}

type FileBasedAuthenticationBackend struct {
	suite.Suite
	configuration schema.AuthenticationBackendConfiguration
	validator     *schema.StructValidator
}

func (suite *FileBasedAuthenticationBackend) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration = schema.AuthenticationBackendConfiguration{}
	suite.configuration.File = &schema.FileAuthenticationBackendConfiguration{Path: "/a/path", PasswordHashing: &schema.PasswordHashingConfiguration{
		Algorithm:   schema.DefaultPasswordOptionsConfiguration.Algorithm,
		Iterations:  schema.DefaultPasswordOptionsConfiguration.Iterations,
		Parallelism: schema.DefaultPasswordOptionsConfiguration.Parallelism,
		Memory:      schema.DefaultPasswordOptionsConfiguration.Memory,
		KeyLength:   schema.DefaultPasswordOptionsConfiguration.KeyLength,
		SaltLength:  schema.DefaultPasswordOptionsConfiguration.SaltLength,
	}}
	suite.configuration.File.PasswordHashing.Algorithm = schema.DefaultPasswordOptionsConfiguration.Algorithm
}

func (suite *FileBasedAuthenticationBackend) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenNoPathProvided() {
	suite.configuration.File.Path = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Please provide a `path` for the users database in `authentication_backend`")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenMemoryNotMoreThanEightTimesParallelism() {
	suite.configuration.File.PasswordHashing.Memory = 8
	suite.configuration.File.PasswordHashing.Parallelism = 2
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Memory for argon2id must be 16 or more (parallelism * 8), you configured memory as 8 and parallelism as 2")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenKeyLengthTooLow() {
	suite.configuration.File.PasswordHashing.KeyLength = 1
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Key length for argon2id must be 16, you configured 1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSaltLengthTooLow() {
	suite.configuration.File.PasswordHashing.SaltLength = -1
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "The salt length must 2 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBadAlgorithmDefined() {
	suite.configuration.File.PasswordHashing.Algorithm = "bogus"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Unknown hashing algorithm supplied, valid values are argon2id and sha512, you configured 'bogus'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenIterationsTooLow() {
	suite.configuration.File.PasswordHashing.Iterations = -1
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "The number of iterations specified is invalid, must be 1 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenParallelismTooLow() {
	suite.configuration.File.PasswordHashing.Parallelism = -1
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Parallelism for argon2id must be 1 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultValues() {
	suite.configuration.File.PasswordHashing.Algorithm = ""
	suite.configuration.File.PasswordHashing.Iterations = 0
	suite.configuration.File.PasswordHashing.SaltLength = 0
	suite.configuration.File.PasswordHashing.Memory = 0
	suite.configuration.File.PasswordHashing.Parallelism = 0
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), schema.DefaultPasswordOptionsConfiguration.Algorithm, suite.configuration.File.PasswordHashing.Algorithm)
	assert.Equal(suite.T(), schema.DefaultPasswordOptionsConfiguration.Iterations, suite.configuration.File.PasswordHashing.Iterations)
	assert.Equal(suite.T(), schema.DefaultPasswordOptionsConfiguration.SaltLength, suite.configuration.File.PasswordHashing.SaltLength)
	assert.Equal(suite.T(), schema.DefaultPasswordOptionsConfiguration.Memory, suite.configuration.File.PasswordHashing.Memory)
	assert.Equal(suite.T(), schema.DefaultPasswordOptionsConfiguration.Parallelism, suite.configuration.File.PasswordHashing.Parallelism)
}

func TestFileBasedAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(FileBasedAuthenticationBackend))
}

type LdapAuthenticationBackendSuite struct {
	suite.Suite
	configuration schema.AuthenticationBackendConfiguration
	validator     *schema.StructValidator
}

func (suite *LdapAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration = schema.AuthenticationBackendConfiguration{}
	suite.configuration.Ldap = &schema.LDAPAuthenticationBackendConfiguration{}
	suite.configuration.Ldap.URL = "ldap://ldap"
	suite.configuration.Ldap.User = "user"
	suite.configuration.Ldap.Password = "password"
	suite.configuration.Ldap.BaseDN = "base_dn"
}

func (suite *LdapAuthenticationBackendSuite) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseErrorWhenURLNotProvided() {
	suite.configuration.Ldap.URL = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Please provide a URL to the LDAP server")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseErrorWhenUserNotProvided() {
	suite.configuration.Ldap.User = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Please provide a user name to connect to the LDAP server")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseErrorWhenPasswordNotProvided() {
	suite.configuration.Ldap.Password = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Please provide a password to connect to the LDAP server")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseErrorWhenBaseDNNotProvided() {
	suite.configuration.Ldap.BaseDN = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Please provide a base DN to connect to the LDAP server")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultUsersFilter() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), "(cn={0})", suite.configuration.Ldap.UsersFilter)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultGroupsFilter() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), "(member={dn})", suite.configuration.Ldap.GroupsFilter)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultGroupNameAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), "cn", suite.configuration.Ldap.GroupNameAttribute)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultMailAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), "mail", suite.configuration.Ldap.MailAttribute)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainEnclosingParenthesis() {
	suite.configuration.Ldap.UsersFilter = "cn={0}"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "The users filter should contain enclosing parenthesis. For instance cn={0} should be (cn={0})")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseWhenGroupsFilterDoesNotContainEnclosingParenthesis() {
	suite.configuration.Ldap.UsersFilter = "cn={0}"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "The users filter should contain enclosing parenthesis. For instance cn={0} should be (cn={0})")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldAdaptLDAPURL() {
	assert.Equal(suite.T(), "", validateLdapURL("127.0.0.1", suite.validator))
	require.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Unknown scheme for ldap url, should be ldap:// or ldaps://")

	assert.Equal(suite.T(), "", validateLdapURL("127.0.0.1:636", suite.validator))
	require.Len(suite.T(), suite.validator.Errors(), 2)
	assert.EqualError(suite.T(), suite.validator.Errors()[1], "Unable to parse URL to ldap server. The scheme is probably missing: ldap:// or ldaps://")

	assert.Equal(suite.T(), "ldap://127.0.0.1:389", validateLdapURL("ldap://127.0.0.1", suite.validator))
	assert.Equal(suite.T(), "ldap://127.0.0.1:390", validateLdapURL("ldap://127.0.0.1:390", suite.validator))
	assert.Equal(suite.T(), "ldap://127.0.0.1:389/abc", validateLdapURL("ldap://127.0.0.1/abc", suite.validator))
	assert.Equal(suite.T(), "ldap://127.0.0.1:389/abc?test=abc&x=y", validateLdapURL("ldap://127.0.0.1/abc?test=abc&x=y", suite.validator))

	assert.Equal(suite.T(), "ldaps://127.0.0.1:390", validateLdapURL("ldaps://127.0.0.1:390", suite.validator))
	assert.Equal(suite.T(), "ldaps://127.0.0.1:636", validateLdapURL("ldaps://127.0.0.1", suite.validator))
}

func TestLdapAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(LdapAuthenticationBackendSuite))
}
