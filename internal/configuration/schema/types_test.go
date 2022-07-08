package schema

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAddressFromString(t *testing.T) {
	testCases := []struct {
		name                                          string
		have                                          string
		expected                                      *Address
		expectedHostPort, expectedString, expectedErr string
	}{
		{"ShouldParseBasicAddress", "tcp://0.0.0.0:9091", &Address{true, "tcp", net.ParseIP("0.0.0.0"), 9091}, "0.0.0.0:9091", "tcp://0.0.0.0:9091", ""},
		{"ShouldParseEmptyAddress", "", &Address{true, "tcp", net.ParseIP("0.0.0.0"), 0}, "0.0.0.0:0", "tcp://0.0.0.0:0", ""},
		{"ShouldParseAddressMissingScheme", "0.0.0.0:9091", &Address{true, "tcp", net.ParseIP("0.0.0.0"), 9091}, "0.0.0.0:9091", "tcp://0.0.0.0:9091", ""},
		{"ShouldParseAddressMissingPort", "tcp://0.0.0.0", &Address{true, "tcp", net.ParseIP("0.0.0.0"), 0}, "0.0.0.0:0", "tcp://0.0.0.0:0", ""},
		{"ShouldNotParseUnknownScheme", "a://0.0.0.0", nil, "", "", "could not parse scheme for address 'a://0.0.0.0': scheme 'a' is not valid, expected to be one of 'tcp://', 'udp://'"},
		{"ShouldNotParseInvalidPort", "tcp://0.0.0.0:abc", nil, "", "", "could not parse string 'tcp://0.0.0.0:abc' as address: expected format is [<scheme>://]<ip>[:<port>]: parse \"tcp://0.0.0.0:abc\": invalid port \":abc\" after host"},
		{"ShouldNotParseInvalidIP", "tcp://example.com:9091", nil, "", "", "could not parse ip for address 'tcp://example.com:9091': example.com does not appear to be an IP"},
		{"ShouldNotParseInvalidAddress", "@$@#%@#$@@", nil, "", "", "could not parse string '@$@#%@#$@@' as address: expected format is [<scheme>://]<ip>[:<port>]: parse \"tcp://@$@#%@#$@@\": invalid URL escape \"%@#\""},
		{"ShouldNotParseInvalidAddressWithScheme", "tcp://@$@#%@#$@@", nil, "", "", "could not parse string 'tcp://@$@#%@#$@@' as address: expected format is [<scheme>://]<ip>[:<port>]: parse \"tcp://@$@#%@#$@@\": invalid URL escape \"%@#\""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, actualErr := NewAddressFromString(tc.have)

			if len(tc.expectedErr) != 0 {
				assert.EqualError(t, actualErr, tc.expectedErr)
			} else {
				assert.Nil(t, actualErr)

				assert.Equal(t, actual.HostPort(), tc.expectedHostPort)
				assert.Equal(t, actual.String(), tc.expectedString)
			}

			assert.Equal(t, tc.expected, actual)

			if actual != nil {
				assert.True(t, actual.Valid())
				assert.NotEmpty(t, actual.String())
				assert.NotEmpty(t, actual.HostPort())
			}
		})
	}
}
