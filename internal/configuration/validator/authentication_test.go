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

	require.Len(t, validator.Errors(), 1)
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

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenNoPathProvided() {
	suite.configuration.File.Path = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a `path` for the users database in `authentication_backend`")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenMemoryNotMoreThanEightTimesParallelism() {
	suite.configuration.File.Password.Memory = 8
	suite.configuration.File.Password.Parallelism = 2

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Memory for argon2id must be 16 or more (parallelism * 8), you configured memory as 8 and parallelism as 2")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenBlank() {
	suite.configuration.File.Password = &schema.PasswordConfiguration{}

	suite.Assert().Equal(0, suite.configuration.File.Password.KeyLength)
	suite.Assert().Equal(0, suite.configuration.File.Password.Iterations)
	suite.Assert().Equal(0, suite.configuration.File.Password.SaltLength)
	suite.Assert().Equal("", suite.configuration.File.Password.Algorithm)
	suite.Assert().Equal(0, suite.configuration.File.Password.Memory)
	suite.Assert().Equal(0, suite.configuration.File.Password.Parallelism)

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.DefaultPasswordConfiguration.KeyLength, suite.configuration.File.Password.KeyLength)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Iterations, suite.configuration.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.SaltLength, suite.configuration.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Algorithm, suite.configuration.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Memory, suite.configuration.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Parallelism, suite.configuration.File.Password.Parallelism)
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenOnlySHA512Set() {
	suite.configuration.File.Password = &schema.PasswordConfiguration{}
	suite.Assert().Equal("", suite.configuration.File.Password.Algorithm)
	suite.configuration.File.Password.Algorithm = "sha512"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.KeyLength, suite.configuration.File.Password.KeyLength)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Iterations, suite.configuration.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.SaltLength, suite.configuration.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Algorithm, suite.configuration.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Memory, suite.configuration.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Parallelism, suite.configuration.File.Password.Parallelism)
}
func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenKeyLengthTooLow() {
	suite.configuration.File.Password.KeyLength = 1

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Key length for argon2id must be 16, you configured 1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSaltLengthTooLow() {
	suite.configuration.File.Password.SaltLength = -1

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "The salt length must be 2 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBadAlgorithmDefined() {
	suite.configuration.File.Password.Algorithm = "bogus"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Unknown hashing algorithm supplied, valid values are argon2id and sha512, you configured 'bogus'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenIterationsTooLow() {
	suite.configuration.File.Password.Iterations = -1

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "The number of iterations specified is invalid, must be 1 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenParallelismTooLow() {
	suite.configuration.File.Password.Parallelism = -1

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Parallelism for argon2id must be 1 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultValues() {
	suite.configuration.File.Password.Algorithm = ""
	suite.configuration.File.Password.Iterations = 0
	suite.configuration.File.Password.SaltLength = 0
	suite.configuration.File.Password.Memory = 0
	suite.configuration.File.Password.Parallelism = 0

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Algorithm, suite.configuration.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Iterations, suite.configuration.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.SaltLength, suite.configuration.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Memory, suite.configuration.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Parallelism, suite.configuration.File.Password.Parallelism)
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
	suite.configuration.Ldap.Implementation = schema.LDAPImplementationCustom
	suite.configuration.Ldap.URL = testLDAPURL
	suite.configuration.Ldap.User = testLDAPUser
	suite.configuration.Ldap.Password = testLDAPPassword
	suite.configuration.Ldap.BaseDN = testLDAPBaseDN
	suite.configuration.Ldap.UsernameAttribute = "uid"
	suite.configuration.Ldap.UsersFilter = "({username_attribute}={input})"
	suite.configuration.Ldap.GroupsFilter = "(cn={input})"
}

func (suite *LdapAuthenticationBackendSuite) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *LdapAuthenticationBackendSuite) TestShouldValidateDefaultImplementationAndUsernameAttribute() {
	suite.configuration.Ldap.Implementation = ""
	suite.configuration.Ldap.UsernameAttribute = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().Equal(schema.LDAPImplementationCustom, suite.configuration.Ldap.Implementation)

	suite.Assert().Equal(suite.configuration.Ldap.UsernameAttribute, schema.DefaultLDAPAuthenticationBackendConfiguration.UsernameAttribute)
	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseErrorWhenImplementationIsInvalidMSAD() {
	suite.configuration.Ldap.Implementation = "masd"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication backend ldap implementation must be blank or one of the following values `custom`, `activedirectory`")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseErrorWhenURLNotProvided() {
	suite.configuration.Ldap.URL = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a URL to the LDAP server")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseErrorWhenUserNotProvided() {
	suite.configuration.Ldap.User = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a user name to connect to the LDAP server")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseErrorWhenPasswordNotProvided() {
	suite.configuration.Ldap.Password = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a password to connect to the LDAP server")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseErrorWhenBaseDNNotProvided() {
	suite.configuration.Ldap.BaseDN = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a base DN to connect to the LDAP server")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseOnEmptyGroupsFilter() {
	suite.configuration.Ldap.GroupsFilter = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a groups filter with `groups_filter` attribute")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseOnEmptyUsersFilter() {
	suite.configuration.Ldap.UsersFilter = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a users filter with `users_filter` attribute")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldNotRaiseOnEmptyUsernameAttribute() {
	suite.configuration.Ldap.UsernameAttribute = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseOnBadRefreshInterval() {
	suite.configuration.RefreshInterval = "blah"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Auth Backend `refresh_interval` is configured to 'blah' but it must be either a duration notation or one of 'disable', or 'always'. Error from parser: Could not convert the input string of blah into a duration")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultImplementation() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.LDAPImplementationCustom, suite.configuration.Ldap.Implementation)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultGroupNameAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("cn", suite.configuration.Ldap.GroupNameAttribute)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultMailAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("mail", suite.configuration.Ldap.MailAttribute)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultDisplayNameAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("displayname", suite.configuration.Ldap.DisplayNameAttribute)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultRefreshInterval() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("5m", suite.configuration.RefreshInterval)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainEnclosingParenthesis() {
	suite.configuration.Ldap.UsersFilter = "{username_attribute}={input}"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "The users filter should contain enclosing parenthesis. For instance {username_attribute}={input} should be ({username_attribute}={input})")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseWhenGroupsFilterDoesNotContainEnclosingParenthesis() {
	suite.configuration.Ldap.GroupsFilter = "cn={input}"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "The groups filter should contain enclosing parenthesis. For instance cn={input} should be (cn={input})")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainUsernameAttribute() {
	suite.configuration.Ldap.UsersFilter = "(&({mail_attribute}={input})(objectClass=person))"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Unable to detect {username_attribute} placeholder in users_filter, your configuration is broken. Please review configuration options listed at https://docs.authelia.com/configuration/authentication/ldap.html")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldHelpDetectNoInputPlaceholder() {
	suite.configuration.Ldap.UsersFilter = "(&({username_attribute}={mail_attribute})(objectClass=person))"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Unable to detect {input} placeholder in users_filter, your configuration might be broken. Please review configuration options listed at https://docs.authelia.com/configuration/authentication/ldap.html")
}

func (suite *LdapAuthenticationBackendSuite) TestShouldAdaptLDAPURL() {
	suite.Assert().Equal("", validateLdapURLSimple("127.0.0.1", suite.validator))

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Unknown scheme for ldap url, should be ldap:// or ldaps://")

	suite.Assert().Equal("", validateLdapURLSimple("127.0.0.1:636", suite.validator))

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 2)
	suite.Assert().EqualError(suite.validator.Errors()[1], "Unable to parse URL to ldap server. The scheme is probably missing: ldap:// or ldaps://")

	suite.Assert().Equal("ldap://127.0.0.1", validateLdapURLSimple("ldap://127.0.0.1", suite.validator))
	suite.Assert().Equal("ldap://127.0.0.1:390", validateLdapURLSimple("ldap://127.0.0.1:390", suite.validator))
	suite.Assert().Equal("ldap://127.0.0.1/abc", validateLdapURLSimple("ldap://127.0.0.1/abc", suite.validator))
	suite.Assert().Equal("ldap://127.0.0.1/abc?test=abc&x=y", validateLdapURLSimple("ldap://127.0.0.1/abc?test=abc&x=y", suite.validator))

	suite.Assert().Equal("ldaps://127.0.0.1:390", validateLdapURLSimple("ldaps://127.0.0.1:390", suite.validator))
	suite.Assert().Equal("ldaps://127.0.0.1", validateLdapURLSimple("ldaps://127.0.0.1", suite.validator))
}

func (suite *LdapAuthenticationBackendSuite) TestShouldDefaultTLS12() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.DefaultLDAPAuthenticationBackendConfiguration.MinimumTLSVersion, suite.configuration.Ldap.MinimumTLSVersion)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldNotAllowInvalidTLSValue() {
	suite.configuration.Ldap.TLS = &schema.TLSConfig{
		MinimumVersion: "SSL2.0",
	}

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "error occurred validating the LDAP minimum_tls_version key with value SSL2.0: supplied TLS version isn't supported")
}

