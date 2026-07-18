//nolint:gosec // G115: integer overflow conversion is safe in tests
package ntp

import (
	"bytes"
	"encoding/binary"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldCheckNTPV4(t *testing.T) {
	addr := testServer(t, clock.New())

	ntp := NewProvider(&schema.NTP{
		Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
		Version:       4,
		MaximumDesync: time.Second * 3,
	})

	assert.NoError(t, ntp.StartupCheck())
}

func TestShouldCheckNTPV3(t *testing.T) {
	addr := testServer(t, clock.New())

	ntp := NewProvider(&schema.NTP{
		Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
		Version:       3,
		MaximumDesync: time.Second * 3,
	})

	assert.NoError(t, ntp.StartupCheck())
}

func TestStartupCheck(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T) *Provider
		err   string
	}{
		{
			"ShouldSucceedWithMockNTPServer",
			func(t *testing.T) *Provider {
				addr := testServer(t, clock.New())

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       4,
					MaximumDesync: time.Minute,
				})
			},
			"",
		},
		{
			"ShouldErrWhenOffsetTooLarge",
			func(t *testing.T) *Provider {
				addr := testServer(t, clock.NewFixed(time.Now().Add(time.Minute*10)))

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       4,
					MaximumDesync: time.Second,
				})
			},
			"the system clock is not synchronized accurately enough with the configured NTP server",
		},
		{
			"ShouldNotErrWhenConnectionFails",
			func(t *testing.T) *Provider {
				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, "127.0.0.1", 1)},
					Version:       4,
					MaximumDesync: time.Minute,
				})
			},
			"",
		},
		{
			"ShouldNotErrWhenSpoofedResponseClaimsLargeOffset",
			func(t *testing.T) *Provider {
				addr := testServer(t, clock.NewFixed(time.Now().Add(time.Minute*10)), func(resp *packet) {
					resp.Stratum = 0
				})

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       4,
					MaximumDesync: time.Second,
				})
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := tc.setup(t)

			err := provider.StartupCheck()

			if len(tc.err) > 0 {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetOffset(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T) *Provider
		err   any
	}{
		{
			"ShouldReturnOffsetFromMockServer",
			func(t *testing.T) *Provider {
				addr := testServer(t, clock.New())

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       4,
					MaximumDesync: time.Minute,
				})
			},
			nil,
		},
		{
			"ShouldReturnOffsetWithV3",
			func(t *testing.T) *Provider {
				addr := testServer(t, clock.New())

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       3,
					MaximumDesync: time.Minute,
				})
			},
			nil,
		},
		{
			"ShouldErrWhenServerUnreachable",
			func(t *testing.T) *Provider {
				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, "192.0.2.1", 1)},
					Version:       4,
					MaximumDesync: time.Minute,
				})
			},
			regexp.MustCompile(`^error occurred reading ntp packet response to the connection: read udp \d+.\d+.\d+.\d+:\d+->\d+.\d+.\d+.\d+:\d+: i/o timeout$`),
		},
		{
			"ShouldErrWhenResponseStratumIsZero",
			func(t *testing.T) *Provider {
				addr := testServer(t, clock.New(), func(resp *packet) { resp.Stratum = 0 })

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       4,
					MaximumDesync: time.Minute,
				})
			},
			regexp.MustCompile(`^error occurred validating the ntp packet response: the response has stratum '0' but only values between 1 and 15 are considered valid$`),
		},
		{
			"ShouldErrWhenResponseModeIsNotServer",
			func(t *testing.T) *Provider {
				addr := testServer(t, clock.New(), func(resp *packet) {
					resp.LeapVersionMode = (resp.LeapVersionMode & maskMode) | modeClient
				})

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       4,
					MaximumDesync: time.Minute,
				})
			},
			regexp.MustCompile(`^error occurred validating the ntp packet response: the response has mode '3' but only the server mode '4' is considered valid$`),
		},
		{
			"ShouldErrWhenResponseOriginTimestampDoesNotMatch",
			func(t *testing.T) *Provider {
				addr := testServer(t, clock.New(), func(resp *packet) { resp.OriginTimeSeconds++ })

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       4,
					MaximumDesync: time.Minute,
				})
			},
			regexp.MustCompile(`^error occurred validating the ntp packet response: the response origin timestamp does not match the transmit timestamp of the request$`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := tc.setup(t)

			offset, err := provider.offset()

			if tc.err != nil {
				require.Error(t, err)
				assert.Regexp(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Less(t, offset, time.Minute)
			}
		})
	}
}

func testServer(t *testing.T, clock clock.Provider, mutators ...func(resp *packet)) *net.UDPAddr {
	t.Helper()

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	addr := conn.LocalAddr().(*net.UDPAddr)

	go func() {
		for {
			buf := make([]byte, 48)

			n, clientAddr, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}

			if n < 48 {
				continue
			}

			req := &packet{}

			if err = binary.Read(bytes.NewReader(buf), binary.BigEndian, req); err != nil {
				continue
			}

			now := clock.Now()

			seconds, fraction := timeToSecondsAndFraction(now)

			resp := &packet{
				LeapVersionMode:    (req.LeapVersionMode & maskMode) | modeServer,
				Stratum:            1,
				Poll:               req.Poll,
				Precision:          -20,
				OriginTimeSeconds:  req.TxTimeSeconds,
				OriginTimeFraction: req.TxTimeFraction,
				RxTimeSeconds:      seconds,
				RxTimeFraction:     fraction,
				TxTimeSeconds:      seconds,
				TxTimeFraction:     fraction,
			}

			for _, mutator := range mutators {
				mutator(resp)
			}

			var out bytes.Buffer

			if err = binary.Write(&out, binary.BigEndian, resp); err != nil {
				continue
			}

			_, _ = conn.WriteTo(out.Bytes(), clientAddr)
		}
	}()

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return addr
}
