package configuration

import (
	"strings"

	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/utils"
)

// koanfKeyCallbackBuilder builds a koanf key callback function that creates a map of replacements which
// are populated by looping through the valid keys and for every key with an underscore, adding a key to the map with
// the underscores replaced with the given symbol with the value of the expected key. For example if using the Env Parser
// you can provide the symbol `.` which will allow you to use the underscore separator but still retain underscores in
// your configuration struct.
func koanfKeyCallbackBuilder() func(key, value string) (finalKey string, finalValue interface{}) {
	keyMap := map[string]string{}
	secretMap := map[string]string{}

	for _, key := range validator.ValidKeys {
		if strings.Contains(key, delimiterEnv) {
			originalKey := strings.ReplaceAll(key, delimiterEnv, delimiter)
			if originalKey != key {
				keyMap[originalKey] = key
			}
		}

		if isSecretKey(key) {
			secretName := strings.ReplaceAll(key, delimiterEnv, delimiter) + secretSuffix
			secretMap[secretName] = secretPrefix + key
		}
	}

	return func(key, value string) (finalKey string, finalValue interface{}) {
		prefix, err := getPrefix(key)
		if err != nil {
			return "", nil
		}

		formattedKey := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(key, prefix)), delimiterEnv, delimiter)

		if k, ok := secretMap[formattedKey]; ok {
			return k, value
		}

		if k, ok := keyMap[formattedKey]; ok {
			return k, value
		}

		if utils.IsStringInSlice(formattedKey, validator.ValidKeys) {
			return formattedKey, value
		}

		return "", nil
	}
}

func getPrefix(key string) (prefix string, err error) {
	var doubleUnderscore bool

	switch {
	case strings.HasPrefix(key, envPrefix):
		doubleUnderscore = true
		prefix = envPrefix
	default:
		prefix = envPrefixAlt
	}

	if !doubleUnderscore && strings.HasPrefix(key, prefix) && !strings.HasSuffix(key, secretSuffixEnv) {
		err = errInvalidPrefix
	}

	return prefix, err
}

func isSecretKey(key string) (isSecretKey bool) {
	for _, suffix := range secretSuffixes {
		if strings.HasSuffix(key, suffix) {
			return true
		}
	}

	return false
}
