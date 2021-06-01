package configuration

import (
	"github.com/knadh/koanf"
	"github.com/mitchellh/mapstructure"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func unmarshallConfig(result *schema.Configuration) koanf.UnmarshalConf {
	return koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
			),
			Metadata:         nil,
			Result:           result,
			WeaklyTypedInput: true,
		},
	}
}
