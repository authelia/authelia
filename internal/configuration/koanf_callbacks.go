package configuration

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// koanfEnvironmentCallback returns a koanf callback to map the environment vars to Configuration keys.
func koanfEnvironmentCallback(keyMap map[string]string, ignoredKeys []string, prefix, delimiter string) func(key, value string) (finalKey string, finalValue any) {
	return func(key, value string) (finalKey string, finalValue any) {
		if k, ok := keyMap[key]; ok {
			return k, value
		}

		if utils.IsStringInSlice(key, ignoredKeys) {
			return "", nil
		}

		formattedKey := strings.TrimPrefix(key, prefix)
		formattedKey = strings.ReplaceAll(strings.ToLower(formattedKey), delimiter, constDelimiter)

		if utils.IsStringInSlice(formattedKey, schema.Keys) {
			return formattedKey, value
		}

		return key, value
	}
}

// koanfEnvironmentSecretsCallback returns a koanf callback to map the environment vars to Configuration keys.
func koanfEnvironmentSecretsCallback(keyMap map[string]string, validator *schema.StructValidator) func(key, value string) (finalKey string, finalValue any) {
	return func(key, value string) (finalKey string, finalValue any) {
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

func koanfCommandLineWithMappingCallback(mapping map[string]string, includeValidKeys, includeUnchangedKeys bool) func(flag *pflag.Flag) (string, any) {
	return func(flag *pflag.Flag) (string, any) {
		if !includeUnchangedKeys && !flag.Changed {
			return "", nil
		}

		if actualKey, ok := mapping[flag.Name]; ok {
			return actualKey, flag.Value.String()
		}

		if includeValidKeys {
			formattedKey := strings.ReplaceAll(flag.Name, "-", "_")

			if utils.IsStringInSlice(formattedKey, schema.Keys) {
				return formattedKey, flag.Value.String()
			}
		}

		return "", nil
	}
}
