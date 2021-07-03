package configuration

import (
	"io/ioutil"
	"strings"

	"github.com/authelia/authelia/internal/utils"
)

func getEnvConfigMap(keys []string) (keyMap map[string]string, ignoredKeys []string) {
	keyMap = make(map[string]string)

	for _, key := range keys {
		if strings.Contains(key, constDelimiterEnv) {
			originalKey := constEnvPrefix + strings.ToUpper(strings.ReplaceAll(key, constDelimiter, constDelimiterEnv))
			keyMap[originalKey] = key
		}

		// Secret envs should be ignored by the env parser.
		if isSecretKey(key) {
			originalKey := strings.ToUpper(strings.ReplaceAll(key, constDelimiter, constDelimiterEnv)) + constSecretSuffix

			ignoredKeys = append(ignoredKeys, constEnvPrefix+originalKey)
			ignoredKeys = append(ignoredKeys, constEnvPrefixAlt+originalKey)
		}
	}

	return keyMap, ignoredKeys
}

func getSecretConfigMap(keys []string) (keyMap map[string]string) {
	keyMap = make(map[string]string)

	for _, key := range keys {
		if isSecretKey(key) {
			originalKey := strings.ToUpper(strings.ReplaceAll(key, constDelimiter, constDelimiterEnv)) + constSecretSuffix

			keyMap[constEnvPrefix+originalKey] = key
			keyMap[constEnvPrefixAlt+originalKey] = key
		}
	}

	return keyMap
}

func getEnvSecretPrefix(key string) (prefix string, err error) {
	var doubleUnderscore bool

	switch {
	case strings.HasPrefix(key, constEnvPrefix):
		doubleUnderscore = true
		prefix = constEnvPrefix
	case strings.HasPrefix(key, constEnvPrefixAlt):
		prefix = constEnvPrefixAlt
	default:
		prefix = ""
	}

	if prefix == "" || !doubleUnderscore && strings.HasPrefix(key, prefix) && !strings.HasSuffix(key, constSecretSuffix) {
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
