package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
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
	suite.configuration.File = &schema.FileAuthenticationBackendConfiguration{Path: "/a/path", Password: &schema.PasswordConfiguration{
		Algorithm:   schema.DefaultPasswordConfiguration.Algorithm,
		Iterations:  schema.DefaultPasswordConfiguration.Iterations,
		Parallelism: schema.DefaultPasswordConfiguration.Parallelism,
		Memory:      schema.DefaultPasswordConfiguration.Memory,
		KeyLength:   schema.DefaultPasswordConfiguration.KeyLength,
		SaltLength:  schema.DefaultPasswordConfiguration.SaltLength,
	}}
	suite.configuration.File.Password.Algorithm = schema.DefaultPasswordConfiguration.Algorithm
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
	suite.configuration.File.Password.Memory = 8
	suite.configuration.File.Password.Parallelism = 2
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Memory for argon2id must be 16 or more (parallelism * 8), you configured memory as 8 and parallelism as 2")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenBlank() {
	suite.configuration.File.Password = &schema.PasswordConfiguration{}

	assert.Equal(suite.T(), 0, suite.configuration.File.Password.KeyLength)
	assert.Equal(suite.T(), 0, suite.configuration.File.Password.Iterations)
	assert.Equal(suite.T(), 0, suite.configuration.File.Password.SaltLength)
	assert.Equal(suite.T(), "", suite.configuration.File.Password.Algorithm)
	assert.Equal(suite.T(), 0, suite.configuration.File.Password.Memory)
	assert.Equal(suite.T(), 0, suite.configuration.File.Password.Parallelism)

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.KeyLength, suite.configuration.File.Password.KeyLength)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.Iterations, suite.configuration.File.Password.Iterations)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.SaltLength, suite.configuration.File.Password.SaltLength)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.Algorithm, suite.configuration.File.Password.Algorithm)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.Memory, suite.configuration.File.Password.Memory)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.Parallelism, suite.configuration.File.Password.Parallelism)
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenOnlySHA512Set() {
	suite.configuration.File.Password = &schema.PasswordConfiguration{}
	assert.Equal(suite.T(), "", suite.configuration.File.Password.Algorithm)
	suite.configuration.File.Password.Algorithm = "sha512"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), schema.DefaultPasswordSHA512Configuration.KeyLength, suite.configuration.File.Password.KeyLength)
	assert.Equal(suite.T(), schema.DefaultPasswordSHA512Configuration.Iterations, suite.configuration.File.Password.Iterations)
	assert.Equal(suite.T(), schema.DefaultPasswordSHA512Configuration.SaltLength, suite.configuration.File.Password.SaltLength)
	assert.Equal(suite.T(), schema.DefaultPasswordSHA512Configuration.Algorithm, suite.configuration.File.Password.Algorithm)
	assert.Equal(suite.T(), schema.DefaultPasswordSHA512Configuration.Memory, suite.configuration.File.Password.Memory)
	assert.Equal(suite.T(), schema.DefaultPasswordSHA512Configuration.Parallelism, suite.configuration.File.Password.Parallelism)
}
func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenKeyLengthTooLow() {
	suite.configuration.File.Password.KeyLength = 1
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Key length for argon2id must be 16, you configured 1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSaltLengthTooLow() {
	suite.configuration.File.Password.SaltLength = -1
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "The salt length must be 2 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBadAlgorithmDefined() {
	suite.configuration.File.Password.Algorithm = "bogus"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Unknown hashing algorithm supplied, valid values are argon2id and sha512, you configured 'bogus'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenIterationsTooLow() {
	suite.configuration.File.Password.Iterations = -1
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "The number of iterations specified is invalid, must be 1 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenParallelismTooLow() {
	suite.configuration.File.Password.Parallelism = -1
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Parallelism for argon2id must be 1 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultValues() {
	suite.configuration.File.Password.Algorithm = ""
	suite.configuration.File.Password.Iterations = 0
	suite.configuration.File.Password.SaltLength = 0
	suite.configuration.File.Password.Memory = 0
	suite.configuration.File.Password.Parallelism = 0
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.Algorithm, suite.configuration.File.Password.Algorithm)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.Iterations, suite.configuration.File.Password.Iterations)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.SaltLength, suite.configuration.File.Password.SaltLength)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.Memory, suite.configuration.File.Password.Memory)
	assert.Equal(suite.T(), schema.DefaultPasswordConfiguration.Parallelism, suite.configuration.File.Password.Parallelism)
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
	suite.configuration.Ldap.Implementation = "rfc"
	suite.configuration.Ldap.URL = "ldap://ldap"
	suite.configuration.Ldap.User = "user"
	suite.configuration.Ldap.Password = "password"
	suite.configuration.Ldap.BaseDN = "base_dn"
	suite.configuration.Ldap.UsernameAttribute = "uid"
	suite.configuration.Ldap.UsersFilter = "(uid={input})"
	suite.configuration.Ldap.GroupsFilter = "(cn={input})"
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

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseOnEmptyGroupsFilter() {
	suite.configuration.Ldap.GroupsFilter = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	require.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Please provide a groups filter with `groups_filter` attribute")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseOnEmptyUsersFilter() {
	suite.configuration.Ldap.UsersFilter = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	require.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Please provide a users filter with `users_filter` attribute")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseOnEmptyUsernameAttribute() {
	suite.configuration.Ldap.UsernameAttribute = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	require.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Please provide a username attribute with `username_attribute`")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseOnBadRefreshInterval() {
	suite.configuration.RefreshInterval = "blah"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	require.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Auth Backend `refresh_interval` is configured to 'blah' but it must be either a duration notation or one of 'disable', or 'always'. Error from parser: Could not convert the input string of blah into a duration")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultImplementation() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), "rfc", suite.configuration.Ldap.Implementation)
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

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultDisplayNameAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), "displayname", suite.configuration.Ldap.DisplayNameAttribute)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultRefreshInterval() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), "5m", suite.configuration.RefreshInterval)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainEnclosingParenthesis() {
	suite.configuration.Ldap.UsersFilter = "uid={input}"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "The users filter should contain enclosing parenthesis. For instance uid={input} should be (uid={input})")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseWhenGroupsFilterDoesNotContainEnclosingParenthesis() {
	suite.configuration.Ldap.GroupsFilter = "cn={input}"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "The groups filter should contain enclosing parenthesis. For instance cn={input} should be (cn={input})")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldHelpDetectNoInputPlaceholder() {
	suite.configuration.Ldap.UsersFilter = "(objectClass=person)"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 1)
	assert.EqualError(suite.T(), suite.validator.Errors()[0], "Unable to detect {input} placeholder in users_filter, your configuration might be broken. Please review configuration options listed at https://docs.authelia.com/configuration/authentication/ldap.html")
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
