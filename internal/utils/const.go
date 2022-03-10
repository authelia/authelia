package utils

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const (
	windows         = "windows"
	testStringInput = "abcdefghijkl"

	// RFC3339Zero is the default value for time.Time.Unix().
	RFC3339Zero = int64(-62135596800)

	// TLS13 is the textual representation of TLS 1.3.
	TLS13 = "1.3"

	// TLS12 is the textual representation of TLS 1.2.
	TLS12 = "1.2"

	// TLS11 is the textual representation of TLS 1.1.
	TLS11 = "1.1"

	// TLS10 is the textual representation of TLS 1.0.
	TLS10 = "1.0"

	clean   = "clean"
	tagged  = "tagged"
	unknown = "unknown"
)

const (
	// Hour is an int based representation of the time unit.
	Hour = time.Minute * 60 //nolint:revive

	// Day is an int based representation of the time unit.
	Day = Hour * 24

	// Week is an int based representation of the time unit.
	Week = Day * 7

	// Year is an int based representation of the time unit.
	Year = Day * 365

	// Month is an int based representation of the time unit.
	Month = Year / 12
)

const (
	errFmtLinuxNotFound = "open %s: no such file or directory"
)

var (
	standardDurationUnits = []string{"ns", "us", "µs", "μs", "ms", "s", "m", "h"}
	reDurationSeconds     = regexp.MustCompile(`^\d+$`)
	reDurationStandard    = regexp.MustCompile(`(?P<Duration>[1-9]\d*?)(?P<Unit>[^\d\s]+)`)
)

// Duration unit types.
const (
	DurationUnitDays   = "d"
	DurationUnitWeeks  = "w"
	DurationUnitMonths = "M"
	DurationUnitYears  = "y"
)

// Number of hours in particular measurements of time.
const (
	HoursInDay   = 24
	HoursInWeek  = HoursInDay * 7
	HoursInMonth = HoursInDay * 30
	HoursInYear  = HoursInDay * 365
)

var (
	// AlphaNumericCharacters are literally just valid alphanumeric chars.
	AlphaNumericCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var htmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&#34;",
	"'", "&#39;",
)

// ErrTimeoutReached error thrown when a timeout is reached.
var ErrTimeoutReached = errors.New("timeout reached")

// ErrTLSVersionNotSupported returned when an unknown TLS version supplied.
var ErrTLSVersionNotSupported = errors.New("supplied tls version isn't supported")
