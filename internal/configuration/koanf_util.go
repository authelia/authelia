package configuration

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func getKoanfKeys(ko *koanf.Koanf) (keys []string) {
	keys = ko.Keys()

	for key, value := range ko.All() {
		slc, ok := value.([]interface{})
		if !ok {
			continue
		}

		for _, item := range slc {
			m, mok := item.(map[string]interface{})
			if !mok {
				continue
			}

			for k := range m {
				full := fmt.Sprintf("%s[].%s", key, k)
				if !utils.IsStringInSlice(full, keys) {
					keys = append(keys, full)
				}
			}
		}
	}

	return keys
}

func remapKoanfKeys(val *schema.StructValidator, ko *koanf.Koanf) (final *koanf.Koanf, err error) {
	keys := ko.All()

	keys = remapKoanfKeysStandard(keys, val)
	keys = remapKoanfKeysMapped(keys, val)

	final = koanf.New(".")

	if err = final.Load(confmap.Provider(keys, "."), nil); err != nil {
		return nil, err
	}

	return final, nil
}

func remapKoanfKeysMapped(keys map[string]interface{}, val *schema.StructValidator) (keysFinal map[string]interface{}) {
	keysFinal = make(map[string]interface{})

	var (
		slc []interface{}
		ok  bool
		m   map[string]interface{}
		d   Deprecation
	)

	for key, value := range keys {
		if slc, ok = value.([]interface{}); !ok {
			keysFinal[key] = value

			continue
		}

		sliceFinal := make([]interface{}, len(slc))

		for i, item := range slc {
			if m, ok = item.(map[string]interface{}); !ok {
				sliceFinal[i] = item

				continue
			}

			itemFinal := make(map[string]interface{})

			for subkey, element := range m {
				prefix := fmt.Sprintf("%s[].", key)

				fullKey := prefix + subkey

				if d, ok = deprecations[fullKey]; ok {
					if !d.AutoMap {
						val.Push(fmt.Errorf("invalid configuration key '%s' was replaced by '%s'", d.Key, d.NewKey))

						itemFinal[subkey] = element

						continue
					} else {
						val.PushWarning(fmt.Errorf("configuration key '%s' is deprecated in %s and has been replaced by '%s': "+
							"this has been automatically mapped for you but you will need to adjust your configuration to remove this message", d.Key, d.Version.String(), d.NewKey))
					}

					newkey := strings.Replace(d.NewKey, prefix, "", 1)

					if !mapHasKey(newkey, m) && !mapHasKey(newkey, itemFinal) {
						if d.MapFunc != nil {
							itemFinal[newkey] = d.MapFunc(element)
						} else {
							itemFinal[newkey] = element
						}
					}
				} else {
					itemFinal[subkey] = element
				}
			}

			sliceFinal[i] = itemFinal
		}

		keysFinal[key] = sliceFinal
	}

	return keysFinal
}

func remapKoanfKeysStandard(keys map[string]interface{}, val *schema.StructValidator) (keysFinal map[string]interface{}) {
	keysFinal = make(map[string]interface{})

	for key, value := range keys {
		if d, ok := deprecations[key]; ok {
			if !d.AutoMap {
				val.Push(fmt.Errorf("invalid configuration key '%s' was replaced by '%s'", d.Key, d.NewKey))

				keysFinal[key] = value

				continue
			} else {
				val.PushWarning(fmt.Errorf("configuration key '%s' is deprecated in %s and has been replaced by '%s': "+
					"this has been automatically mapped for you but you will need to adjust your configuration to remove this message", d.Key, d.Version.String(), d.NewKey))
			}

			if !mapHasKey(d.NewKey, keys) && !mapHasKey(d.NewKey, keysFinal) {
				if d.MapFunc != nil {
					keysFinal[d.NewKey] = d.MapFunc(value)
				} else {
					keysFinal[d.NewKey] = value
				}
			}

			continue
		}

		keysFinal[key] = value
	}

	return keysFinal
}
