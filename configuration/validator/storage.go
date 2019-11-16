package validator

import (
	"errors"

	"github.com/clems4ever/authelia/configuration/schema"
)

// ValidateSQLStorage validates storage configuration.
func ValidateSQLStorage(configuration *schema.StorageConfiguration, validator *schema.StructValidator) {
	if configuration.Local == nil && configuration.SQL == nil {
		validator.Push(errors.New("A storage configuration must be provided. It could be 'local' or 'sql'"))
	}

	if configuration.SQL != nil {
		validateSQLConfiguration(configuration.SQL, validator)
	} else if configuration.Local != nil {
		validateLocalStorageConfiguration(configuration.Local, validator)
	}
}

func validateSQLConfiguration(configuration *schema.SQLStorageConfiguration, validator *schema.StructValidator) {
	if configuration.Password != "" && configuration.Username == "" {
		validator.Push(errors.New("Username and password must be provided"))
	}

	if configuration.Database == "" {
		validator.Push(errors.New("A database must be provided"))
	}
}

func validateLocalStorageConfiguration(configuration *schema.LocalStorageConfiguration, validator *schema.StructValidator) {
	if configuration.Path == "" {
		validator.Push(errors.New("A file path must be provided with key 'path'"))
	}
}
