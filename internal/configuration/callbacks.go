package configuration

import (
	"strings"

	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/utils"
)

var excludedKeys = []string{
	"config",
}

func isKeyExcluded(key string) (isExcluded bool) {
	return utils.IsStringInSlice(key, excludedKeys)
}

func koanfPosFlagCallbackFunc(key string, value string) (string, interface{}) {
	if isKeyExcluded(key) {
		return "", nil
	}

	return strings.ReplaceAll(key, "-", "_"), value
}

func koanfEnvCallback() func(key string, value string) (string, interface{}) {
	keyReplacements := map[string]string{}

	for _, key := range validator.ValidKeys {
		if strings.Contains(key, "_") {
			keyReplacements[strings.ReplaceAll(key, "_", ".")] = key
		}
	}

	for _, key := range validator.SecretNames {
		if strings.Contains(key, "_") {
			keyReplacements[strings.ReplaceAll(key, "_", ".")] = key
			keyReplacements[validator.SecretNameToEnvName(strings.ReplaceAll(key, "_", "."))] = validator.SecretNameToEnvName(key)
		}
	}

	return func(key string, value string) (string, interface{}) {
		if isKeyExcluded(key) {
			return "", nil
		}

		formattedKey := strings.ReplaceAll(strings.ToLower(key), "_", ".")

		if replacedKey, ok := keyReplacements[formattedKey]; ok {
			formattedKey = replacedKey
		}

		if validator.IsSecretKey(formattedKey) {
			return validator.KoanfEnvSecretKeyReplacer(formattedKey), value
		}

		return strings.TrimPrefix(formattedKey, "authelia."), value
	}
}
