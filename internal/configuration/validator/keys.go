package validator

import (
	"errors"
	"fmt"
	"strings"

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

		replacedKey := reKeyReplacer.ReplaceAllString(key, "[]")
		if err, ok := specificErrorKeys[replacedKey]; ok {
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
