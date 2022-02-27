package validator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldRaiseErrorWhenBothBackendsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	backendConfig := schema.AuthenticationBackendConfiguration{}

	backendConfig.LDAP = &schema.LDAPAuthenticationBackendConfiguration{}
	backendConfig.File = &schema.FileAuthenticationBackendConfiguration{
		Path: "/tmp",
	}

	ValidateAuthenticationBackend(&backendConfig, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "authentication_backend: please ensure only one of the 'file' or 'ldap' backend is configured")
}

func TestShouldRaiseErrorWhenNoBackendProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	backendConfig := schema.AuthenticationBackendConfiguration{}

	ValidateAuthenticationBackend(&backendConfig, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "authentication_backend: you must ensure either the 'file' or 'ldap' authentication backend is configured")
}

type FileBasedAuthenticationBackend struct {
	suite.Suite
	config    schema.AuthenticationBackendConfiguration
	validator *schema.StructValidator
}

func (suite *FileBasedAuthenticationBackend) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = schema.AuthenticationBackendConfiguration{}
	suite.config.File = &schema.FileAuthenticationBackendConfiguration{Path: "/a/path", Password: &schema.PasswordConfiguration{
		Algorithm:   schema.DefaultPasswordConfiguration.Algorithm,
		Iterations:  schema.DefaultPasswordConfiguration.Iterations,
		Parallelism: schema.DefaultPasswordConfiguration.Parallelism,
		Memory:      schema.DefaultPasswordConfiguration.Memory,
		KeyLength:   schema.DefaultPasswordConfiguration.KeyLength,
		SaltLength:  schema.DefaultPasswordConfiguration.SaltLength,
	}}
	suite.config.File.Password.Algorithm = schema.DefaultPasswordConfiguration.Algorithm
}
func (suite *FileBasedAuthenticationBackend) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenNoPathProvided() {
	suite.config.File.Path = ""

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: option 'path' is required")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenMemoryNotMoreThanEightTimesParallelism() {
	suite.config.File.Password.Memory = 8
	suite.config.File.Password.Parallelism = 2

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: option 'memory' must at least be parallelism multiplied by 8 when using algorithm 'argon2id' with parallelism 2 it should be at least 16 but it is configured as '8'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenBlank() {
	suite.config.File.Password = &schema.PasswordConfiguration{}

	suite.Assert().Equal(0, suite.config.File.Password.KeyLength)
	suite.Assert().Equal(0, suite.config.File.Password.Iterations)
	suite.Assert().Equal(0, suite.config.File.Password.SaltLength)
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)
	suite.Assert().Equal(0, suite.config.File.Password.Memory)
	suite.Assert().Equal(0, suite.config.File.Password.Parallelism)

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(schema.DefaultPasswordConfiguration.KeyLength, suite.config.File.Password.KeyLength)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Iterations, suite.config.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.SaltLength, suite.config.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Algorithm, suite.config.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Memory, suite.config.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Parallelism, suite.config.File.Password.Parallelism)
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenOnlySHA512Set() {
	suite.config.File.Password = &schema.PasswordConfiguration{}
	suite.Assert().Equal("", suite.config.File.Password.Algorithm)
	suite.config.File.Password.Algorithm = "sha512"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.KeyLength, suite.config.File.Password.KeyLength)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Iterations, suite.config.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.SaltLength, suite.config.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Algorithm, suite.config.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Memory, suite.config.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Parallelism, suite.config.File.Password.Parallelism)
}
func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenKeyLengthTooLow() {
	suite.config.File.Password.KeyLength = 1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: option 'key_length' must be 16 or more when using algorithm 'argon2id' but it is configured as '1'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSaltLengthTooLow() {
	suite.config.File.Password.SaltLength = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: option 'salt_length' must be 2 or more but it is configured a '-1'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBadAlgorithmDefined() {
	suite.config.File.Password.Algorithm = "bogus"

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: option 'algorithm' must be either 'argon2id' or 'sha512' but it is configured as 'bogus'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenIterationsTooLow() {
	suite.config.File.Password.Iterations = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: option 'iterations' must be 1 or more but it is configured as '-1'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenParallelismTooLow() {
	suite.config.File.Password.Parallelism = -1

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: file: password: option 'parallelism' must be 1 or more when using algorithm 'argon2id' but it is configured as '-1'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultValues() {
	suite.config.File.Password.Algorithm = ""
	suite.config.File.Password.Iterations = 0
	suite.config.File.Password.SaltLength = 0
	suite.config.File.Password.Memory = 0
	suite.config.File.Password.Parallelism = 0

	ValidateAuthenticationBackend(&suite.config, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Algorithm, suite.config.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Iterations, suite.config.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.SaltLength, suite.config.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Memory, suite.config.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Parallelism, suite.config.File.Password.Parallelism)
}

func TestFileBasedAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(FileBasedAuthenticationBackend))
}

type LDAPAuthenticationBackendSuite struct {
	suite.Suite
	configuration schema.AuthenticationBackendConfiguration
	validator     *schema.StructValidator
}

func (suite *LDAPAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration = schema.AuthenticationBackendConfiguration{}
	suite.configuration.LDAP = &schema.LDAPAuthenticationBackendConfiguration{}
	suite.configuration.LDAP.Implementation = schema.LDAPImplementationCustom
	suite.configuration.LDAP.URL = testLDAPURL
	suite.configuration.LDAP.User = testLDAPUser
	suite.configuration.LDAP.Password = testLDAPPassword
	suite.configuration.LDAP.BaseDN = testLDAPBaseDN
	suite.configuration.LDAP.UsernameAttribute = "uid"
	suite.configuration.LDAP.UsersFilter = "({username_attribute}={input})"
	suite.configuration.LDAP.GroupsFilter = "(cn={input})"
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateDefaultImplementationAndUsernameAttribute() {
	suite.configuration.LDAP.Implementation = ""
	suite.configuration.LDAP.UsernameAttribute = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().Equal(schema.LDAPImplementationCustom, suite.configuration.LDAP.Implementation)

	suite.Assert().Equal(suite.configuration.LDAP.UsernameAttribute, schema.DefaultLDAPAuthenticationBackendConfiguration.UsernameAttribute)
	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenImplementationIsInvalidMSAD() {
	suite.configuration.LDAP.Implementation = "masd"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'implementation' is configured as 'masd' but must be one of the following values: 'custom', 'activedirectory'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenURLNotProvided() {
	suite.configuration.LDAP.URL = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'url' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenUserNotProvided() {
	suite.configuration.LDAP.User = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'user' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenPasswordNotProvided() {
	suite.configuration.LDAP.Password = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'password' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenBaseDNNotProvided() {
	suite.configuration.LDAP.BaseDN = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'base_dn' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnEmptyGroupsFilter() {
	suite.configuration.LDAP.GroupsFilter = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'groups_filter' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnEmptyUsersFilter() {
	suite.configuration.LDAP.UsersFilter = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseOnEmptyUsernameAttribute() {
	suite.configuration.LDAP.UsernameAttribute = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnBadRefreshInterval() {
	suite.configuration.RefreshInterval = "blah"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: option 'refresh_interval' is configured to 'blah' but it must be either a duration notation or one of 'disable', or 'always': could not convert the input string of blah into a duration")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultImplementation() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.LDAPImplementationCustom, suite.configuration.LDAP.Implementation)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorOnBadFilterPlaceholders() {
	suite.configuration.LDAP.UsersFilter = "(&({username_attribute}={0})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.configuration.LDAP.GroupsFilter = "(&(member={0})(objectClass=group)(objectCategory=group))"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().True(suite.validator.HasErrors())

	suite.Require().Len(suite.validator.Errors(), 3)
	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' has an invalid placeholder: '{0}' has been removed, please use '{input}' instead")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication_backend: ldap: option 'groups_filter' has an invalid placeholder: '{0}' has been removed, please use '{input}' instead")
	suite.Assert().EqualError(suite.validator.Errors()[2], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{input}' but it is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultGroupNameAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("cn", suite.configuration.LDAP.GroupNameAttribute)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultMailAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("mail", suite.configuration.LDAP.MailAttribute)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultDisplayNameAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("displayName", suite.configuration.LDAP.DisplayNameAttribute)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultRefreshInterval() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("5m", suite.configuration.RefreshInterval)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainEnclosingParenthesis() {
	suite.configuration.LDAP.UsersFilter = "{username_attribute}={input}"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain enclosing parenthesis: '{username_attribute}={input}' should probably be '({username_attribute}={input})'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenGroupsFilterDoesNotContainEnclosingParenthesis() {
	suite.configuration.LDAP.GroupsFilter = "cn={input}"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'groups_filter' must contain enclosing parenthesis: 'cn={input}' should probably be '(cn={input})'")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainUsernameAttribute() {
	suite.configuration.LDAP.UsersFilter = "(&({mail_attribute}={input})(objectClass=person))"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{username_attribute}' but it is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldHelpDetectNoInputPlaceholder() {
	suite.configuration.LDAP.UsersFilter = "(&({username_attribute}={mail_attribute})(objectClass=person))"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: option 'users_filter' must contain the placeholder '{input}' but it is required")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultTLSMinimumVersion() {
	suite.configuration.LDAP.TLS = &schema.TLSConfig{MinimumVersion: ""}

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.DefaultLDAPAuthenticationBackendConfiguration.TLS.MinimumVersion, suite.configuration.LDAP.TLS.MinimumVersion)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotAllowInvalidTLSValue() {
	suite.configuration.LDAP.TLS = &schema.TLSConfig{
		MinimumVersion: "SSL2.0",
	}

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication_backend: ldap: tls: option 'minimum_tls_version' is invalid: SSL2.0: supplied tls version isn't supported")
}

func TestLdapAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(LDAPAuthenticationBackendSuite))
}

type ActiveDirectoryAuthenticationBackendSuite struct {
	suite.Suite
	configuration schema.AuthenticationBackendConfiguration
	validator     *schema.StructValidator
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration = schema.AuthenticationBackendConfiguration{}
	suite.configuration.LDAP = &schema.LDAPAuthenticationBackendConfiguration{}
	suite.configuration.LDAP.Implementation = schema.LDAPImplementationActiveDirectory
	suite.configuration.LDAP.URL = testLDAPURL
	suite.configuration.LDAP.User = testLDAPUser
	suite.configuration.LDAP.Password = testLDAPPassword
	suite.configuration.LDAP.BaseDN = testLDAPBaseDN
	suite.configuration.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldSetActiveDirectoryDefaults() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendConfiguration.Timeout,
		suite.configuration.LDAP.Timeout)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsersFilter,
		suite.configuration.LDAP.UsersFilter)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsernameAttribute,
		suite.configuration.LDAP.UsernameAttribute)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.DisplayNameAttribute,
		suite.configuration.LDAP.DisplayNameAttribute)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.MailAttribute,
		suite.configuration.LDAP.MailAttribute)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupsFilter,
		suite.configuration.LDAP.GroupsFilter)
	suite.Assert().Equal(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupNameAttribute,
		suite.configuration.LDAP.GroupNameAttribute)
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.configuration.LDAP.Timeout = time.Second * 2
	suite.configuration.LDAP.UsersFilter = "(&({username_attribute}={input})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.configuration.LDAP.UsernameAttribute = "cn"
	suite.configuration.LDAP.MailAttribute = "userPrincipalName"
	suite.configuration.LDAP.DisplayNameAttribute = "name"
	suite.configuration.LDAP.GroupsFilter = "(&(member={dn})(objectClass=group)(objectCategory=group))"
	suite.configuration.LDAP.GroupNameAttribute = "distinguishedName"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendConfiguration.Timeout,
		suite.configuration.LDAP.Timeout)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsersFilter,
		suite.configuration.LDAP.UsersFilter)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsernameAttribute,
		suite.configuration.LDAP.UsernameAttribute)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.DisplayNameAttribute,
		suite.configuration.LDAP.DisplayNameAttribute)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.MailAttribute,
		suite.configuration.LDAP.MailAttribute)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupsFilter,
		suite.configuration.LDAP.GroupsFilter)
	suite.Assert().NotEqual(
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupNameAttribute,
		suite.configuration.LDAP.GroupNameAttribute)
}

func TestActiveDirectoryAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(ActiveDirectoryAuthenticationBackendSuite))
}
