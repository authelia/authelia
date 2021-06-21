package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/knadh/koanf"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateKeys determines if all provided keys are valid.
func ValidateKeys(validator *schema.StructValidator, keys []string) {
	var errStrings []string

	for _, key := range keys {
		if utils.IsStringInSlice(key, ValidKeys) {
			continue
		}

		if expectedKey := strings.TrimPrefix(key, "secret."); expectedKey != key {
			if utils.IsStringInSlice(expectedKey, ValidKeys) {
				continue
			}
		}

		if newKey, ok := replacedKeys[key]; ok {
			validator.Push(fmt.Errorf(errFmtReplacedConfigurationKey, key, newKey))
			continue
		}

		if err, ok := specificErrorKeys[key]; ok {
			if !utils.IsStringInSlice(err, errStrings) {
				errStrings = append(errStrings, err)
			}
		} else {
			validator.Push(fmt.Errorf("config key not expected: %s", key))
		}
	}

	for _, err := range errStrings {
		validator.Push(errors.New(err))
	}
}

// ValidateAccessControlRuleKeys determines if a provided keys are valid for an access control rule.
func ValidateAccessControlRuleKeys(validator *schema.StructValidator, koanfs []*koanf.Koanf) {
	for i, k := range koanfs {
		for _, key := range k.Keys() {
			if utils.IsStringInSlice(key, validACLKeys) {
				continue
			}

			validator.Push(fmt.Errorf("config key not expected: access_control.rules[%d].%s", i, key))
		}
	}
}

// ValidateOpenIDConnectClientKeys determines if a provided keys are valid for an OpenID Connect client.
func ValidateOpenIDConnectClientKeys(validator *schema.StructValidator, koanfs []*koanf.Koanf) {
	for i, k := range koanfs {
		for _, key := range k.Keys() {
			if utils.IsStringInSlice(key, validOpenIDConnectClientKeys) {
				continue
			}

			validator.Push(fmt.Errorf("config key not expected: identity_providers.oidc.clients[%d].%s", i, key))
		}
	}
}
