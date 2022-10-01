package configuration_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestStringToMailAddressHookFunc(t *testing.T) {
	testCases := []struct {
		desc   string
		have   interface{}
		want   interface{}
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeMailAddress",
			have:   "james@example.com",
			want:   mail.Address{Name: "", Address: "james@example.com"},
			decode: true,
		},
		{
			desc:   "ShouldDecodeMailAddressWithName",
			have:   "James <james@example.com>",
			want:   mail.Address{Name: "James", Address: "james@example.com"},
			decode: true,
		},
		{
			desc:   "ShouldDecodeMailAddressWithEmptyString",
			have:   "",
			want:   mail.Address{},
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeInvalidMailAddress",
			have:   "fred",
			want:   mail.Address{},
			err:    "could not decode 'fred' to a mail.Address (RFC5322): mail: missing '@' or angle-addr",
			decode: true,
		},
	}

	hook := configuration.StringToMailAddressHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, result)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.want, result)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestStringToMailAddressHookFuncPointer(t *testing.T) {
	testCases := []struct {
		desc   string
		have   interface{}
		want   interface{}
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeMailAddress",
			have:   "james@example.com",
			want:   &mail.Address{Name: "", Address: "james@example.com"},
			decode: true,
		},
		{
			desc:   "ShouldDecodeMailAddressWithName",
			have:   "James <james@example.com>",
			want:   &mail.Address{Name: "James", Address: "james@example.com"},
			decode: true,
		},
		{
			desc:   "ShouldDecodeMailAddressWithEmptyString",
			have:   "",
			want:   (*mail.Address)(nil),
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeInvalidMailAddress",
			have:   "fred",
			want:   &mail.Address{},
			err:    "could not decode 'fred' to a *mail.Address (RFC5322): mail: missing '@' or angle-addr",
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeToInt",
			have:   "fred",
			want:   testInt32Ptr(4),
			decode: false,
		},
	}

	hook := configuration.StringToMailAddressHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, result)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.want, result)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestStringToURLHookFunc(t *testing.T) {
	testCases := []struct {
		desc   string
		have   interface{}
		want   interface{}
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeURL",
			have:   "https://www.example.com:9090/abc?test=true",
			want:   url.URL{Scheme: "https", Host: "www.example.com:9090", Path: "/abc", RawQuery: "test=true"},
			decode: true,
		},
		{
			desc:   "ShouldDecodeURLEmptyString",
			have:   "",
			want:   url.URL{},
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeToString",
			have:   "abc",
			want:   "",
			decode: false,
		},
		{
			desc:   "ShouldDecodeURLWithUserAndPassword",
			have:   "https://john:abc123@www.example.com:9090/abc?test=true",
			want:   url.URL{Scheme: "https", Host: "www.example.com:9090", Path: "/abc", RawQuery: "test=true", User: url.UserPassword("john", "abc123")},
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeInt",
			have:   5,
			want:   url.URL{},
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeBool",
			have:   true,
			want:   url.URL{},
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeBadURL",
			have:   "*(!&@#(!*^$%",
			want:   url.URL{},
			err:    "could not decode '*(!&@#(!*^$%' to a url.URL: parse \"*(!&@#(!*^$%\": invalid URL escape \"%\"",
			decode: true,
		},
	}

	hook := configuration.StringToURLHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, result)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.want, result)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestStringToURLHookFuncPointer(t *testing.T) {
	testCases := []struct {
		desc   string
		have   interface{}
		want   interface{}
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeURL",
			have:   "https://www.example.com:9090/abc?test=true",
			want:   &url.URL{Scheme: "https", Host: "www.example.com:9090", Path: "/abc", RawQuery: "test=true"},
			decode: true,
		},
		{
			desc:   "ShouldDecodeURLEmptyString",
			have:   "",
			want:   (*url.URL)(nil),
			decode: true,
		},
		{
			desc:   "ShouldDecodeURLWithUserAndPassword",
			have:   "https://john:abc123@www.example.com:9090/abc?test=true",
			want:   &url.URL{Scheme: "https", Host: "www.example.com:9090", Path: "/abc", RawQuery: "test=true", User: url.UserPassword("john", "abc123")},
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeInt",
			have:   5,
			want:   &url.URL{},
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeBool",
			have:   true,
			want:   &url.URL{},
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeBadURL",
			have:   "*(!&@#(!*^$%",
			want:   &url.URL{},
			err:    "could not decode '*(!&@#(!*^$%' to a *url.URL: parse \"*(!&@#(!*^$%\": invalid URL escape \"%\"",
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeToInt",
			have:   "fred",
			want:   testInt32Ptr(4),
			decode: false,
		},
	}

	hook := configuration.StringToURLHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, result)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.want, result)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestToTimeDurationHookFunc(t *testing.T) {
	testCases := []struct {
		desc   string
		have   interface{}
		want   interface{}
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeFourtyFiveSeconds",
			have:   "45s",
			want:   time.Second * 45,
			decode: true,
		},
		{
			desc:   "ShouldDecodeOneMinute",
			have:   "1m",
			want:   time.Minute,
			decode: true,
		},
		{
			desc:   "ShouldDecodeTwoHours",
			have:   "2h",
			want:   time.Hour * 2,
			decode: true,
		},
		{
			desc:   "ShouldDecodeThreeDays",
			have:   "3d",
			want:   time.Hour * 24 * 3,
			decode: true,
		},
		{
			desc:   "ShouldDecodeFourWeeks",
			have:   "4w",
			want:   time.Hour * 24 * 7 * 4,
			decode: true,
		},
		{
			desc:   "ShouldDecodeFiveMonths",
			have:   "5M",
			want:   time.Hour * 24 * 30 * 5,
			decode: true,
		},
		{
			desc:   "ShouldDecodeSixYears",
			have:   "6y",
			want:   time.Hour * 24 * 365 * 6,
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeInvalidString",
			have:   "abc",
			want:   time.Duration(0),
			err:    "could not decode 'abc' to a time.Duration: could not parse 'abc' as a duration",
			decode: true,
		},
		{
			desc:   "ShouldDecodeIntToSeconds",
			have:   60,
			want:   time.Second * 60,
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt32ToSeconds",
			have:   int32(90),
			want:   time.Second * 90,
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt64ToSeconds",
			have:   int64(120),
			want:   time.Second * 120,
			decode: true,
		},
		{
			desc:   "ShouldDecodeTimeDuration",
			have:   time.Second * 30,
			want:   time.Second * 30,
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeToString",
			have:   int64(30),
			want:   "",
			decode: false,
		},
		{
			desc:   "ShouldDecodeFromIntZero",
			have:   0,
			want:   time.Duration(0),
			decode: true,
		},
		{
			desc: "ShouldNotDecodeFromBool",
			have: true,
			want: true,
		},
	}

	hook := configuration.ToTimeDurationHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, result)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.want, result)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestToTimeDurationHookFuncPointer(t *testing.T) {
	testCases := []struct {
		desc   string
		have   interface{}
		want   interface{}
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeFourtyFiveSeconds",
			have:   "45s",
			want:   testTimeDurationPtr(time.Second * 45),
			decode: true,
		},
		{
			desc:   "ShouldDecodeOneMinute",
			have:   "1m",
			want:   testTimeDurationPtr(time.Minute),
			decode: true,
		},
		{
			desc:   "ShouldDecodeTwoHours",
			have:   "2h",
			want:   testTimeDurationPtr(time.Hour * 2),
			decode: true,
		},
		{
			desc:   "ShouldDecodeThreeDays",
			have:   "3d",
			want:   testTimeDurationPtr(time.Hour * 24 * 3),
			decode: true,
		},
		{
			desc:   "ShouldDecodeFourWeeks",
			have:   "4w",
			want:   testTimeDurationPtr(time.Hour * 24 * 7 * 4),
			decode: true,
		},
		{
			desc:   "ShouldDecodeFiveMonths",
			have:   "5M",
			want:   testTimeDurationPtr(time.Hour * 24 * 30 * 5),
			decode: true,
		},
		{
			desc:   "ShouldDecodeSixYears",
			have:   "6y",
			want:   testTimeDurationPtr(time.Hour * 24 * 365 * 6),
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeInvalidString",
			have:   "abc",
			want:   testTimeDurationPtr(time.Duration(0)),
			err:    "could not decode 'abc' to a *time.Duration: could not parse 'abc' as a duration",
			decode: true,
		},
		{
			desc:   "ShouldDecodeIntToSeconds",
			have:   60,
			want:   testTimeDurationPtr(time.Second * 60),
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt32ToSeconds",
			have:   int32(90),
			want:   testTimeDurationPtr(time.Second * 90),
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt64ToSeconds",
			have:   int64(120),
			want:   testTimeDurationPtr(time.Second * 120),
			decode: true,
		},
		{
			desc:   "ShouldDecodeTimeDuration",
			have:   time.Second * 30,
			want:   testTimeDurationPtr(time.Second * 30),
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeToString",
			have:   int64(30),
			want:   &testString,
			decode: false,
		},
		{
			desc:   "ShouldDecodeFromIntZero",
			have:   0,
			want:   testTimeDurationPtr(time.Duration(0)),
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeFromBool",
			have:   true,
			want:   &testTrue,
			decode: false,
		},
	}

	hook := configuration.ToTimeDurationHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, result)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.want, result)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestStringToRegexpFunc(t *testing.T) {
	testCases := []struct {
		desc     string
		have     interface{}
		want     interface{}
		err      string
		decode   bool
		wantGrps []string
	}{
		{
			desc:   "ShouldNotDecodeRegexpWithOpenParenthesis",
			have:   "hello(test one two",
			want:   regexp.Regexp{},
			err:    "could not decode 'hello(test one two' to a regexp.Regexp: error parsing regexp: missing closing ): `hello(test one two`",
			decode: true,
		},
		{
			desc:   "ShouldDecodeValidRegex",
			have:   "^(api|admin)$",
			want:   *regexp.MustCompile(`^(api|admin)$`),
			decode: true,
		},
		{
			desc:     "ShouldDecodeValidRegexWithGroupNames",
			have:     "^(?P<area>api|admin)(one|two)$",
			want:     *regexp.MustCompile(`^(?P<area>api|admin)(one|two)$`),
			decode:   true,
			wantGrps: []string{"area"},
		},
		{
			desc:   "ShouldNotDecodeFromInt32",
			have:   int32(20),
			want:   regexp.Regexp{},
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeFromBool",
			have:   false,
			want:   regexp.Regexp{},
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeToBool",
			have:   "^(?P<area>api|admin)(one|two)$",
			want:   testTrue,
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeToInt32",
			have:   "^(?P<area>api|admin)(one|two)$",
			want:   testInt32Ptr(0),
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeToMailAddress",
			have:   "^(?P<area>api|admin)(one|two)$",
			want:   mail.Address{},
			decode: false,
		},
		{
			desc:   "ShouldErrOnDecodeEmptyString",
			have:   "",
			want:   regexp.Regexp{},
			err:    "could not decode an empty value to a regexp.Regexp: must have a non-empty value",
			decode: true,
		},
	}

	hook := configuration.StringToRegexpHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, result)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.want, result)

				pattern := result.(regexp.Regexp)

				var names []string
				for _, name := range pattern.SubexpNames() {
					if name != "" {
						names = append(names, name)
					}
				}

				if len(tc.wantGrps) != 0 {
					t.Run("MustHaveAllExpectedSubexpGroupNames", func(t *testing.T) {
						for _, name := range tc.wantGrps {
							assert.Contains(t, names, name)
						}
					})
					t.Run("MustNotHaveUnexpectedSubexpGroupNames", func(t *testing.T) {
						for _, name := range names {
							assert.Contains(t, tc.wantGrps, name)
						}
					})
				} else {
					t.Run("MustHaveNoSubexpGroupNames", func(t *testing.T) {
						assert.Len(t, names, 0)
					})
				}
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestStringToRegexpFuncPointers(t *testing.T) {
	testCases := []struct {
		desc     string
		have     interface{}
		want     interface{}
		err      string
		decode   bool
		wantGrps []string
	}{
		{
			desc:   "ShouldNotDecodeRegexpWithOpenParenthesis",
			have:   "hello(test one two",
			want:   &regexp.Regexp{},
			err:    "could not decode 'hello(test one two' to a *regexp.Regexp: error parsing regexp: missing closing ): `hello(test one two`",
			decode: true,
		},
		{
			desc:   "ShouldDecodeValidRegex",
			have:   "^(api|admin)$",
			want:   regexp.MustCompile(`^(api|admin)$`),
			decode: true,
		},
		{
			desc:     "ShouldDecodeValidRegexWithGroupNames",
			have:     "^(?P<area>api|admin)(one|two)$",
			want:     regexp.MustCompile(`^(?P<area>api|admin)(one|two)$`),
			decode:   true,
			wantGrps: []string{"area"},
		},
		{
			desc:   "ShouldNotDecodeFromInt32",
			have:   int32(20),
			want:   &regexp.Regexp{},
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeFromBool",
			have:   false,
			want:   &regexp.Regexp{},
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeToBool",
			have:   "^(?P<area>api|admin)(one|two)$",
			want:   &testTrue,
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeToInt32",
			have:   "^(?P<area>api|admin)(one|two)$",
			want:   &testZero,
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeToMailAddress",
			have:   "^(?P<area>api|admin)(one|two)$",
			want:   &mail.Address{},
			decode: false,
		},
		{
			desc:   "ShouldDecodeEmptyStringToNil",
			have:   "",
			want:   (*regexp.Regexp)(nil),
			decode: true,
		},
	}

	hook := configuration.StringToRegexpHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, result)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.want, result)

				pattern := result.(*regexp.Regexp)

				if tc.want == (*regexp.Regexp)(nil) {
					assert.Nil(t, pattern)
				} else {
					var names []string
					for _, name := range pattern.SubexpNames() {
						if name != "" {
							names = append(names, name)
						}
					}

					if len(tc.wantGrps) != 0 {
						t.Run("MustHaveAllExpectedSubexpGroupNames", func(t *testing.T) {
							for _, name := range tc.wantGrps {
								assert.Contains(t, names, name)
							}
						})
						t.Run("MustNotHaveUnexpectedSubexpGroupNames", func(t *testing.T) {
							for _, name := range names {
								assert.Contains(t, tc.wantGrps, name)
							}
						})
					} else {
						t.Run("MustHaveNoSubexpGroupNames", func(t *testing.T) {
							assert.Len(t, names, 0)
						})
					}
				}
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestStringToAddressHookFunc(t *testing.T) {
	mustParseAddress := func(a string) (addr schema.Address) {
		addrs, err := schema.NewAddressFromString(a)
		if err != nil {
			panic(err)
		}

		return *addrs
	}

	mustParseAddressPtr := func(a string) (addr *schema.Address) {
		addr, err := schema.NewAddressFromString(a)
		if err != nil {
			panic(err)
		}

		return addr
	}

	testCases := []struct {
		name     string
		have     interface{}
		expected interface{}
		err      string
		decode   bool
	}{
		{
			name:     "ShouldDecodeNonPtr",
			have:     "tcp://0.0.0.0:2020",
			expected: mustParseAddress("tcp://0.0.0.0:2020"),
			decode:   true,
		},
		{
			name:     "ShouldDecodePtr",
			have:     "tcp://0.0.0.0:2020",
			expected: mustParseAddressPtr("tcp://0.0.0.0:2020"),
			decode:   true,
		},
		{
			name:     "ShouldNotDecodeIntegerToCorrectType",
			have:     1,
			expected: schema.Address{},
			decode:   false,
		},
		{
			name:     "ShouldNotDecodeIntegerToCorrectTypePtr",
			have:     1,
			expected: &schema.Address{},
			decode:   false,
		},
		{
			name:     "ShouldNotDecodeIntegerPtrToCorrectType",
			have:     testInt32Ptr(1),
			expected: schema.Address{},
			decode:   false,
		},
		{
			name:     "ShouldNotDecodeIntegerPtrToCorrectTypePtr",
			have:     testInt32Ptr(1),
			expected: &schema.Address{},
			decode:   false,
		},
		{
			name:     "ShouldNotDecodeToString",
			have:     "tcp://0.0.0.0:2020",
			expected: "",
			decode:   false,
		},
		{
			name:     "ShouldNotDecodeToIntPtr",
			have:     "tcp://0.0.0.0:2020",
			expected: testInt32Ptr(1),
			decode:   false,
		},
		{
			name:     "ShouldNotDecodeToIntPtr",
			have:     "tcp://0.0.0.0:2020",
			expected: testInt32Ptr(1),
			decode:   false,
		},
		{
			name:     "ShouldFailDecode",
			have:     "tcp://&!@^#*&!@#&*@!:2020",
			expected: schema.Address{},
			err:      "could not decode 'tcp://&!@^#*&!@#&*@!:2020' to a schema.Address: could not parse string 'tcp://&!@^#*&!@#&*@!:2020' as address: expected format is [<scheme>://]<ip>[:<port>]: parse \"tcp://&!@^\": invalid character \"^\" in host name",
			decode:   false,
		},
	}

	hook := configuration.StringToAddressHookFunc()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.expected), tc.have)
			if tc.err != "" {
				assert.EqualError(t, err, tc.err)

				if !tc.decode {
					assert.Nil(t, actual)
				}
			} else {
				assert.NoError(t, err)

				if tc.decode {
					assert.Equal(t, tc.expected, actual)
				} else {
					assert.Equal(t, tc.have, actual)
				}
			}
		})
	}
}

