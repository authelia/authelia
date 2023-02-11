package ntp

const ntpEpochOffset = 2208988800

const (
	ntpV3 ntpVersion = iota
	ntpV4
)

const (
	maskMode    = 0xf8
	maskVersion = 0xc7
	maskLeap    = 0x3f
)

const (
	modeClient = 3
)

const (
	version3 = 3
	version4 = 4
)

const (
	leapUnknown = 3
)
