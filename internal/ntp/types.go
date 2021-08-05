package ntp

import (
	"github.com/authelia/authelia/internal/configuration/schema"
	"time"
)

type Configuration struct {
	Address             string
	Version             int
	MaximumDesync       time.Duration
	DisableStartupCheck bool
}

type Provider struct {
	config *schema.NtpConfiguration
}

type ntpVersion int

type ntpPacket struct {
	LeapVersionMode       uint8
	Stratum               uint8
	Poll                  int8
	Precision             int8
	RootDelay             uint32
	RootDispersion        uint32
	ReferenceID           uint32
	ReferenceTimeSeconds  uint32
	ReferenceTimeFraction uint32
	OriginTimeSeconds     uint32
	OriginTimeFraction    uint32
	RxTimeSeconds         uint32
	RxTimeFraction        uint32
	TxTimeSeconds         uint32
	TxTimeFraction        uint32
}
