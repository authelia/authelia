package validator

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type StorageSuite struct {
	suite.Suite
	config schema.Storage
	val    *schema.StructValidator
}

func (suite *StorageSuite) SetupTest() {
	suite.val = schema.NewStructValidator()
	suite.config.EncryptionKey = testEncryptionKey
	suite.config.Local = nil
	suite.config.PostgreSQL = nil
	suite.config.MySQL = nil
}

func (suite *StorageSuite) TestShouldValidateOneStorageIsConfigured() {
	suite.config.Local = nil
	suite.config.PostgreSQL = nil
	suite.config.MySQL = nil

	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)
	suite.Assert().EqualError(suite.val.Errors()[0], "storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided")
}

func (suite *StorageSuite) TestShouldValidateLocalPathIsProvided() {
	suite.config.Local = &schema.StorageLocal{
		Path: "",
	}

	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.Assert().EqualError(suite.val.Errors()[0], "storage: local: option 'path' is required")

	suite.val.Clear()
	suite.config.Local.Path = "/myapth"

	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 0)
}

func (suite *StorageSuite) TestShouldValidateMySQLHostUsernamePasswordAndDatabaseAreProvided() {
	suite.config.MySQL = &schema.StorageMySQL{}
	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Errors(), 3)
	suite.Assert().EqualError(suite.val.Errors()[0], "storage: mysql: option 'address' is required")
	suite.Assert().EqualError(suite.val.Errors()[1], "storage: mysql: option 'username' and 'password' are required")
	suite.Assert().EqualError(suite.val.Errors()[2], "storage: mysql: option 'database' is required")

	suite.val.Clear()
	suite.config.MySQL = &schema.StorageMySQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://localhost:3306"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}
	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 0)
}

func (suite *StorageSuite) TestShouldSetDefaultMySQLTLSServerName() {
	suite.config.MySQL = &schema.StorageMySQL{
		StorageSQL: schema.StorageSQL{
			Address:  &schema.AddressTCP{Address: MustParseAddress("tcp://mysql:1234")},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		TLS: &schema.TLS{
			MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Assert().Len(suite.val.Errors(), 0)

	suite.Assert().Equal(suite.config.MySQL.Address.Hostname(), suite.config.MySQL.TLS.ServerName)
	suite.Assert().Equal("mysql", suite.config.MySQL.TLS.ServerName)
}

func (suite *StorageSuite) TestShouldRaiseErrorOnInvalidMySQLTLSVersion() {
	suite.config.MySQL = &schema.StorageMySQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://db1:3306"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		TLS: &schema.TLS{
			MinimumVersion: schema.TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.Assert().EqualError(suite.val.Errors()[0], "storage: mysql: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
}

func (suite *StorageSuite) TestShouldRaiseErrorOnInvalidMySQLTLSMinVersionGreaterThanMaximum() {
	suite.config.MySQL = &schema.StorageMySQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://db1:3306"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		TLS: &schema.TLS{
			MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
			MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS11},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.Assert().EqualError(suite.val.Errors()[0], "storage: mysql: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS1.3 is greater than the maximum version TLS1.1")
}

func (suite *StorageSuite) TestShouldValidatePostgreSQLHostUsernamePasswordAndDatabaseAreProvided() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{}
	suite.config.MySQL = nil
	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Errors(), 3)
	suite.Assert().EqualError(suite.val.Errors()[0], "storage: postgres: option 'address' is required")
	suite.Assert().EqualError(suite.val.Errors()[1], "storage: postgres: option 'username' and 'password' are required")
	suite.Assert().EqualError(suite.val.Errors()[2], "storage: postgres: option 'database' is required")

	suite.val.Clear()
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}
	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Assert().Len(suite.val.Errors(), 0)
}

func (suite *StorageSuite) TestShouldValidatePostgresSchemaDefault() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Assert().Len(suite.val.Errors(), 0)

	suite.Assert().Nil(suite.config.PostgreSQL.SSL) //nolint:staticcheck
	suite.Assert().Nil(suite.config.PostgreSQL.TLS)

	suite.Assert().Equal("public", suite.config.PostgreSQL.Schema)
}

