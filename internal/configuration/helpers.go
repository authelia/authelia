package configuration

import (
	"os"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

func getEnvConfigMap(keys []string, prefix, delimiter string) (keyMap map[string]string, ignoredKeys []string) {
	keyMap = make(map[string]string)

	for _, key := range keys {
		if strings.Contains(key, delimiter) {
			originalKey := prefix + strings.ToUpper(strings.ReplaceAll(key, constDelimiter, delimiter))
			keyMap[originalKey] = key
		}

		// Secret envs should be ignored by the env parser.
		if isSecretKey(key) {
			originalKey := strings.ToUpper(strings.ReplaceAll(key, constDelimiter, delimiter)) + constSecretSuffix

			ignoredKeys = append(ignoredKeys, prefix+originalKey)
		}
	}

	return keyMap, ignoredKeys
}

func getSecretConfigMap(keys []string, prefix, delimiter string) (keyMap map[string]string) {
	keyMap = make(map[string]string)

	for _, key := range keys {
		if isSecretKey(key) {
			originalKey := strings.ToUpper(strings.ReplaceAll(key, constDelimiter, delimiter)) + constSecretSuffix

			keyMap[prefix+originalKey] = key
		}
	}

	return keyMap
}

func isSecretKey(key string) (isSecretKey bool) {
	return utils.IsStringInSliceSuffix(key, secretSuffixes)
}

func loadSecret(path string) (value string, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(content), "\n"), err
}
