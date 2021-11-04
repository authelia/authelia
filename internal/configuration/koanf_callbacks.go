package configuration

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/utils"
)

// koanfEnvironmentCallback returns a koanf callback to map the environment vars to Configuration keys.
func koanfEnvironmentCallback(keyMap map[string]string, ignoredKeys []string, prefix, delimiter string) func(key, value string) (finalKey string, finalValue interface{}) {
	return func(key, value string) (finalKey string, finalValue interface{}) {
		if k, ok := keyMap[key]; ok {
			return k, value
		}

		if utils.IsStringInSlice(key, ignoredKeys) {
			return "", nil
		}

		formattedKey := strings.TrimPrefix(key, prefix)
		formattedKey = strings.ReplaceAll(strings.ToLower(formattedKey), delimiter, constDelimiter)

		if utils.IsStringInSlice(formattedKey, validator.ValidKeys) {
			return formattedKey, value
		}

		return key, value
	}
}

// koanfEnvironmentSecretsCallback returns a koanf callback to map the environment vars to Configuration keys.
func koanfEnvironmentSecretsCallback(keyMap map[string]string, validator *schema.StructValidator) func(key, value string) (finalKey string, finalValue interface{}) {
	return func(key, value string) (finalKey string, finalValue interface{}) {
		k, ok := keyMap[key]
		if !ok {
			return "", nil
		}

		v, err := loadSecret(value)
		if err != nil {
			validator.Push(fmt.Errorf(errFmtSecretIOIssue, value, k, err))
			return k, ""
		}

		return k, v
	}
}

func koanfCommandLineCallback(key, value string) (string, interface{}) {
	formattedKey := strings.ReplaceAll(key, "-", "_")

	if !utils.IsStringInSlice(formattedKey, validator.ValidKeys) {
		return "", nil
	}

	return formattedKey, value
}

func koanfCommandLineWithPrefixesCallback(delimiter string, prefixes []string) func(key, value string) (string, interface{}) {
	return func(key, value string) (string, interface{}) {
		formattedKey := strings.ReplaceAll(key, "-", "_")
		for _, prefix := range prefixes {
			actualKey := fmt.Sprintf("%s%s%s", prefix, delimiter, formattedKey)

			if utils.IsStringInSlice(actualKey, validator.ValidKeys) {
				return actualKey, value
			}
		}

		return "", nil
	}
}
