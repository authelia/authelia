package configuration

import (
	"strings"

	"github.com/authelia/authelia/internal/configuration/validator"
)

// koanfKeyCallbackBuilder builds a koanf key callback function that creates a map of replacements which
// are populated by looping through the valid keys and for every key with an underscore, adding a key to the map with
// the underscores replaced with the given symbol with the value of the expected key. For example if using the Env Parser
// you can provide the symbol `.` which will allow you to use the underscore separator but still retain underscores in
// your configuration struct.
func koanfKeyCallbackBuilder(old, new, prefix string) func(key, value string) (finalKey string, finalValue interface{}) {
	keyMap := map[string]string{}
	secretMap := map[string]string{}

	for _, key := range validator.ValidKeys {
		if strings.Contains(key, "_") {
			originalKey := strings.ReplaceAll(key, old, new)
			if originalKey != key {
				keyMap[originalKey] = key
			}
		}

		if strings.HasSuffix(key, "password") || strings.HasSuffix(key, "secret") ||
			strings.HasSuffix(key, "key") || strings.HasSuffix(key, "token") {

			secretName := strings.ReplaceAll(key, old, new) + ".file"
			secretMap[secretName] = "secret." + key
		}
	}

	return func(key, value string) (finalKey string, finalValue interface{}) {
		formattedKey := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(key, prefix)), "_", ".")

		if k, ok := secretMap[formattedKey]; ok {
			return k, value
		}

		if k, ok := keyMap[formattedKey]; ok {
			return k, value
		}

		return formattedKey, value
	}
}
