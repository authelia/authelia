package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateStorage validates storage configuration.
func ValidateStorage(config schema.StorageConfiguration, validator *schema.StructValidator) {
	if config.Local == nil && config.MySQL == nil && config.PostgreSQL == nil {
		validator.Push(errors.New(errStrStorage))
	}

	switch {
	case config.MySQL != nil:
		validateSQLConfiguration(&config.MySQL.SQLStorageConfiguration, validator, "mysql")
	case config.PostgreSQL != nil:
		validatePostgreSQLConfiguration(config.PostgreSQL, validator)
	case config.Local != nil:
		validateLocalStorageConfiguration(config.Local, validator)
	}

	if config.EncryptionKey == "" {
		validator.Push(errors.New(errStrStorageEncryptionKeyMustBeProvided))
	} else if len(config.EncryptionKey) < 20 {
		validator.Push(errors.New(errStrStorageEncryptionKeyTooShort))
	}
}

func validateSQLConfiguration(config *schema.SQLStorageConfiguration, validator *schema.StructValidator, provider string) {
	if config.Timeout == 0 {
		config.Timeout = schema.DefaultSQLStorageConfiguration.Timeout
	}

	if config.Host == "" {
		validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, provider, "host"))
	}

	if config.Username == "" || config.Password == "" {
		validator.Push(fmt.Errorf(errFmtStorageUserPassMustBeProvided, provider))
	}

	if config.Database == "" {
		validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, provider, "database"))
	}
}

func validatePostgreSQLConfiguration(config *schema.PostgreSQLStorageConfiguration, validator *schema.StructValidator) {
	validateSQLConfiguration(&config.SQLStorageConfiguration, validator, "postgres")

	if config.Schema == "" {
		config.Schema = schema.DefaultPostgreSQLStorageConfiguration.Schema
	}

	switch {
	case config.TLS != nil && config.SSL != nil:
		validator.Push(fmt.Errorf(errFmtStoragePostgreSQLInvalidSSLAndTLSConfig))
	case config.TLS != nil:
		if config.TLS.MinimumVersion.Value == 0 {
			config.TLS.MinimumVersion = schema.DefaultPostgreSQLStorageConfiguration.TLS.MinimumVersion
		}
	case config.SSL != nil:
		validator.PushWarning(fmt.Errorf(warnFmtStoragePostgreSQLInvalidSSLDeprecated))

		switch {
		case config.SSL.Mode == "":
			config.SSL.Mode = schema.DefaultPostgreSQLStorageConfiguration.SSL.Mode
		case !utils.IsStringInSlice(config.SSL.Mode, validStoragePostgreSQLSSLModes):
			validator.Push(fmt.Errorf(errFmtStoragePostgreSQLInvalidSSLMode, strings.Join(validStoragePostgreSQLSSLModes, "', '"), config.SSL.Mode))
		}
	}
}

func validateLocalStorageConfiguration(config *schema.LocalStorageConfiguration, validator *schema.StructValidator) {
	if config.Path == "" {
		validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, "local", "path"))
	}
}