func TestStringToRSAPrivateKeyHookFunc(t *testing.T) {
	var nilkey *rsa.PrivateKey

	testCases := []struct {
		desc   string
		have   interface{}
		want   interface{}
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeRSAPrivateKey",
			have:   x509PrivateKeyRSA1,
			want:   mustParseRSAPrivateKey(x509PrivateKeyRSA1),
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeToECDSAPrivateKey",
			have:   x509PrivateKeyRSA1,
			want:   &ecdsa.PrivateKey{},
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeEmptyKey",
			have:   "",
			want:   nilkey,
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeECDSAKeyToRSAKey",
			have:   x509PrivateKeyEC1,
			want:   nilkey,
			decode: true,
			err:    "could not decode to a *rsa.PrivateKey: the data is for a *ecdsa.PrivateKey not a *rsa.PrivateKey",
		},
		{
			desc:   "ShouldNotDecodeBadRSAPrivateKey",
			have:   x509PrivateKeyRSA2,
			want:   nilkey,
			decode: true,
			err:    "could not decode to a *rsa.PrivateKey: failed to parse PEM block containing the key",
		},
	}

	hook := configuration.StringToRSAPrivateKeyHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, result)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.want, result)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestStringToX509CertificateHookFunc(t *testing.T) {
	var nilkey *x509.Certificate

	testCases := []struct {
		desc   string
		have   interface{}
		want   interface{}
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeRSACertificate",
			have:   x509CertificateRSA1,
			want:   mustParseX509Certificate(x509CertificateRSA1),
			decode: true,
		},
		{
			desc:   "ShouldDecodeECDSACertificate",
			have:   x509CACertificateECDSA,
			want:   mustParseX509Certificate(x509CACertificateECDSA),
			decode: true,
		},
		{
			desc:   "ShouldDecodeRSACACertificate",
			have:   x509CACertificateRSA,
			want:   mustParseX509Certificate(x509CACertificateRSA),
			decode: true,
		},
		{
			desc:   "ShouldDecodeECDSACACertificate",
			have:   x509CACertificateECDSA,
			want:   mustParseX509Certificate(x509CACertificateECDSA),
			decode: true,
		},
		{
			desc:   "ShouldDecodeEmptyCertificateToNil",
			have:   "",
			want:   nilkey,
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeECDSAKeyToCertificate",
			have:   x509PrivateKeyEC1,
			want:   nilkey,
			decode: true,
			err:    "could not decode to a *x509.Certificate: the data is for a *ecdsa.PrivateKey not a *x509.Certificate",
		},
		{
			desc:   "ShouldNotDecodeBadRSAPrivateKeyToCertificate",
			have:   x509PrivateKeyRSA2,
			want:   nilkey,
			decode: true,
			err:    "could not decode to a *x509.Certificate: failed to parse PEM block containing the key",
		},
	}

	hook := configuration.StringToX509CertificateHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, result)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.want, result)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, result)
			}
		})
	}
}

