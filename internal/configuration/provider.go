package configuration

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
)

var provider *Provider

// Provider holds the koanf.Koanf instance and the schema.Configuration.
type Provider struct {
	*koanf.Koanf
	*schema.StructValidator
	Configuration *schema.Configuration
	fileKeys      []string
}

// LoadFile loads the provided paths into the configuration.
func (p *Provider) LoadFile(paths []string) (err error) {
	for _, path := range paths {
		if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
			if err := p.Load(file.Provider(path), yaml.Parser()); err != nil {
				return err
			}

			continue
		}

		if info, err := os.Stat(path); err == nil && info.IsDir() {
			files, err := ioutil.ReadDir(path)
			if err != nil {
				return err
			}

			noConfigs := true

			for _, f := range files {
				if f.IsDir() {
					continue
				}

				name := f.Name()

				if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
					if err := p.Load(file.Provider(name), yaml.Parser()); err != nil {
						return err
					}

					noConfigs = false
				}
			}

			if noConfigs {
				return fmt.Errorf("path does not contain Configuration files: %s", path)
			}
		}
	}

	copy(p.fileKeys, p.Keys())

	return nil
}

// LoadEnvironment loads the environment variables to the configuration.
func (p *Provider) LoadEnvironment() (err error) {
	return p.Load(env.ProviderWithValue("AUTHELIA_", ".", koanfEnvCallback()), nil)
}

// LoadCommandLineArguments loads the CLI args to the configuration.
func (p *Provider) LoadCommandLineArguments(flags *pflag.FlagSet) (err error) {
	if flags != nil {
		if err := p.Load(posflag.ProviderWithValue(flags, ".", p.Koanf, koanfPosFlagCallbackFunc), nil); err != nil {
			return err
		}
	}

	return nil
}

// Validate runs all of the configuration validation tasks.
func (p *Provider) Validate() {
	validator.ValidateKeys(p.StructValidator, p.fileKeys)
	validator.ValidateAccessControlRuleKeys(p.StructValidator, p.Slices("access_control.rules"))
	validator.ValidateOpenIDConnectClientKeys(p.StructValidator, p.Slices("identity_providers.oidc.clients"))

	validator.ValidateSecrets(p.Configuration, p.StructValidator, p.Koanf)
	validator.ValidateConfiguration(p.Configuration, p.StructValidator)
	p.fileKeys = nil
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

// GetProvider returns the Configuration provider and the Configuration.
func GetProvider() *Provider {
	if provider == nil {
		provider = &Provider{
			Koanf: koanf.NewWithConf(koanf.Conf{
				Delim:       ".",
				StrictMerge: true,
			}),
			StructValidator: schema.NewStructValidator(),
			Configuration:   &schema.Configuration{},
		}
	}

	return provider
}
