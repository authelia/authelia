package configuration

import (
	_ "embed" // Embed config.template.yml.
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/logging"
)

// Read a YAML configuration and create a Configuration object out of it.
func Read(configPath string) (*schema.Configuration, []error) {
	logger := logging.Logger()

	if configPath == "" {
		return nil, []error{errors.New("No config file path provided")}
	}

	_, err := os.Stat(configPath)
	if err != nil {
		errs := []error{
			fmt.Errorf("Unable to find config file: %v", configPath),
			fmt.Errorf("Generating config file: %v", configPath),
		}

		err = generateConfigFromTemplate(configPath)
		if err != nil {
			errs = append(errs, err)
		} else {
			errs = append(errs, fmt.Errorf("Generated configuration at: %v", configPath))
		}

		return nil, errs
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
	viper.BindEnv("authelia.oauth.oidc_server.hmac_secret.file")        //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

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

	if val.HasWarnings() {
		for _, warn := range val.Warnings() {
			logger.Warnf(warn.Error())
		}
	}

	return &configuration, nil
}

//go:embed config.template.yml
var cfg []byte

func generateConfigFromTemplate(configPath string) error {
	err := ioutil.WriteFile(configPath, cfg, 0600)
	if err != nil {
		return fmt.Errorf("Unable to generate %v: %v", configPath, err)
	}

	return nil
}
