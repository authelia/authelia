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
			assert.Equal(t, tc.expected, isOffsetTooLarge(tc.max, tc.first, tc.second))
		})
	}
}

func TestNTPPacketToTime(t *testing.T) {
	testCases := []struct {
		name     string
		packet   *packet
		expected time.Time
	}{
		{
			"ShouldConvertPacketWithZeroFraction",
			&packet{TxTimeSeconds: 60, TxTimeFraction: 0},
			time.Unix(int64(60)-epochOffset, 0),
		},
		{
			"ShouldConvertPacketWithNonZeroFraction",
			&packet{TxTimeSeconds: 3900000000, TxTimeFraction: 2147483648},
			time.Unix(int64(3900000000)-epochOffset, (int64(2147483648)*1e9)>>32),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, secondsAndFractionToTime(tc.packet.TxTimeSeconds, tc.packet.TxTimeFraction))
		})
	}
}

func TestLeapVersionClientMode(t *testing.T) {
	testCases := []struct {
		name     string
		version  version
		expected uint8
	}{
		{"ShouldReturnV3NoLeap", v3, 0xdb},
		{"ShouldReturnV4NoLeap", v4, 0xe3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, leapVersionClientMode(tc.version))
		})
	}
}
