package ntp

const (
	ntpClientModeValue  uint8 = 3  // 00000011.
	ntpLeapEnabledValue uint8 = 64 // 01000000.
	ntpVersion3Value    uint8 = 24 // 00011000.
	ntpVersion4Value    uint8 = 40 // 00101000.
)

const ntpEpochOffset = 2208988800

const (
	ntpV3 ntpVersion = iota
	ntpV4
)
