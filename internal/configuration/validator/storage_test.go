package validator

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
)

type StorageSuite struct {
	suite.Suite
	configuration schema.StorageConfiguration
	validator     *schema.StructValidator
}

func (suite *StorageSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration.Local = &schema.LocalStorageConfiguration{
		Path: "/this/is/a/path",
	}
}

func (suite *StorageSuite) TestShouldValidateOneStorageIsConfigured() {
	suite.configuration.Local = nil

	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "A storage configuration must be provided. It could be 'local', 'mysql' or 'postgres'")
}

func (suite *StorageSuite) TestShouldValidateLocalPathIsProvided() {
	suite.configuration.Local.Path = ""

	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "A file path must be provided with key 'path'")

	suite.validator.Clear()
	suite.configuration.Local.Path = "/myapth"

	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *StorageSuite) TestShouldValidateSQLUsernamePasswordAndDatabaseAreProvided() {
	suite.configuration.MySQL = &schema.MySQLStorageConfiguration{}
	ValidateStorage(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 2)
	suite.Assert().EqualError(suite.validator.Errors()[0], "Username and password must be provided")
	suite.Assert().EqualError(suite.validator.Errors()[1], "A database must be provided")

	suite.validator.Clear()
	suite.configuration.MySQL = &schema.MySQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}
	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *StorageSuite) TestShouldValidatePostgresSSLModeIsDisableByDefault() {
	suite.configuration.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}

	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("disable", suite.configuration.PostgreSQL.SSLMode)
}

func (suite *StorageSuite) TestShouldValidatePostgresSSLModeMustBeValid() {
	suite.configuration.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		SSLMode: "unknown",
	}

	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "SSL mode must be 'disable', 'require', 'verify-ca', or 'verify-full'")
}

func TestShouldRunStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}
