package configuration

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/spf13/viper"
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
	viper.AutomaticEnv()

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