// Deprecated: Temporary Test. TODO: Remove in 4.28 (Whole Test).
func (suite *LdapAuthenticationBackendSuite) TestShouldReturnDeprecationWarningsAndNoMappingFor428() {
	var skipVerify = true

	suite.configuration.Ldap.MinimumTLSVersion = "TLS1.0"
	suite.configuration.Ldap.SkipVerify = &skipVerify
	suite.configuration.Ldap.TLS = nil
	suite.configuration.Ldap.TLS = &schema.TLSConfig{
		ServerName:     "golang.org",
		MinimumVersion: "",
	}

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	// Should not override since TLS schema is defined
	suite.Assert().Equal(false, suite.configuration.Ldap.TLS.SkipVerify)
	suite.Assert().Equal(schema.DefaultLDAPAuthenticationBackendConfiguration.TLS.MinimumVersion, suite.configuration.Ldap.TLS.MinimumVersion)

	suite.Assert().False(suite.validator.HasErrors())
	suite.Require().Len(suite.validator.Warnings(), 2)

	warnings := suite.validator.Warnings()

	suite.Assert().EqualError(warnings[0], "DEPRECATED: LDAP Auth Backend `skip_verify` option has been replaced by `authentication_backend.ldap.tls.skip_verify` (will be removed in 4.28.0)")
	suite.Assert().EqualError(warnings[1], "DEPRECATED: LDAP Auth Backend `minimum_tls_version` option has been replaced by `authentication_backend.ldap.tls.minimum_version` (will be removed in 4.28.0)")
}

