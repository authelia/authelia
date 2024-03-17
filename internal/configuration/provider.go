package configuration

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Load the configuration given the provided options and sources.
func Load(val *schema.StructValidator, sources ...Source) (keys []string, configuration *schema.Configuration, err error) {
	configuration = &schema.Configuration{}

	keys, err = LoadAdvanced(val, "", configuration, sources...)

	return keys, configuration, err
}

// LoadAdvanced is intended to give more flexibility over loading a particular path to a specific interface.
func LoadAdvanced(val *schema.StructValidator, path string, result any, sources ...Source) (keys []string, err error) {
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

	if final, err = koanfRemapKeys(val, ko, deprecations, deprecationsMKM); err != nil {
		return koanfGetKeys(ko), err
	}

	unmarshal(final, val, path, result)

	return koanfGetKeys(final), nil
}

func mapHasKey(k string, m map[string]any) bool {
	if _, ok := m[k]; ok {
		return true
	}

	return false
}

func unmarshal(ko *koanf.Koanf, val *schema.StructValidator, path string, o any) {
	c := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToSliceHookFunc(","),
				StringToMailAddressHookFunc(),
				StringToURLHookFunc(),
				StringToRegexpHookFunc(),
				StringToAddressHookFunc(),
				StringToX509CertificateHookFunc(),
				StringToX509CertificateChainHookFunc(),
				StringToPrivateKeyHookFunc(),
				StringToCryptoPrivateKeyHookFunc(),
				StringToCryptographicKeyHookFunc(),
				StringToTLSVersionHookFunc(),
				StringToPasswordDigestHookFunc(),
				ToTimeDurationHookFunc(),
				ToRefreshIntervalDurationHookFunc(),
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
		if err = source.Load(val); err != nil {
			val.Push(fmt.Errorf("failed to load configuration from %s source: %+v", source.Name(), err))

			continue
		}

		if err = source.Merge(ko, val); err != nil {
			val.Push(fmt.Errorf("failed to merge configuration from %s source: %+v", source.Name(), err))

			continue
		}
	}

	return nil
}
