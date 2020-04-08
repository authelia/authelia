package configuration

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Read a YAML configuration and create a Configuration object out of it.
func Read(configPath string) (*schema.Configuration, []error) {
	viper.SetEnvPrefix("AUTHELIA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// we need to bind all env variables as long as https://github.com/spf13/viper/issues/761
	// is not resolved.
	viper.BindEnv("jwt_secret")
	viper.BindEnv("duo_api.secret_key")
	viper.BindEnv("session.secret")
	viper.BindEnv("authentication_backend.ldap.password")
	viper.BindEnv("notifier.smtp.password")
	viper.BindEnv("session.redis.password")
	viper.BindEnv("storage.mysql.password")
	viper.BindEnv("storage.postgres.password")

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, []error{fmt.Errorf("unable to find config file %s", configPath)}
		}
	}

	var configuration schema.Configuration
	viper.Unmarshal(&configuration)

	val := schema.NewStructValidator()
	validator.Validate(&configuration, val)

	if val.HasErrors() {
		return nil, val.Errors()
	}

	return &configuration, nil
}
