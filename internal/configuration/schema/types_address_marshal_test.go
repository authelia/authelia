// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressMarshal(t *testing.T) {
	testCases := []struct {
		name     string
		have     testAddressMarshaler
		expected string
	}{
		{"ShouldMarshalAddressTCP", AddressTCP{Address: NewAddressFromNetworkValues(AddressSchemeTCP, "127.0.0.1", 6379)}, "tcp://127.0.0.1:6379"},
		{"ShouldMarshalAddressTCPInvalid", AddressTCP{}, ""},
		{"ShouldMarshalAddressUDP", AddressUDP{Address: NewAddressFromNetworkValues(AddressSchemeUDP, "127.0.0.1", 53)}, "udp://127.0.0.1:53"},
		{"ShouldMarshalAddressUDPInvalid", AddressUDP{}, ""},
		{"ShouldMarshalAddressLDAP", AddressLDAP{Address: NewAddressFromNetworkValues(AddressSchemeLDAP, "127.0.0.1", 389)}, "ldap://127.0.0.1:389"},
		{"ShouldMarshalAddressLDAPInvalid", AddressLDAP{}, ""},
		{"ShouldMarshalAddressSMTP", AddressSMTP{Address: NewAddressFromNetworkValues(AddressSchemeSMTP, "127.0.0.1", 25)}, "smtp://127.0.0.1:25"},
		{"ShouldMarshalAddressSMTPInvalid", AddressSMTP{}, ""},
		{"ShouldMarshalAddress", NewAddressFromNetworkValues(AddressSchemeTCP, "127.0.0.1", 9091), "tcp://127.0.0.1:9091"},
		{"ShouldMarshalAddressInvalid", Address{}, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			yaml, err := tc.have.MarshalYAML()

			assert.NoError(t, err)

			text, err := tc.have.MarshalText()

			assert.NoError(t, err)

			if tc.expected == "" {
				assert.Nil(t, yaml)
				assert.Nil(t, text)
			} else {
				assert.Equal(t, tc.expected, yaml)
				assert.Equal(t, []byte(tc.expected), text)
			}
		})
	}
}

type testAddressMarshaler interface {
	MarshalYAML() (any, error)
	MarshalText() ([]byte, error)
}
