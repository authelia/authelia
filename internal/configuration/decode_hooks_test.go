package configuration

import (
	"net/url"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToURLHookFunc_ShouldNotParseStrings(t *testing.T) {
	hook := StringToURLHookFunc()

	var (
		from = "https://google.com/abc?a=123"

		result interface{}
		err    error

		resultTo    string
		resultPtrTo *time.Time
		ok          bool
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(resultTo), from)
	assert.NoError(t, err)

	resultTo, ok = result.(string)
	assert.True(t, ok)
	assert.Equal(t, from, resultTo)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(resultPtrTo), from)
	assert.NoError(t, err)

	resultTo, ok = result.(string)
	assert.True(t, ok)
	assert.Equal(t, from, resultTo)
}

func TestStringToURLHookFunc_ShouldParseEmptyString(t *testing.T) {
	hook := StringToURLHookFunc()

	var (
		from = ""

		result interface{}
		err    error

		resultTo    url.URL
		resultPtrTo *url.URL

		ok bool
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(resultTo), from)
	assert.NoError(t, err)

	resultTo, ok = result.(url.URL)
	assert.True(t, ok)
	assert.Equal(t, "", resultTo.String())

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(resultPtrTo), from)
	assert.NoError(t, err)

	resultPtrTo, ok = result.(*url.URL)
	assert.True(t, ok)
	assert.Nil(t, resultPtrTo)
}

func TestStringToURLHookFunc_ShouldNotParseBadURLs(t *testing.T) {
	hook := StringToURLHookFunc()

	var (
		from = "*(!&@#(!*^$%"

		result interface{}
		err    error

		resultTo    url.URL
		resultPtrTo *url.URL
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(resultTo), from)
	assert.EqualError(t, err, "could not parse '*(!&@#(!*^$%' as a URL: parse \"*(!&@#(!*^$%\": invalid URL escape \"%\"")
	assert.Nil(t, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(resultPtrTo), from)
	assert.EqualError(t, err, "could not parse '*(!&@#(!*^$%' as a URL: parse \"*(!&@#(!*^$%\": invalid URL escape \"%\"")
	assert.Nil(t, result)
}

func TestStringToURLHookFunc_ShouldParseURLs(t *testing.T) {
	hook := StringToURLHookFunc()

	var (
		from = "https://google.com/abc?a=123"

		result interface{}
		err    error

		resultTo    url.URL
		resultPtrTo *url.URL

		ok bool
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(resultTo), from)
	assert.NoError(t, err)

	resultTo, ok = result.(url.URL)
	assert.True(t, ok)
	assert.Equal(t, "https", resultTo.Scheme)
	assert.Equal(t, "google.com", resultTo.Host)
	assert.Equal(t, "/abc", resultTo.Path)
	assert.Equal(t, "a=123", resultTo.RawQuery)

	resultPtrTo, ok = result.(*url.URL)
	assert.False(t, ok)
	assert.Nil(t, resultPtrTo)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(resultPtrTo), from)
	assert.NoError(t, err)

	resultPtrTo, ok = result.(*url.URL)
	assert.True(t, ok)
	assert.NotNil(t, resultPtrTo)

	assert.Equal(t, "https", resultTo.Scheme)
	assert.Equal(t, "google.com", resultTo.Host)
	assert.Equal(t, "/abc", resultTo.Path)
	assert.Equal(t, "a=123", resultTo.RawQuery)

	resultTo, ok = result.(url.URL)
	assert.False(t, ok)
}

func TestToTimeDurationHookFunc_ShouldParse_String(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from     = "1h"
		expected = time.Hour

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, &expected, result)
}

func TestToTimeDurationHookFunc_ShouldParse_String_Years(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from     = "1y"
		expected = time.Hour * 24 * 365

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, &expected, result)
}

func TestToTimeDurationHookFunc_ShouldParse_String_Months(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from     = "1M"
		expected = time.Hour * 24 * 30

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, &expected, result)
}

func TestToTimeDurationHookFunc_ShouldParse_String_Weeks(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from     = "1w"
		expected = time.Hour * 24 * 7

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, &expected, result)
}

func TestToTimeDurationHookFunc_ShouldParse_String_Days(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from     = "1d"
		expected = time.Hour * 24

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, &expected, result)
}

func TestToTimeDurationHookFunc_ShouldNotParseAndRaiseErr_InvalidString(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from = "abc"

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.EqualError(t, err, "could not parse 'abc' as a duration")
	assert.Nil(t, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.EqualError(t, err, "could not parse 'abc' as a duration")
	assert.Nil(t, result)
}

func TestToTimeDurationHookFunc_ShouldParse_Int(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from     = 60
		expected = time.Second * 60

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, &expected, result)
}

