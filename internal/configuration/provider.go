package configuration

import (
	"fmt"

	"github.com/knadh/koanf"
	"github.com/mitchellh/mapstructure"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// Load the configuration given the provided options and sources.
func Load(val *schema.StructValidator, sources ...Source) (keys []string, configuration *schema.Configuration) {
	ko := koanf.NewWithConf(koanf.Conf{
		Delim:       constDelimiter,
		StrictMerge: false,
	})

	loadSources(ko, val, sources...)

	configuration = &schema.Configuration{}

	unmarshal(ko, val, "", configuration)

	return ko.Keys(), configuration
}

func unmarshal(ko *koanf.Koanf, val *schema.StructValidator, path string, o interface{}) {
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

	if err := ko.UnmarshalWithConf(path, o, c); err != nil {
		val.Push(fmt.Errorf("error occurred during unmarshalling configuration: %w", err))
	}
}

func loadSources(ko *koanf.Koanf, val *schema.StructValidator, sources ...Source) {
	if len(sources) == 0 {
		val.Push(errNoSources)
		return
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
}
