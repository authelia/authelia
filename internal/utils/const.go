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

// ErrTLSVersionNotSupported returned when an unknown TLS version supplied.
var ErrTLSVersionNotSupported = errors.New("supplied TLS version isn't supported")

// TLS13 is the textual representation of TLS 1.3.
const TLS13 = "1.3"

// TLS12 is the textual representation of TLS 1.2.
const TLS12 = "1.2"

// TLS11 is the textual representation of TLS 1.1.
const TLS11 = "1.1"

// TLS10 is the textual representation of TLS 1.0.
const TLS10 = "1.0"
