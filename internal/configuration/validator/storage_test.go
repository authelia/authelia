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
	suite.EqualError(suite.val.Errors()[0], "storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided")
}

func (suite *StorageSuite) TestShouldValidateMultipleStorageIsConfigured() {
	suite.config.Local = &schema.StorageLocal{}
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{}
	suite.config.MySQL = &schema.StorageMySQL{}

	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)
	suite.EqualError(suite.val.Errors()[0], "storage: option 'local', 'mysql' and 'postgres' are mutually exclusive but 'local', 'mysql', and 'postgres' have been configured")
}

func (suite *StorageSuite) TestShouldValidateLocalPathIsProvided() {
	suite.config.Local = &schema.StorageLocal{
		Path: "",
	}

	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.EqualError(suite.val.Errors()[0], "storage: local: option 'path' is required")

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
	suite.EqualError(suite.val.Errors()[0], "storage: mysql: option 'address' is required")
	suite.EqualError(suite.val.Errors()[1], "storage: mysql: option 'username' is required")
	suite.EqualError(suite.val.Errors()[2], "storage: mysql: option 'database' is required")

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
			TLS: &schema.TLS{
				MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
			},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Len(suite.val.Errors(), 0)

	suite.Equal(suite.config.MySQL.Address.Hostname(), suite.config.MySQL.TLS.ServerName)
	suite.Equal("mysql", suite.config.MySQL.TLS.ServerName)
}

func (suite *StorageSuite) TestShouldRaiseErrorOnInvalidMySQLAddressScheme() {
	suite.config.MySQL = &schema.StorageMySQL{
		StorageSQL: schema.StorageSQL{
			Address:  &schema.AddressTCP{Address: MustParseAddress("udp://mysql:1234")},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.EqualError(suite.val.Errors()[0], "storage: mysql: option 'address' with value 'udp://mysql:1234' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', 'unix', or 'fd' but is configured as 'udp'")
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
			TLS: &schema.TLS{
				MinimumVersion: schema.TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
			},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.EqualError(suite.val.Errors()[0], "storage: mysql: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
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
			TLS: &schema.TLS{
				MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
				MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS11},
			},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.EqualError(suite.val.Errors()[0], "storage: mysql: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS 1.3 is greater than the maximum version TLS 1.1")
}

func (suite *StorageSuite) TestShouldValidatePostgreSQLHostUsernamePasswordAndDatabaseAreProvided() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{}
	suite.config.MySQL = nil
	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Errors(), 3)
	suite.EqualError(suite.val.Errors()[0], "storage: postgres: option 'address' is required")
	suite.EqualError(suite.val.Errors()[1], "storage: postgres: option 'username' is required")
	suite.EqualError(suite.val.Errors()[2], "storage: postgres: option 'database' is required")

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

	suite.Len(suite.val.Warnings(), 0)
	suite.Len(suite.val.Errors(), 0)
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

	suite.Len(suite.val.Warnings(), 0)
	suite.Len(suite.val.Errors(), 0)

	suite.Nil(suite.config.PostgreSQL.SSL) //nolint:staticcheck
	suite.Nil(suite.config.PostgreSQL.TLS)

	suite.Equal("public", suite.config.PostgreSQL.Schema)
}

func (suite *StorageSuite) TestShouldValidatePostgresPortDefault() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		Schema: "public",
	}

	ValidateStorage(suite.config, suite.val)

	suite.Assert().Len(suite.val.Warnings(), 0)
	suite.Assert().Len(suite.val.Errors(), 0)

	suite.Equal(uint16(5432), suite.config.PostgreSQL.Address.Port())
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
			TLS:      &schema.TLS{},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Len(suite.val.Errors(), 0)

	suite.Nil(suite.config.PostgreSQL.SSL) //nolint:staticcheck
	suite.Require().NotNil(suite.config.PostgreSQL.TLS)

	suite.Equal(uint16(tls.VersionTLS12), suite.config.PostgreSQL.TLS.MinimumVersion.Value)
}

func (suite *StorageSuite) TestShouldValidatePostgresServers() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		Servers: []schema.StoragePostgreSQLServer{
			{
				Address: &schema.AddressTCP{
					Address: MustParseAddress("udp://server1:4321"),
				},
			},
			{},
			{
				Address: &schema.AddressTCP{
					Address: MustParseAddress("tcp://server1"),
				},
			},
			{
				Address: &schema.AddressTCP{
					Address: MustParseAddress("tcp://server1:5432"),
				},
				TLS: &schema.TLS{},
			},
			{
				Address: &schema.AddressTCP{
					Address: MustParseAddress("tcp://server1:5432"),
				},
				TLS: &schema.TLS{
					MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
					MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS10},
				},
			},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.config.PostgreSQL.Servers, 5)

	suite.Equal(uint16(5432), suite.config.PostgreSQL.Servers[2].Address.Port())
	suite.Equal("server1", suite.config.PostgreSQL.Servers[3].TLS.ServerName)
	suite.Equal(schema.TLSVersion{Value: tls.VersionTLS12}, suite.config.PostgreSQL.Servers[3].TLS.MinimumVersion)

	errors := suite.val.Errors()

	suite.Require().Len(errors, 3)
	suite.EqualError(errors[0], "storage: postgres: servers: #1: option 'address' with value 'udp://server1:4321' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', 'unix', or 'fd' but is configured as 'udp'")
	suite.EqualError(errors[1], "storage: postgres: servers: #2: option 'address' is required")
	suite.EqualError(errors[2], "storage: postgres: servers: #5: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS 1.3 is greater than the maximum version TLS 1.0")
}

