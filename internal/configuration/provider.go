package configuration

import (
	"errors"

	"github.com/knadh/koanf"
	"github.com/mitchellh/mapstructure"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
)

// LoadSources is a variadic function that takes Source structs.
func (p *Provider) LoadSources(sources ...Source) (err error) {
	for _, source := range sources {
		if err = p.Load(source.Provider, source.Parser); err != nil {
			return err
		}
	}

	if p.HasErrors() {
		return errors.New("errors occurred during loading configuration sources")
	}

	return nil
}

// Configuration returns the configuration.
func (p *Provider) Configuration() (configuration *schema.Configuration) {
	return p.configuration
}

// Validate runs the validation tasks.
func (p *Provider) Validate() {
	validator.ValidateKeys(p.StructValidator, p.Keys())
	validator.ValidateConfiguration(p.configuration, p.StructValidator)
}

// UnmarshalToConfiguration unmarshalls the koanf.Koanf to the global configuration struct ptr.
func (p *Provider) UnmarshalToConfiguration() (err error) {
	return p.unmarshal("", p.configuration)
}

func (p *Provider) unmarshal(path string, o interface{}) (err error) {
	c := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
			),
			Metadata:         nil,
			Result:           o,
			WeaklyTypedInput: true,
		},
	}

	return p.UnmarshalWithConf(path, o, c)
}

/*
// LoadAll loads all of the configuration sources.
func (p *Provider) LoadAll(paths []string) (err error) {
	err = p.LoadPaths(paths)
	if err != nil {
		return err
	}

	err = p.LoadEnvironment()
	if err != nil {
		return err
	}

	err = p.LoadSecrets()
	if err != nil {
		return err
	}

	return nil
}

// LoadEnvironment loads the environment variables to the configuration.
func (p *Provider) LoadEnvironment() (err error) {
	keyMap := getEnvConfigMap(validator.ValidKeys)

	return p.Load(env.ProviderWithValue(envPrefix, delimiter, koanfEnvironmentCallback(keyMap)), nil)
}

// LoadSecrets loads the secrets into the struct from the path values.
func (p *Provider) LoadSecrets() (err error) {
	keyMap := getSecretConfigMap(validator.ValidKeys)

	return p.Load(env.ProviderWithValue(envPrefixAlt, delimiter, koanfEnvironmentSecretsCallback(keyMap)), nil)
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

func (p *Provider) loadFile(path string) (err error) {
	return p.Load(file.Provider(path), yaml.Parser())
}.

*/

var provider *Provider

// GetProvider returns the global provider.
func GetProvider() *Provider {
	if provider == nil {
		provider = NewProvider()
	}

	return provider
}

// NewProvider creates a new Configuration provider. This is *not* the global configuration provider and generally
// should not be used for anything other than just validating configurations.
func NewProvider() (p *Provider) {
	return &Provider{
		Koanf: koanf.NewWithConf(koanf.Conf{
			Delim:       delimiter,
			StrictMerge: false,
		}),
		StructValidator: schema.NewStructValidator(),
		configuration:   &schema.Configuration{},
	}
}
