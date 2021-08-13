package ntp

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewProvider instantiate a ntp provider given a configuration.
func NewProvider(config *schema.NTPConfiguration) *Provider {
	return &Provider{config}
}

// StartupCheck checks if the system clock is not out of sync.
func (p *Provider) StartupCheck() (failed bool, err error) {
	conn, err := net.Dial("udp", p.config.Address)
	if err != nil {
		return false, fmt.Errorf("could not connect to ntp server to validate the time desync: %w", err)
	}

	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return false, fmt.Errorf("could not connect to ntp server to validate the time desync: %w", err)
	}

	version := ntpV4
	if p.config.Version == 3 {
		version = ntpV3
	}

	req := &ntpPacket{LeapVersionMode: ntpLeapVersionClientMode(false, version)}

	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		return false, fmt.Errorf("could not write to the ntp server socket to validate the time desync: %w", err)
	}

	now := time.Now()

	resp := &ntpPacket{}

	if err := binary.Read(conn, binary.BigEndian, resp); err != nil {
		return false, fmt.Errorf("could not read from the ntp server socket to validate the time desync: %w", err)
	}

	maxOffset, err := utils.ParseDurationString(p.config.MaximumDesync)
	if err != nil {
		return false, fmt.Errorf("Error ocuured: %w", err)
	}

	ntpTime := ntpPacketToTime(resp)

	return ntpIsOffsetTooLarge(maxOffset, now, ntpTime), nil
}
