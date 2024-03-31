package validator

import (
	"errors"
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateStorage validates storage configuration.
func ValidateStorage(config schema.Storage, validator *schema.StructValidator) {
	if config.Local == nil && config.MySQL == nil && config.PostgreSQL == nil {
		validator.Push(errors.New(errStrStorage))
	}

	switch {
	case config.MySQL != nil:
		validateMySQLConfiguration(config.MySQL, validator)
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

func validateSQLConfiguration(config, defaults *schema.StorageSQL, validator *schema.StructValidator, provider string) {
	if config.Address == nil {
		validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, provider, "address"))
	} else {
		var err error

		if err = config.Address.ValidateSQL(); err != nil {
			validator.Push(fmt.Errorf(errFmtServerAddress, config.Address.String(), err))
		}
	}

	if config.Address != nil && config.Address.IsTCP() && config.Address.Port() == 0 {
		config.Address.SetPort(defaults.Address.Port())
	}

	if config.Username == "" || config.Password == "" {
		validator.Push(fmt.Errorf(errFmtStorageUserPassMustBeProvided, provider))
	}

	if config.Database == "" {
		validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, provider, "database"))
	}

	if config.Timeout == 0 {
		config.Timeout = schema.DefaultSQLStorageConfiguration.Timeout
	}
}

func validateMySQLConfiguration(config *schema.StorageMySQL, validator *schema.StructValidator) {
	validateSQLConfiguration(&config.StorageSQL, &schema.DefaultMySQLStorageConfiguration.StorageSQL, validator, "mysql")

	if config.TLS != nil {
		configDefaultTLS := &schema.TLS{
			MinimumVersion: schema.DefaultMySQLStorageConfiguration.TLS.MinimumVersion,
			MaximumVersion: schema.DefaultMySQLStorageConfiguration.TLS.MaximumVersion,
		}

		if config.Address != nil {
			configDefaultTLS.ServerName = config.Address.Hostname()
		}

		if err := ValidateTLSConfig(config.TLS, configDefaultTLS); err != nil {
			validator.Push(fmt.Errorf(errFmtStorageTLSConfigInvalid, "mysql", err))
		}
	}
}

func validatePostgreSQLConfiguration(config *schema.StoragePostgreSQL, validator *schema.StructValidator) {
	validateSQLConfiguration(&config.StorageSQL, &schema.DefaultPostgreSQLStorageConfiguration.StorageSQL, validator, "postgres")

	if config.Schema == "" {
		config.Schema = schema.DefaultPostgreSQLStorageConfiguration.Schema
	}

	switch {
	case config.TLS != nil && config.SSL != nil: //nolint:staticcheck
		validator.Push(fmt.Errorf(errFmtStoragePostgreSQLInvalidSSLAndTLSConfig))
	case config.TLS != nil:
		configDefaultTLS := &schema.TLS{
			ServerName:     config.Address.Hostname(),
			MinimumVersion: schema.DefaultPostgreSQLStorageConfiguration.TLS.MinimumVersion,
			MaximumVersion: schema.DefaultPostgreSQLStorageConfiguration.TLS.MaximumVersion,
		}

		if err := ValidateTLSConfig(config.TLS, configDefaultTLS); err != nil {
			validator.Push(fmt.Errorf(errFmtStorageTLSConfigInvalid, "postgres", err))
		}
	case config.SSL != nil: //nolint:staticcheck
		validator.PushWarning(fmt.Errorf(warnFmtStoragePostgreSQLInvalidSSLDeprecated))

		switch {
		case config.SSL.Mode == "": //nolint:staticcheck
			config.SSL.Mode = schema.DefaultPostgreSQLStorageConfiguration.SSL.Mode //nolint:staticcheck
		case !utils.IsStringInSlice(config.SSL.Mode, validStoragePostgreSQLSSLModes): //nolint:staticcheck
			validator.Push(fmt.Errorf(errFmtStoragePostgreSQLInvalidSSLMode, utils.StringJoinOr(validStoragePostgreSQLSSLModes), config.SSL.Mode)) //nolint:staticcheck
		}
	}
}

func validateLocalStorageConfiguration(config *schema.StorageLocal, validator *schema.StructValidator) {
	if config.Path == "" {
		validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, "local", "path"))
	}
}
