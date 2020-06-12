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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.BindEnv("authelia.jwt_secret.file")                           //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.duo_api.secret_key.file")                   //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.session.secret.file")                       //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.authentication_backend.ldap.password.file") //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.notifier.smtp.password.file")               //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.session.redis.password.file")               //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.storage.mysql.password.file")               //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.storage.postgres.password.file")            //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.openid_connect.oauth2_hmac_secret.file")    //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, []error{fmt.Errorf("unable to find config file %s", configPath)}
		}
	}

	var configuration schema.Configuration

	viper.Unmarshal(&configuration) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	val := schema.NewStructValidator()
	validator.ValidateSecrets(&configuration, val, viper.GetViper())
	validator.ValidateConfiguration(&configuration, val)
	validator.ValidateKeys(val, viper.AllKeys())

	if val.HasErrors() {
		return nil, val.Errors()
	}

	return &configuration, nil
}
