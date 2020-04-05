package validator

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
)

type StorageSuite struct {
	suite.Suite

	configuration schema.StorageConfiguration
}

func (s *StorageSuite) SetupTest() {
	s.configuration.Local = &schema.LocalStorageConfiguration{
		Path: "/this/is/a/path",
	}
}

func (s *StorageSuite) TestShouldValidateOneStorageIsConfigured() {
	validator := schema.NewStructValidator()
	s.configuration.Local = nil

	ValidateStorage(s.configuration, validator)

	s.Require().Len(validator.Errors(), 1)
	s.Assert().EqualError(validator.Errors()[0], "A storage configuration must be provided. It could be 'local', 'mysql' or 'postgres'")
}

func (s *StorageSuite) TestShouldValidateLocalPathIsProvided() {
	validator := schema.NewStructValidator()
	s.configuration.Local.Path = ""

	ValidateStorage(s.configuration, validator)

	s.Require().Len(validator.Errors(), 1)
	s.Assert().EqualError(validator.Errors()[0], "A file path must be provided with key 'path'")

	validator = schema.NewStructValidator()
	s.configuration.Local.Path = "/myapth"
	ValidateStorage(s.configuration, validator)
	s.Require().Len(validator.Errors(), 0)
}

func (s *StorageSuite) TestShouldValidateSQLUsernamePasswordAndDatabaseAreProvided() {
	validator := schema.NewStructValidator()
	s.configuration.MySQL = &schema.MySQLStorageConfiguration{}
	ValidateStorage(s.configuration, validator)

	s.Require().Len(validator.Errors(), 2)
	s.Assert().EqualError(validator.Errors()[0], "Username and password must be provided")
	s.Assert().EqualError(validator.Errors()[1], "A database must be provided")

	validator = schema.NewStructValidator()
	s.configuration.MySQL = &schema.MySQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}
	ValidateStorage(s.configuration, validator)

	s.Require().Len(validator.Errors(), 0)
}

func (s *StorageSuite) TestShouldValidatePostgresSSLModeIsDisableByDefault() {
	validator := schema.NewStructValidator()
	s.configuration.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
	}
	ValidateStorage(s.configuration, validator)

	s.Assert().Equal("disable", s.configuration.PostgreSQL.SSLMode)
}

func (s *StorageSuite) TestShouldValidatePostgresSSLModeMustBeValid() {
	validator := schema.NewStructValidator()
	s.configuration.PostgreSQL = &schema.PostgreSQLStorageConfiguration{
		SQLStorageConfiguration: schema.SQLStorageConfiguration{
			Username: "myuser",
			Password: "pass",
			Database: "database",
		},
		SSLMode: "unknown",
	}
	ValidateStorage(s.configuration, validator)

	s.Require().Len(validator.Errors(), 1)
	s.Assert().EqualError(validator.Errors()[0], "SSL mode must be 'disable', 'require', 'verify-ca' or 'verify-full'")
}

func TestShouldRunStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}
