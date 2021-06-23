package configuration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/knadh/koanf/maps"
)

// SecretsProvider implements the koanf.Provider interface.
type SecretsProvider struct {
	conf *Provider
}

// ReadBytes is not supported by this provider.
func (p *SecretsProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("provider does not support this method")
}

// Read reads the struct and returns a nested config map.
func (p *SecretsProvider) Read() (k map[string]interface{}, err error) {
	keys := make(map[string]interface{})

	sortedKeys := p.conf.Keys()
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		if expectedKey := strings.TrimPrefix(key, secretPrefix); expectedKey != key {
			currentValue, ok := p.conf.Get(expectedKey).(string)
			if ok && currentValue != "" {
				p.conf.Push(fmt.Errorf(errFmtSecretAlreadyDefined, expectedKey))
				continue
			}

			fileName, ok := p.conf.Get(key).(string)
			if !ok {
				continue
			}

			content, err := ioutil.ReadFile(fileName)
			if err != nil {
				p.conf.Push(fmt.Errorf(errFmtSecretIOIssue, fileName, expectedKey, err))
				continue
			}

			value := strings.TrimRight(string(content), "\n")

			keys[expectedKey] = value
		}
	}

	return maps.Unflatten(keys, delimiter), nil
}

// NewSecretsProvider returns a new SecretsProvider.
func NewSecretsProvider(p *Provider) *SecretsProvider {
	return &SecretsProvider{p}
}
