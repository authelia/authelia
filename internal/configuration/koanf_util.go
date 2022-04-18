package configuration

import (
	"fmt"

	"github.com/knadh/koanf"

	"github.com/authelia/authelia/v4/internal/utils"
)

func getAllKoanfKeys(ko *koanf.Koanf) (keys []string) {
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
