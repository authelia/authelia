package ntp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// NewProvider instantiate a ntp provider given a configuration.
func NewProvider(config *schema.NTP) *Provider {
	return &Provider{
		config: config,
		log:    logging.Logger(),
	}
}

// StartupCheck implements the startup check provider interface.
func (p *Provider) StartupCheck() (err error) {
	var offset time.Duration

	if offset, err = p.GetOffset(); err != nil {
		p.log.WithError(err).Warnf("Could not determine the clock offset due to an error")

		return nil
	}

	if offset > p.config.MaximumDesync {
		return errors.New("the system clock is not synchronized accurately enough with the configured NTP server")
	}

	return nil
}

// GetOffset returns the current offset for this provider.
func (p *Provider) GetOffset() (offset time.Duration, err error) {
	var conn net.Conn

	if conn, err = net.Dial(p.config.Address.Network(), p.config.Address.NetworkAddress()); err != nil {
		return offset, fmt.Errorf("error occurred during dial: %w", err)
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			p.log.WithError(err).Error("Error occurred closing connection with NTP sever")
		}
	}()

	if err = conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return offset, fmt.Errorf("error occurred setting connection deadline: %w", err)
	}

	version := ntpV4
	if p.config.Version == 3 {
		version = ntpV3
	}

	req := &ntpPacket{LeapVersionMode: ntpLeapVersionClientMode(version)}

	if err = binary.Write(conn, binary.BigEndian, req); err != nil {
		return offset, fmt.Errorf("error occurred writing ntp packet request to the connection: %w", err)
	}

	now := time.Now()

	resp := &ntpPacket{}

	if err = binary.Read(conn, binary.BigEndian, resp); err != nil {
		return offset, fmt.Errorf("error occurred reading ntp packet response to the connection: %w", err)
	}

	ntpTime := ntpPacketToTime(resp)

	return calcOffset(now, ntpTime), nil
}
