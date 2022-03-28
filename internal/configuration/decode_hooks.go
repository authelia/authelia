package configuration

import (
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/authelia/authelia/v4/internal/utils"
)

// StringToMailAddressHookFunc decodes a string into a mail.Address.
func StringToMailAddressHookFunc() mapstructure.DecodeHookFuncType {
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

// StringToURLHookFunc converts string types into a url.URL.
func StringToURLHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		ptr = t.Kind() == reflect.Ptr

		typeURL := reflect.TypeOf(url.URL{})

		if ptr && t.Elem() != typeURL {
			return data, nil
		} else if !ptr && t != typeURL {
			return data, nil
		}

		dataStr := data.(string)

		var parsedURL *url.URL

		// Return an empty URL if there is an empty string.
		if dataStr != "" {
			if parsedURL, err = url.Parse(dataStr); err != nil {
				return nil, fmt.Errorf("could not parse '%s' as a URL: %w", dataStr, err)
			}
		}

		if ptr {
			return parsedURL, nil
		}

		// Return an empty URL if there is an empty string.
		if parsedURL == nil {
			return url.URL{}, nil
		}

		return *parsedURL, nil
	}
}

// ToTimeDurationHookFunc converts string and integer types to a time.Duration.
func ToTimeDurationHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var ptr bool

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
			dataStr := data.(string)

			if duration, err = utils.ParseDurationString(dataStr); err != nil {
				return nil, err
			}
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

		if ptr {
			return &duration, nil
		}

		return duration, nil
	}
}

// StringToRegexpFunc decodes a string into a *regexp.Regexp or regexp.Regexp.
func StringToRegexpFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		ptr = t.Kind() == reflect.Ptr

		typeRegexp := reflect.TypeOf(regexp.Regexp{})

		if ptr && t.Elem() != typeRegexp {
			return data, nil
		} else if !ptr && t != typeRegexp {
			return data, nil
		}

		regexStr := data.(string)

		pattern, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, fmt.Errorf("could not parse '%s' as regexp: %w", regexStr, err)
		}

		if ptr {
			return pattern, nil
		}

		return *pattern, nil
	}
}
