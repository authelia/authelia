package configuration

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToTimeDurationFunc_ShouldParse_String(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldParse_String_Years(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldParse_String_Months(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldParse_String_Weeks(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldParse_String_Days(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldNotParseAndRaiseErr_InvalidString(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldParse_Int(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldParse_Int32(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldParse_Int64(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldParse_Duration(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldNotParse_Int64ToString(t *testing.T) {
	hook := ToTimeDurationFunc()

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

func TestToTimeDurationFunc_ShouldNotParse_FromBool(t *testing.T) {
	hook := ToTimeDurationFunc()

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
