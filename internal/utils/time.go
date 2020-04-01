package utils

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Parses a string and additively adds all duration notations together
// Duration notations are an integer followed by a unit
// Units are s = second, M = minute, d = day, w = week, m = month, y = year
// Example 1y10d1M is the same as 1 year, 10 days, 1 minute
func ParseDurationString(input string) (duration time.Duration, err error) {
	duration = 0
	err = nil

	sanitizedInput := whitespace.ReplaceAllString(input, "")
	if sanitizedInput != "" {
		inputBytes := []byte(sanitizedInput)
		if !parseDurationFullRegexp.Match(inputBytes) {
			if parseDurationSecondsRegexp.Match(inputBytes) {
				seconds, err := strconv.Atoi(input)
				if err != nil {
					err = errors.New(fmt.Sprintf("could not convert the input string of %s into a duration: %s", input, err))
				} else {
					duration = time.Duration(seconds) * time.Second
				}
			} else {
				err = errors.New(fmt.Sprintf("could not convert the input string of %s into a duration", input))
			}
		} else {
			matches := parseDurationRegexp.FindAllString(input, -1)
			if len(matches) != 0 {
				var d int
				var u string
				for _, match := range matches {
					_, err := fmt.Sscanf(match, "%d%s", &d, &u)
					if err != nil {
						err = errors.New(fmt.Sprintf("error parsing a part of the duration string: %s", match))
						break
					}
					switch u {
					case "y":
						duration += time.Duration(d) * Year
					case "m":
						duration += time.Duration(d) * Month
					case "w":
						duration += time.Duration(d) * Week
					case "d":
						duration += time.Duration(d) * Day
					case "h":
						duration += time.Duration(d) * Hour
					case "M":
						duration += time.Duration(d) * time.Minute
					case "s":
						duration += time.Duration(d) * time.Second
					}
				}
			}
		}
	}
	return
}
