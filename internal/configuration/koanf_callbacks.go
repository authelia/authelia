package configuration

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/utils"
)

// koanfEnvironmentCallback returns a koanf callback to map the environment vars to configuration keys.
func koanfEnvironmentCallback(keyMap map[string]string, ignoredKeys []string) func(key, value string) (finalKey string, finalValue interface{}) {
	return func(key, value string) (finalKey string, finalValue interface{}) {
		if k, ok := keyMap[key]; ok {
			return k, value
		}

		if utils.IsStringInSlice(key, ignoredKeys) {
			return "", nil
		}

		formattedKey := strings.TrimPrefix(key, envPrefix)
		formattedKey = strings.ReplaceAll(strings.ToLower(formattedKey), delimiterEnv, delimiter)

		if utils.IsStringInSlice(formattedKey, validator.ValidKeys) {
			return formattedKey, value
		}

		return key, value
	}
}

// koanfEnvironmentSecretsCallback returns a koanf callback to map the environment vars to configuration keys.
func koanfEnvironmentSecretsCallback(keyMap map[string]string, provider *Provider) func(key, value string) (finalKey string, finalValue interface{}) {
	return func(key, value string) (finalKey string, finalValue interface{}) {
		k, ok := keyMap[key]

		if !ok {
			return "", nil
		}

		if v, ok := provider.Get(k).(string); ok && v != "" {
			provider.Push(fmt.Errorf(errFmtSecretAlreadyDefined, k))
			return "", nil
		}

		v, err := loadSecret(value)
		if err != nil {
			provider.Push(fmt.Errorf(errFmtSecretIOIssue, value, k, err))
			return "", nil
		}

		return k, v
	}
}
