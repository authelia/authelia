package ntp

import "time"

// leapVersionClientMode does the mathematics to configure the leap/version/mode value of an NTP client packet.
func leapVersionClientMode(version version) (lvm uint8) {
	lvm = (lvm & maskMode) | uint8(modeClient)

	switch version {
	case v3:
		lvm = (lvm & maskVersion) | uint8(version3)<<3
	case v4:
		lvm = (lvm & maskVersion) | uint8(version4)<<3
	}

	lvm = (lvm & maskLeap) | uint8(leapUnknown)<<6

	return lvm
}

func secondsAndFractionToTime(seconds, fraction uint32) time.Time {
	nanoseconds := (int64(fraction) * 1e9) >> 32

	return time.Unix(int64(seconds)-epochOffset, nanoseconds)
}

func timeToSecondsAndFraction(t time.Time) (seconds, fraction uint32) {
	//nolint:gosec // G115: Overflow is intentional; uint32 truncation implements NTP era wrapping per RFC 5905 Section 6.
	return uint32(t.Unix() + epochOffset), uint32((int64(t.Nanosecond()) << 32) / 1e9)
}

// isOffsetTooLarge return true if there is offset of "offset" between two times.
func isOffsetTooLarge(maxOffset time.Duration, first, second time.Time) (tooLarge bool) {
	var offset time.Duration

	if first.After(second) {
		offset = first.Sub(second)
	} else {
		offset = second.Sub(first)
	}

	return offset > maxOffset
}

// calcOffset calculates the clock offset using the SNTP four-timestamp formula:
// offset = ((T2 - T1) + (T3 - T4)) / 2
// where T1 is the client send time, T2 is the server receive time,
// T3 is the server transmit time, and T4 is the client receive time.
func calcOffset(t1, t2, t3, t4 time.Time) time.Duration {
	offset := (t2.Sub(t1) + t3.Sub(t4)) / 2

	if offset < 0 {
		return -offset
	}

	return offset
}
