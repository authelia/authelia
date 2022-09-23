package configuration

import (
	"crypto/rsa"
	"crypto/x509"
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
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		prefixType := ""

		if t.Kind() == reflect.Ptr {
			ptr = true
			prefixType = "*"
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
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType.String()+" (RFC5322)", err)
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
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		prefixType := ""

		if t.Kind() == reflect.Ptr {
			ptr = true
			prefixType = "*"
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
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
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
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var ptr bool

		switch f.Kind() {
		case reflect.String, reflect.Int, reflect.Int32, reflect.Int64:
			// We only allow string and integer from kinds to match.
			break
		default:
			return data, nil
		}

		prefixType := ""

		if t.Kind() == reflect.Ptr {
			ptr = true
			prefixType = "*"
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
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
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
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		prefixType := ""

		if t.Kind() == reflect.Ptr {
			ptr = true
			prefixType = "*"
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
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
			}
		}

		if ptr {
			return result, nil
		}

		if result == nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseEmptyValue, prefixType, expectedType, errDecodeNonPtrMustHaveValue)
		}

		return *result, nil
	}
}

// StringToAddressHookFunc decodes a string into an Address or *Address.
func StringToAddressHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		prefixType := ""

		if t.Kind() == reflect.Ptr {
			ptr = true
			prefixType = "*"
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
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
		}

		if ptr {
			return result, nil
		}

		return *result, nil
	}
}

// StringToX509CertificateHookFunc decodes strings to x509.Certificate's.
func StringToX509CertificateHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t.Kind() != reflect.Ptr {
			return data, nil
		}

		expectedType := reflect.TypeOf(x509.Certificate{})

		if t.Elem() != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		var result *x509.Certificate

		if dataStr == "" {
			return result, nil
		}

		var i interface{}

		if i, err = utils.ParseX509FromPEM([]byte(dataStr)); err != nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "*", expectedType, err)
		}

		switch r := i.(type) {
		case *x509.Certificate:
			return r, nil
		default:
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "*", expectedType, fmt.Errorf("the data is for a %T not a *%s", r, expectedType))
		}
	}
}

// StringToX509CertificateChainHookFunc decodes strings to schema.X509CertificateChain's.
func StringToX509CertificateChainHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		prefixType := ""

		if t.Kind() == reflect.Ptr {
			ptr = true
			prefixType = "*"
		}

		expectedType := reflect.TypeOf(schema.X509CertificateChain{})

		if ptr && t.Elem() != expectedType {
			return data, nil
		} else if !ptr && t != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		var result *schema.X509CertificateChain

		if dataStr == "" && ptr {
			return result, nil
		}

		if result, err = schema.NewX509CertificateChain(dataStr); err != nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, prefixType, expectedType, err)
		}

		if ptr {
			return result, nil
		}

		if result == nil {
			return schema.X509CertificateChain{}, nil
		}

		return *result, nil
	}
}

// StringToRSAPrivateKeyHookFunc decodes strings to rsa.PrivateKey's.
func StringToRSAPrivateKeyHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t.Kind() != reflect.Ptr {
			return data, nil
		}

		expectedType := reflect.TypeOf(rsa.PrivateKey{})

		if t.Elem() != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		var result *rsa.PrivateKey

		if dataStr == "" {
			return result, nil
		}

		var i interface{}

		if i, err = utils.ParseX509FromPEM([]byte(dataStr)); err != nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "*", expectedType, err)
		}

		switch r := i.(type) {
		case *rsa.PrivateKey:
			return r, nil
		default:
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "*", expectedType, fmt.Errorf("the data is for a %T not a *%s", r, expectedType))
		}
	}
}

// StringToTLSVersionHookFunc decodes strings to schema.TLSVersion's.
func StringToTLSVersionHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (value interface{}, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		prefixType := ""

		if t.Kind() == reflect.Ptr {
			ptr = true
			prefixType = "*"
		}

		expectedType := reflect.TypeOf(schema.TLSVersion{})

		if ptr && t.Elem() != expectedType {
			return data, nil
		} else if !ptr && t != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		var result *schema.TLSVersion

		if dataStr == "" && ptr {
			return result, nil
		}

		if result, err = schema.NewTLSVersion(dataStr); err != nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, prefixType, expectedType, err)
		}

		if ptr {
			return result, nil
		}

		if result == nil {
			return schema.TLSVersion{}, nil
		}

		return *result, nil
	}
}
