package validator

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type StorageSuite struct {
	suite.Suite
	config    schema.StorageConfiguration
	validator *schema.StructValidator
}

func (suite *StorageSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config.EncryptionKey = testEncryptionKey
	suite.config.Local = nil
	suite.config.PostgreSQL = nil
	suite.config.MySQL = nil
}

func (suite *StorageSuite) TestShouldValidateOneStorageIsConfigured() {
	suite.config.Local = nil
	suite.config.PostgreSQL = nil
	suite.config.MySQL = nil

	ValidateStorage(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided")
}

func (suite *StorageSuite) TestShouldValidateLocalPathIsProvided() {
	suite.config.Local = &schema.LocalStorageConfiguration{
		Path: "",
	}

	ValidateStorage(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: local: option 'path' is required")

	suite.validator.Clear()
	suite.config.Local.Path = "/myapth"

	ValidateStorage(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 0)
}

func (suite *StorageSuite) TestShouldValidateMySQLHostUsernamePasswordAndDatabaseAreProvided() {
	suite.config.MySQL = &schema.MySQLStorageConfiguration{}
	ValidateStorage(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 3)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: mysql: option 'host' is required")
	suite.Assert().EqualError(suite.validator.Errors()[1], "storage: mysql: option 'username' and 'password' are required")
	suite.Assert().EqualError(suite.validator.Errors()[2], "storage: mysql: option 'database' is required")

	suite.validator.Clear()
	suite.config.MySQL = &schema.MySQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "localhost",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}
	ValidateStorage(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 0)
}

func (suite *StorageSuite) TestShouldValidatePostgreSQLHostUsernamePasswordAndDatabaseAreProvided() {
	suite.config.PostgreSQL = &schema.PostgreSQLStorageConfiguration{}
	suite.config.MySQL = nil
	ValidateStorage(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Errors(), 3)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: postgres: option 'host' is required")
	suite.Assert().EqualError(suite.validator.Errors()[1], "storage: postgres: option 'username' and 'password' are required")
	suite.Assert().EqualError(suite.validator.Errors()[2], "storage: postgres: option 'database' is required")

	suite.validator.Clear()
	suite.config.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "postgre",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}
	ValidateStorage(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *StorageSuite) TestShouldValidatePostgresSchemaDefault() {
	suite.config.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "db1",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}

	ValidateStorage(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Nil(suite.config.PostgreSQL.SSL)
	suite.Assert().Nil(suite.config.PostgreSQL.TLS)

	suite.Assert().Equal("public", suite.config.PostgreSQL.Schema)
}

func (suite *StorageSuite) TestShouldValidatePostgresTLSDefaults() {
	suite.config.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "db1",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		TLS: &schema.TLSConfig{},
	}

	ValidateStorage(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Nil(suite.config.PostgreSQL.SSL)
	suite.Require().NotNil(suite.config.PostgreSQL.TLS)

	suite.Assert().Equal(uint16(tls.VersionTLS12), suite.config.PostgreSQL.TLS.MinimumVersion.Value)
}

func (suite *StorageSuite) TestShouldValidatePostgresSSLDefaults() {
	suite.config.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "db1",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		SSL: &schema.PostgreSQLSSLStorageConfiguration{},
	}

	ValidateStorage(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 1)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().NotNil(suite.config.PostgreSQL.SSL)
	suite.Require().Nil(suite.config.PostgreSQL.TLS)

	suite.Assert().Equal(schema.DefaultPostgreSQLStorageConfiguration.SSL.Mode, suite.config.PostgreSQL.SSL.Mode)
}

func (suite *StorageSuite) TestShouldRaiseErrorOnTLSAndLegacySSL() {
	suite.config.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "db1",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		SSL: &schema.PostgreSQLSSLStorageConfiguration{},
		TLS: &schema.TLSConfig{},
	}

	ValidateStorage(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: postgres: can't define both 'tls' and 'ssl' configuration options")
}

func (suite *StorageSuite) TestShouldValidatePostgresDefaultsDontOverrideConfiguration() {
	suite.config.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "db1",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		Schema: "authelia",
		SSL: &schema.PostgreSQLSSLStorageConfiguration{
			Mode: "require",
		},
	}

	ValidateStorage(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 1)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("require", suite.config.PostgreSQL.SSL.Mode)
	suite.Assert().Equal("authelia", suite.config.PostgreSQL.Schema)

	suite.Assert().EqualError(suite.validator.Warnings()[0], "storage: postgres: ssl: the ssl configuration options are deprecated and we recommend the tls options instead")
}

func (suite *StorageSuite) TestShouldValidatePostgresSSLModeMustBeValid() {
	suite.config.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Host:     "db2",
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		SSL: &schema.PostgreSQLSSLStorageConfiguration{
			Mode: "unknown",
		},
	}

	ValidateStorage(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 1)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: postgres: ssl: option 'mode' must be one of 'disable', 'require', 'verify-ca', 'verify-full' but it is configured as 'unknown'")
}

func (suite *StorageSuite) TestShouldRaiseErrorOnNoEncryptionKey() {
	suite.config.EncryptionKey = ""
	suite.config.Local = &schema.LocalStorageConfiguration{
		Path: "/this/is/a/path",
	}

	ValidateStorage(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: option 'encryption_key' is required")
}

func (suite *StorageSuite) TestShouldRaiseErrorOnShortEncryptionKey() {
	suite.config.EncryptionKey = "abc"
	suite.config.Local = &schema.LocalStorageConfiguration{
		Path: "/this/is/a/path",
	}

	ValidateStorage(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)
	suite.Assert().EqualError(suite.validator.Errors()[0], "storage: option 'encryption_key' must be 20 characters or longer")
}

func TestShouldRunStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}
