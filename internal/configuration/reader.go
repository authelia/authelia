package configuration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
)

// Read a YAML configuration and create a Configuration object out of it.
func Read(configPath string) (*schema.Configuration, []error) {
	if configPath == "" {
		return nil, []error{errors.New("No config file path provided")}
	}

	_, err := os.Stat(configPath)
	if err != nil {
		return nil, []error{fmt.Errorf("Unable to find config file: %v", configPath)}
	}

	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, []error{fmt.Errorf("Failed to %v", err)}
	}

	var data interface{}

	err = yaml.Unmarshal(file, &data)
	if err != nil {
		return nil, []error{fmt.Errorf("Error malformed %v", err)}
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.BindEnv("authelia.jwt_secret.file")                           //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.duo_api.secret_key.file")                   //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.session.secret.file")                       //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.authentication_backend.ldap.password.file") //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.notifier.smtp.password.file")               //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.session.redis.password.file")               //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.storage.mysql.password.file")               //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	viper.BindEnv("authelia.storage.postgres.password.file")            //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	viper.SetConfigFile(configPath)

	_ = viper.ReadInConfig()

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
