package logging

import (
	"strings"
	"time"
)

// FormatFilePath formats a file path with the given time.
func FormatFilePath(in string, now time.Time) (out string) {
	matches := reFormatFilePath.FindStringSubmatch(in)

	if len(matches) == 0 {
		return in
	}

	layout := time.RFC3339

	if len(matches[3]) != 0 {
		layout = matches[3]
	}

	return strings.Replace(in, matches[0], now.Format(layout), 1)
}
