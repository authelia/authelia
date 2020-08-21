package utils

import (
	"errors"
	"regexp"
	"time"
)

// ErrTimeoutReached error thrown when a timeout is reached.
var ErrTimeoutReached = errors.New("timeout reached")
var parseDurationRegexp = regexp.MustCompile(`^(?P<Duration>[1-9]\d*?)(?P<Unit>[smhdwMy])?$`)

// Hour is an int based representation of the time unit.
const Hour = time.Minute * 60

// Day is an int based representation of the time unit.
const Day = Hour * 24

// Week is an int based representation of the time unit.
const Week = Day * 7

// Year is an int based representation of the time unit.
const Year = Day * 365

// Month is an int based representation of the time unit.
const Month = Year / 12

// RFC3339Zero is the default value for time.Time.Unix().
const RFC3339Zero = int64(-62135596800)

const testStringInput = "abcdefghijkl"

// AlphaNumericCharacters are literally just valid alphanumeric chars.
var AlphaNumericCharacters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
