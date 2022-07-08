package configuration_test

import (
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
			err:      "could not decode 'tcp://&!@^#*&!@#&*@!:2020' to a Address: could not parse string 'tcp://&!@^#*&!@#&*@!:2020' as address: expected format is [<scheme>://]<ip>[:<port>]: parse \"tcp://&!@^\": invalid character \"^\" in host name",
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
