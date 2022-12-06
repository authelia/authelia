package utils

import (
	"fmt"
	"strconv"
	"time"
)

// StandardizeDurationString converts units of time that stdlib is unaware of to hours.
func StandardizeDurationString(input string) (output string, err error) {
	if input == "" {
		return "0s", nil
	}

	matches := reDurationStandard.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return "", fmt.Errorf("could not parse '%s' as a duration", input)
	}

	var d int

	for _, match := range matches {
		if d, err = strconv.Atoi(match[1]); err != nil {
			return "", fmt.Errorf("could not parse the numeric portion of '%s' in duration string '%s': %w", match[0], input, err)
		}

		unit := match[2]

		switch {
		case IsStringInSlice(unit, standardDurationUnits):
			output += fmt.Sprintf("%d%s", d, unit)
		case unit == DurationUnitDays:
			output += fmt.Sprintf("%dh", d*HoursInDay)
		case unit == DurationUnitWeeks:
			output += fmt.Sprintf("%dh", d*HoursInWeek)
		case unit == DurationUnitMonths:
			output += fmt.Sprintf("%dh", d*HoursInMonth)
		case unit == DurationUnitYears:
			output += fmt.Sprintf("%dh", d*HoursInYear)
		default:
			return "", fmt.Errorf("could not parse the units portion of '%s' in duration string '%s': the unit '%s' is not valid", match[0], input, unit)
		}
	}

	return output, nil
}

// ParseDurationString standardizes a duration string with StandardizeDurationString then uses time.ParseDuration to
// convert it into a time.Duration.
func ParseDurationString(input string) (duration time.Duration, err error) {
	if reDurationSeconds.MatchString(input) {
		var seconds int

		if seconds, err = strconv.Atoi(input); err != nil {
			return 0, nil
		}

		return time.Second * time.Duration(seconds), nil
	}

	var out string

	if out, err = StandardizeDurationString(input); err != nil {
		return 0, err
	}

	return time.ParseDuration(out)
}

// UnixNanoTimeToWin32Epoch converts a unix timestamp in nanosecond format to win32 epoch format.
func UnixNanoTimeToWin32Epoch(nano int64) (t uint64) {
	return uint64(nano/100) + timeUnixEpochAsWin32Epoch
}