func TestStringToX509CertificateChainHookFunc(t *testing.T) {
	var nilkey *schema.X509CertificateChain

	testCases := []struct {
		desc      string
		have      interface{}
		expected  interface{}
		err, verr string
		decode    bool
	}{
		{
			desc:     "ShouldDecodeRSACertificate",
			have:     x509CertificateRSA1,
			expected: mustParseX509CertificateChain(x509CertificateRSA1),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSACertificateNoPtr",
			have:     x509CertificateRSA1,
			expected: *mustParseX509CertificateChain(x509CertificateRSA1),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSACertificateChain",
			have:     buildChain(x509CertificateRSA1, x509CACertificateRSA),
			expected: mustParseX509CertificateChain(x509CertificateRSA1, x509CACertificateRSA),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSACertificateChainNoPtr",
			have:     buildChain(x509CertificateRSA1, x509CACertificateRSA),
			expected: *mustParseX509CertificateChain(x509CertificateRSA1, x509CACertificateRSA),
			decode:   true,
		},
		{
			desc:     "ShouldNotDecodeBadRSACertificateChain",
			have:     buildChain(x509CertificateRSA1, x509CACertificateECDSA),
			expected: mustParseX509CertificateChain(x509CertificateRSA1, x509CACertificateECDSA),
			verr:     "certificate #1 in chain is not signed properly by certificate #2 in chain: x509: signature algorithm specifies an RSA public key, but have public key of type *ecdsa.PublicKey",
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSACertificate",
			have:     x509CACertificateECDSA,
			expected: mustParseX509CertificateChain(x509CACertificateECDSA),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSACACertificate",
			have:     x509CACertificateRSA,
			expected: mustParseX509CertificateChain(x509CACertificateRSA),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSACACertificate",
			have:     x509CACertificateECDSA,
			expected: mustParseX509CertificateChain(x509CACertificateECDSA),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeEmptyCertificateToNil",
			have:     "",
			expected: nilkey,
			decode:   true,
		},
		{
			desc:     "ShouldDecodeEmptyCertificateToEmptyStruct",
			have:     "",
			expected: schema.X509CertificateChain{},
			decode:   true,
		},
		{
			desc:     "ShouldNotDecodeECDSAKeyToCertificate",
			have:     x509PrivateKeyEC1,
			expected: nilkey,
			decode:   true,
			err:      "could not decode to a *schema.X509CertificateChain: the PEM data chain contains a EC PRIVATE KEY but only certificates are expected",
		},
		{
			desc:     "ShouldNotDecodeBadRSAPrivateKeyToCertificate",
			have:     x509PrivateKeyRSA2,
			expected: nilkey,
			decode:   true,
			err:      "could not decode to a *schema.X509CertificateChain: invalid PEM block",
		},
	}

	hook := configuration.StringToX509CertificateChainHookFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.expected), tc.have)
			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, actual)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.expected, actual)

				if tc.expected == nilkey {
					break
				}

				switch chain := actual.(type) {
				case *schema.X509CertificateChain:
					require.NotNil(t, chain)
					if tc.verr == "" {
						assert.NoError(t, chain.Validate())
					} else {
						assert.EqualError(t, chain.Validate(), tc.verr)
					}
				case schema.X509CertificateChain:
					require.NotNil(t, chain)
					if tc.verr == "" {
						assert.NoError(t, chain.Validate())
					} else {
						assert.EqualError(t, chain.Validate(), tc.verr)
					}
				}
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			}
		})
	}
}