func (suite *StorageSuite) TestShouldValidatePostgresTLSDefaults() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		TLS: &schema.TLS{},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Assert().Len(suite.val.Errors(), 0)

	suite.Assert().Nil(suite.config.PostgreSQL.SSL) //nolint:staticcheck
	suite.Require().NotNil(suite.config.PostgreSQL.TLS)

	suite.Assert().Equal(uint16(tls.VersionTLS12), suite.config.PostgreSQL.TLS.MinimumVersion.Value)
}

func (suite *StorageSuite) TestShouldSetDefaultPostgreSQLTLSServerName() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		TLS: &schema.TLS{
			MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Assert().Len(suite.val.Errors(), 0)

	suite.Assert().Equal(suite.config.PostgreSQL.Address.Hostname(), suite.config.PostgreSQL.TLS.ServerName)
}

func (suite *StorageSuite) TestShouldRaiseErrorOnInvalidPostgreSQLTLSVersion() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		TLS: &schema.TLS{
			MinimumVersion: schema.TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.Assert().EqualError(suite.val.Errors()[0], "storage: postgres: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
}

func (suite *StorageSuite) TestShouldRaiseErrorOnInvalidPostgreSQLMinVersionGreaterThanMaximum() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		TLS: &schema.TLS{
			MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
			MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS11},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.Assert().EqualError(suite.val.Errors()[0], "storage: postgres: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS1.3 is greater than the maximum version TLS1.1")
}

func (suite *StorageSuite) TestShouldValidatePostgresSSLDefaults() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		SSL: &schema.StoragePostgreSQLSSL{},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 1)
	suite.Assert().Len(suite.val.Errors(), 0)

	suite.Assert().NotNil(suite.config.PostgreSQL.SSL) //nolint:staticcheck
	suite.Require().Nil(suite.config.PostgreSQL.TLS)

	suite.Assert().Equal(schema.DefaultPostgreSQLStorageConfiguration.SSL.Mode, suite.config.PostgreSQL.SSL.Mode) //nolint:staticcheck
}

func (suite *StorageSuite) TestShouldRaiseErrorOnTLSAndLegacySSL() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		SSL: &schema.StoragePostgreSQLSSL{},
		TLS: &schema.TLS{},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.Assert().EqualError(suite.val.Errors()[0], "storage: postgres: can't define both 'tls' and 'ssl' configuration options")
}

func (suite *StorageSuite) TestShouldValidatePostgresDefaultsDontOverrideConfiguration() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		Schema: "authelia",
		SSL: &schema.StoragePostgreSQLSSL{
			Mode: "require",
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 1)
	suite.Assert().Len(suite.val.Errors(), 0)

	suite.Assert().Equal("require", suite.config.PostgreSQL.SSL.Mode) //nolint:staticcheck
	suite.Assert().Equal("authelia", suite.config.PostgreSQL.Schema)

	suite.Assert().EqualError(suite.val.Warnings()[0], "storage: postgres: ssl: the ssl configuration options are deprecated and we recommend the tls options instead")
}

func (suite *StorageSuite) TestShouldValidatePostgresSSLModeMustBeValid() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		SSL: &schema.StoragePostgreSQLSSL{
			Mode: "unknown",
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 1)
	suite.Require().Len(suite.val.Errors(), 1)
	suite.Assert().EqualError(suite.val.Errors()[0], "storage: postgres: ssl: option 'mode' must be one of 'disable', 'require', 'verify-ca', or 'verify-full' but it's configured as 'unknown'")
}

func (suite *StorageSuite) TestShouldRaiseErrorOnNoEncryptionKey() {
	suite.config.EncryptionKey = ""
	suite.config.Local = &schema.StorageLocal{
		Path: "/this/is/a/path",
	}

	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)
	suite.Assert().EqualError(suite.val.Errors()[0], "storage: option 'encryption_key' is required")
}

func (suite *StorageSuite) TestShouldRaiseErrorOnShortEncryptionKey() {
	suite.config.EncryptionKey = "abc"
	suite.config.Local = &schema.StorageLocal{
		Path: "/this/is/a/path",
	}

	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)
	suite.Assert().EqualError(suite.val.Errors()[0], "storage: option 'encryption_key' must be 20 characters or longer")
}

func TestShouldRunStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}
