package configuration

import (
	"fmt"

	"github.com/knadh/koanf"
	"github.com/mitchellh/mapstructure"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
)

// LoadSources is a variadic function that takes types that implement the Source interface.
func (p *Provider) LoadSources(sources ...Source) (errs []error) {
	for _, source := range sources {
		err := source.Load()
		if err != nil {
			p.Validator.Push(fmt.Errorf("failed to load configuration from %s source: %+v", source.Name(), err))

			continue
		}

		err = source.Merge(p.Koanf)
		if err != nil {
			p.Validator.Push(fmt.Errorf("failed to merge configuration from %s source: %+v", source.Name(), err))

			continue
		}

		val := source.Validator()

		if val != nil {
			for _, err := range val.Errors() {
				p.Validator.Push(fmt.Errorf("%s: %+v", source.Name(), err))
			}
		}
	}

	return p.Validator.Errors()
}

// Unmarshal the koanf.Koanf.
func (p *Provider) Unmarshal() (warns []error, errs []error) {
	if p.validation {
		validator.ValidateKeys(p.Koanf.Keys(), p.Validator)
	}

	p.Configuration = &schema.Configuration{}

	err := p.unmarshal("", p.Configuration)
	if err != nil {
		p.Validator.Push(fmt.Errorf("error occurred during unmarshalling configuration: %+v", err))
	} else if p.validation {
		validator.ValidateConfiguration(p.Configuration, p.Validator)
	}

	return p.Validator.Warnings(), p.Validator.Errors()
}

// SetValidation changes the validation state of the provider. Disabling it allows manual checking.
func (p *Provider) SetValidation(validation bool) {
	p.validation = validation
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

	return p.Koanf.UnmarshalWithConf(path, o, c)
}

var provider *Provider

// GetProvider returns the global provider.
func GetProvider() *Provider {
	if provider == nil {
		provider = NewProvider()
	}

	return provider
}

// NewProvider creates a new Configuration provider. This is *not* the global Configuration provider and generally
// should not be used for anything other than just validating configurations.
func NewProvider() (p *Provider) {
	return &Provider{
		Koanf: koanf.NewWithConf(koanf.Conf{
			Delim:       constDelimiter,
			StrictMerge: false,
		}),
		Validator:  schema.NewStructValidator(),
		validation: true,
	}
}
