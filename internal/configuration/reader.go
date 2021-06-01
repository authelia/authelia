package configuration

import (
	_ "embed" // Embed config.template.yml.
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/logging"
)

// Read a YAML configuration and create a Configuration object out of it.
func Read(configPath string) (configuration *schema.Configuration, errs []error) {
	if configPath == "" {
		return nil, []error{errors.New("No config file path provided")}
	}

	errs = ensureConfigFileExists(configPath)
	if len(errs) != 0 {
		return nil, errs
	}

	konfig := koanf.NewWithConf(koanf.Conf{
		Delim:       ".",
		StrictMerge: false,
	})

	if err := konfig.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		errs = append(errs, err)
	}

	if err := konfig.Load(env.ProviderWithValue("AUTHELIA_", ".", koanfSecretEnvParser()), nil); err != nil {
		errs = append(errs, err)
	}

	configuration = &schema.Configuration{}

	if err := konfig.UnmarshalWithConf("", configuration, unmarshallConfig(configuration)); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return nil, errs
	}

	val := schema.NewStructValidator()
	validator.ValidateSecrets(configuration, val, konfig)
	validator.ValidateConfiguration(configuration, val)
	validator.ValidateKeys(val, konfig.Keys())

	if val.HasErrors() {
		return nil, val.Errors()
	}

	if val.HasWarnings() {
		logger := logging.Logger()
		for _, warn := range val.Warnings() {
			logger.Warnf(warn.Error())
		}
	}

	return configuration, nil
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

func ensureConfigFileExists(path string) (errs []error) {
	if _, err := os.Stat(path); err != nil {
		errs = []error{
			fmt.Errorf("Unable to find config file: %s", path),
			fmt.Errorf("Generating config file: %s", path),
		}

		if err := generateConfigFromTemplate(path); err != nil {
			errs = append(errs, err)
		} else {
			errs = append(errs, fmt.Errorf("Generated configuration at: %v", path))
		}

		return errs
	}

	if _, err := ioutil.ReadFile(path); err != nil {
		errs = append(errs, fmt.Errorf("Failed to read file: %+v", err))
		return errs
	}

	return errs
}
