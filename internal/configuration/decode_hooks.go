package configuration

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-crypt/crypt/algorithm/plaintext"
	"github.com/go-viper/mapstructure/v2"

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
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
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

func DecodeTimeDuration(f, expectedType reflect.Type, prefixType string, data any) (result time.Duration, err error) {
	e := reflect.TypeOf(time.Duration(0))

	switch {
	case f.Kind() == reflect.String:
		dataStr := data.(string)

		if result, err = utils.ParseDurationString(dataStr); err != nil {
			return time.Duration(0), fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
		}
	case f.Kind() == reflect.Int:
		seconds := data.(int)

		result = time.Second * time.Duration(seconds)
	case f.Kind() == reflect.Int8:
		seconds := data.(int8)

		result = time.Second * time.Duration(seconds)
	case f.Kind() == reflect.Int16:
		seconds := data.(int16)

		result = time.Second * time.Duration(seconds)
	case f.Kind() == reflect.Int32:
		seconds := data.(int32)

		result = time.Second * time.Duration(seconds)
	case f.Kind() == reflect.Float64:
		fseconds := data.(float64)

		if fseconds > durationMax.Seconds() {
			result = durationMax
		} else {
			seconds, _ := strconv.Atoi(fmt.Sprintf("%.0f", fseconds))

			result = time.Second * time.Duration(seconds)
		}
	case f == e:
		result = data.(time.Duration)
	case f.Kind() == reflect.Int64:
		seconds := data.(int64)

		result = time.Second * time.Duration(seconds)
	}

	return result, nil
}

// ToRefreshIntervalDurationHookFunc converts string and integer types to a schema.RefreshIntervalDuration.
func ToRefreshIntervalDurationHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
		var ptr bool

		switch f.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float64:
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

		expectedType := reflect.TypeOf(schema.RefreshIntervalDuration{})

		if ptr && t.Elem() != expectedType {
			return data, nil
		} else if !ptr && t != expectedType {
			return data, nil
		}

		var (
			result  schema.RefreshIntervalDuration
			decoded bool
		)

		if f.Kind() == reflect.String {
			dataStr, ok := data.(string)
			if ok {
				switch dataStr {
				case schema.ProfileRefreshAlways:
					result, decoded = schema.NewRefreshIntervalDurationAlways(), true
				case schema.ProfileRefreshDisabled:
					result, decoded = schema.NewRefreshIntervalDurationNever(), true
				}
			}
		}

		if !decoded {
			var resultv time.Duration

			if resultv, err = DecodeTimeDuration(f, expectedType, prefixType, data); err != nil {
				return nil, err
			}

			result = schema.NewRefreshIntervalDuration(resultv)
		}

		if ptr {
			return &result, nil
		}

		return result, nil
	}
}

// ToTimeDurationHookFunc converts string and integer types to a time.Duration.
func ToTimeDurationHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
		var (
			ptr        bool
			prefixType string
		)

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

		switch f.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float64:
			// We only allow string and integer from kinds to match.
			break
		default:
			return data, nil
		}

		var result time.Duration

		if result, err = DecodeTimeDuration(f, expectedType, prefixType, data); err != nil {
			return nil, err
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
//
//nolint:gocyclo // This is an adequately clear function even with the complexity.
func StringToAddressHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
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
		expectedTypeTCP := reflect.TypeOf(schema.AddressTCP{})
		expectedTypeUDP := reflect.TypeOf(schema.AddressUDP{})
		expectedTypeLDAP := reflect.TypeOf(schema.AddressLDAP{})
		expectedTypeSMTP := reflect.TypeOf(schema.AddressSMTP{})

		switch {
		case ptr:
			switch t.Elem() {
			case expectedType:
			case expectedTypeTCP:
				expectedType = expectedTypeTCP
			case expectedTypeUDP:
				expectedType = expectedTypeUDP
			case expectedTypeLDAP:
				expectedType = expectedTypeLDAP
			case expectedTypeSMTP:
				expectedType = expectedTypeSMTP
			default:
				return data, nil
			}
		default:
			switch t {
			case expectedType:
				break
			case expectedTypeTCP:
				expectedType = expectedTypeTCP
			case expectedTypeUDP:
				expectedType = expectedTypeUDP
			case expectedTypeLDAP:
				expectedType = expectedTypeLDAP
			case expectedTypeSMTP:
				expectedType = expectedTypeSMTP
			default:
				return data, nil
			}
		}

		dataStr := data.(string)

		var result *schema.Address

		switch expectedType {
		case expectedTypeTCP:
			if result, err = schema.NewAddressDefault(dataStr, schema.AddressSchemeTCP, schema.AddressSchemeUnix); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
			}

			if ptr {
				return &schema.AddressTCP{Address: *result}, nil
			}

			return schema.AddressTCP{Address: *result}, nil
		case expectedTypeUDP:
			if result, err = schema.NewAddressDefault(dataStr, schema.AddressSchemeUDP, schema.AddressSchemeUnix); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
			}

			if ptr {
				return &schema.AddressUDP{Address: *result}, nil
			}

			return schema.AddressUDP{Address: *result}, nil
		case expectedTypeLDAP:
			if result, err = schema.NewAddressDefault(dataStr, schema.AddressSchemeLDAPS, schema.AddressSchemeLDAPI); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
			}

			if ptr {
				return &schema.AddressLDAP{Address: *result}, nil
			}

			return schema.AddressLDAP{Address: *result}, nil
		case expectedTypeSMTP:
			if result, err = schema.NewAddressDefault(dataStr, schema.AddressSchemeSMTP, schema.AddressSchemeUnix); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
			}

			if ptr {
				return &schema.AddressSMTP{Address: *result}, nil
			}

			return schema.AddressSMTP{Address: *result}, nil
		default:
			if result, err = schema.NewAddress(dataStr); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
			}

			if ptr {
				return result, nil
			}

			return *result, nil
		}
	}
}

