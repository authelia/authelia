package configuration

import (
	"fmt"
	"net/mail"
	"reflect"
	"regexp"

	"github.com/mitchellh/mapstructure"
)

// StringToMailAddressFunc decodes a string into a mail.Address.
func StringToMailAddressFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(mail.Address{}) {
			return data, nil
		}

		dataStr := data.(string)

		if dataStr == "" {
			return mail.Address{}, nil
		}

		var (
			mailAddress *mail.Address
		)

		mailAddress, err = mail.ParseAddress(dataStr)
		if err != nil {
			return nil, fmt.Errorf("could not parse '%s' as a RFC5322 address: %w", dataStr, err)
		}

		return *mailAddress, nil
	}
}

// StringToRegexpPtrFunc decodes a string into a regexp.Regexp.
func StringToRegexpPtrFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(&regexp.Regexp{}) {
			return data, nil
		}

		regexStr := data.(string)

		pattern, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, fmt.Errorf("could not parse '%s' as regexp: %w", regexStr, err)
		}

		return pattern, nil
	}
}
