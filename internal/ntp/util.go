package ntp

import "time"

// ntpLeapVersionClientMode does the mathematics to configure the leap/version/mode value of an NTP client packet.
func ntpLeapVersionClientMode(leap bool, version ntpVersion) (lvm uint8) {
	lvm = ntpClientModeValue

	if leap {
		lvm += ntpLeapEnabledValue
	}

	switch version {
	case ntpV3:
		lvm += ntpVersion3Value
	case ntpV4:
		lvm += ntpVersion4Value
	}

	return lvm
}

// ntpPacketToTime converts a NTP server response into a time.Time.
func ntpPacketToTime(packet *ntpPacket) time.Time {
	seconds := float64(packet.TxTimeSeconds) - ntpEpochOffset
	nanoseconds := (int64(packet.TxTimeFraction) * 1e9) >> 32

	return time.Unix(int64(seconds), nanoseconds)
}

// ntpIsOffsetTooLarge return true if there is offset of "offset" between two times.
func ntpIsOffsetTooLarge(maxOffset time.Duration, first, second time.Time) (tooLarge bool) {
	var offset time.Duration

	if first.After(second) {
		offset = first.Sub(second)
	} else {
		offset = second.Sub(first)
	}

	return offset > maxOffset
}
