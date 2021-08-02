package utils

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// ParseDurationString parses a string to a duration
// Duration notations are an integer followed by a unit
// Units are s = second, m = minute, d = day, w = week, M = month, y = year
// Example 1y is the same as 1 year.
func ParseDurationString(input string) (time.Duration, error) {
	var duration time.Duration

	matches := parseDurationRegexp.FindStringSubmatch(input)

	switch {
	case len(matches) == 3 && matches[2] != "":
		d, _ := strconv.Atoi(matches[1])

		switch matches[2] {
		case "y":
			duration = time.Duration(d) * Year
		case "M":
			duration = time.Duration(d) * Month
		case "w":
			duration = time.Duration(d) * Week
		case "d":
			duration = time.Duration(d) * Day
		case "h":
			duration = time.Duration(d) * Hour
		case "m":
			duration = time.Duration(d) * time.Minute
		case "s":
			duration = time.Duration(d) * time.Second
		}
	case input == "0" || len(matches) == 3:
		seconds, err := strconv.Atoi(input)
		if err != nil {
			return 0, fmt.Errorf("could not convert the input string of %s into a duration: %s", input, err)
		}

		duration = time.Duration(seconds) * time.Second
	case input != "":
		// Throw this error if input is anything other than a blank string, blank string will default to a duration of nothing
		return 0, fmt.Errorf("could not convert the input string of %s into a duration", input)
	}

	return duration, nil
}

// UnixNanoTimeToWin32Epoch converts a unix timestamp in nanosecond format to win32 epoch format.
func UnixNanoTimeToWin32Epoch(nano int64) (t uint64) {
	return uint64(nano/100) + unixEpochAsWin32Epoch
}

// Win32EpochToTime returns the relevant time.Time for a given win32 epoch time.
func Win32EpochToTime(epoch uint64) (result time.Time, err error) {
	if epoch < unixEpochAsWin32Epoch {
		return result, errors.New("can't convert that epoch to native time as it is before the unix epoch")
	}

	offset := epoch - unixEpochAsWin32Epoch
	seconds := offset / 10000000
	remainder := offset % 10000000

	return time.Unix(int64(seconds), int64(remainder)), nil
}
