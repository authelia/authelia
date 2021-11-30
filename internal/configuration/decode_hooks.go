package configuration

import (
	"fmt"
	"net/mail"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// StringToMailAddressFunc decodes a string into a mail.Address.
func StringToMailAddressFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Kind, t reflect.Kind, data interface{}) (value interface{}, err error) {
		if f != reflect.String || t != reflect.TypeOf(mail.Address{}).Kind() {
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
