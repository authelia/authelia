package configuration

import (
	"fmt"
	"net/mail"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/authelia/authelia/v4/internal/utils"
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

// ToTimeDurationFunc converts string and integer types to a time.Duration.
func ToTimeDurationFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var (
			ptr bool
		)

		switch f.Kind() {
		case reflect.String, reflect.Int, reflect.Int32, reflect.Int64:
			// We only allow string and integer from kinds to match.
			break
		default:
			return data, nil
		}

		typeTimeDuration := reflect.TypeOf(time.Hour)

		if t.Kind() == reflect.Ptr {
			if t.Elem() != typeTimeDuration {
				return data, nil
			}

			ptr = true
		} else if t != typeTimeDuration {
			return data, nil
		}

		var duration time.Duration

		switch {
		case f.Kind() == reflect.String:
			break
		case f.Kind() == reflect.Int:
			seconds := data.(int)

			duration = time.Second * time.Duration(seconds)
		case f.Kind() == reflect.Int32:
			seconds := data.(int32)

			duration = time.Second * time.Duration(seconds)
		case f == typeTimeDuration:
			duration = data.(time.Duration)
		case f.Kind() == reflect.Int64:
			seconds := data.(int64)

			duration = time.Second * time.Duration(seconds)
		}

		if duration == 0 {
			dataStr := data.(string)

			if duration, err = utils.ParseDurationString(dataStr); err != nil {
				return nil, err
			}
		}

		if ptr {
			return &duration, nil
		}

		return duration, nil
	}
}