func TestToTimeDurationHookFunc_ShouldParse_Int32(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from     = int32(120)
		expected = time.Second * 120

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, &expected, result)
}

func TestToTimeDurationHookFunc_ShouldParse_Int64(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from     = int64(30)
		expected = time.Second * 30

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, &expected, result)
}

func TestToTimeDurationHookFunc_ShouldParse_Duration(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from     = time.Second * 30
		expected = time.Second * 30

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, &expected, result)
}

func TestToTimeDurationHookFunc_ShouldNotParse_Int64ToString(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from = int64(30)

		to     string
		ptrTo  *string
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, from, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, from, result)
}

func TestToTimeDurationHookFunc_ShouldNotParse_FromBool(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from = true

		to     string
		ptrTo  *string
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, from, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, from, result)
}

func TestToTimeDurationHookFunc_ShouldParse_FromZero(t *testing.T) {
	hook := ToTimeDurationHookFunc()

	var (
		from     = 0
		expected = time.Duration(0)

		to     time.Duration
		ptrTo  *time.Duration
		result interface{}
		err    error
	)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(to), from)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = hook(reflect.TypeOf(from), reflect.TypeOf(ptrTo), from)
	assert.NoError(t, err)
	assert.Equal(t, &expected, result)
}

func TestStringToRegexpFunc(t *testing.T) {
	wantRegexp := func(regexpStr string) regexp.Regexp {
		pattern := regexp.MustCompile(regexpStr)

		return *pattern
	}

	testCases := []struct {
		desc           string
		have           interface{}
		want           regexp.Regexp
		wantPtr        *regexp.Regexp
		wantErr        string
		wantGroupNames []string
	}{
		{
			desc:    "should not parse regexp with open paren",
			have:    "hello(test one two",
			wantErr: "could not parse 'hello(test one two' as regexp: error parsing regexp: missing closing ): `hello(test one two`",
		},
		{
			desc:    "should parse valid regex",
			have:    "^(api|admin)$",
			want:    wantRegexp("^(api|admin)$"),
			wantPtr: regexp.MustCompile("^(api|admin)$"),
		},
		{
			desc:           "should parse valid regex with named groups",
			have:           "^(?P<area>api|admin)$",
			want:           wantRegexp("^(?P<area>api|admin)$"),
			wantPtr:        regexp.MustCompile("^(?P<area>api|admin)$"),
			wantGroupNames: []string{"area"},
		},
	}

	hook := StringToRegexpFunc()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Run("non-ptr", func(t *testing.T) {
				result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.want), tc.have)
				if tc.wantErr != "" {
					assert.EqualError(t, err, tc.wantErr)
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					require.Equal(t, tc.want, result)

					resultRegexp := result.(regexp.Regexp)

					actualNames := resultRegexp.SubexpNames()
					var names []string

					for _, name := range actualNames {
						if name != "" {
							names = append(names, name)
						}
					}

					if len(tc.wantGroupNames) != 0 {
						t.Run("must_have_all_expected_subexp_group_names", func(t *testing.T) {
							for _, name := range tc.wantGroupNames {
								assert.Contains(t, names, name)
							}
						})
						t.Run("must_not_have_unexpected_subexp_group_names", func(t *testing.T) {
							for _, name := range names {
								assert.Contains(t, tc.wantGroupNames, name)
							}
						})
					} else {
						t.Run("must_have_no_subexp_group_names", func(t *testing.T) {
							assert.Len(t, names, 0)
						})
					}
				}
			})
			t.Run("ptr", func(t *testing.T) {
				result, err := hook(reflect.TypeOf(tc.have), reflect.TypeOf(tc.wantPtr), tc.have)
				if tc.wantErr != "" {
					assert.EqualError(t, err, tc.wantErr)
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tc.wantPtr, result)

					resultRegexp := result.(*regexp.Regexp)

					actualNames := resultRegexp.SubexpNames()
					var names []string

					for _, name := range actualNames {
						if name != "" {
							names = append(names, name)
						}
					}

					if len(tc.wantGroupNames) != 0 {
						t.Run("must_have_all_expected_names", func(t *testing.T) {
							for _, name := range tc.wantGroupNames {
								assert.Contains(t, names, name)
							}
						})

						t.Run("must_not_have_unexpected_names", func(t *testing.T) {
							for _, name := range names {
								assert.Contains(t, tc.wantGroupNames, name)
							}
						})
					} else {
						assert.Len(t, names, 0)
					}
				}
			})
		})
	}
}
