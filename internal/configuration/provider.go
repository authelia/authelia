package configuration

import (
	"errors"
	"fmt"
	"os"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/mitchellh/mapstructure"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
)

// Provider holds the koanf.Koanf instance and the schema.Configuration.
type Provider struct {
	*koanf.Koanf
	*schema.StructValidator

	Configuration *schema.Configuration
}

func (p *Provider) loadFile(path string) (err error) {
	return p.Load(file.Provider(path), yaml.Parser())
}

// LoadPaths loads the provided paths into the configuration.
func (p *Provider) LoadPaths(paths []string) (err error) {
	errs := false

	for _, path := range paths {
		info, osErr := os.Stat(path)

		switch {
		case osErr == nil:
			if info.IsDir() {
				p.Push(fmt.Errorf("error loading path '%s': is not a file", path))

				errs = true

				continue
			}

			err = p.loadFile(path)
			if err != nil {
				p.Push(fmt.Errorf("configuration file could not be loaded due to an error: %v", err))

				errs = true

				continue
			}
		case os.IsNotExist(osErr):
			switch len(paths) {
			case 1:
				errs = true

				err = generateConfigFromTemplate(path)
				if err != nil {
					p.Push(fmt.Errorf("configuration file could not be generated at %s: %v", path, err))

					continue
				}

				p.Push(fmt.Errorf("configuration file did not exist at %s and generated with defaults but you will need to configure it", path))
			default:
				p.Push(fmt.Errorf("configuration file does not exist at %s", path))

				errs = true

				continue
			}
		default:
			p.Push(fmt.Errorf("configuration file could not be loaded due to an error: %v", osErr))

			errs = true

			continue
		}
	}

	if errs {
		return errors.New("one or more errors occurred while loading configuration files")
	}

	return nil
}

// LoadEnvironment loads the environment variables to the configuration.
func (p *Provider) LoadEnvironment() (err error) {
	return p.Load(env.ProviderWithValue(envPrefixAlt, delimiter, koanfKeyCallbackBuilder()), nil)
}

// LoadSecrets loads the secrets into the struct from the path values.
func (p *Provider) LoadSecrets() (err error) {
	return p.Load(NewSecretsProvider(p), nil)
}

// ValidateConfiguration runs the configuration validation tasks.
func (p *Provider) ValidateConfiguration() {
	validator.ValidateKeys(p.StructValidator, p.Keys())
	validator.ValidateConfiguration(p.Configuration, p.StructValidator)
}

// UnmarshalToStruct unmarshalls the configuration to the struct.
func (p *Provider) UnmarshalToStruct() (err error) {
	conf := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
			),
			Metadata:         nil,
			Result:           p.Configuration,
			WeaklyTypedInput: true,
		},
	}

	return p.UnmarshalWithConf("", p.Configuration, conf)
}

var confProvider *Provider

// GetProvider returns the global Configuration provider.
func GetProvider() *Provider {
	if confProvider == nil {
		confProvider = NewProvider()
	}

	return confProvider
}

// NewProvider creates a new Configuration provider.
func NewProvider() (p *Provider) {
	return &Provider{
		Koanf: koanf.NewWithConf(koanf.Conf{
			Delim:       delimiter,
			StrictMerge: false,
		}),
		StructValidator: schema.NewStructValidator(),
		Configuration:   &schema.Configuration{},
	}
}
