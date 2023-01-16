package ntp

import (
	"encoding/binary"
	"errors"
	"net"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// NewProvider instantiate a ntp provider given a configuration.
func NewProvider(config *schema.NTPConfiguration) *Provider {
	return &Provider{
		config: config,
		log:    logging.Logger(),
	}
}

// StartupCheck implements the startup check provider interface.
func (p *Provider) StartupCheck() (err error) {
	conn, err := net.Dial("udp", p.config.Address)
	if err != nil {
		p.log.Warnf("Could not connect to NTP server to validate the system time is properly synchronized: %+v", err)

		return nil
	}

	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		p.log.Warnf("Could not connect to NTP server to validate the system time is properly synchronized: %+v", err)

		return nil
	}

	version := ntpV4
	if p.config.Version == 3 {
		version = ntpV3
	}

	req := &ntpPacket{LeapVersionMode: ntpLeapVersionClientMode(version)}

	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		p.log.Warnf("Could not write to the NTP server socket to validate the system time is properly synchronized: %+v", err)

		return nil
	}

	now := time.Now()

	resp := &ntpPacket{}

	if err := binary.Read(conn, binary.BigEndian, resp); err != nil {
		p.log.Warnf("Could not read from the NTP server socket to validate the system time is properly synchronized: %+v", err)

		return nil
	}

	ntpTime := ntpPacketToTime(resp)

	if result := ntpIsOffsetTooLarge(p.config.MaximumDesync, now, ntpTime); result {
		return errors.New("the system clock is not synchronized accurately enough with the configured NTP server")
	}

	return nil
}
