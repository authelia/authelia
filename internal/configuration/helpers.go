package configuration

import (
	"os"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

func getEnvConfigMap(keys []string, prefix, delimiter string, ds map[string]Deprecation, dms []MultiKeyMappedDeprecation) (keyMap map[string]string, ignoredKeys []string) {
	keyMap = make(map[string]string)

	for _, key := range keys {
		keyMap[ToEnvironmentKey(key, prefix, delimiter)] = key

		// Secret envs should be ignored by the env parser.
		if IsSecretKey(key) {
			ignoredKeys = append(ignoredKeys, ToEnvironmentSecretKey(key, prefix, delimiter))
		}
	}

	for key, deprecation := range ds {
		if IsSecretKey(key) {
			ignoredKeys = append(ignoredKeys, ToEnvironmentSecretKey(key, prefix, delimiter))
		}

		if !deprecation.AutoMap {
			continue
		}

		d := ToEnvironmentKey(deprecation.Key, prefix, delimiter)

		if _, ok := keyMap[d]; ok {
			continue
		}

		keyMap[d] = deprecation.Key
	}

	for _, deprecation := range dms {
		for _, key := range deprecation.Keys {
			if IsSecretKey(key) {
				ignoredKeys = append(ignoredKeys, ToEnvironmentSecretKey(key, prefix, delimiter))
			}

			d := ToEnvironmentKey(key, prefix, delimiter)

			if _, ok := keyMap[d]; ok {
				continue
			}

			keyMap[d] = key
		}
	}

	return keyMap, ignoredKeys
}

func getSecretConfigMap(keys []string, prefix, delimiter string, ds map[string]Deprecation) (keyMap map[string]string) {
	keyMap = make(map[string]string)

	for _, key := range keys {
		if IsSecretKey(key) {
			originalKey := strings.ToUpper(strings.ReplaceAll(key, constDelimiter, delimiter)) + constSecretSuffix

			keyMap[prefix+originalKey] = key
		}
	}

	for key := range ds {
		if IsSecretKey(key) {
			originalKey := strings.ToUpper(strings.ReplaceAll(key, constDelimiter, delimiter)) + constSecretSuffix

			keyMap[prefix+originalKey] = key
		}
	}

	return keyMap
}

// ToEnvironmentKey converts a key into the environment variable name.
func ToEnvironmentKey(key, prefix, delimiter string) string {
	return prefix + strings.ToUpper(strings.ReplaceAll(key, constDelimiter, delimiter))
}

// ToEnvironmentSecretKey converts a key into the environment variable name.
func ToEnvironmentSecretKey(key, prefix, delimiter string) string {
	return prefix + strings.ToUpper(strings.ReplaceAll(key, constDelimiter, delimiter)) + constSecretSuffix
}

// IsSecretKey returns true if the provided key is a secret enabled key.
func IsSecretKey(key string) (isSecretKey bool) {
	if strings.Contains(key, "[]") {
		return false
	}

	if strings.Contains(key, ".*.") {
		return false
	}

	if utils.IsStringInSlice(key, secretExclusionExact) {
		return false
	}

	if utils.IsStringInSliceF(key, secretExclusionPrefix, strings.HasPrefix) {
		return false
	}

	return utils.IsStringInSliceF(key, secretSuffix, strings.HasSuffix)
}

func loadSecret(path string) (value string, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(content), "\n"), err
}
