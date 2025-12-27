package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StandardizeDurationString converts units of time that stdlib is unaware of to hours.
func StandardizeDurationString(input string) (output string, err error) {
	if input == "" {
		return "0s", nil
	}

	input = strings.ReplaceAll(input, "and", "")

	matches := reDurationStandard.FindAllStringSubmatch(strings.ReplaceAll(input, " ", ""), -1)

	if len(matches) == 0 {
		return "", fmt.Errorf("could not parse '%s' as a duration", input)
	}

	var (
		o string
		q int
	)

	for _, match := range matches {
		if q, err = strconv.Atoi(match[1]); err != nil {
			return "", err
		}

		if o, err = standardizeQuantityAndUnits(q, match[2]); err != nil {
			return "", fmt.Errorf("could not parse the units portion of '%s' in duration string '%s': %w", match[0], input, err)
		}

		output += o
	}

	return output, nil
}

func standardizeQuantityAndUnits(qty int, unit string) (output string, err error) {
	switch {
	case IsStringInSlice(unit, standardDurationUnits):
		return fmt.Sprintf("%d%s", qty, unit), nil
	case len(unit) == 1:
		switch unit {
		case DurationUnitDays:
			return fmt.Sprintf("%dh", qty*HoursInDay), nil
		case DurationUnitWeeks:
			return fmt.Sprintf("%dh", qty*HoursInWeek), nil
		case DurationUnitMonths:
			return fmt.Sprintf("%dh", qty*HoursInMonth), nil
		case DurationUnitYears:
			return fmt.Sprintf("%dh", qty*HoursInYear), nil
		}
	default:
		switch unit {
		case "millisecond", "milliseconds":
			return fmt.Sprintf("%dms", qty), nil
		case "second", "seconds":
			return fmt.Sprintf("%ds", qty), nil
		case "minute", "minutes":
			return fmt.Sprintf("%dm", qty), nil
		case "hour", "hours":
			return fmt.Sprintf("%dh", qty), nil
		case "day", "days":
			return fmt.Sprintf("%dh", qty*HoursInDay), nil
		case "week", "weeks":
			return fmt.Sprintf("%dh", qty*HoursInWeek), nil
		case "month", "months":
			return fmt.Sprintf("%dh", qty*HoursInMonth), nil
		case "year", "years":
			return fmt.Sprintf("%dh", qty*HoursInYear), nil
		}
	}

	return "", fmt.Errorf("the unit '%s' is not valid", unit)
}

// ParseDurationString standardizes a duration string with StandardizeDurationString then uses time.ParseDuration to
// convert it into a time.Duration.
func ParseDurationString(input string) (duration time.Duration, err error) {
	if reOnlyNumeric.MatchString(input) {
		var seconds int

		if seconds, err = strconv.Atoi(input); err != nil {
			return 0, err
		}

		return time.Second * time.Duration(seconds), nil
	}

	var out string

	if out, err = StandardizeDurationString(input); err != nil {
		return 0, err
	}

	return time.ParseDuration(out)
}

// ParseTimeString attempts to parse a string with several time formats.
func ParseTimeString(input string) (t time.Time, err error) {
	return ParseTimeStringWithLayouts(input, StandardTimeLayouts)
}

// ParseTimeStringWithLayouts attempts to parse a string with several time formats. The format with the most matching
// characters is returned.
func ParseTimeStringWithLayouts(input string, layouts []string) (match time.Time, err error) {
	_, match, err = matchParseTimeStringWithLayouts(input, layouts)

	return
}

func matchParseTimeStringWithLayouts(input string, layouts []string) (index int, match time.Time, err error) {
	if reOnlyNumeric.MatchString(input) {
		var u int64

		if u, err = strconv.ParseInt(input, 10, 64); err != nil {
			return -999, match, fmt.Errorf("time value was detected as an integer but the integer could not be parsed: %w", err)
		}

		switch {
		case u > 32503554000000: // 2999-12-31 00:00:00 in unix time (milliseconds).
			return -3, time.UnixMicro(u), nil
		case u > 946645200000: // 2000-01-01 00:00:00 in unix time (milliseconds).
			return -2, time.UnixMilli(u), nil
		default:
			return -1, time.Unix(u, 0), nil
		}
	}

	var layout string

	for index, layout = range layouts {
		if match, err = time.Parse(layout, input); err == nil {
			if len(match.Format(layout))-len(input) == 0 {
				return index, match, nil
			}
		}
	}

	return -998, time.UnixMilli(0), fmt.Errorf("failed to find a suitable time layout for time '%s'", input)
}

// UnixNanoTimeToMicrosoftNTEpoch converts a unix timestamp in nanosecond format to win32 epoch format.
func UnixNanoTimeToMicrosoftNTEpoch(nano int64) (t uint64) {
	if nano >= 0 {
		return uint64(nano/100) + timeUnixEpochAsMicrosoftNTEpoch //nolint:gosec // This is a gated condition and is checked,
	}

	return timeUnixEpochAsMicrosoftNTEpoch
}
