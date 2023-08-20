package validator

import (
	"errors"
	"fmt"

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

func validateSQLConfiguration(config *schema.SQLStorageConfiguration, validator *schema.StructValidator, provider string) {
	if config.Address == nil {
		if config.Host == "" { //nolint:staticcheck
			validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, provider, "address"))
		} else {
			host := config.Host //nolint:staticcheck
			port := config.Port //nolint:staticcheck

			if address, err := schema.NewAddressFromNetworkValuesDefault(host, port, schema.AddressSchemeTCP, schema.AddressSchemeUnix); err != nil {
				validator.Push(fmt.Errorf(errFmtStorageFailedToConvertHostPortToAddress, provider, err))
			} else {
				config.Address = &schema.AddressTCP{Address: *address}
			}
		}
	} else {
		if config.Host != "" || config.Port != 0 { //nolint:staticcheck
			validator.Push(fmt.Errorf(errFmtStorageOptionAddressConflictWithHostPort, provider))
		}

		var err error

		if err = config.Address.ValidateSQL(); err != nil {
			validator.Push(fmt.Errorf(errFmtServerAddress, config.Address.String(), err))
		}
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

func validateMySQLConfiguration(config *schema.MySQLStorageConfiguration, validator *schema.StructValidator) {
	validateSQLConfiguration(&config.SQLStorageConfiguration, validator, "mysql")

	if config.TLS != nil {
		configDefaultTLS := &schema.TLSConfig{
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

func validatePostgreSQLConfiguration(config *schema.PostgreSQLStorageConfiguration, validator *schema.StructValidator) {
	validateSQLConfiguration(&config.SQLStorageConfiguration, validator, "postgres")

	if config.Schema == "" {
		config.Schema = schema.DefaultPostgreSQLStorageConfiguration.Schema
	}

	switch {
	case config.TLS != nil && config.SSL != nil:
		validator.Push(fmt.Errorf(errFmtStoragePostgreSQLInvalidSSLAndTLSConfig))
	case config.TLS != nil:
		configDefaultTLS := &schema.TLSConfig{
			ServerName:     config.Address.Hostname(),
			MinimumVersion: schema.DefaultPostgreSQLStorageConfiguration.TLS.MinimumVersion,
			MaximumVersion: schema.DefaultPostgreSQLStorageConfiguration.TLS.MaximumVersion,
		}

		if err := ValidateTLSConfig(config.TLS, configDefaultTLS); err != nil {
			validator.Push(fmt.Errorf(errFmtStorageTLSConfigInvalid, "postgres", err))
		}
	case config.SSL != nil:
		validator.PushWarning(fmt.Errorf(warnFmtStoragePostgreSQLInvalidSSLDeprecated))

		switch {
		case config.SSL.Mode == "":
			config.SSL.Mode = schema.DefaultPostgreSQLStorageConfiguration.SSL.Mode
		case !utils.IsStringInSlice(config.SSL.Mode, validStoragePostgreSQLSSLModes):
			validator.Push(fmt.Errorf(errFmtStoragePostgreSQLInvalidSSLMode, strJoinOr(validStoragePostgreSQLSSLModes), config.SSL.Mode))
		}
	}
}

func validateLocalStorageConfiguration(config *schema.LocalStorageConfiguration, validator *schema.StructValidator) {
	if config.Path == "" {
		validator.Push(fmt.Errorf(errFmtStorageOptionMustBeProvided, "local", "path"))
	}
}
