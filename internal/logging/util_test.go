package logging

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatFilePath(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		now      time.Time
		expected string
	}{
		{
			"ShouldReturnInput",
			"abc 123",
			time.Unix(0, 0).UTC(),
			"abc 123",
		},
		{
			"ShouldReturnStandardWithDateTime",
			"abc %d 123",
			time.Unix(0, 0).UTC(),
			"abc 1970-01-01T00:00:00Z 123",
		},
		{
			"ShouldReturnStandardWithDateTimeFormatter",
			"abc {datetime} 123",
			time.Unix(0, 0).UTC(),
			"abc 1970-01-01T00:00:00Z 123",
		},
		{
			"ShouldReturnStandardWithDateTimeCustomLayout",
			"abc {datetime:Mon Jan 2 15:04:05 MST 2006} 123",
			time.Unix(0, 0).UTC(),
			"abc Thu Jan 1 00:00:00 UTC 1970 123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FormatFilePath(tc.have, tc.now))
		})
	}
}
