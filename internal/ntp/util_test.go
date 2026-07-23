package ntp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateResponse(t *testing.T) {
	req := &packet{
		LeapVersionMode: leapVersionClientMode(v4),
		TxTimeSeconds:   3900000000,
		TxTimeFraction:  2147483648,
	}

	response := func(mutate func(resp *packet)) *packet {
		resp := &packet{
			LeapVersionMode:    (leapUnknown << 6) | (version4 << 3) | modeServer,
			Stratum:            1,
			OriginTimeSeconds:  req.TxTimeSeconds,
			OriginTimeFraction: req.TxTimeFraction,
		}

		if mutate != nil {
			mutate(resp)
		}

		return resp
	}

	testCases := []struct {
		name   string
		mutate func(resp *packet)
		err    string
	}{
		{
			"ShouldAcceptValidResponse",
			nil,
			"",
		},
		{
			"ShouldAcceptStratumFifteen",
			func(resp *packet) { resp.Stratum = 15 },
			"",
		},
		{
			"ShouldRejectStratumZero",
			func(resp *packet) { resp.Stratum = 0 },
			"the response has stratum '0' but only values between 1 and 15 are considered valid",
		},
		{
			"ShouldRejectStratumUnsynchronized",
			func(resp *packet) { resp.Stratum = 16 },
			"the response has stratum '16' but only values between 1 and 15 are considered valid",
		},
		{
			"ShouldRejectClientMode",
			func(resp *packet) { resp.LeapVersionMode = (leapUnknown << 6) | (version4 << 3) | modeClient },
			"the response has mode '3' but only the server mode '4' is considered valid",
		},
		{
			"ShouldRejectBroadcastMode",
			func(resp *packet) { resp.LeapVersionMode = (leapUnknown << 6) | (version4 << 3) | 5 },
			"the response has mode '5' but only the server mode '4' is considered valid",
		},
		{
			"ShouldRejectMismatchedOriginSeconds",
			func(resp *packet) { resp.OriginTimeSeconds = req.TxTimeSeconds + 1 },
			"the response origin timestamp does not match the transmit timestamp of the request",
		},
		{
			"ShouldRejectMismatchedOriginFraction",
			func(resp *packet) { resp.OriginTimeFraction = req.TxTimeFraction + 1 },
			"the response origin timestamp does not match the transmit timestamp of the request",
		},
		{
			"ShouldRejectZeroedOriginTimestamp",
			func(resp *packet) { resp.OriginTimeSeconds, resp.OriginTimeFraction = 0, 0 },
			"the response origin timestamp does not match the transmit timestamp of the request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateResponse(req, response(tc.mutate))

			if len(tc.err) > 0 {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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
