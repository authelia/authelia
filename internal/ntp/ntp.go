package ntp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// NewProvider instantiate a ntp provider given a configuration.
func NewProvider(config *schema.NTP) *Provider {
	return &Provider{
		config: config,
		clock:  clock.New(),
		log:    logging.Logger(),
	}
}

// StartupCheck implements the startup check provider interface.
func (p *Provider) StartupCheck() (err error) {
	var offset time.Duration

	if offset, err = p.offset(); err != nil {
		p.log.WithError(err).Warnf("Could not determine the clock offset due to an error")

		return nil
	}

	if offset > p.config.MaximumDesync {
		return errors.New("the system clock is not synchronized accurately enough with the configured NTP server")
	}

	return nil
}

func (p *Provider) offset() (offset time.Duration, err error) {
	var conn net.Conn

	if conn, err = net.Dial(p.config.Address.Network(), p.config.Address.NetworkAddress()); err != nil {
		return offset, fmt.Errorf("error occurred during dial: %w", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			p.log.WithError(err).Error("Error occurred closing connection with NTP server")
		}
	}()

	if err = conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return offset, fmt.Errorf("error occurred setting connection deadline: %w", err)
	}

	v := v4
	if p.config.Version == 3 {
		v = v3
	}

	t1 := p.clock.Now()

	t1Seconds, t1Fraction := timeToSecondsAndFraction(t1)

	req := &packet{
		LeapVersionMode: leapVersionClientMode(v),
		TxTimeSeconds:   t1Seconds,
		TxTimeFraction:  t1Fraction,
	}

	if err = binary.Write(conn, binary.BigEndian, req); err != nil {
		return offset, fmt.Errorf("error occurred writing ntp packet request to the connection: %w", err)
	}

	r := &packet{}

	if err = binary.Read(conn, binary.BigEndian, r); err != nil {
		return offset, fmt.Errorf("error occurred reading ntp packet response to the connection: %w", err)
	}

	t2 := secondsAndFractionToTime(r.RxTimeSeconds, r.RxTimeFraction)
	t3 := secondsAndFractionToTime(r.TxTimeSeconds, r.TxTimeFraction)
	t4 := p.clock.Now()

	return calcOffset(t1, t2, t3, t4), nil
}
