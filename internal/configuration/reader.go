package configuration

import (
	_ "embed" // Embed config.template.yml.
	"errors"
	"fmt"
	"github.com/knadh/koanf/providers/posflag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/logging"
)

var config *schema.Configuration
var konfiguration *koanf.Koanf

// GetKoanf returns the Configuration provider.
func GetKoanf() *koanf.Koanf {
	if konfiguration == nil {
		konfiguration = koanf.NewWithConf(koanf.Conf{
			Delim:       ".",
			StrictMerge: true,
		})
	}

	return konfiguration
}

func GetConfiguration(fallback *schema.Configuration) *schema.Configuration {
	if config == nil {
		if fallback == nil {
			config = &schema.Configuration{}
		} else {
			config = fallback
		}
	}

	return config
}

func Load(paths []string, flags *pflag.FlagSet) (configuration *schema.Configuration, err error) {
	konfig := GetKoanf()

	configuration = GetConfiguration(nil)

	val := schema.NewStructValidator()
	if len(paths) != 0 {
		validator.ValidateKeys(val, konfig.Keys())
	}

	if err := konfig.Load(env.ProviderWithValue("AUTHELIA_", ".", koanfEnvCallback()), nil); err != nil {
		return configuration, err
	}

	if flags != nil {
		if err := konfig.Load(posflag.ProviderWithValue(flags, ".", konfig, koanfPosFlagCallbackFunc), nil); err != nil {
			return configuration, err
		}
	}

	if err := konfig.UnmarshalWithConf("", configuration, unmarshallConfig(configuration)); err != nil {
		return configuration, err
	}

	validator.ValidateSecrets(configuration, val, konfig)
	validator.ValidateConfiguration(configuration, val)

	if val.HasErrors() {
		s := strings.Builder{}
		s.WriteString("Errors during Configuration validation: \n")
		for _, err := range val.Errors() {
			s.WriteString(err.Error())
		}

		return configuration, errors.New(s.String())
	}

	if val.HasWarnings() {
		logger := logging.Logger()
		for _, warn := range val.Warnings() {
			logger.Warnf(warn.Error())
		}
	}

	return configuration, nil
}

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

	val := schema.NewStructValidator()
	validator.ValidateKeys(val, konfig.Keys())

	if err := konfig.Load(env.ProviderWithValue("AUTHELIA_", ".", koanfEnvCallback()), nil); err != nil {
		errs = append(errs, err)
	}

	configuration = &schema.Configuration{}

	if err := konfig.UnmarshalWithConf("", configuration, unmarshallConfig(configuration)); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return nil, errs
	}

	validator.ValidateSecrets(configuration, val, konfig)
	validator.ValidateConfiguration(configuration, val)

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
