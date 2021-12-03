package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateStorage validates storage configuration.
func ValidateStorage(configuration schema.StorageConfiguration, validator *schema.StructValidator) {
	if configuration.Local == nil && configuration.MySQL == nil && configuration.PostgreSQL == nil {
		validator.Push(errors.New(errStrStorage))
	}

	switch {
	case configuration.MySQL != nil:
		validateSQLConfiguration(&configuration.MySQL.SQLStorageConfiguration, validator, "mysql")
	case configuration.PostgreSQL != nil:
		validatePostgreSQLConfiguration(configuration.PostgreSQL, validator)
	case configuration.Local != nil:
		validateLocalStorageConfiguration(configuration.Local, validator)
	}

	if configuration.EncryptionKey == "" {
		validator.Push(errors.New(errStrStorageEncryptionKeyMustBeProvided))
	} else if len(configuration.EncryptionKey) < 20 {
		validator.Push(errors.New(errStrStorageEncryptionKeyTooShort))
	}
}

func validateSQLConfiguration(configuration *schema.SQLStorageConfiguration, validator *schema.StructValidator, provider string) {
	if configuration.Timeout == 0 {
		configuration.Timeout = schema.DefaultSQLStorageConfiguration.Timeout
	}

	if configuration.Host == "" {
		validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, provider, "host"))
	}

	if configuration.Username == "" || configuration.Password == "" {
		validator.Push(fmt.Errorf(errFmtStorageUserPassMustBeProvided, provider))
	}

	if configuration.Database == "" {
		validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, provider, "database"))
	}
}

func validatePostgreSQLConfiguration(configuration *schema.PostgreSQLStorageConfiguration, validator *schema.StructValidator) {
	validateSQLConfiguration(&configuration.SQLStorageConfiguration, validator, "postgres")

	if configuration.Schema == "" {
		configuration.Schema = schema.DefaultPostgreSQLStorageConfiguration.Schema
	}

	// Deprecated. TODO: Remove in v4.36.0.
	if configuration.SSLMode != "" && configuration.SSL.Mode == "" {
		configuration.SSL.Mode = configuration.SSLMode
	}

	if configuration.SSL.Mode == "" {
		configuration.SSL.Mode = schema.DefaultPostgreSQLStorageConfiguration.SSL.Mode
	} else if !utils.IsStringInSlice(configuration.SSL.Mode, storagePostgreSQLValidSSLModes) {
		validator.Push(fmt.Errorf(errFmtStoragePostgreSQLInvalidSSLMode, configuration.SSL.Mode, strings.Join(storagePostgreSQLValidSSLModes, "', '")))
	}
}

func validateLocalStorageConfiguration(configuration *schema.LocalStorageConfiguration, validator *schema.StructValidator) {
	if configuration.Path == "" {
		validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, "local", "path"))
	}
}