// Deprecated: Temporary Test. TODO: Remove in 4.28 (Whole Test).
func (suite *LdapAuthenticationBackendSuite) TestShouldReturnDeprecationWarningsAndMappingFor428() {
	var skipVerify = true

	tlsVersion := "TLS1.1"

	suite.configuration.Ldap.MinimumTLSVersion = tlsVersion
	suite.configuration.Ldap.SkipVerify = &skipVerify

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	// Should override since TLS schema is not defined
	suite.Assert().Equal(true, suite.configuration.Ldap.TLS.SkipVerify)
	suite.Assert().Equal(tlsVersion, suite.configuration.Ldap.TLS.MinimumVersion)

	suite.Assert().False(suite.validator.HasErrors())
	suite.Require().Len(suite.validator.Warnings(), 2)

	warnings := suite.validator.Warnings()

	suite.Assert().EqualError(warnings[0], "DEPRECATED: LDAP Auth Backend `skip_verify` option has been replaced by `authentication_backend.ldap.tls.skip_verify` (will be removed in 4.28.0)")
	suite.Assert().EqualError(warnings[1], "DEPRECATED: LDAP Auth Backend `minimum_tls_version` option has been replaced by `authentication_backend.ldap.tls.minimum_version` (will be removed in 4.28.0)")
}

func TestLdapAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(LdapAuthenticationBackendSuite))
}

type ActiveDirectoryAuthenticationBackendSuite struct {
	suite.Suite
	configuration schema.AuthenticationBackendConfiguration
	validator     *schema.StructValidator
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration = schema.AuthenticationBackendConfiguration{}
	suite.configuration.Ldap = &schema.LDAPAuthenticationBackendConfiguration{}
	suite.configuration.Ldap.Implementation = schema.LDAPImplementationActiveDirectory
	suite.configuration.Ldap.URL = testLDAPURL
	suite.configuration.Ldap.User = testLDAPUser
	suite.configuration.Ldap.Password = testLDAPPassword
	suite.configuration.Ldap.BaseDN = testLDAPBaseDN
	suite.configuration.Ldap.TLS = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldSetActiveDirectoryDefaults() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(
		suite.configuration.Ldap.UsersFilter,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsersFilter)
	suite.Assert().Equal(
		suite.configuration.Ldap.UsernameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsernameAttribute)
	suite.Assert().Equal(
		suite.configuration.Ldap.DisplayNameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.DisplayNameAttribute)
	suite.Assert().Equal(
		suite.configuration.Ldap.MailAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.MailAttribute)
	suite.Assert().Equal(
		suite.configuration.Ldap.GroupsFilter,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupsFilter)
	suite.Assert().Equal(
		suite.configuration.Ldap.GroupNameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupNameAttribute)
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.configuration.Ldap.UsersFilter = "(&({username_attribute}={input})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.configuration.Ldap.UsernameAttribute = "cn"
	suite.configuration.Ldap.MailAttribute = "userPrincipalName"
	suite.configuration.Ldap.DisplayNameAttribute = "name"
	suite.configuration.Ldap.GroupsFilter = "(&(member={dn})(objectClass=group)(objectCategory=group))"
	suite.configuration.Ldap.GroupNameAttribute = "distinguishedName"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().NotEqual(
		suite.configuration.Ldap.UsersFilter,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsersFilter)
	suite.Assert().NotEqual(
		suite.configuration.Ldap.UsernameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsernameAttribute)
	suite.Assert().NotEqual(
		suite.configuration.Ldap.DisplayNameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.DisplayNameAttribute)
	suite.Assert().NotEqual(
		suite.configuration.Ldap.MailAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.MailAttribute)
	suite.Assert().NotEqual(
		suite.configuration.Ldap.GroupsFilter,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupsFilter)
	suite.Assert().NotEqual(
		suite.configuration.Ldap.GroupNameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupNameAttribute)
}

func TestActiveDirectoryAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(ActiveDirectoryAuthenticationBackendSuite))
}
