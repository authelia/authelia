package configuration

import (
	"fmt"
	"net/mail"
	"net/url"
	"reflect"

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
			parsedAddress *mail.Address
		)

		if parsedAddress, err = mail.ParseAddress(dataStr); err != nil {
			return nil, fmt.Errorf("could not parse '%s' as a RFC5322 address: %w", dataStr, err)
		}

		return *parsedAddress, nil
	}
}

// StringToURLFunc creates a mapstructure.DecodeHookFunc that decodes strings into a url.URL.
func StringToURLFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Kind, t reflect.Kind, data interface{}) (value interface{}, err error) {
		if f != reflect.String || t != reflect.TypeOf(url.URL{}).Kind() {
			return data, nil
		}

		dataStr := data.(string)

		if dataStr == "" {
			return url.URL{}, nil
		}

		var (
			parsedURL *url.URL
		)

		if parsedURL, err = url.Parse(dataStr); err != nil {
			return nil, fmt.Errorf("could not parse '%s' as a URL: %w", dataStr, err)
		}

		return *parsedURL, nil
	}
}
