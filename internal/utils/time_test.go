package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestShouldParseDurationString(t *testing.T) {
	input := "1h10M"
	duration, err := ParseDurationString(input)
	assert.NoError(t, err)
	assert.Equal(t, 70*time.Minute, duration)
}

func TestShouldParseDurationStringWithZeroValues(t *testing.T) {
	input := "0h10M"
	duration, err := ParseDurationString(input)
	assert.NoError(t, err)
	assert.Equal(t, 10*time.Minute, duration)
}

func TestShouldParseDurationStringWithRepeatingUnits(t *testing.T) {
	input := "10M10M"
	duration, err := ParseDurationString(input)
	assert.NoError(t, err)
	assert.Equal(t, 20*time.Minute, duration)
}

func TestShouldParseDurationStringWithSpacingBetweenItems(t *testing.T) {
	input := "1h 10M"
	duration, err := ParseDurationString(input)
	assert.NoError(t, err)
	assert.Equal(t, 70*time.Minute, duration)
}

func TestShouldNotParseDurationStringWithOutOfOrderQuantitiesAndUnits(t *testing.T) {
	input := "h1M10"
	duration, err := ParseDurationString(input)
	assert.EqualError(t, err, "could not convert the input string of h1M10 into a duration")
	assert.Equal(t, time.Duration(0), duration)
}

func TestShouldNotParseBadDurationString(t *testing.T) {
	input := "1h10x"
	duration, err := ParseDurationString(input)
	assert.EqualError(t, err, "could not convert the input string of 1h10x into a duration")
	assert.Equal(t, time.Duration(0), duration)
}

func TestShouldParseSecondsString(t *testing.T) {
	input := "100"
	duration, err := ParseDurationString(input)
	assert.NoError(t, err)
	assert.Equal(t, 100*time.Second, duration)
}
