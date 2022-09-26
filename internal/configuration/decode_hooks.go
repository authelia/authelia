package configuration

import (
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// StringToMailAddressHookFunc decodes a string into a mail.Address or *mail.Address.
func StringToMailAddressHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		kindStr := "mail.Address (RFC5322)"

		if t.Kind() == reflect.Ptr {
			ptr = true
			kindStr = "*" + kindStr
		}

		expectedType := reflect.TypeOf(mail.Address{})

		if ptr && t.Elem() != expectedType {
			return data, nil
		} else if !ptr && t != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		var result *mail.Address

		if dataStr != "" {
			if result, err = mail.ParseAddress(dataStr); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, kindStr, err)
			}
		}

		if ptr {
			return result, nil
		}

		if result == nil {
			return mail.Address{}, nil
		}

		return *result, nil
	}
}

// StringToURLHookFunc converts string types into a url.URL or *url.URL.
func StringToURLHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		kindStr := "url.URL"

		if t.Kind() == reflect.Ptr {
			ptr = true
			kindStr = "*" + kindStr
		}

		expectedType := reflect.TypeOf(url.URL{})

		if ptr && t.Elem() != expectedType {
			return data, nil
		} else if !ptr && t != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		var result *url.URL

		if dataStr != "" {
			if result, err = url.Parse(dataStr); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, kindStr, err)
			}
		}

		if ptr {
			return result, nil
		}

		if result == nil {
			return url.URL{}, nil
		}

		return *result, nil
	}
}

// ToTimeDurationHookFunc converts string and integer types to a time.Duration.
func ToTimeDurationHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
		var ptr bool

		switch f.Kind() {
		case reflect.String, reflect.Int, reflect.Int32, reflect.Int64:
			// We only allow string and integer from kinds to match.
			break
		default:
			return data, nil
		}

		kindStr := "time.Duration"

		if t.Kind() == reflect.Ptr {
			ptr = true
			kindStr = "*" + kindStr
		}

		expectedType := reflect.TypeOf(time.Duration(0))

		if ptr && t.Elem() != expectedType {
			return data, nil
		} else if !ptr && t != expectedType {
			return data, nil
		}

		var result time.Duration

		switch {
		case f.Kind() == reflect.String:
			dataStr := data.(string)

			if result, err = utils.ParseDurationString(dataStr); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, kindStr, err)
			}
		case f.Kind() == reflect.Int:
			seconds := data.(int)

			result = time.Second * time.Duration(seconds)
		case f.Kind() == reflect.Int32:
			seconds := data.(int32)

			result = time.Second * time.Duration(seconds)
		case f == expectedType:
			result = data.(time.Duration)
		case f.Kind() == reflect.Int64:
			seconds := data.(int64)

			result = time.Second * time.Duration(seconds)
		}

		if ptr {
			return &result, nil
		}

		return result, nil
	}
}

// StringToRegexpHookFunc decodes a string into a *regexp.Regexp or regexp.Regexp.
func StringToRegexpHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		kindStr := "regexp.Regexp"

		if t.Kind() == reflect.Ptr {
			ptr = true
			kindStr = "*" + kindStr
		}

		expectedType := reflect.TypeOf(regexp.Regexp{})

		if ptr && t.Elem() != expectedType {
			return data, nil
		} else if !ptr && t != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		var result *regexp.Regexp

		if dataStr != "" {
			if result, err = regexp.Compile(dataStr); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, kindStr, err)
			}
		}

		if ptr {
			return result, nil
		}

		if result == nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseEmptyValue, kindStr, errDecodeNonPtrMustHaveValue)
		}

		return *result, nil
	}
}

// StringToAddressHookFunc decodes a string into an Address or *Address.
func StringToAddressHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value interface{}, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		kindStr := "Address"

		if t.Kind() == reflect.Ptr {
			ptr = true
			kindStr = "*" + kindStr
		}

		expectedType := reflect.TypeOf(schema.Address{})

		if ptr && t.Elem() != expectedType {
			return data, nil
		} else if !ptr && t != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		var result *schema.Address

		if result, err = schema.NewAddressFromString(dataStr); err != nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, kindStr, err)
		}

		if ptr {
			return result, nil
		}

		return *result, nil
	}
}
