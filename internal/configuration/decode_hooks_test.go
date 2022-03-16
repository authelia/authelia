package configuration

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStringToURLHookFunc(t *testing.T) {
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
