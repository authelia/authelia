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

	// we need to bind all env variables as long as https://github.com/spf13/viper/issues/761
	// is not resolved.
	viper.BindEnv("authelia.jwt_secret")
	viper.BindEnv("authelia.duo_api.secret_key")
	viper.BindEnv("authelia.session.secret")
	viper.BindEnv("authelia.authentication_backend.ldap.password")
	viper.BindEnv("authelia.notifier.smtp.password")
	viper.BindEnv("authelia.session.redis.password")
	viper.BindEnv("authelia.storage.mysql.password")
	viper.BindEnv("authelia.storage.postgres.password")

	viper.BindEnv("authelia.jwt_secret.file")
	viper.BindEnv("authelia.duo_api.secret_key.file")
	viper.BindEnv("authelia.session.secret.file")
	viper.BindEnv("authelia.authentication_backend.ldap.password.file")
	viper.BindEnv("authelia.notifier.smtp.password.file")
	viper.BindEnv("authelia.session.redis.password.file")
	viper.BindEnv("authelia.storage.mysql.password.file")
	viper.BindEnv("authelia.storage.postgres.password.file")

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, []error{fmt.Errorf("unable to find config file %s", configPath)}
		}
	}

	var configuration schema.Configuration
	viper.Unmarshal(&configuration)

	val := schema.NewStructValidator()
	validator.ValidateSecrets(&configuration, val, viper.GetViper())
	validator.Validate(&configuration, val)

	if val.HasErrors() {
		return nil, val.Errors()
	}

	return &configuration, nil
}
