package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShouldParseDurationString(t *testing.T) {
	duration, err := ParseDurationString("1h")
	assert.NoError(t, err)
	assert.Equal(t, 60*time.Minute, duration)
}

func TestShouldParseDurationStringAllUnits(t *testing.T) {
	duration, err := ParseDurationString("1y")
	assert.NoError(t, err)
	assert.Equal(t, Year, duration)

	duration, err = ParseDurationString("1M")
	assert.NoError(t, err)
	assert.Equal(t, Month, duration)

	duration, err = ParseDurationString("1w")
	assert.NoError(t, err)
	assert.Equal(t, Week, duration)

	duration, err = ParseDurationString("1d")
	assert.NoError(t, err)
	assert.Equal(t, Day, duration)

	duration, err = ParseDurationString("1h")
	assert.NoError(t, err)
	assert.Equal(t, Hour, duration)

	duration, err = ParseDurationString("1s")
	assert.NoError(t, err)
	assert.Equal(t, time.Second, duration)
}

func TestShouldParseSecondsString(t *testing.T) {
	duration, err := ParseDurationString("100")
	assert.NoError(t, err)
	assert.Equal(t, 100*time.Second, duration)
}

func TestShouldNotParseDurationStringWithOutOfOrderQuantitiesAndUnits(t *testing.T) {
	duration, err := ParseDurationString("h1")
	assert.EqualError(t, err, "could not convert the input string of h1 into a duration")
	assert.Equal(t, time.Duration(0), duration)
}

func TestShouldNotParseBadDurationString(t *testing.T) {
	duration, err := ParseDurationString("10x")
	assert.EqualError(t, err, "could not convert the input string of 10x into a duration")
	assert.Equal(t, time.Duration(0), duration)
}

func TestShouldNotParseDurationStringWithMultiValueUnits(t *testing.T) {
	duration, err := ParseDurationString("10ms")
	assert.EqualError(t, err, "could not convert the input string of 10ms into a duration")
	assert.Equal(t, time.Duration(0), duration)
}

func TestShouldNotParseDurationStringWithLeadingZero(t *testing.T) {
	duration, err := ParseDurationString("005h")
	assert.EqualError(t, err, "could not convert the input string of 005h into a duration")
	assert.Equal(t, time.Duration(0), duration)
}

func TestShouldTimeIntervalsMakeSense(t *testing.T) {
	assert.Equal(t, Hour, time.Minute*60)
	assert.Equal(t, Day, Hour*24)
	assert.Equal(t, Week, Day*7)
	assert.Equal(t, Year, Day*365)
	assert.Equal(t, Month, Year/12)
}
