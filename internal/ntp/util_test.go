package ntp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNTPIsOffsetTooLarge(t *testing.T) {
	testCases := []struct {
		name     string
		max      time.Duration
		first    time.Time
		second   time.Time
		expected bool
	}{
		{"ShouldReturnTrueWhenSecondIsAhead", time.Second, time.Now(), time.Now().Add(2 * time.Second), true},
		{"ShouldReturnTrueWhenFirstIsAhead", time.Second, time.Now().Add(2 * time.Second), time.Now(), true},
		{"ShouldReturnFalseWhenEqual", time.Second, time.Now(), time.Now(), false},
		{"ShouldReturnFalseWhenWithinOffset", time.Second, time.Now(), time.Now().Add(500 * time.Millisecond), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ntpIsOffsetTooLarge(tc.max, tc.first, tc.second))
		})
	}
}

func TestNTPPacketToTime(t *testing.T) {
	testCases := []struct {
		name     string
		packet   *ntpPacket
		expected time.Time
	}{
		{
			"ShouldConvertPacketWithZeroFraction",
			&ntpPacket{TxTimeSeconds: 60, TxTimeFraction: 0},
			time.Unix(int64(float64(60)-ntpEpochOffset), 0),
		},
		{
			"ShouldConvertPacketWithNonZeroFraction",
			&ntpPacket{TxTimeSeconds: 3900000000, TxTimeFraction: 2147483648},
			time.Unix(int64(float64(3900000000)-ntpEpochOffset), (int64(2147483648)*1e9)>>32),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ntpPacketToTime(tc.packet))
		})
	}
}

func TestLeapVersionClientMode(t *testing.T) {
	testCases := []struct {
		name     string
		version  ntpVersion
		expected uint8
	}{
		{"ShouldReturnV3NoLeap", ntpV3, 0xdb},
		{"ShouldReturnV4NoLeap", ntpV4, 0xe3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ntpLeapVersionClientMode(tc.version))
		})
	}
}

func TestCalcOffset(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name     string
		first    time.Time
		second   time.Time
		expected time.Duration
	}{
		{"ShouldReturnOffsetWhenFirstAfterSecond", now.Add(5 * time.Second), now, 5 * time.Second},
		{"ShouldReturnOffsetWhenSecondAfterFirst", now, now.Add(3 * time.Second), 3 * time.Second},
		{"ShouldReturnZeroWhenEqual", now, now, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, calcOffset(tc.first, tc.second))
		})
	}
}
