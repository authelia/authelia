package configuration_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math"
	"net"
	"net/mail"
	"net/url"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestStringToMailAddressHookFunc(t *testing.T) {
	testCases := []struct {
		desc   string
		have   any
		want   any
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
		have   any
		want   any
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
		have   any
		want   any
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
		have   any
		want   any
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
		have   any
		want   any
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
			desc:   "ShouldDecodeInt8ToSeconds",
			have:   int8(90),
			want:   time.Second * 90,
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt16ToSeconds",
			have:   int16(90),
			want:   time.Second * 90,
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt32ToSeconds",
			have:   int32(90),
			want:   time.Second * 90,
			decode: true,
		},
		{
			desc:   "ShouldDecodeFloat64ToSeconds",
			have:   float64(90),
			want:   time.Second * 90,
			decode: true,
		},
		{
			desc:   "ShouldDecodeFloat64ToSeconds",
			have:   math.MaxFloat64,
			want:   time.Duration(math.MaxInt64),
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
			desc:   "ShouldSkipParsingBoolean",
			have:   true,
			want:   time.Duration(0),
			decode: false,
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
		have   any
		want   any
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

func TestToRefreshIntervalDurationHookFunc(t *testing.T) {
	testCases := []struct {
		desc   string
		have   any
		want   any
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeFourtyFiveSeconds",
			have:   "45s",
			want:   schema.NewRefreshIntervalDuration(time.Second * 45),
			decode: true,
		},
		{
			desc:   "ShouldDecodeOneMinute",
			have:   "1m",
			want:   schema.NewRefreshIntervalDuration(time.Minute),
			decode: true,
		},
		{
			desc:   "ShouldDecodeTwoHours",
			have:   "2h",
			want:   schema.NewRefreshIntervalDuration(time.Hour * 2),
			decode: true,
		},
		{
			desc:   "ShouldDecodeThreeDays",
			have:   "3d",
			want:   schema.NewRefreshIntervalDuration(time.Hour * 24 * 3),
			decode: true,
		},
		{
			desc:   "ShouldDecodeFourWeeks",
			have:   "4w",
			want:   schema.NewRefreshIntervalDuration(time.Hour * 24 * 7 * 4),
			decode: true,
		},
		{
			desc:   "ShouldDecodeFiveMonths",
			have:   "5M",
			want:   schema.NewRefreshIntervalDuration(time.Hour * 24 * 30 * 5),
			decode: true,
		},
		{
			desc:   "ShouldDecodeSixYears",
			have:   "6y",
			want:   schema.NewRefreshIntervalDuration(time.Hour * 24 * 365 * 6),
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeInvalidString",
			have:   "abc",
			want:   schema.RefreshIntervalDuration{},
			err:    "could not decode 'abc' to a schema.RefreshIntervalDuration: could not parse 'abc' as a duration",
			decode: true,
		},
		{
			desc:   "ShouldDecodeIntToSeconds",
			have:   60,
			want:   schema.NewRefreshIntervalDuration(time.Second * 60),
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt8ToSeconds",
			have:   int8(90),
			want:   schema.NewRefreshIntervalDuration(time.Second * 90),
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt16ToSeconds",
			have:   int16(90),
			want:   schema.NewRefreshIntervalDuration(time.Second * 90),
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt32ToSeconds",
			have:   int32(90),
			want:   schema.NewRefreshIntervalDuration(time.Second * 90),
			decode: true,
		},
		{
			desc:   "ShouldDecodeFloat64ToSeconds",
			have:   float64(90),
			want:   schema.NewRefreshIntervalDuration(time.Second * 90),
			decode: true,
		},
		{
			desc:   "ShouldDecodeFloat64ToSeconds",
			have:   math.MaxFloat64,
			want:   schema.NewRefreshIntervalDuration(time.Duration(math.MaxInt64)),
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt64ToSeconds",
			have:   int64(120),
			want:   schema.NewRefreshIntervalDuration(time.Second * 120),
			decode: true,
		},
		{
			desc:   "ShouldDecodeTimeDuration",
			have:   time.Second * 30,
			want:   schema.NewRefreshIntervalDuration(time.Second * 30),
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
			want:   schema.NewRefreshIntervalDuration(time.Duration(0)),
			decode: true,
		},
		{
			desc:   "ShouldSkipParsingBoolean",
			have:   true,
			want:   schema.RefreshIntervalDuration{},
			decode: false,
		},
		{
			desc: "ShouldNotDecodeFromBool",
			have: true,
			want: true,
		},
	}

	hook := configuration.ToRefreshIntervalDurationHookFunc()

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

func TestTestToRefreshIntervalDurationHookFuncPointer(t *testing.T) {
	testCases := []struct {
		desc   string
		have   any
		want   any
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeFourtyFiveSeconds",
			have:   "45s",
			want:   testRefreshIntervalDurationPtr(time.Second * 45),
			decode: true,
		},
		{
			desc:   "ShouldDecodeOneMinute",
			have:   "1m",
			want:   testRefreshIntervalDurationPtr(time.Minute),
			decode: true,
		},
		{
			desc:   "ShouldDecodeTwoHours",
			have:   "2h",
			want:   testRefreshIntervalDurationPtr(time.Hour * 2),
			decode: true,
		},
		{
			desc:   "ShouldDecodeThreeDays",
			have:   "3d",
			want:   testRefreshIntervalDurationPtr(time.Hour * 24 * 3),
			decode: true,
		},
		{
			desc:   "ShouldDecodeFourWeeks",
			have:   "4w",
			want:   testRefreshIntervalDurationPtr(time.Hour * 24 * 7 * 4),
			decode: true,
		},
		{
			desc:   "ShouldDecodeFiveMonths",
			have:   "5M",
			want:   testRefreshIntervalDurationPtr(time.Hour * 24 * 30 * 5),
			decode: true,
		},
		{
			desc:   "ShouldDecodeSixYears",
			have:   "6y",
			want:   testRefreshIntervalDurationPtr(time.Hour * 24 * 365 * 6),
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeInvalidString",
			have:   "abc",
			want:   testRefreshIntervalDurationPtr(time.Duration(0)),
			err:    "could not decode 'abc' to a *schema.RefreshIntervalDuration: could not parse 'abc' as a duration",
			decode: true,
		},
		{
			desc:   "ShouldDecodeIntToSeconds",
			have:   60,
			want:   testRefreshIntervalDurationPtr(time.Second * 60),
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt32ToSeconds",
			have:   int32(90),
			want:   testRefreshIntervalDurationPtr(time.Second * 90),
			decode: true,
		},
		{
			desc:   "ShouldDecodeInt64ToSeconds",
			have:   int64(120),
			want:   testRefreshIntervalDurationPtr(time.Second * 120),
			decode: true,
		},
		{
			desc:   "ShouldDecodeTimeDuration",
			have:   time.Second * 30,
			want:   testRefreshIntervalDurationPtr(time.Second * 30),
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
			want:   testRefreshIntervalDurationPtr(time.Duration(0)),
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeFromBool",
			have:   true,
			want:   &testTrue,
			decode: false,
		},
	}

	hook := configuration.ToRefreshIntervalDurationHookFunc()

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
		have     any
		want     any
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

				var names []string

				pattern := result.(regexp.Regexp)
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
		have     any
		want     any
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
	testCases := []struct {
		name     string
		have     any
		expected any
		err      string
		decode   bool
	}{
		{
			name:     "ShouldDecodeNonPtr",
			have:     "tcp://0.0.0.0:2020",
			expected: MustParseAddress("tcp://0.0.0.0:2020"),
			decode:   true,
		},
		{
			name:     "ShouldDecodePtr",
			have:     "tcp://0.0.0.0:2020",
			expected: MustParseAddressPtr("tcp://0.0.0.0:2020"),
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
			err:      "could not decode 'tcp://&!@^#*&!@#&*@!:2020' to a schema.Address: could not parse string 'tcp://&!@^#*&!@#&*@!:2020' as address: expected format is [<scheme>://]<hostname>[:<port>]: parse \"tcp://&!@^\": invalid character \"^\" in host name",
			decode:   false,
		},
		{
			name:     "ShouldDecodeTCP",
			have:     "tcp://127.0.0.1",
			expected: schema.AddressTCP{Address: MustParseAddress("tcp://127.0.0.1")},
			err:      "",
			decode:   true,
		},
		{
			name:     "ShouldDecodeTCPPtr",
			have:     "tcp://127.0.0.1",
			expected: &schema.AddressTCP{Address: MustParseAddress("tcp://127.0.0.1")},
			err:      "",
			decode:   true,
		},
		{
			name:     "ShouldDecodeUDP",
			have:     "udp://127.0.0.1",
			expected: schema.AddressUDP{Address: MustParseAddress("udp://127.0.0.1")},
			err:      "",
			decode:   true,
		},
		{
			name:     "ShouldDecodeUDPPtr",
			have:     "udp://127.0.0.1",
			expected: &schema.AddressUDP{Address: MustParseAddress("udp://127.0.0.1")},
			err:      "",
			decode:   true,
		},
		{
			name:     "ShouldDecodeLDAP",
			have:     "ldap://127.0.0.1",
			expected: schema.AddressLDAP{Address: MustParseAddress("ldap://127.0.0.1")},
			err:      "",
			decode:   true,
		},
		{
			name:     "ShouldDecodeLDAPPtr",
			have:     "ldap://127.0.0.1",
			expected: &schema.AddressLDAP{Address: MustParseAddress("ldap://127.0.0.1")},
			err:      "",
			decode:   true,
		},
		{
			name:     "ShouldDecodeSMTP",
			have:     "smtp://127.0.0.1",
			expected: schema.AddressSMTP{Address: MustParseAddress("smtp://127.0.0.1")},
			err:      "",
			decode:   true,
		},
		{
			name:     "ShouldDecodeSMTPPtr",
			have:     "smtp://127.0.0.1",
			expected: &schema.AddressSMTP{Address: MustParseAddress("smtp://127.0.0.1")},
			err:      "",
			decode:   true,
		},
		{
			name:     "ShouldFailDecodeTCP",
			have:     "@@@@@@@",
			expected: schema.AddressTCP{Address: MustParseAddress("tcp://127.0.0.1")},
			err:      "could not decode '@@@@@@@' to a schema.AddressTCP: error validating the address: the url 'tcp://%40%40%40%40%40%40@' appears to have user info but this is not valid for addresses",
			decode:   false,
		},
		{
			name:     "ShouldFailDecodeUDP",
			have:     "@@@@@@@",
			expected: schema.AddressUDP{Address: MustParseAddress("udp://127.0.0.1")},
			err:      "could not decode '@@@@@@@' to a schema.AddressUDP: error validating the address: the url 'udp://%40%40%40%40%40%40@' appears to have user info but this is not valid for addresses",
			decode:   false,
		},
		{
			name:     "ShouldFailDecodeLDAP",
			have:     "@@@@@@@",
			expected: schema.AddressLDAP{Address: MustParseAddress("ldap://127.0.0.1")},
			err:      "could not decode '@@@@@@@' to a schema.AddressLDAP: error validating the address: the url 'ldaps://%40%40%40%40%40%40@' appears to have user info but this is not valid for addresses",
			decode:   false,
		},
		{
			name:     "ShouldFailDecodeSMTP",
			have:     "@@@@@@@",
			expected: schema.AddressSMTP{Address: MustParseAddress("smtp://127.0.0.1")},
			err:      "could not decode '@@@@@@@' to a schema.AddressSMTP: error validating the address: the url 'smtp://%40%40%40%40%40%40@' appears to have user info but this is not valid for addresses",
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

func TestStringToPrivateKeyHookFunc(t *testing.T) {
	var (
		nilRSA   *rsa.PrivateKey
		nilECDSA *ecdsa.PrivateKey
		nilCert  *x509.Certificate
	)

	testCases := []struct {
		desc   string
		have   any
		want   any
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeRSA1024PrivateKey",
			have:   x509PrivateKeyRSA1024,
			want:   MustParsePKCS8RSAPrivateKey(x509PrivateKeyRSA1024),
			decode: true,
		},
		{
			desc:   "ShouldDecodeRSA2048PrivateKey",
			have:   x509PrivateKeyRSA2048,
			want:   MustParsePKCS8RSAPrivateKey(x509PrivateKeyRSA2048),
			decode: true,
		},
		{
			desc:   "ShouldDecodeRSA4096PrivateKey",
			have:   x509PrivateKeyRSA4096,
			want:   MustParsePKCS8RSAPrivateKey(x509PrivateKeyRSA4096),
			decode: true,
		},
		{
			desc:   "ShouldDecodeECDSAP224PrivateKey",
			have:   x509PrivateKeyECDSAP224,
			want:   MustParsePKCS8ECDSAPrivateKey(x509PrivateKeyECDSAP224),
			decode: true,
		},
		{
			desc:   "ShouldDecodeECDSAP256PrivateKey",
			have:   x509PrivateKeyECDSAP256,
			want:   MustParsePKCS8ECDSAPrivateKey(x509PrivateKeyECDSAP256),
			decode: true,
		},
		{
			desc:   "ShouldDecodeECDSAP384PrivateKey",
			have:   x509PrivateKeyECDSAP384,
			want:   MustParsePKCS8ECDSAPrivateKey(x509PrivateKeyECDSAP384),
			decode: true,
		},
		{
			desc:   "ShouldDecodeECDSAP521PrivateKey",
			have:   x509PrivateKeyECDSAP521,
			want:   MustParsePKCS8ECDSAPrivateKey(x509PrivateKeyECDSAP521),
			decode: true,
		},
		{
			desc:   "ShouldDecodeEd25519PrivateKey",
			have:   x509PrivateKeyEd25519,
			want:   MustParsePKCS8Ed25519PrivateKey(x509PrivateKeyEd25519),
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeToECDSAPrivateKey",
			have:   x509PrivateKeyRSA2048,
			want:   &ecdsa.PrivateKey{},
			decode: true,
			err:    "could not decode to a *ecdsa.PrivateKey: the data is for a *rsa.PrivateKey not a *ecdsa.PrivateKey",
		},
		{
			desc:   "ShouldNotDecodeEmptyRSAKey",
			have:   "",
			want:   nilRSA,
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeEmptyECDSAKey",
			have:   "",
			want:   nilECDSA,
			decode: true,
		},
		{
			desc:   "ShouldNotDecodeECDSAKeyToRSAKey",
			have:   x509PrivateKeyECDSAP521,
			want:   nilRSA,
			decode: true,
			err:    "could not decode to a *rsa.PrivateKey: the data is for a *ecdsa.PrivateKey not a *rsa.PrivateKey",
		},
		{
			desc:   "ShouldNotDecodeRSAKeyToECDSAKey",
			have:   x509PrivateKeyRSA2048,
			want:   nilECDSA,
			decode: true,
			err:    "could not decode to a *ecdsa.PrivateKey: the data is for a *rsa.PrivateKey not a *ecdsa.PrivateKey",
		},
		{
			desc:   "ShouldNotDecodeBadRSAPrivateKey",
			have:   x509PrivateKeyRSABad,
			want:   nilRSA,
			decode: true,
			err:    "could not decode to a *rsa.PrivateKey: error occurred attempting to parse PEM block: either no PEM block was supplied or it was malformed",
		},
		{
			desc:   "ShouldNotDecodeBadRSAPrivateKeyTrailingData",
			have:   strings.ReplaceAll(x509PrivateKeyRSA2048, "END PRIVATE KEY-----", "END PRIVATE KEY---------"),
			want:   nilRSA,
			decode: true,
			err:    "could not decode to a *rsa.PrivateKey: error occurred attempting to parse PEM block: either no PEM block was supplied or it was malformed",
		},
		{
			desc:   "ShouldNotDecodeBadRSAPrivateKeyTrailingDataDoubled",
			have:   x509PrivateKeyRSA2048 + x509PrivateKeyRSA2048,
			want:   nilRSA,
			decode: true,
			err:    "could not decode to a *rsa.PrivateKey: error occurred attempting to parse PEM block: the block either had trailing data or was otherwise malformed",
		},
		{
			desc:   "ShouldNotDecodeBadECDSAPrivateKey",
			have:   x509PrivateKeyECBad,
			want:   nilECDSA,
			decode: true,
			err:    "could not decode to a *ecdsa.PrivateKey: error occurred attempting to parse PEM block: either no PEM block was supplied or it was malformed",
		},
		{
			desc:   "ShouldNotDecodeCertificateToRSAPrivateKey",
			have:   x509CertificateRSA2048,
			want:   nilRSA,
			decode: true,
			err:    "could not decode to a *rsa.PrivateKey: the data is for a *x509.Certificate not a *rsa.PrivateKey",
		},
		{
			desc:   "ShouldNotDecodeCertificateToECDSAPrivateKey",
			have:   x509CertificateRSA2048,
			want:   nilECDSA,
			decode: true,
			err:    "could not decode to a *ecdsa.PrivateKey: the data is for a *x509.Certificate not a *ecdsa.PrivateKey",
		},
		{
			desc:   "ShouldNotDecodeRSAKeyToCertificate",
			have:   x509PrivateKeyRSA2048,
			want:   nilCert,
			decode: false,
		},
		{
			desc:   "ShouldNotDecodeECDSAKeyToCertificate",
			have:   x509PrivateKeyECDSAP521,
			want:   nilCert,
			decode: false,
		},
	}

	hook := configuration.StringToPrivateKeyHookFunc()

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
		have   any
		want   any
		err    string
		decode bool
	}{
		{
			desc:   "ShouldDecodeRSACertificate",
			have:   x509CertificateRSA2048,
			want:   MustParseX509Certificate(x509CertificateRSA2048),
			decode: true,
		},
		{
			desc:   "ShouldDecodeECDSACertificate",
			have:   x509CertificateECDSAP521,
			want:   MustParseX509Certificate(x509CertificateECDSAP521),
			decode: true,
		},
		{
			desc:   "ShouldDecodeEd25519Certificate",
			have:   x509CertificateEd25519,
			want:   MustParseX509Certificate(x509CertificateEd25519),
			decode: true,
		},
		{
			desc:   "ShouldDecodeRSACACertificate",
			have:   x509CACertificateRSA2048,
			want:   MustParseX509Certificate(x509CACertificateRSA2048),
			decode: true,
		},
		{
			desc:   "ShouldDecodeECDSACACertificate",
			have:   x509CACertificateECDSAP521,
			want:   MustParseX509Certificate(x509CACertificateECDSAP521),
			decode: true,
		},
		{
			desc:   "ShouldDecodeEd25519CACertificate",
			have:   x509CACertificateEd25519,
			want:   MustParseX509Certificate(x509CACertificateEd25519),
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
			have:   x509PrivateKeyECDSAP224,
			want:   nilkey,
			decode: true,
			err:    "could not decode to a *x509.Certificate: the data is for a *ecdsa.PrivateKey not a *x509.Certificate",
		},
		{
			desc:   "ShouldNotDecodeBadRSAPrivateKeyToCertificate",
			have:   x509PrivateKeyRSABad,
			want:   nilkey,
			decode: true,
			err:    "could not decode to a *x509.Certificate: error occurred attempting to parse PEM block: either no PEM block was supplied or it was malformed",
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

func TestStringToPasswordDigestHookFunc(t *testing.T) {
	var nilvalue *schema.PasswordDigest

	testCases := []struct {
		name     string
		have     any
		expected any
		err      string
		decode   bool
	}{
		{
			"ShouldParse",
			"$plaintext$example",
			MustParsePasswordDigest("$plaintext$example"),
			"",
			true,
		},
		{
			"ShouldParsePtr",
			"$plaintext$example",
			MustParsePasswordDigestPtr("$plaintext$example"),
			"",
			true,
		},
		{
			"ShouldNotParseUnknown",
			"$abc$example",
			schema.PasswordDigest{},
			"could not decode '$abc$example' to a schema.PasswordDigest: provided encoded hash has an invalid identifier: the identifier 'abc' is unknown to the decoder",
			false,
		},
		{
			"ShouldNotParseWrongType",
			"$abc$example",
			schema.TLSVersion{},
			"",
			false,
		},
		{
			"ShouldNotParseWrongTypePtr",
			"$abc$example",
			&schema.TLSVersion{},
			"",
			false,
		},
		{
			"ShouldNotParseEmptyString",
			"",
			schema.PasswordDigest{},
			"could not decode an empty value to a schema.PasswordDigest: must have a non-empty value",
			false,
		},
		{
			"ShouldParseEmptyStringPtr",
			"",
			nilvalue,
			"",
			true,
		},
	}

	hook := configuration.StringToPasswordDigestHookFunc()

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

func TestStringToTLSVersionHookFunc(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		expected any
		err      string
		decode   bool
	}{
		{
			"ShouldParseTLS1.3",
			"TLS1.3",
			schema.TLSVersion{Value: tls.VersionTLS13},
			"",
			true,
		},
		{
			"ShouldParseTLS1.3PTR",
			"TLS1.3",
			&schema.TLSVersion{Value: tls.VersionTLS13},
			"",
			true,
		},
		{
			"ShouldParseTLS1.2",
			"TLS1.2",
			schema.TLSVersion{Value: tls.VersionTLS12},
			"",
			true,
		},
		{
			"ShouldNotParseInt",
			1,
			&schema.TLSVersion{},
			"",
			false,
		},
		{
			"ShouldNotParseNonVersion",
			"1",
			&schema.TLSVersion{},
			"could not decode '1' to a *schema.TLSVersion: supplied tls version isn't supported",
			false,
		},
	}

	hook := configuration.StringToTLSVersionHookFunc()

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

func TestStringToX509CertificateChainHookFunc(t *testing.T) {
	var nilkey *schema.X509CertificateChain

	testCases := []struct {
		desc      string
		have      any
		expected  any
		err, verr string
		decode    bool
	}{
		{
			desc:     "ShouldDecodeRSA1024Certificate",
			have:     x509CertificateRSA1024,
			expected: MustParseX509CertificateChain(x509CertificateRSA1024),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA1024CertificateNoPtr",
			have:     x509CertificateRSA1024,
			expected: *MustParseX509CertificateChain(x509CertificateRSA1024),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA1024CACertificate",
			have:     x509CACertificateRSA1024,
			expected: MustParseX509CertificateChain(x509CACertificateRSA1024),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA1024CACertificateNoPtr",
			have:     x509CACertificateRSA1024,
			expected: *MustParseX509CertificateChain(x509CACertificateRSA1024),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA1024CertificateChain",
			have:     BuildChain(x509CertificateRSA1024, x509CACertificateRSA1024),
			expected: MustParseX509CertificateChain(x509CertificateRSA1024, x509CACertificateRSA1024),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA1024CertificateChainNoPtr",
			have:     BuildChain(x509CertificateRSA1024, x509CACertificateRSA1024),
			expected: *MustParseX509CertificateChain(x509CertificateRSA1024, x509CACertificateRSA1024),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA2048Certificate",
			have:     x509CertificateRSA2048,
			expected: MustParseX509CertificateChain(x509CertificateRSA2048),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA2048CertificateNoPtr",
			have:     x509CertificateRSA2048,
			expected: *MustParseX509CertificateChain(x509CertificateRSA2048),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA2048CACertificate",
			have:     x509CACertificateRSA2048,
			expected: MustParseX509CertificateChain(x509CACertificateRSA2048),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA2048CACertificateNoPtr",
			have:     x509CACertificateRSA2048,
			expected: *MustParseX509CertificateChain(x509CACertificateRSA2048),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA2048CertificateChain",
			have:     BuildChain(x509CertificateRSA2048, x509CACertificateRSA2048),
			expected: MustParseX509CertificateChain(x509CertificateRSA2048, x509CACertificateRSA2048),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA2048CertificateChainNoPtr",
			have:     BuildChain(x509CertificateRSA2048, x509CACertificateRSA2048),
			expected: *MustParseX509CertificateChain(x509CertificateRSA2048, x509CACertificateRSA2048),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA4096Certificate",
			have:     x509CertificateRSA4096,
			expected: MustParseX509CertificateChain(x509CertificateRSA4096),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA4096CertificateNoPtr",
			have:     x509CertificateRSA4096,
			expected: *MustParseX509CertificateChain(x509CertificateRSA4096),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA4096CACertificate",
			have:     x509CACertificateRSA4096,
			expected: MustParseX509CertificateChain(x509CACertificateRSA4096),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA4096CACertificateNoPtr",
			have:     x509CACertificateRSA4096,
			expected: *MustParseX509CertificateChain(x509CACertificateRSA4096),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA4096CertificateChain",
			have:     BuildChain(x509CertificateRSA4096, x509CACertificateRSA4096),
			expected: MustParseX509CertificateChain(x509CertificateRSA4096, x509CACertificateRSA4096),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeRSA4096CertificateChainNoPtr",
			have:     BuildChain(x509CertificateRSA4096, x509CACertificateRSA4096),
			expected: *MustParseX509CertificateChain(x509CertificateRSA4096, x509CACertificateRSA4096),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP224Certificate",
			have:     x509CertificateECDSAP224,
			expected: MustParseX509CertificateChain(x509CertificateECDSAP224),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP224CertificateNoPtr",
			have:     x509CertificateECDSAP224,
			expected: *MustParseX509CertificateChain(x509CertificateECDSAP224),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP224CACertificate",
			have:     x509CACertificateECDSAP224,
			expected: MustParseX509CertificateChain(x509CACertificateECDSAP224),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP224CACertificateNoPtr",
			have:     x509CACertificateECDSAP224,
			expected: *MustParseX509CertificateChain(x509CACertificateECDSAP224),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP224CertificateChain",
			have:     BuildChain(x509CertificateECDSAP224, x509CACertificateECDSAP224),
			expected: MustParseX509CertificateChain(x509CertificateECDSAP224, x509CACertificateECDSAP224),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP224CertificateChainNoPtr",
			have:     BuildChain(x509CertificateECDSAP224, x509CACertificateECDSAP224),
			expected: *MustParseX509CertificateChain(x509CertificateECDSAP224, x509CACertificateECDSAP224),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP256Certificate",
			have:     x509CertificateECDSAP256,
			expected: MustParseX509CertificateChain(x509CertificateECDSAP256),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP256CertificateNoPtr",
			have:     x509CertificateECDSAP256,
			expected: *MustParseX509CertificateChain(x509CertificateECDSAP256),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP256CACertificate",
			have:     x509CACertificateECDSAP256,
			expected: MustParseX509CertificateChain(x509CACertificateECDSAP256),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP256CACertificateNoPtr",
			have:     x509CACertificateECDSAP256,
			expected: *MustParseX509CertificateChain(x509CACertificateECDSAP256),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP256CertificateChain",
			have:     BuildChain(x509CertificateECDSAP256, x509CACertificateECDSAP256),
			expected: MustParseX509CertificateChain(x509CertificateECDSAP256, x509CACertificateECDSAP256),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP256CertificateChainNoPtr",
			have:     BuildChain(x509CertificateECDSAP256, x509CACertificateECDSAP256),
			expected: *MustParseX509CertificateChain(x509CertificateECDSAP256, x509CACertificateECDSAP256),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP384Certificate",
			have:     x509CertificateECDSAP384,
			expected: MustParseX509CertificateChain(x509CertificateECDSAP384),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP384CertificateNoPtr",
			have:     x509CertificateECDSAP384,
			expected: *MustParseX509CertificateChain(x509CertificateECDSAP384),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP384CACertificate",
			have:     x509CACertificateECDSAP384,
			expected: MustParseX509CertificateChain(x509CACertificateECDSAP384),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP384CACertificateNoPtr",
			have:     x509CACertificateECDSAP384,
			expected: *MustParseX509CertificateChain(x509CACertificateECDSAP384),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP384CertificateChain",
			have:     BuildChain(x509CertificateECDSAP384, x509CACertificateECDSAP384),
			expected: MustParseX509CertificateChain(x509CertificateECDSAP384, x509CACertificateECDSAP384),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP384CertificateChainNoPtr",
			have:     BuildChain(x509CertificateECDSAP384, x509CACertificateECDSAP384),
			expected: *MustParseX509CertificateChain(x509CertificateECDSAP384, x509CACertificateECDSAP384),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP521Certificate",
			have:     x509CertificateECDSAP521,
			expected: MustParseX509CertificateChain(x509CertificateECDSAP521),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP521CertificateNoPtr",
			have:     x509CertificateECDSAP521,
			expected: *MustParseX509CertificateChain(x509CertificateECDSAP521),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP521CACertificate",
			have:     x509CACertificateECDSAP521,
			expected: MustParseX509CertificateChain(x509CACertificateECDSAP521),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP521CACertificateNoPtr",
			have:     x509CACertificateECDSAP521,
			expected: *MustParseX509CertificateChain(x509CACertificateECDSAP521),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP521CertificateChain",
			have:     BuildChain(x509CertificateECDSAP521, x509CACertificateECDSAP521),
			expected: MustParseX509CertificateChain(x509CertificateECDSAP521, x509CACertificateECDSAP521),
			decode:   true,
		},
		{
			desc:     "ShouldDecodeECDSAP521CertificateChainNoPtr",
			have:     BuildChain(x509CertificateECDSAP521, x509CACertificateECDSAP521),
			expected: *MustParseX509CertificateChain(x509CertificateECDSAP521, x509CACertificateECDSAP521),
			decode:   true,
		},
		{
			desc:     "ShouldNotDecodeBadRSACertificateChain",
			have:     BuildChain(x509CertificateRSA2048, x509CACertificateECDSAP521),
			expected: MustParseX509CertificateChain(x509CertificateRSA2048, x509CACertificateECDSAP521),
			verr:     "certificate #1 in chain is not signed properly by certificate #2 in chain: x509: signature algorithm specifies an RSA public key, but have public key of type *ecdsa.PublicKey",
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
			have:     x509PrivateKeyECDSAP224,
			expected: nilkey,
			decode:   true,
			err:      "could not decode to a *schema.X509CertificateChain: the PEM data chain contains a PRIVATE KEY but only certificates are expected",
		},
		{
			desc:     "ShouldNotDecodeBadRSAPrivateKeyToCertificate",
			have:     x509PrivateKeyRSABad,
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

func TestStringToUUIDHookFunc(t *testing.T) {
	var nilkey *uuid.UUID

	toPtr := func(value uuid.UUID) *uuid.UUID {
		return &value
	}

	testCases := []struct {
		name      string
		have      any
		expected  any
		err, verr string
		decode    bool
	}{
		{
			name:     "ShouldNotDecodeEmptyString",
			have:     "",
			expected: uuid.UUID{},
			decode:   true,
			err:      "could not decode an empty value to a uuid.UUID: must have a non-empty value",
		},
		{
			name:     "ShouldDecodeEmptyStringNil",
			have:     "",
			expected: nilkey,
			decode:   true,
			err:      "",
		},
		{
			name:     "ShouldDecodeValid",
			have:     "cb69481e-8ff7-4039-93ec-0a2729a154a8",
			expected: uuid.MustParse("cb69481e-8ff7-4039-93ec-0a2729a154a8"),
			decode:   true,
			err:      "",
		},
		{
			name:     "ShouldDecodeValidPtr",
			have:     "cb69481e-8ff7-4039-93ec-0a2729a154a8",
			expected: toPtr(uuid.MustParse("cb69481e-8ff7-4039-93ec-0a2729a154a8")),
			decode:   true,
			err:      "",
		},
		{
			name:     "ShouldNotDecodeParseError",
			have:     "cb69481e-4039-93ec-0a2729a154a8",
			expected: uuid.UUID{},
			decode:   true,
			err:      "could not decode 'cb69481e-4039-93ec-0a2729a154a8' to a uuid.UUID: invalid UUID length: 31",
		},
		{
			name:     "ShouldNotDecodeParseErrorPtr",
			have:     "cb69481e-4039-93ec-0a2729a154a8",
			expected: nilkey,
			decode:   true,
			err:      "could not decode 'cb69481e-4039-93ec-0a2729a154a8' to a *uuid.UUID: invalid UUID length: 31",
		},
	}

	hook := configuration.StringToUUIDHookFunc()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.expected), tc.have)

			switch {
			case !tc.decode:
				assert.NoError(t, err)
				assert.Equal(t, tc.have, actual)
			case tc.err == "":
				assert.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			}
		})
	}
}

func TestStringToIPNetworksHookFunc(t *testing.T) {
	mustParseNet := func(in string) *net.IPNet {
		_, n, err := net.ParseCIDR(in)
		if err != nil {
			panic(err)
		}

		return n
	}

	testCases := []struct {
		name     string
		path     string
		have     *schema.Definitions
		expected TestConfigDefinitions
		err      string
	}{
		{
			"ShouldDecode",
			"decode_networks.yml",
			&schema.Definitions{},
			TestConfigDefinitions{
				Definitions: schema.Definitions{
					Network: map[string][]*net.IPNet{
						"single": {
							mustParseNet("192.168.0.1/32"),
						},
						"example": {
							mustParseNet("192.168.1.20/32"),
							mustParseNet("192.168.2.0/24"),
						},
					},
				},
			},
			"",
		},
		{
			"ShouldDecodeDefinitions",
			"decode_networks_abc.yml",
			&schema.Definitions{
				Network: map[string][]*net.IPNet{
					"abc": {
						mustParseNet("1.1.1.1/32"),
						mustParseNet("1.1.1.2/32"),
					},
				},
			},
			TestConfigDefinitions{
				Definitions: schema.Definitions{
					Network: map[string][]*net.IPNet{
						"example": {
							mustParseNet("192.168.1.20/32"),
							mustParseNet("192.168.2.0/24"),
							mustParseNet("1.1.1.1/32"),
							mustParseNet("1.1.1.2/32"),
						},
					},
				},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TestConfigDefinitions{}

			val := schema.NewStructValidator()
			_, err := configuration.LoadAdvanced(val, "", &result, tc.have, configuration.NewDefaultSourcesFiltered([]string{path.Join("./test_resources", tc.path)}, nil, configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter)...)

			if tc.err == "" {
				assert.NoError(t, err)

				assert.Equal(t, tc.expected, result)

				assert.Len(t, val.Errors(), 0)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

type TestConfigDefinitions struct {
	Definitions schema.Definitions `koanf:"definitions"`
}

var (
	x509PrivateKeyRSABad = `
-----BEGIN RSA PRIVATE KEY-----
bad key
-----END RSA PRIVATE KEY-----`

	x509PrivateKeyECBad = `
-----BEGIN EC PRIVATE KEY-----
bad key
-----END EC PRIVATE KEY-----`
)

func MustParsePKCS8RSAPrivateKey(data string) *rsa.PrivateKey {
	return MustParsePKCS8PrivateKey(data).(*rsa.PrivateKey)
}

func MustParsePKCS8ECDSAPrivateKey(data string) *ecdsa.PrivateKey {
	return MustParsePKCS8PrivateKey(data).(*ecdsa.PrivateKey)
}

func MustParsePKCS8Ed25519PrivateKey(data string) *ed25519.PrivateKey {
	key := MustParsePKCS8PrivateKey(data).(ed25519.PrivateKey)

	return &key
}

func MustParsePKCS8PrivateKey(data string) schema.CryptographicPrivateKey {
	block, _ := pem.Decode([]byte(data))
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		panic("not pem encoded")
	}

	if block.Type != "PRIVATE KEY" {
		panic("not PKCS8 private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	if pkey, ok := key.(schema.CryptographicPrivateKey); ok {
		return pkey
	}

	panic("key does not implement the required members")
}

func MustParseX509Certificate(data string) *x509.Certificate {
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

func BuildChain(pems ...string) string {
	buf := bytes.Buffer{}

	for i, data := range pems {
		if i != 0 {
			buf.WriteString("\n")
		}

		buf.WriteString(data)
	}

	return buf.String()
}

func MustParseX509CertificateChain(datas ...string) *schema.X509CertificateChain {
	chain, err := schema.NewX509CertificateChain(BuildChain(datas...))
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

func testRefreshIntervalDurationPtr(t time.Duration) *schema.RefreshIntervalDuration {
	x := schema.NewRefreshIntervalDuration(t)

	return &x
}

var (
	testTrue   = true
	testZero   int32
	testString = ""
)

func MustParseAddress(input string) schema.Address {
	address, err := schema.NewAddress(input)
	if err != nil {
		panic(err)
	}

	addr := *address

	return addr
}

func MustParseAddressPtr(input string) *schema.Address {
	address, err := schema.NewAddress(input)
	if err != nil {
		panic(err)
	}

	return address
}

func MustParsePasswordDigest(input string) schema.PasswordDigest {
	digest, err := schema.DecodePasswordDigest(input)
	if err != nil {
		panic(err)
	}

	return *digest
}

func MustParsePasswordDigestPtr(input string) *schema.PasswordDigest {
	digest, err := schema.DecodePasswordDigest(input)
	if err != nil {
		panic(err)
	}

	return digest
}

const (
	pathCrypto = "./test_resources/crypto/%s.%s"
)

func MustLoadCryptoSet(alg string, legacy bool, extra ...string) (certCA, keyCA, cert, key string) {
	extraAlt := make([]string, len(extra))

	copy(extraAlt, extra)

	if legacy {
		extraAlt = append(extraAlt, "legacy")
	}

	return MustLoadCryptoRaw(true, alg, "crt", extra...), MustLoadCryptoRaw(true, alg, "pem", extra...), MustLoadCryptoRaw(false, alg, "crt", extraAlt...), MustLoadCryptoRaw(false, alg, "pem", extraAlt...)
}

func MustLoadCryptoRaw(ca bool, alg, ext string, extra ...string) string {
	var fparts []string

	if ca {
		fparts = append(fparts, "ca")
	}

	fparts = append(fparts, strings.ToLower(alg))

	if len(extra) != 0 {
		fparts = append(fparts, extra...)
	}

	var (
		data []byte
		err  error
	)
	if data, err = os.ReadFile(fmt.Sprintf(pathCrypto, strings.Join(fparts, "."), ext)); err != nil {
		panic(err)
	}

	return string(data)
}

var (
	x509CertificateRSA1024, x509CertificateRSA2048, x509CertificateRSA4096, x509CertificateEd25519                 string
	x509CACertificateRSA1024, x509CACertificateRSA2048, x509CACertificateRSA4096, x509CACertificateEd25519         string
	x509PrivateKeyRSA1024, x509PrivateKeyRSA2048, x509PrivateKeyRSA4096, x509PrivateKeyEd25519                     string
	x509CAPrivateKeyRSA1024, x509CAPrivateKeyRSA2048, x509CAPrivateKeyRSA4096, x509CAPrivateKeyEd25519             string
	x509CertificateECDSAP224, x509CertificateECDSAP256, x509CertificateECDSAP384, x509CertificateECDSAP521         string
	x509CACertificateECDSAP224, x509CACertificateECDSAP256, x509CACertificateECDSAP384, x509CACertificateECDSAP521 string
	x509PrivateKeyECDSAP224, x509PrivateKeyECDSAP256, x509PrivateKeyECDSAP384, x509PrivateKeyECDSAP521             string
	x509CAPrivateKeyECDSAP224, x509CAPrivateKeyECDSAP256, x509CAPrivateKeyECDSAP384, x509CAPrivateKeyECDSAP521     string
)

func init() {
	x509CACertificateRSA1024, x509CAPrivateKeyRSA1024, x509CertificateRSA1024, x509PrivateKeyRSA1024 = MustLoadCryptoSet("RSA", false, "1024")
	x509CACertificateRSA2048, x509CAPrivateKeyRSA2048, x509CertificateRSA2048, x509PrivateKeyRSA2048 = MustLoadCryptoSet("RSA", false, "2048")
	x509CACertificateRSA4096, x509CAPrivateKeyRSA4096, x509CertificateRSA4096, x509PrivateKeyRSA4096 = MustLoadCryptoSet("RSA", false, "4096")

	x509CACertificateECDSAP224, x509CAPrivateKeyECDSAP224, x509CertificateECDSAP224, x509PrivateKeyECDSAP224 = MustLoadCryptoSet("ECDSA", false, "P224")
	x509CACertificateECDSAP256, x509CAPrivateKeyECDSAP256, x509CertificateECDSAP256, x509PrivateKeyECDSAP256 = MustLoadCryptoSet("ECDSA", false, "P256")
	x509CACertificateECDSAP384, x509CAPrivateKeyECDSAP384, x509CertificateECDSAP384, x509PrivateKeyECDSAP384 = MustLoadCryptoSet("ECDSA", false, "P384")
	x509CACertificateECDSAP521, x509CAPrivateKeyECDSAP521, x509CertificateECDSAP521, x509PrivateKeyECDSAP521 = MustLoadCryptoSet("ECDSA", false, "P521")

	x509CACertificateEd25519, x509CAPrivateKeyEd25519, x509CertificateEd25519, x509PrivateKeyEd25519 = MustLoadCryptoSet("Ed25519", false)
}
