package ntp

import (
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Provider type is the NTP provider.
type Provider struct {
	config *schema.NTP
	log    *logrus.Logger
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
