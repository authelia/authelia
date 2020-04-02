package utils

import (
	"errors"
	"regexp"
	"time"
)

// ErrTimeoutReached error thrown when a timeout is reached
var ErrTimeoutReached = errors.New("timeout reached")

var parseDurationRegexp = regexp.MustCompile(`\d+[smhdwMy]`)
var parseDurationFullRegexp = regexp.MustCompile(`^(\d+[smhdwMy])+$`)
var parseDurationSecondsRegexp = regexp.MustCompile(`^\d+$`)
var whitespace = regexp.MustCompile(`\s+`)

const Hour = time.Minute * 60
const Day = Hour * 24
const Week = Day * 7
const Year = Day * 365
const Month = Year / 12
