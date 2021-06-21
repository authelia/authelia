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
	fileKeys      []string
}

func (p *Provider) loadFile(path string) (err error) {
	if err := p.Load(file.Provider(path), yaml.Parser()); err != nil {
		return err
	}

	return nil
}

// LoadPaths loads the provided paths into the configuration.
func (p *Provider) LoadPaths(paths []string) (err error) {
	errs := false

	for _, path := range paths {
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				p.Push(fmt.Errorf("error loading path '%s': is not a file", path))
				errs = true
				continue
			}

			err = p.loadFile(path)
			if err != nil {
				p.Push(err)
				errs = true
				continue
			}
		} else if os.IsNotExist(err) {
			switch len(paths) {
			case 1:
				err = generateConfigFromTemplate(path)
				if err != nil {
					p.Push(fmt.Errorf("configuration file could not be generated at %s: %v", path, err))
					errs = true
					continue
				}
			default:
				p.Push(fmt.Errorf("configuration file does not exist at %s", path))
				errs = true
				continue
			}

			errs = true
			p.Push(fmt.Errorf("configuration file did not exist a default one has been generated at %s", path))
		}
	}

	p.fileKeys = p.Keys()

	if errs {
		return errors.New("one or more errors occurred while loading configuration files")
	}

	return nil
}

// LoadEnvironment loads the environment variables to the configuration.
func (p *Provider) LoadEnvironment() (err error) {
	return p.Load(env.ProviderWithValue("AUTHELIA_", ".", koanfKeyCallbackBuilder("_", ".", "AUTHELIA_")), nil)
}

// LoadSecrets loads the secrets into the struct from the path values.
func (p *Provider) LoadSecrets() (err error) {
	err = p.Load(NewSecretsProvider(".", p), nil)

	if err != nil {
		return err
	}

	return nil
}

// ValidateConfiguration runs the configuration validation tasks.
func (p *Provider) ValidateConfiguration() {
	validator.ValidateConfiguration(p.Configuration, p.StructValidator)
}

// ValidateFileAuthenticationBackend runs the configuration validation tasks on specifically the file authentication backend.
func (p *Provider) ValidateFileAuthenticationBackend() {
	validator.ValidateFileAuthenticationBackend(p.Configuration.AuthenticationBackend.File, p.StructValidator)
}

// ValidateKeys runs key validation tasks.
func (p *Provider) ValidateKeys() {
	validator.ValidateKeys(p.StructValidator, p.fileKeys)
	validator.ValidateAccessControlRuleKeys(p.StructValidator, p.Slices("access_control.rules"))
	validator.ValidateOpenIDConnectClientKeys(p.StructValidator, p.Slices("identity_providers.oidc.clients"))
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
			Delim:       ".",
			StrictMerge: false,
		}),
		StructValidator: schema.NewStructValidator(),
		Configuration:   &schema.Configuration{},
	}
}
