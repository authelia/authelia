//nolint:gosec // G115: integer overflow conversion is safe in tests
package ntp

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldCheckNTPV4(t *testing.T) {
	addr := startMockNTPServer(t, time.Now())

	ntp := NewProvider(&schema.NTP{
		Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
		Version:       4,
		MaximumDesync: time.Second * 3,
	})

	assert.NoError(t, ntp.StartupCheck())
}

func TestShouldCheckNTPV3(t *testing.T) {
	addr := startMockNTPServer(t, time.Now())

	ntp := NewProvider(&schema.NTP{
		Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
		Version:       3,
		MaximumDesync: time.Second * 3,
	})

	assert.NoError(t, ntp.StartupCheck())
}

func TestStartupCheck(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(t *testing.T) *Provider
		expectErr   bool
		errContains string
	}{
		{
			"ShouldSucceedWithMockNTPServer",
			func(t *testing.T) *Provider {
				addr := startMockNTPServer(t, time.Now())

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       4,
					MaximumDesync: time.Minute,
				})
			},
			false,
			"",
		},
		{
			"ShouldErrWhenOffsetTooLarge",
			func(t *testing.T) *Provider {
				addr := startMockNTPServer(t, time.Now().Add(time.Hour))

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       4,
					MaximumDesync: time.Second,
				})
			},
			true,
			"the system clock is not synchronized accurately enough",
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
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := tc.setup(t)

			err := provider.StartupCheck()

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetOffset(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func(t *testing.T) *Provider
		expectErr bool
	}{
		{
			"ShouldReturnOffsetFromMockServer",
			func(t *testing.T) *Provider {
				addr := startMockNTPServer(t, time.Now())

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       4,
					MaximumDesync: time.Minute,
				})
			},
			false,
		},
		{
			"ShouldReturnOffsetWithV3",
			func(t *testing.T) *Provider {
				addr := startMockNTPServer(t, time.Now())

				return NewProvider(&schema.NTP{
					Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, addr.IP.String(), uint16(addr.Port))},
					Version:       3,
					MaximumDesync: time.Minute,
				})
			},
			false,
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
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := tc.setup(t)

			offset, err := provider.GetOffset()

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Less(t, offset, time.Minute)
			}
		})
	}
}

func startMockNTPServer(t *testing.T, respondTime time.Time) *net.UDPAddr {
	t.Helper()

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	addr := conn.LocalAddr().(*net.UDPAddr)

	go func() {
		buf := make([]byte, 48)

		for {
			n, clientAddr, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}

			if n < 48 {
				continue
			}

			seconds := uint32(respondTime.Unix() + ntpEpochOffset)
			fraction := uint32(0)

			resp := make([]byte, 48)
			resp[0] = buf[0]

			binary.BigEndian.PutUint32(resp[40:44], seconds)
			binary.BigEndian.PutUint32(resp[44:48], fraction)

			_, _ = conn.WriteTo(resp, clientAddr)
		}
	}()

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return addr
}
