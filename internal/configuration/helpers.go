package configuration

import (
	"io/ioutil"
	"strings"

	"github.com/authelia/authelia/internal/utils"
)

func getEnvConfigMap(keys []string) (keyMap map[string]string, ignoredKeys []string) {
	keyMap = make(map[string]string)

	for _, key := range keys {
		if strings.Contains(key, delimiterEnv) {
			originalKey := envPrefix + strings.ToUpper(strings.ReplaceAll(key, delimiter, delimiterEnv))
			keyMap[originalKey] = key
		}

		// Secret envs should be ignored by the env parser.
		if isSecretKey(key) {
			originalKey := strings.ToUpper(strings.ReplaceAll(key, delimiter, delimiterEnv)) + secretSuffix

			ignoredKeys = append(ignoredKeys, envPrefix+originalKey)
			ignoredKeys = append(ignoredKeys, envPrefixAlt+originalKey)
		}
	}

	return keyMap, ignoredKeys
}

func getSecretConfigMap(keys []string) (keyMap map[string]string) {
	keyMap = make(map[string]string)

	for _, key := range keys {
		if isSecretKey(key) {
			originalKey := strings.ToUpper(strings.ReplaceAll(key, delimiter, delimiterEnv)) + secretSuffix

			keyMap[envPrefix+originalKey] = key
			keyMap[envPrefixAlt+originalKey] = key
		}
	}

	return keyMap
}

func getEnvSecretPrefix(key string) (prefix string, err error) {
	var doubleUnderscore bool

	switch {
	case strings.HasPrefix(key, envPrefix):
		doubleUnderscore = true
		prefix = envPrefix
	case strings.HasPrefix(key, envPrefixAlt):
		prefix = envPrefixAlt
	default:
		prefix = ""
	}

	if prefix == "" || !doubleUnderscore && strings.HasPrefix(key, prefix) && !strings.HasSuffix(key, secretSuffix) {
		err = errInvalidPrefix
	}

	return prefix, err
}

func isSecretKey(key string) (isSecretKey bool) {
	return utils.IsStringInSliceSuffix(key, secretSuffixes)
}

func loadSecret(path string) (value string, err error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(content), "\n"), err
}
