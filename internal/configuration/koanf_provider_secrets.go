package configuration

import (
	"errors"
	"fmt"
	"github.com/knadh/koanf/maps"
	"io/ioutil"
	"sort"
	"strings"
)

func NewSecretsProvider(delim string, p *Provider) *SecretsProvider {
	return &SecretsProvider{delim, p}
}

type SecretsProvider struct {
	delim string
	conf  *Provider
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

	for _, k := range sortedKeys {

		if expectedKey := strings.TrimPrefix(k, "secret."); expectedKey != k {
			currentValue, ok := p.conf.Get(expectedKey).(string)
			if ok && currentValue != "" {
				p.conf.Push(fmt.Errorf(errFmtSecretAlreadyDefined, expectedKey))
				continue
			}

			fileName, ok := p.conf.Get(k).(string)
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

	if p.conf.HasErrors() {
		return maps.Unflatten(keys, p.delim), errSecretOneOrMoreErrors
	}

	return maps.Unflatten(keys, p.delim), nil
}

// Watch is not supported by this provider.
func (p *SecretsProvider) Watch(_ func(event interface{}, err error)) error {
	return errors.New("provider does not support this method")
}
