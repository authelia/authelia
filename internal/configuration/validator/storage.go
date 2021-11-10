package validator

import (
	"errors"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateStorage validates storage configuration.
func ValidateStorage(configuration schema.StorageConfiguration, validator *schema.StructValidator) {
	if configuration.Local == nil && configuration.MySQL == nil && configuration.PostgreSQL == nil {
		validator.Push(errors.New("A storage configuration must be provided. It could be 'local', 'mysql' or 'postgres'"))
	}

	switch {
	case configuration.MySQL != nil:
		validateMySQLConfiguration(&configuration.MySQL.SQLStorageConfiguration, validator)
	case configuration.PostgreSQL != nil:
		validatePostgreSQLConfiguration(configuration.PostgreSQL, validator)
	case configuration.Local != nil:
		validateLocalStorageConfiguration(configuration.Local, validator)
	}

	if configuration.EncryptionKey == "" || len(configuration.EncryptionKey) < 20 {
		validator.Push(errors.New("the configuration option storage.encryption_key must be provided and must have a length above 20"))
	}
}

func validateMySQLConfiguration(configuration *schema.SQLStorageConfiguration, validator *schema.StructValidator) {
	if configuration.Timeout == 0 {
		configuration.Timeout = schema.DefaultMySQLStorageConfiguration.Timeout
	}

	if configuration.Password == "" || configuration.Username == "" {
		validator.Push(errors.New("the SQL username and password must be provided"))
	}

	if configuration.Database == "" {
		validator.Push(errors.New("the SQL database must be provided"))
	}
}

func validatePostgreSQLConfiguration(configuration *schema.PostgreSQLStorageConfiguration, validator *schema.StructValidator) {
	validateMySQLConfiguration(&configuration.SQLStorageConfiguration, validator)

	if configuration.Timeout == 0 {
		configuration.Timeout = schema.DefaultPostgreSQLStorageConfiguration.Timeout
	}

	if configuration.SSLMode == "" {
		configuration.SSLMode = testModeDisabled
	}

	if !(configuration.SSLMode == testModeDisabled || configuration.SSLMode == "require" ||
		configuration.SSLMode == "verify-ca" || configuration.SSLMode == "verify-full") {
		validator.Push(errors.New("SSL mode must be 'disable', 'require', 'verify-ca', or 'verify-full'"))
	}
}

func validateLocalStorageConfiguration(configuration *schema.LocalStorageConfiguration, validator *schema.StructValidator) {
	if configuration.Path == "" {
		validator.Push(errors.New("A file path must be provided with key 'path'"))
	}
}
