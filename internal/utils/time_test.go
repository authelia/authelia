package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDurationString_ShouldParseDurationString(t *testing.T) {
	duration, err := ParseDurationString("1h")

	assert.NoError(t, err)
	assert.Equal(t, 60*time.Minute, duration)
}

func TestParseDurationString_ShouldParseBlankString(t *testing.T) {
	duration, err := ParseDurationString("")

	assert.NoError(t, err)
	assert.Equal(t, time.Second*0, duration)
}

func TestParseDurationString_ShouldParseDurationStringAllUnits(t *testing.T) {
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

func TestParseDurationString_ShouldParseSecondsString(t *testing.T) {
	duration, err := ParseDurationString("100")

	assert.NoError(t, err)
	assert.Equal(t, 100*time.Second, duration)
}

func TestParseDurationString_ShouldNotParseDurationStringWithOutOfOrderQuantitiesAndUnits(t *testing.T) {
	duration, err := ParseDurationString("h1")

	assert.EqualError(t, err, "could not parse 'h1' as a duration")
	assert.Equal(t, time.Duration(0), duration)
}

func TestParseDurationString_ShouldNotParseBadDurationString(t *testing.T) {
	duration, err := ParseDurationString("10x")

	assert.EqualError(t, err, "could not parse the units portion of '10x' in duration string '10x': the unit 'x' is not valid")
	assert.Equal(t, time.Duration(0), duration)
}

func TestParseDurationString_ShouldParseDurationStringWithMultiValueUnits(t *testing.T) {
	duration, err := ParseDurationString("10ms")

	assert.NoError(t, err)
	assert.Equal(t, time.Duration(10)*time.Millisecond, duration)
}

func TestParseDurationString_ShouldParseDurationStringWithLeadingZero(t *testing.T) {
	duration, err := ParseDurationString("005h")

	assert.NoError(t, err)
	assert.Equal(t, time.Duration(5)*time.Hour, duration)
}

func TestParseDurationString_ShouldParseMultiUnitValues(t *testing.T) {
	duration, err := ParseDurationString("1d3w10ms")

	assert.NoError(t, err)
	assert.Equal(t,
		(time.Hour*time.Duration(24))+
			(time.Hour*time.Duration(24)*time.Duration(7)*time.Duration(3))+
			(time.Millisecond*time.Duration(10)), duration)
}

func TestParseDurationString_ShouldParseDuplicateUnitValues(t *testing.T) {
	duration, err := ParseDurationString("1d4d2d")

	assert.NoError(t, err)
	assert.Equal(t,
		(time.Hour*time.Duration(24))+
			(time.Hour*time.Duration(24)*time.Duration(4))+
			(time.Hour*time.Duration(24)*time.Duration(2)), duration)
}

func TestStandardizeDurationString_ShouldParseStringWithSpaces(t *testing.T) {
	result, err := StandardizeDurationString("1d 1h 20m")

	assert.NoError(t, err)
	assert.Equal(t, result, "24h1h20m")
}

func TestShouldTimeIntervalsMakeSense(t *testing.T) {
	assert.Equal(t, Hour, time.Minute*60)
	assert.Equal(t, Day, Hour*24)
	assert.Equal(t, Week, Day*7)
	assert.Equal(t, Year, Day*365)
	assert.Equal(t, Month, Year/12)
}

func TestShouldConvertKnownUnixNanoTimeToKnownWin32Epoch(t *testing.T) {
	exampleNanoTime := int64(1626234411 * 1000000000)
	win32Epoch := uint64(132707080110000000)

	assert.Equal(t, win32Epoch, UnixNanoTimeToMicrosoftNTEpoch(exampleNanoTime))
	assert.Equal(t, timeUnixEpochAsMicrosoftNTEpoch, UnixNanoTimeToMicrosoftNTEpoch(0))
}
