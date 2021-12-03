package validator

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type StorageSuite struct {
	suite.Suite
	configuration schema.StorageConfiguration
	validator     *schema.StructValidator
}

func (suite *StorageSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration.EncryptionKey = testEncryptionKey
	suite.configuration.Local = nil
	suite.configuration.PostgreSQL = nil
	suite.configuration.MySQL = nil
}

func (suite *StorageSuite) TestShouldValidateOneStorageIsConfigured() {
	suite.configuration.Local = nil
	suite.configuration.PostgreSQL = nil
	suite.configuration.MySQL = nil

	ValidateStorage(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided")
}

func (suite *StorageSuite) TestShouldValidateLocalPathIsProvided() {
	suite.configuration.Local = &schema.LocalStorageConfiguration{
		Path: "",
	}

	ValidateStorage(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: local: 'path' configuration option must be provided")

	suite.validator.Clear()
	suite.configuration.Local.Path = "/myapth"

	ValidateStorage(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 0)
}

func (suite *StorageSuite) TestShouldValidateMySQLHostUsernamePasswordAndDatabaseAreProvided() {
	suite.configuration.MySQL = &schema.MySQLStorageConfiguration{}
	ValidateStorage(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 3)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: mysql: 'host' configuration option must be provided")
	suite.Assert().EqualError(suite.validator.Errors()[1], "storage: mysql: 'username' and 'password' configuration options must be provided")
	suite.Assert().EqualError(suite.validator.Errors()[2], "storage: mysql: 'database' configuration option must be provided")

	suite.validator.Clear()
	suite.configuration.MySQL = &schema.MySQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "localhost",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}
	ValidateStorage(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 0)
}

func (suite *StorageSuite) TestShouldValidatePostgreSQLHostUsernamePasswordAndDatabaseAreProvided() {
	suite.configuration.PostgreSQL = &schema.PostgreSQLStorageConfiguration{}
	suite.configuration.MySQL = nil
	ValidateStorage(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 3)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: postgres: 'host' configuration option must be provided")
	suite.Assert().EqualError(suite.validator.Errors()[1], "storage: postgres: 'username' and 'password' configuration options must be provided")
	suite.Assert().EqualError(suite.validator.Errors()[2], "storage: postgres: 'database' configuration option must be provided")

	suite.validator.Clear()
	suite.configuration.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "postgre",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}
	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *StorageSuite) TestShouldValidatePostgresSSLModeAndSchemaDefaults() {
	suite.configuration.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "db1",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}

	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("disable", suite.configuration.PostgreSQL.SSL.Mode)
	suite.Assert().Equal("public", suite.configuration.PostgreSQL.Schema)
}

func (suite *StorageSuite) TestShouldValidatePostgresDefaultsDontOverrideConfiguration() {
	suite.configuration.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "db1",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		Schema: "authelia",
		SSL: schema.PostgreSQLSSLStorageConfiguration{
			Mode: "require",
		},
	}

	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("require", suite.configuration.PostgreSQL.SSL.Mode)
	suite.Assert().Equal("authelia", suite.configuration.PostgreSQL.Schema)
}

func (suite *StorageSuite) TestShouldValidatePostgresSSLModeMustBeValid() {
	suite.configuration.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "db2",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		SSL: schema.PostgreSQLSSLStorageConfiguration{
			Mode: "unknown",
		},
	}

	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: postgres: ssl: 'mode' configuration option 'unknown' is invalid: must be one of 'disable', 'require', 'verify-ca', 'verify-full'")
}

// Deprecated. TODO: Remove in v4.36.0.
func (suite *StorageSuite) TestShouldValidatePostgresSSLModeMustBeMappedForDeprecations() {
	suite.configuration.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "pg",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		SSLMode: "require",
	}

	ValidateStorage(suite.configuration, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal(suite.configuration.PostgreSQL.SSL.Mode, "require")
}

func (suite *StorageSuite) TestShouldRaiseErrorOnNoEncryptionKey() {
	suite.configuration.EncryptionKey = ""
	suite.configuration.Local = &schema.LocalStorageConfiguration{
		Path: "/this/is/a/path",
	}

	ValidateStorage(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: 'encryption_key' configuration option must be provided")
}

func (suite *StorageSuite) TestShouldRaiseErrorOnShortEncryptionKey() {
	suite.configuration.EncryptionKey = "abc"
	suite.configuration.Local = &schema.LocalStorageConfiguration{
		Path: "/this/is/a/path",
	}

	ValidateStorage(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: 'encryption_key' configuration option must be 20 characters or longer")
}

func TestShouldRunStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}