// StringToX509CertificateHookFunc decodes strings to x509.Certificate's.
func StringToX509CertificateHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
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

		var i any

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
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
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

// StringToTLSVersionHookFunc decodes strings to schema.TLSVersion's.
func StringToTLSVersionHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
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

		if result, err = schema.NewTLSVersion(dataStr); err != nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType, err)
		}

		if ptr {
			return result, nil
		}

		return *result, nil
	}
}

// StringToCryptoPrivateKeyHookFunc decodes strings to schema.CryptographicPrivateKey's.
func StringToCryptoPrivateKeyHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		field, _ := reflect.TypeOf(schema.TLS{}).FieldByName("PrivateKey")
		expectedType := field.Type

		if t != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		var i any

		if i, err = utils.ParseX509FromPEM([]byte(dataStr)); err != nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "", expectedType, err)
		}

		if result, ok := i.(schema.CryptographicPrivateKey); !ok {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "", expectedType, fmt.Errorf("the data is for a %T not a %s", i, expectedType))
		} else {
			return result, nil
		}
	}
}

// StringToCryptographicKeyHookFunc decodes strings to schema.CryptographicKey's.
func StringToCryptographicKeyHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		field, _ := reflect.TypeOf(schema.JWK{}).FieldByName("Key")
		expectedType := field.Type

		if t != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		if value, err = utils.ParseX509FromPEM([]byte(dataStr)); err != nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "", expectedType, err)
		}

		return value, nil
	}
}

// StringToPrivateKeyHookFunc decodes strings to rsa.PrivateKey's and ecdsa.PrivateKey's.
//
//nolint:gocyclo
func StringToPrivateKeyHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t.Kind() != reflect.Ptr {
			return data, nil
		}

		expectedTypeRSA := reflect.TypeOf(rsa.PrivateKey{})
		expectedTypeECDSA := reflect.TypeOf(ecdsa.PrivateKey{})
		expectedTypeEd25519 := reflect.TypeOf(ed25519.PrivateKey{})

		var (
			i            any
			expectedType reflect.Type
		)

		dataStr := data.(string)

		switch t.Elem() {
		case expectedTypeRSA:
			var result *rsa.PrivateKey

			if dataStr == "" {
				return result, nil
			}

			expectedType = expectedTypeRSA
		case expectedTypeECDSA:
			var result *ecdsa.PrivateKey

			if dataStr == "" {
				return result, nil
			}

			expectedType = expectedTypeECDSA
		case expectedTypeEd25519:
			var result *ed25519.PrivateKey

			if dataStr == "" {
				return result, nil
			}

			expectedType = expectedTypeEd25519
		default:
			return data, nil
		}

		if i, err = utils.ParseX509FromPEM([]byte(dataStr)); err != nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "*", expectedType, err)
		}

		switch r := i.(type) {
		case *rsa.PrivateKey:
			if expectedType != expectedTypeRSA {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "*", expectedType, fmt.Errorf("the data is for a %T not a *%s", r, expectedType))
			}

			if err = r.Validate(); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "*", expectedType, err)
			}

			return r, nil
		case *ecdsa.PrivateKey:
			if expectedType != expectedTypeECDSA {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "*", expectedType, fmt.Errorf("the data is for a %T not a *%s", r, expectedType))
			}

			return r, nil
		case ed25519.PrivateKey:
			if expectedType != expectedTypeEd25519 {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "*", expectedType, fmt.Errorf("the data is for a %T not a *%s", r, expectedType))
			}

			return &r, nil
		default:
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseBasic, "*", expectedType, fmt.Errorf("the data is for a %T not a *%s", r, expectedType))
		}
	}
}

// StringToPasswordDigestHookFunc decodes a string into a crypt.Digest.
func StringToPasswordDigestHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (value any, err error) {
		var ptr bool

		if f.Kind() != reflect.String {
			return data, nil
		}

		prefixType := ""

		if t.Kind() == reflect.Ptr {
			ptr = true
			prefixType = "*"
		}

		expectedType := reflect.TypeOf(schema.PasswordDigest{})

		if ptr && t.Elem() != expectedType {
			return data, nil
		} else if !ptr && t != expectedType {
			return data, nil
		}

		dataStr := data.(string)

		var result *schema.PasswordDigest

		if dataStr != "" {
			if !strings.HasPrefix(dataStr, "$") {
				dataStr = fmt.Sprintf(plaintext.EncodingFmt, plaintext.AlgIdentifierPlainText, dataStr)
			}

			if result, err = schema.DecodePasswordDigest(dataStr); err != nil {
				return nil, fmt.Errorf(errFmtDecodeHookCouldNotParse, dataStr, prefixType, expectedType.String(), err)
			}
		}

		if ptr {
			return result, nil
		}

		if result == nil {
			return nil, fmt.Errorf(errFmtDecodeHookCouldNotParseEmptyValue, prefixType, expectedType.String(), errDecodeNonPtrMustHaveValue)
		}

		return *result, nil
	}
}
