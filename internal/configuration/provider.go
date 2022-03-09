package configuration

import (
	"fmt"

	"github.com/knadh/koanf"
	"github.com/mitchellh/mapstructure"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Load the configuration given the provided options and sources.
func Load(val *schema.StructValidator, sources ...Source) (keys []string, configuration *schema.Configuration, err error) {
	configuration = &schema.Configuration{}

	keys, err = LoadAdvanced(val, "", configuration, sources...)

	return keys, configuration, err
}

// LoadAdvanced is intended to give more flexibility over loading a particular path to a specific interface.
func LoadAdvanced(val *schema.StructValidator, path string, result interface{}, sources ...Source) (keys []string, err error) {
	if val == nil {
		return keys, errNoValidator
	}

	ko := koanf.NewWithConf(koanf.Conf{
		Delim:       constDelimiter,
		StrictMerge: false,
	})

	err = loadSources(ko, val, sources...)
	if err != nil {
		return ko.Keys(), err
	}

	unmarshal(ko, val, path, result)

	return ko.Keys(), nil
}

func unmarshal(ko *koanf.Koanf, val *schema.StructValidator, path string, o interface{}) {
	c := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToSliceHookFunc(","),
				StringToMailAddressFunc(),
				ToTimeDurationFunc(),
			),
			Metadata:         nil,
			Result:           o,
			WeaklyTypedInput: true,
		},
	}

	if err := ko.UnmarshalWithConf(path, o, c); err != nil {
		val.Push(fmt.Errorf("error occurred during unmarshalling configuration: %w", err))
	}
}

func loadSources(ko *koanf.Koanf, val *schema.StructValidator, sources ...Source) (err error) {
	if len(sources) == 0 {
		return errNoSources
	}

	for _, source := range sources {
		err := source.Load(val)
		if err != nil {
			val.Push(fmt.Errorf("failed to load configuration from %s source: %+v", source.Name(), err))

			continue
		}

		err = source.Merge(ko, val)
		if err != nil {
			val.Push(fmt.Errorf("failed to merge configuration from %s source: %+v", source.Name(), err))

			continue
		}
	}

	return nil
}