var (
	x509PrivateKeyRSA1 = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA6z1LOg1ZCqb0lytXWZ+MRBpMHEXOoTOLYgfZXt1IYyE3Z758
cyalk0NYQhY5cZDsXPYWPvAHiPMUxutWkoxFwby56S+AbIMa3/Is+ILrHRJs8Exn
ZkpyrYFxPX12app2kErdmAkHSx0Z5/kuXiz96PHs8S8/ZbyZolLHzdfLtSzjvRm5
Zue5iFzsf19NJz5CIBfv8g5lRwtE8wNJoRSpn1xq7fqfuA0weDNFPzjlNWRLy6aa
rK7qJexRkmkCs4sLgyl+9NODYJpvmN8E1yhyC27E0joI6rBFVW7Ihv+cSPCdDzGp
EWe81x3AeqAa3mjVqkiq4u4Z2i8JDgBaPboqJwIDAQABAoIBAAFdLZ58jVOefDSU
L8F5R1rtvBs93GDa56f926jNJ6pLewLC+/2+757W+SAI+PRLntM7Kg3bXm/Q2QH+
Q1Y+MflZmspbWCdI61L5GIGoYKyeers59i+FpvySj5GHtLQRiTZ0+Kv1AXHSDWBm
9XneUOqU3IbZe0ifu1RRno72/VtjkGXbW8Mkkw+ohyGbIeTx/0/JQ6sSNZTT3Vk7
8i4IXptq3HSF0/vqZuah8rShoeNq72pD1YLM9YPdL5by1QkDLnqATDiCpLBTCaNV
I8sqYEun+HYbQzBj8ZACG2JVZpEEidONWQHw5BPWO95DSZYrVnEkuCqeH+u5vYt7
CHuJ3AECgYEA+W3v5z+j91w1VPHS0VB3SCDMouycAMIUnJPAbt+0LPP0scUFsBGE
hPAKddC54pmMZRQ2KIwBKiyWfCrJ8Xz8Yogn7fJgmwTHidJBr2WQpIEkNGlK3Dzi
jXL2sh0yC7sHvn0DqiQ79l/e7yRbSnv2wrTJEczOOH2haD7/tBRyCYECgYEA8W+q
E9YyGvEltnPFaOxofNZ8LHVcZSsQI5b6fc0iE7fjxFqeXPXEwGSOTwqQLQRiHn9b
CfPmIG4Vhyq0otVmlPvUnfBZ2OK+tl5X2/mQFO3ROMdvpi0KYa994uqfJdSTaqLn
jjoKFB906UFHnDQDLZUNiV1WwnkTglgLc+xrd6cCgYEAqqthyv6NyBTM3Tm2gcio
Ra9Dtntl51LlXZnvwy3IkDXBCd6BHM9vuLKyxZiziGx+Vy90O1xI872cnot8sINQ
Am+dur/tAEVN72zxyv0Y8qb2yfH96iKy9gxi5s75TnOEQgAygLnYWaWR2lorKRUX
bHTdXBOiS58S0UzCFEslGIECgYBqkO4SKWYeTDhoKvuEj2yjRYyzlu28XeCWxOo1
otiauX0YSyNBRt2cSgYiTzhKFng0m+QUJYp63/wymB/5C5Zmxi0XtWIDADpLhqLj
HmmBQ2Mo26alQ5YkffBju0mZyhVzaQop1eZi8WuKFV1FThPlB7hc3E0SM5zv2Grd
tQnOWwKBgQC40yZY0PcjuILhy+sIc0Wvh7LUA7taSdTye149kRvbvsCDN7Jh75lM
USjhLXY0Nld2zBm9r8wMb81mXH29uvD+tDqqsICvyuKlA/tyzXR+QTr7dCVKVwu0
1YjCJ36UpTsLre2f8nOSLtNmRfDPtbOE2mkOoO9dD9UU0XZwnvn9xw==
-----END RSA PRIVATE KEY-----`

	x509PrivateKeyRSA2 = `
-----BEGIN RSA PRIVATE KEY-----
bad key
-----END RSA PRIVATE KEY-----`

	x509PrivateKeyEC1 = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIMn970LSn8aKVhBM4vyUmpZyEdCT4riN+Lp4QU04zUhYoAoGCCqGSM49
AwEHoUQDQgAEMD69n22nd78GmaRDzy/s7muqhbc/OEnFS2mNtiRAA5FaX+kbkCB5
8pu/k2jkaSVNZtBYKPVAibHkhvakjVb66A==
-----END EC PRIVATE KEY-----`

	x509CertificateRSA1 = `
-----BEGIN CERTIFICATE-----
MIIC5TCCAc2gAwIBAgIQfBUmKLmEvMqS6S9auKCY2DANBgkqhkiG9w0BAQsFADAT
MREwDwYDVQQKEwhBdXRoZWxpYTAeFw0yMjA5MDgxMDA5MThaFw0yMzA5MDgxMDA5
MThaMBMxETAPBgNVBAoTCEF1dGhlbGlhMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEApqno1cOpDcgKmOJqeDQGIGH5/ZnqcJ4xud6eOUfbDqel3b0RkAQX
mFYWEDO/PDOAOjYk/xSwZGo3jDofOHGhrKstQqLdweHGfme5NXYHJda7nGv/OY5q
zUuEG4xBVgUsvbshWZ18H+bIQpwiP6tDAabxc0B7J15F1pArK8QN4pDTfsqZDwMi
Qyo638XfUbDzEVZRbdDKxHz5g0w2vFdXon8uOxRRb0+zlHF9nM4PiESNgiUIYeua
8Q5yP10SY2k9zlQ/OFJ4XhQmioCJvNjJE/TSc5/ECub2n7hTZhN5TGKagukZ5NAy
KgbvNYW+CN+H4pFJt/9WptiDfBqhlUvjnwIDAQABozUwMzAOBgNVHQ8BAf8EBAMC
BaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIwADANBgkqhkiG9w0B
AQsFAAOCAQEAH9veGzfWqXxsa5s2KHV2Jzed9V8KSs1Qy9QKRez1i2OMvMPh2DRM
RLzAAp/XigjxLQF/LFXuoFW0Qg8BRb44iRgZrCiqVOUnd3xTrS/CcFExnpQI4F12
/U70o97rkTonCOHmUUW6vQfWSXR/GU3/faRLJjiqcpWLZhTQNrnsip1ym3B2NMdk
gMKkT8Acx1DX48MvTE4+DyqCS8TlJbacBJ2RFFELKu3jYnVNyrb0ywLxoCtWqBBE
veVj+VMn9hNY1u5uydLsUDOlT5QyQcEuUzjjdhsJKEgDE5daNtB2OJJnd9IOMzUA
hasPZETCCKabTpWiEPw1Cn/ZRqya0SZqFg==
-----END CERTIFICATE-----
`

	x509CACertificateRSA = `
-----BEGIN CERTIFICATE-----
MIIDBDCCAeygAwIBAgIRAJfz0dHS9UkDngE55lUPdu4wDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNMjIwOTA4MDk1OTI1WhcNMjMwOTA4MDk1
OTI1WjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBALfivbwq9r5X8N+NSbNHVuKbCb9f9vD5Xw2pOjSVvVjFkWQ1YKJu
JGx9yskhHBZTBt76cInipA+0PqCBrBrjij1lh2StvzRVuQwgFG6H01LxBPi0JyYv
Is94F6PHr6fSBgFWB5GNQ797KQIOdIr057uEFbp0eBMxxqiQ9gdyD0HPretrx1Uy
kHuF6jck958combn9luHW0i53mt8706j7UAhxFqu9YUeklTM1VqUiRm5+nJKIdNA
LiDMGVAuoxjhF6aIgY0yh5mL5mKtYYzhtA8WryrMzBgFRUGzHCSI1TNisA8wSf2T
Z2JhbFHrFPR5fiSqAEHok3UXu++wsfl/lisCAwEAAaNTMFEwDgYDVR0PAQH/BAQD
AgKkMA8GA1UdJQQIMAYGBFUdJQAwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU
OSXG42bCPNuWeP0ahScUMVjxe/wwDQYJKoZIhvcNAQELBQADggEBAFRnubHiYy1H
PGODKA++UY9eAmmaCJxzuWpY8FY9fBz8/VBzcp8vaURPmmQ/34QcqfaSHDM2jIaL
dQ2o9Ae5NjbRzLB6a5DcVO50oHG4BHP1ix4Bt3POr8J80KgA9pOIyAQqbAlFBSzQ
l9yrzVULyf+qpUmByRf5qy2kQJOBfMJbn5j+BprWKwbcI8OAZWWSLItTXqJDrFTk
OMZK4wZ6KiZM07KWMlwW/CE0QRzDk5MXfbwRt4D8pyx6rGKqI7QRusjm5osIpHZV
26FdBdBvEhq4i8UsmDsQqH3iMY1AKmojZToZb5rStOZWHO/BZZ7nT2bscNjwm0E8
6E2l6czk8ss=
-----END CERTIFICATE-----`

	x509CACertificateECDSA = `
-----BEGIN CERTIFICATE-----
MIIBdzCCAR2gAwIBAgIQUzb62irYb/7B2H0c1AbriDAKBggqhkjOPQQDAjATMREw
DwYDVQQKEwhBdXRoZWxpYTAeFw0yMjA5MDgxMDEzNDZaFw0yMzA5MDgxMDEzNDZa
MBMxETAPBgNVBAoTCEF1dGhlbGlhMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
b/EiIBpmifCI34JdI7luetygue2rTtoNH0QXhtrjMuZNugT29LUz+DobZQxvGsOY
4TXzAQXq4gnTb7enNWFgsaNTMFEwDgYDVR0PAQH/BAQDAgKkMA8GA1UdJQQIMAYG
BFUdJQAwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUxlDPBKHKawuvhtQTN874
TeCEKjkwCgYIKoZIzj0EAwIDSAAwRQIgAQeV01FZ/VkSERwaRKTeXAXxmKyc/05O
uDv6M2spMi0CIQC8uOSMcv11vp1ylsGg38N6XYA+GQa1BHRd79+91hC+7w==
-----END CERTIFICATE-----`
)

func mustParseRSAPrivateKey(data string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(data))
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		panic("not pem encoded")
	}

	if block.Type != "RSA PRIVATE KEY" {
		panic("not private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return key
}

func mustParseX509Certificate(data string) *x509.Certificate {
	block, _ := pem.Decode([]byte(data))
	if block == nil || len(block.Bytes) == 0 {
		panic("not a PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}

	return cert
}

func buildChain(pems ...string) string {
	buf := bytes.Buffer{}

	for i, data := range pems {
		if i != 0 {
			buf.WriteString("\n")
		}

		buf.WriteString(data)
	}

	return buf.String()
}

func mustParseX509CertificateChain(datas ...string) *schema.X509CertificateChain {
	chain, err := schema.NewX509CertificateChain(buildChain(datas...))
	if err != nil {
		panic(err)
	}

	return chain
}

func testInt32Ptr(i int32) *int32 {
	return &i
}

func testTimeDurationPtr(t time.Duration) *time.Duration {
	return &t
}

var (
	testTrue   = true
	testZero   int32
	testString = ""
)