func (suite *StorageSuite) TestShouldValidatePostgresServersTLSMinMaxFromPrimary() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
			TLS: &schema.TLS{
				MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS10},
				MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
			},
		},
		Servers: []schema.StoragePostgreSQLServer{
			{
				Address: &schema.AddressTCP{
					Address: MustParseAddress("tcp://server1:5432"),
				},
				TLS: &schema.TLS{},
			},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Len(suite.val.Errors(), 0)
	suite.Require().Len(suite.config.PostgreSQL.Servers, 1)

	suite.Equal("server1", suite.config.PostgreSQL.Servers[0].TLS.ServerName)
	suite.Equal(schema.TLSVersion{Value: tls.VersionTLS10}, suite.config.PostgreSQL.Servers[0].TLS.MinimumVersion)
	suite.Equal(schema.TLSVersion{Value: tls.VersionTLS13}, suite.config.PostgreSQL.Servers[0].TLS.MaximumVersion)
}

func (suite *StorageSuite) TestShouldValidatePostgresUDP() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("udp://postgre:4321"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
			TLS: &schema.TLS{
				MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS10},
				MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
			},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.EqualError(suite.val.Errors()[0], "storage: postgres: option 'address' with value 'udp://postgre:4321' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', 'unix', or 'fd' but is configured as 'udp'")
}

func (suite *StorageSuite) TestShouldValidatePostgresSetPort() {
	suite.config.PostgreSQL = &schema.StoragePostgreSQL{
		StorageSQL: schema.StorageSQL{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://postgre"),
			},
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Len(suite.val.Errors(), 0)

	suite.Equal(uint16(5432), suite.config.PostgreSQL.Address.Port())
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
			TLS: &schema.TLS{
				MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
			},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Len(suite.val.Errors(), 0)

	suite.Equal(suite.config.PostgreSQL.Address.Hostname(), suite.config.PostgreSQL.TLS.ServerName)
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
			TLS: &schema.TLS{
				MinimumVersion: schema.TLSVersion{
					Value: tls.VersionSSL30, //nolint:staticcheck
				},
			},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.EqualError(suite.val.Errors()[0], "storage: postgres: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
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
			TLS: &schema.TLS{
				MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
				MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS11},
			},
		},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.EqualError(suite.val.Errors()[0], "storage: postgres: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS 1.3 is greater than the maximum version TLS 1.1")
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

	suite.Len(suite.val.Warnings(), 1)
	suite.Len(suite.val.Errors(), 0)

	suite.NotNil(suite.config.PostgreSQL.SSL) //nolint:staticcheck
	suite.Require().Nil(suite.config.PostgreSQL.TLS)

	suite.Equal(schema.DefaultPostgreSQLStorageConfiguration.SSL.Mode, suite.config.PostgreSQL.SSL.Mode) //nolint:staticcheck
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
			TLS:      &schema.TLS{},
		},
		SSL: &schema.StoragePostgreSQLSSL{},
	}

	ValidateStorage(suite.config, suite.val)

	suite.Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)

	suite.EqualError(suite.val.Errors()[0], "storage: postgres: can't define both 'tls' and 'ssl' configuration options")
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
	suite.Len(suite.val.Errors(), 0)

	suite.Equal("require", suite.config.PostgreSQL.SSL.Mode) //nolint:staticcheck
	suite.Equal("authelia", suite.config.PostgreSQL.Schema)

	suite.EqualError(suite.val.Warnings()[0], "storage: postgres: ssl: the ssl configuration options are deprecated and we recommend the tls options instead")
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

	suite.Len(suite.val.Warnings(), 1)
	suite.Require().Len(suite.val.Errors(), 1)
	suite.EqualError(suite.val.Errors()[0], "storage: postgres: ssl: option 'mode' must be one of 'disable', 'require', 'verify-ca', or 'verify-full' but it's configured as 'unknown'")
}

func (suite *StorageSuite) TestShouldRaiseErrorOnNoEncryptionKey() {
	suite.config.EncryptionKey = ""
	suite.config.Local = &schema.StorageLocal{
		Path: "/this/is/a/path",
	}

	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)
	suite.EqualError(suite.val.Errors()[0], "storage: option 'encryption_key' is required")
}

func (suite *StorageSuite) TestShouldRaiseErrorOnShortEncryptionKey() {
	suite.config.EncryptionKey = "abc"
	suite.config.Local = &schema.StorageLocal{
		Path: "/this/is/a/path",
	}

	ValidateStorage(suite.config, suite.val)

	suite.Require().Len(suite.val.Warnings(), 0)
	suite.Require().Len(suite.val.Errors(), 1)
	suite.EqualError(suite.val.Errors()[0], "storage: option 'encryption_key' must be 20 characters or longer")
}

func TestShouldRunStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}
