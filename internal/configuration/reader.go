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
	_ = viper.BindEnv("authelia.jwt_secret")
	_ = viper.BindEnv("authelia.duo_api.secret_key")
	_ = viper.BindEnv("authelia.session.secret")
	_ = viper.BindEnv("authelia.authentication_backend.ldap.password")
	_ = viper.BindEnv("authelia.notifier.smtp.password")
	_ = viper.BindEnv("authelia.session.redis.password")
	_ = viper.BindEnv("authelia.storage.mysql.password")
	_ = viper.BindEnv("authelia.storage.postgres.password")

	_ = viper.BindEnv("authelia.jwt_secret.file")
	_ = viper.BindEnv("authelia.duo_api.secret_key.file")
	_ = viper.BindEnv("authelia.session.secret.file")
	_ = viper.BindEnv("authelia.authentication_backend.ldap.password.file")
	_ = viper.BindEnv("authelia.notifier.smtp.password.file")
	_ = viper.BindEnv("authelia.session.redis.password.file")
	_ = viper.BindEnv("authelia.storage.mysql.password.file")
	_ = viper.BindEnv("authelia.storage.postgres.password.file")

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, []error{fmt.Errorf("unable to find config file %s", configPath)}
		}
	}

	var configuration schema.Configuration
	_ = viper.Unmarshal(&configuration)

	val := schema.NewStructValidator()
	validator.ValidateSecrets(&configuration, val, viper.GetViper())
	validator.Validate(&configuration, val)

	if val.HasErrors() {
		return nil, val.Errors()
	}

	return &configuration, nil
}
