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
			mailAddress *mail.Address
		)

		mailAddress, err = mail.ParseAddress(dataStr)
		if err != nil {
			return nil, fmt.Errorf("could not parse '%s' as a RFC5322 address: %w", dataStr, err)
		}

		return *mailAddress, nil
	}
}

// ToTimeDurationFunc converts string and integer types to a time.Duration.
func ToTimeDurationFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		o := time.Minute
		ptr := false

		switch t {
		case reflect.TypeOf(o):
			break
		case reflect.TypeOf(&o):
			ptr = true
		default:
			return data, nil
		}

		var duration time.Duration

		switch f.Kind() {
		case reflect.String:
			break
		case reflect.Int:
			seconds := data.(int)

			duration = time.Second * time.Duration(seconds)
		case reflect.Int32:
			seconds := data.(int32)

			duration = time.Second * time.Duration(seconds)
		case reflect.Int64:
			seconds := data.(int64)

			duration = time.Second * time.Duration(seconds)
		default:
			return data, nil
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
