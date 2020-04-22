package configuration

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
)

// Read a YAML configuration and create a Configuration object out of it.
func Read(configPath string) (*schema.Configuration, []error) {
	viper.SetEnvPrefix("AUTHELIA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// we need to bind all env variables as long as https://github.com/spf13/viper/issues/761
	// is not resolved.
	viper.BindEnv("jwt_secret")                           //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("duo_api.secret_key")                   //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("session.secret")                       //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authentication_backend.ldap.password") //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("notifier.smtp.password")               //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("session.redis.password")               //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("storage.mysql.password")               //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("storage.postgres.password")            //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, []error{fmt.Errorf("unable to find config file %s", configPath)}
		}
	}

	var configuration schema.Configuration
	viper.Unmarshal(&configuration) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	val := schema.NewStructValidator()
	validator.Validate(&configuration, val)
	validator.ValidateKeys(val, viper.AllKeys())

	if val.HasErrors() {
		return nil, val.Errors()
	}

	return &configuration, nil
}
