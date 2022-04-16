package configuration

import (
	"fmt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
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

	if err = loadSources(ko, val, sources...); err != nil {
		return ko.Keys(), err
	}

	var final *koanf.Koanf

	if final, err = remap(val, ko); err != nil {
		return ko.Keys(), err
	}

	unmarshal(final, val, path, result)

	return getAllKoanfKeys(final), nil
}

func remap(val *schema.StructValidator, ko *koanf.Koanf) (final *koanf.Koanf, err error) {
	keysFinal := make(map[string]interface{})

	keysCurrent := ko.All()

	for key, value := range keysCurrent {
		if deprecation, ok := deprecations[key]; ok {
			if !deprecation.AutoMap {
				val.Push(fmt.Errorf("invalid configuration key '%s' was replaced by '%s'", deprecation.Key, deprecation.NewKey))
			} else {
				val.PushWarning(fmt.Errorf("configuration key '%s' is deprecated in %s and has been replaced by '%s': "+
					"this has been automatically mapped for you but you will need to adjust your configuration to remove this message", deprecation.Key, deprecation.Version.String(), deprecation.NewKey))
			}

			if !mapHasKey(deprecation.NewKey, keysCurrent) && !mapHasKey(deprecation.NewKey, keysFinal) {
				if deprecation.MapFunc != nil {
					keysFinal[deprecation.NewKey] = deprecation.MapFunc(value)
				} else {
					keysFinal[deprecation.NewKey] = value
				}
			}

			continue
		}

		keysFinal[key] = value
	}

	final = koanf.New(".")

	if err = final.Load(confmap.Provider(keysFinal, "."), nil); err != nil {
		return nil, err
	}

	return final, nil
}

func mapHasKey(k string, m map[string]interface{}) bool {
	if _, ok := m[k]; ok {
		return true
	}

	return false
}

func unmarshal(ko *koanf.Koanf, val *schema.StructValidator, path string, o interface{}) {
	c := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToSliceHookFunc(","),
				StringToMailAddressHookFunc(),
				ToTimeDurationHookFunc(),
				StringToURLHookFunc(),
				StringToRegexpFunc(),
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
