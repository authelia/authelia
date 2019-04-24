package validator

import (
	"testing"

	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/stretchr/testify/assert"
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
	suite.configuration.File = &schema.FileAuthenticationBackendConfiguration{Path: "/a/path"}
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
	assert.Equal(suite.T(), "cn={0}", suite.configuration.Ldap.UsersFilter)
}

func (suite *LdapAuthenticationBackendSuite) TestShouldSetDefaultGroupsFilter() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)
	assert.Len(suite.T(), suite.validator.Errors(), 0)
	assert.Equal(suite.T(), "member={dn}", suite.configuration.Ldap.GroupsFilter)
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

func TestLdapAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(LdapAuthenticationBackendSuite))
}
