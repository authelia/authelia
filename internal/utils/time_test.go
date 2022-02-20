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
	assert.Equal(t, time.Hour*24*365, duration)

	duration, err = ParseDurationString("1M")

	assert.NoError(t, err)
	assert.Equal(t, time.Hour*24*30, duration)

	duration, err = ParseDurationString("1w")

	assert.NoError(t, err)
	assert.Equal(t, time.Hour*24*7, duration)

	duration, err = ParseDurationString("1d")

	assert.NoError(t, err)
	assert.Equal(t, time.Hour*24, duration)

	duration, err = ParseDurationString("1h")

	assert.NoError(t, err)
	assert.Equal(t, time.Hour, duration)

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

	assert.EqualError(t, err, "could not parse 'h1' as a duration string")
	assert.Equal(t, time.Duration(0), duration)
}

func TestShouldNotParseBadDurationString(t *testing.T) {
	duration, err := ParseDurationString("10x")

	assert.EqualError(t, err, "could not parse the units portion of '10x' in duration string '10x': the unit 'x' is not valid")
	assert.Equal(t, time.Duration(0), duration)
}

func TestShouldParseDurationStringWithMultiValueUnits(t *testing.T) {
	duration, err := ParseDurationString("10ms")

	assert.NoError(t, err)
	assert.Equal(t, time.Duration(10)*time.Millisecond, duration)
}

func TestShouldParseDurationStringWithLeadingZero(t *testing.T) {
	duration, err := ParseDurationString("005h")

	assert.NoError(t, err)
	assert.Equal(t, time.Duration(5)*time.Hour, duration)
}

func TestShouldParseMultiUnitValues(t *testing.T) {
	duration, err := ParseDurationString("1d3w10ms")

	assert.NoError(t, err)
	assert.Equal(t,
		(time.Hour*time.Duration(24))+
			(time.Hour*time.Duration(24)*time.Duration(7)*time.Duration(3))+
			(time.Millisecond*time.Duration(10)), duration)
}

func TestShouldParseDuplicateUnitValues(t *testing.T) {
	duration, err := ParseDurationString("1d4d2d")

	assert.NoError(t, err)
	assert.Equal(t,
		(time.Hour*time.Duration(24))+
			(time.Hour*time.Duration(24)*time.Duration(4))+
			(time.Hour*time.Duration(24)*time.Duration(2)), duration)
}

func TestShouldTimeIntervalsMakeSense(t *testing.T) {
	assert.Equal(t, Hour, time.Minute*60)
	assert.Equal(t, Day, Hour*24)
	assert.Equal(t, Week, Day*7)
	assert.Equal(t, Year, Day*365)
	assert.Equal(t, Month, Year/12)
}
