package regulation

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReturnBanResult(t *testing.T) {
	type testCase struct {
		name          string
		ban           BanType
		value         string
		inputTime     sql.NullTime
		expectExpires *time.Time
	}

	loc := time.FixedZone("TEST", 2*60*60)
	ts := time.Date(2025, time.August, 12, 15, 4, 5, 0, loc)

	testCases := []testCase{
		{
			name:          "ShouldReturnBanResultWithNilExpiresWhenTimeIsInvalid",
			ban:           BanType(0),
			value:         "test-value",
			inputTime:     sql.NullTime{Valid: false},
			expectExpires: nil,
		},
		{
			name:          "ShouldReturnBanResultWithExpiresWhenTimeIsValid",
			ban:           BanType(0),
			value:         "another-value",
			inputTime:     sql.NullTime{Time: ts, Valid: true},
			expectExpires: &ts,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ban, value, expires, err := returnBanResult(tc.ban, tc.value, tc.inputTime)

			require.ErrorIs(t, err, ErrUserIsBanned)
			assert.Equal(t, tc.ban, ban)
			assert.Equal(t, tc.value, value)

			if tc.expectExpires == nil {
				assert.Nil(t, expires)
			} else {
				require.NotNil(t, expires)
				assert.Equal(t, *tc.expectExpires, *expires)
			}
		})
	}
}

func TestFormatExpiresLong(t *testing.T) {
	loc := time.FixedZone("TEST", 2*60*60)
	ts := time.Date(2025, time.August, 12, 15, 4, 5, 0, loc)

	testCases := []struct {
		name     string
		input    *time.Time
		expected string
	}{
		{
			name:     "ShouldReturnNeverExpiresWhenNil",
			input:    nil,
			expected: "never expires",
		},
		{
			name:     "ShouldFormatTimestampWithOffset",
			input:    &ts,
			expected: "3:04:05PM on August 12 2025 (+02:00)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := FormatExpiresLong(tc.input)
			assert.Equal(t, tc.expected, out)
		})
	}
}

func TestFormatExpiresShort(t *testing.T) {
	ts := time.Date(2025, time.August, 12, 15, 4, 5, 0, time.UTC)

	testCases := []struct {
		name     string
		input    sql.NullTime
		expected string
	}{
		{
			name:     "ShouldReturnNeverWhenNullTimeInvalid",
			input:    sql.NullTime{Valid: false},
			expected: "never",
		},
		{
			name:     "ShouldFormatShortTimeWhenValid",
			input:    sql.NullTime{Time: ts, Valid: true},
			expected: "2025-08-12 15:04:05",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := FormatExpiresShort(tc.input)
			assert.Equal(t, tc.expected, out)
		})
	}
}
