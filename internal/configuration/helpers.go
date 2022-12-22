package configuration

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
)

func getEnvConfigMap(keys []string, prefix, delimiter string) (keyMap map[string]string, ignoredKeys []string) {
	keyMap = make(map[string]string)

	for _, key := range keys {
		if strings.Contains(key, delimiter) {
			keyMap[ToEnvironmentKey(key, prefix, delimiter)] = key
		}

		// Secret envs should be ignored by the env parser.
		if IsSecretKey(key) {
			ignoredKeys = append(ignoredKeys, ToEnvironmentSecretKey(key, prefix, delimiter))
		}
	}

	return keyMap, ignoredKeys
}

func getSecretConfigMap(keys []string, prefix, delimiter string) (keyMap map[string]string) {
	keyMap = make(map[string]string)

	for _, key := range keys {
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

	return utils.IsStringInSliceSuffix(key, secretSuffixes)
}

func loadSecret(path string) (value string, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(content), "\n"), err
}

func pathToParser(path string) (parser koanf.Parser, explicit bool) {
	switch filepath.Ext(path) {
	case ".yml", ".yaml":
		return yaml.Parser(), true
	case ".tml", ".toml":
		return toml.Parser(), true
	default:
		return yaml.Parser(), false
	}
}
