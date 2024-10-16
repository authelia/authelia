package configuration

import (
	"fmt"
	"net"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Load the configuration given the provided options and sources.
func Load(val *schema.StructValidator, sources ...Source) (keys []string, configuration *schema.Configuration, err error) {
	configuration = &schema.Configuration{}

	keys, err = LoadAdvanced(val, "", configuration, nil, sources...)

	return keys, configuration, err
}

// LoadAdvanced is intended to give more flexibility over loading a particular path to a specific interface.
func LoadAdvanced(val *schema.StructValidator, path string, result any, definitions *schema.Definitions, sources ...Source) (keys []string, err error) {
	if val == nil {
		return keys, errNoValidator
	}

	ko := koanf.NewWithConf(koanf.Conf{Delim: constDelimiter, StrictMerge: false})

	if err = loadSources(ko, val, sources...); err != nil {
		return ko.Keys(), err
	}

	var final *koanf.Koanf

	if final, err = koanfRemapKeys(val, ko, deprecations, deprecationsMKM); err != nil {
		return koanfGetKeys(ko), err
	}

	unmarshal(final, val, path, result, definitions)

	mapDefinitionsResult(val, result)

	return koanfGetKeys(final), nil
}

func LoadDefinitions(val *schema.StructValidator, sources ...Source) (definitions *schema.Definitions, err error) {
	ko := koanf.NewWithConf(koanf.Conf{Delim: constDelimiter, StrictMerge: false})

	if err = loadSources(ko, val, sources...); err != nil {
		return nil, err
	}

	var final *koanf.Koanf

	if final, err = koanfRemapKeys(val, ko, deprecations, deprecationsMKM); err != nil {
		return nil, err
	}

	legacy := &legacyDefinitions{}

	c := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToSliceHookFunc(","),
				StringToIPNetworksHookFunc(nil),
			),
			Metadata:         nil,
			Result:           legacy,
			WeaklyTypedInput: true,
		},
	}

	if err = final.UnmarshalWithConf("", legacy, c); err != nil {
		val.Push(fmt.Errorf("error occurred during unmarshalling definitions configuration: %w", err))
	}

	d := legacy.Definitions

	mapDefinitions(val, legacy.AccessControl.Networks, &d)

	return &d, nil
}

func mapDefinitionsResult(val *schema.StructValidator, result any) {
	if config, ok := result.(*schema.Configuration); ok {
		mapDefinitions(val, config.AccessControl.Networks, &config.Definitions)
	}
}

func mapDefinitions(val *schema.StructValidator, networks []schema.AccessControlNetwork, definitions *schema.Definitions) {
	var ok bool

	if definitions.Network == nil {
		definitions.Network = map[string][]*net.IPNet{}
	}

	for _, network := range networks {
		if _, ok = definitions.Network[network.Name]; ok {
			val.Push(fmt.Errorf("error occurred during unmarshalling definitions configuration: the definition for network with name '%s' exists in both the defintions section and access control section which is not permitted", network.Name))

			continue
		}

		definitions.Network[network.Name] = network.Networks
	}
}

type legacyDefinitions struct {
	Definitions   schema.Definitions  `koanf:"definitions"`
	AccessControl legacyAccessControl `koanf:"access_control"`
}

type legacyAccessControl struct {
	Networks []schema.AccessControlNetwork `koanf:"networks"`
}

func mapHasKey(k string, m map[string]any) bool {
	if _, ok := m[k]; ok {
		return true
	}

	return false
}

func unmarshal(ko *koanf.Koanf, val *schema.StructValidator, path string, o any, definitions *schema.Definitions) {
	if definitions == nil {
		definitions = &schema.Definitions{}
	}

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
				StringToIPNetworksHookFunc(definitions.Network),
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
