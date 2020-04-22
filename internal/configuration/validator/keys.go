package validator

import (
	"errors"
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

func ValidateKeys(validator *schema.StructValidator, keys []string) {
	for _, key := range keys {
		if utils.IsStringInSlice(key, validKeys) {
			continue
		}
		var errStrings []string

		if err, ok := specificErrorKeys[key]; ok {
			if !utils.IsStringInSlice(err, errStrings) {
				errStrings = append(errStrings, err)
			}
		} else {
			validator.Push(fmt.Errorf("config key not expected: %s", key))
		}
		for _, err := range errStrings {
			validator.Push(errors.New(err))
		}
	}
}
