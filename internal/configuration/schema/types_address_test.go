package schema

import (
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAddressFromString(t *testing.T) {
	testCases := []struct {
		name                                         string
		have                                         string
		expected                                     *Address
		expectedAddress, expectedString, expectedErr string
	}{
		{
			"ShouldParseBasicAddress",
			"tcp://0.0.0.0:9091",
			&Address{true, false, -1, 9091, &url.URL{Scheme: AddressSchemeTCP, Host: "0.0.0.0:9091"}},
			"0.0.0.0:9091",
			"tcp://0.0.0.0:9091",
			"",
		},
		{
			"ShouldParseEmptyAddress",
			"",
			&Address{true, false, -1, 0, &url.URL{Scheme: AddressSchemeTCP, Host: ":0"}},
			":0",
			"tcp://:0",
			"",
		},
		{
			"ShouldParseAddressMissingScheme",
			"0.0.0.0:9091",
			&Address{true, false, -1, 9091, &url.URL{Scheme: AddressSchemeTCP, Host: "0.0.0.0:9091"}},
			"0.0.0.0:9091",
			"tcp://0.0.0.0:9091",
			"",
		},
		{
			"ShouldParseUnixAddressMissingScheme",
			"/var/run/example.sock",
			&Address{true, true, -1, 0, &url.URL{Scheme: AddressSchemeUnix, Path: "/var/run/example.sock"}},
			"/var/run/example.sock",
			"unix:///var/run/example.sock",
			"",
		},
		{
			"ShouldParseAddressMissingPort",
			"tcp://0.0.0.0",
			&Address{true, false, -1, 0, &url.URL{Scheme: AddressSchemeTCP, Host: "0.0.0.0:0"}},
			"0.0.0.0:0",
			"tcp://0.0.0.0:0",
			"",
		},
		{
			"ShouldNotParseAddressWithQuery",
			"tcp://0.0.0.0?umask=0022",
			nil,
			"0.0.0.0:0",
			"tcp://0.0.0.0:0",
			"error validating the address: the url 'tcp://0.0.0.0?umask=0022' appears to have a query but this is not valid for addresses with the 'tcp' scheme",
		},
		{
			"ShouldParseUnixSocket",
			"unix:///path/to/a/socket.sock",
			&Address{true, true, -1, 0, &url.URL{Scheme: AddressSchemeUnix, Path: "/path/to/a/socket.sock"}},
			"/path/to/a/socket.sock",
			"unix:///path/to/a/socket.sock",
			"",
		},
		{
			"ShouldNotParseUnixSocketWithHost",
			"unix://ahost/path/to/a/socket.sock",
			nil,
			"",
			"",
			"error validating the unix socket address: the url 'unix://ahost/path/to/a/socket.sock' appears to have a host but this is not valid for unix sockets: this may occur if you omit the leading forward slash from the socket path",
		},
		{
			"ShouldNotParseUnixSocketWithoutPath",
			"unix://nopath.com",
			nil,
			"",
			"",
			"error validating the unix socket address: could not determine path from 'unix://nopath.com'",
		},
		{
			"ShouldParseUnixSocketWithQuery",
			"unix:///path/to/a/socket.sock?umask=0022",
			&Address{true, true, 18, 0, &url.URL{Scheme: AddressSchemeUnix, Path: "/path/to/a/socket.sock", RawQuery: "umask=0022"}},
			"/path/to/a/socket.sock",
			"unix:///path/to/a/socket.sock?umask=0022",
			"",
		},
		{
			"ShouldNotParseUnixSocketWithFragment",
			"unix:///path/to/a/socket.sock#example",
			nil,
			"",
			"",
			"error validating the address: the url 'unix:///path/to/a/socket.sock#example' appears to have a fragment but this is not valid for addresses",
		},
		{
			"ShouldNotParseUnixSocketWithUserInfo",
			"unix://user:example@/path/to/a/socket.sock",
			nil,
			"",
			"",
			"error validating the address: the url 'unix://user:example@/path/to/a/socket.sock' appears to have user info but this is not valid for addresses",
		},
		{
			"ShouldParseUnknownScheme",
			"a://0.0.0.0",
			&Address{true, false, -1, 0, &url.URL{Scheme: "a", Host: "0.0.0.0"}},
			"0.0.0.0",
			"a://0.0.0.0",
			"",
		},
		{
			"ShouldNotParseInvalidPort",
			"tcp://0.0.0.0:abc",
			nil,
			"",
			"",
			"could not parse string 'tcp://0.0.0.0:abc' as address: expected format is [<scheme>://]<hostname>[:<port>]: parse \"tcp://0.0.0.0:abc\": invalid port \":abc\" after host",
		},
		{
			"ShouldNotParseInvalidAddress",
			"@$@#%@#$@@",
			nil,
			"",
			"",
			"could not parse string '@$@#%@#$@@' as address: expected format is [<scheme>://]<hostname>[:<port>]: parse \"tcp://@$@#%@#$@@\": invalid URL escape \"%@#\"",
		},
		{
			"ShouldNotParseInvalidAddressWithScheme",
			"tcp://@$@#%@#$@@",
			nil,
			"",
			"",
			"could not parse string 'tcp://@$@#%@#$@@' as address: expected format is [<scheme>://]<hostname>[:<port>]: parse \"tcp://@$@#%@#$@@\": invalid URL escape \"%@#\"",
		},
		{
			"ShouldSetDefaultPortLDAP",
			"ldap://127.0.0.1",
			&Address{true, false, -1, 389, &url.URL{Scheme: AddressSchemeLDAP, Host: "127.0.0.1:389"}},
			"127.0.0.1:389",
			"ldap://127.0.0.1:389",
			"",
		},
		{
			"ShouldSetDefaultPortLDAPS",
			"ldaps://127.0.0.1",
			&Address{true, false, -1, 636, &url.URL{Scheme: AddressSchemeLDAPS, Host: "127.0.0.1:636"}},
			"127.0.0.1:636",
			"ldaps://127.0.0.1:636",
			"",
		},
		{
			"ShouldAllowLDAPI",
			"ldapi:///abc",
			&Address{true, true, -1, 0, &url.URL{Scheme: AddressSchemeLDAPI, Path: "/abc"}},
			"/abc",
			"ldapi:///abc",
			"",
		},
		{
			"ShouldAllowImplicitLDAPI",
			"ldapi://",
			&Address{true, true, -1, 0, &url.URL{Scheme: AddressSchemeLDAPI, Path: ""}},
			"",
			"ldapi:",
			"",
		},
		{
			"ShouldAllowImplicitLDAPINoSlash",
			"ldapi:",
			&Address{true, true, -1, 0, &url.URL{Scheme: AddressSchemeLDAPI, Path: ""}},
			"",
			"ldapi:",
			"",
		},
		{
			"ShouldSetDefaultPortSMTP",
			"smtp://127.0.0.1",
			&Address{true, false, -1, 25, &url.URL{Scheme: AddressSchemeSMTP, Host: "127.0.0.1:25"}},
			"127.0.0.1:25",
			"smtp://127.0.0.1:25",
			"",
		},
		{
			"ShouldSetDefaultPortSUBMISSION",
			"submission://127.0.0.1",
			&Address{true, false, -1, 587, &url.URL{Scheme: AddressSchemeSUBMISSION, Host: "127.0.0.1:587"}},
			"127.0.0.1:587",
			"submission://127.0.0.1:587",
			"",
		},
		{
			"ShouldSetDefaultPortSUBMISSIONS",
			"submissions://127.0.0.1",
			&Address{true, false, -1, 465, &url.URL{Scheme: AddressSchemeSUBMISSIONS, Host: "127.0.0.1:465"}},
			"127.0.0.1:465",
			"submissions://127.0.0.1:465",
			"",
		},
		{
			"ShouldNotOverridePort",
			"ldap://127.0.0.1:123",
			&Address{true, false, -1, 123, &url.URL{Scheme: AddressSchemeLDAP, Host: "127.0.0.1:123"}},
			"127.0.0.1:123",
			"ldap://127.0.0.1:123",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, actualErr := NewAddress(tc.have)

			if len(tc.expectedErr) != 0 {
				assert.EqualError(t, actualErr, tc.expectedErr)
			} else {
				assert.Nil(t, actualErr)

				assert.Equal(t, tc.expectedAddress, actual.NetworkAddress())
				assert.Equal(t, tc.expectedString, actual.String())

				assert.True(t, actual.Valid())
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestAddress_ValidateErrors(t *testing.T) {
	testCases := []struct {
		name                                                                    string
		have                                                                    *Address
		expectedLDAP, expectedSMTP, expectedHTTP, expectedSQL, expectedListener string
	}{
		{
			"ShouldValidateLDAPAddress",
			&Address{true, false, -1, 0, &url.URL{Scheme: AddressSchemeLDAP, Host: "127.0.0.1"}},
			"",
			"scheme must be one of 'smtp', 'submission', or 'submissions' but is configured as 'ldap'",
			"scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'ldap'",
			"scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'ldap'",
			"scheme must be one of 'tcp', 'tcp4', 'tcp6', 'udp', 'udp4', 'udp6', or 'unix' but is configured as 'ldap'",
		},
		{
			"ShouldValidateSMTPAddress",
			&Address{true, false, -1, 0, &url.URL{Scheme: AddressSchemeSMTP, Host: "127.0.0.1"}},
			"scheme must be one of 'ldap', 'ldaps', or 'ldapi' but is configured as 'smtp'",
			"",
			"scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'smtp'",
			"scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'smtp'",
			"scheme must be one of 'tcp', 'tcp4', 'tcp6', 'udp', 'udp4', 'udp6', or 'unix' but is configured as 'smtp'",
		},
		{
			"ShouldValidateTCPAddress",
			&Address{true, false, -1, 0, &url.URL{Scheme: AddressSchemeTCP, Host: "127.0.0.1"}},
			"scheme must be one of 'ldap', 'ldaps', or 'ldapi' but is configured as 'tcp'",
			"scheme must be one of 'smtp', 'submission', or 'submissions' but is configured as 'tcp'",
			"",
			"",
			"",
		},
		{
			"ShouldValidateUnixSocket",
			&Address{true, true, -1, 0, &url.URL{Scheme: AddressSchemeUnix, Path: "/path/to/socket"}},
			"scheme must be one of 'ldap', 'ldaps', or 'ldapi' but is configured as 'unix'",
			"scheme must be one of 'smtp', 'submission', or 'submissions' but is configured as 'unix'",
			"",
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedLDAP == "" {
				assert.NoError(t, tc.have.ValidateLDAP())
			} else {
				assert.EqualError(t, tc.have.ValidateLDAP(), tc.expectedLDAP)
			}

			if tc.expectedSMTP == "" {
				assert.NoError(t, tc.have.ValidateSMTP())
			} else {
				assert.EqualError(t, tc.have.ValidateSMTP(), tc.expectedSMTP)
			}

			if tc.expectedHTTP == "" {
				assert.NoError(t, tc.have.ValidateHTTP())
			} else {
				assert.EqualError(t, tc.have.ValidateHTTP(), tc.expectedHTTP)
			}

			if tc.expectedSQL == "" {
				assert.NoError(t, tc.have.ValidateSQL())
			} else {
				assert.EqualError(t, tc.have.ValidateSQL(), tc.expectedSQL)
			}

			if tc.expectedListener == "" {
				assert.NoError(t, tc.have.ValidateListener())
			} else {
				assert.EqualError(t, tc.have.ValidateListener(), tc.expectedListener)
			}
		})
	}
}

func TestAddress_SetHostname(t *testing.T) {
	address := &Address{true, false, -1, 0, &url.URL{Scheme: AddressSchemeTCP, Host: "0.0.0.0"}}

	assert.Equal(t, "tcp://0.0.0.0", address.String())

	address.SetHostname("127.0.0.1")

	assert.Equal(t, "tcp://127.0.0.1", address.String())
}

func TestAddressOutputValues(t *testing.T) {
	var (
		address  *Address
		listener net.Listener
		err      error
	)

	address = &Address{}
	assert.EqualError(t, address.validate(), "error validating the address: address url was nil")

	address = &Address{false, false, -1, 0, nil}

	assert.Equal(t, "", address.String())
	assert.Equal(t, "", address.Scheme())
	assert.Equal(t, "", address.Host())
	assert.Equal(t, "", address.Hostname())
	assert.Equal(t, "", address.NetworkAddress())
	assert.Equal(t, 0, address.Port())

	listener, err = address.Listener()

	assert.Nil(t, listener)
	assert.EqualError(t, err, "address url is nil")

	address = &Address{true, false, -1, 8080, &url.URL{Scheme: AddressSchemeTCP, Host: "0.0.0.0:8080"}}

	assert.Equal(t, "tcp://0.0.0.0:8080", address.String())
	assert.Equal(t, "tcp", address.Scheme())
	assert.Equal(t, "0.0.0.0:8080", address.Host())
	assert.Equal(t, "0.0.0.0", address.Hostname())
	assert.Equal(t, "0.0.0.0:8080", address.NetworkAddress())
	assert.Equal(t, 8080, address.Port())

	listener, err = address.Listener()

	assert.NotNil(t, listener)
	assert.NoError(t, err)

	address = &Address{true, false, -1, 0, nil}

	assert.Equal(t, "", address.String())
	assert.Equal(t, "", address.Scheme())
	assert.Equal(t, "", address.Host())
	assert.Equal(t, "", address.Hostname())
	assert.Equal(t, "", address.NetworkAddress())
	assert.Equal(t, 0, address.Port())

	listener, err = address.Listener()

	assert.Nil(t, listener)
	assert.EqualError(t, err, "address url is nil")

	address.SetHostname("abc123.com")
	address.SetPort(50)

	assert.Equal(t, "", address.String())
	assert.Equal(t, "", address.Scheme())
	assert.Equal(t, "", address.Host())
	assert.Equal(t, "", address.Hostname())
	assert.Equal(t, "", address.NetworkAddress())
	assert.Equal(t, 0, address.Port())

	listener, err = address.Listener()

	assert.Nil(t, listener)
	assert.EqualError(t, err, "address url is nil")

	address = &Address{true, false, -1, 9091, &url.URL{Scheme: AddressSchemeTCP, Host: "0.0.0.0:9091"}}

	assert.Equal(t, "tcp://0.0.0.0:9091", address.String())
	assert.Equal(t, "tcp", address.Scheme())
	assert.Equal(t, "0.0.0.0:9091", address.Host())
	assert.Equal(t, "0.0.0.0", address.Hostname())
	assert.Equal(t, "0.0.0.0:9091", address.NetworkAddress())
	assert.Equal(t, 9091, address.Port())

	listener, err = address.Listener()

	assert.NotNil(t, listener)
	assert.NoError(t, err)

	assert.NoError(t, listener.Close())

	address.SetPort(9092)

	assert.Equal(t, "tcp://0.0.0.0:9092", address.String())
	assert.Equal(t, "tcp", address.Scheme())
	assert.Equal(t, "0.0.0.0:9092", address.Host())
	assert.Equal(t, "0.0.0.0", address.Hostname())
	assert.Equal(t, "0.0.0.0:9092", address.NetworkAddress())
	assert.Equal(t, 9092, address.Port())

	listener, err = address.Listener()

	assert.NotNil(t, listener)
	assert.NoError(t, err)

	assert.NoError(t, listener.Close())

	address.SetHostname("example.com")

	assert.Equal(t, "tcp://example.com:9092", address.String())
	assert.Equal(t, "tcp", address.Scheme())
	assert.Equal(t, "example.com:9092", address.Host())
	assert.Equal(t, "example.com", address.Hostname())
	assert.Equal(t, "example.com:9092", address.NetworkAddress())
	assert.Equal(t, 9092, address.Port())
}

func TestNewAddressUnix(t *testing.T) {
	have := NewAddressUnix("/abc/123")

	require.NotNil(t, have)
	assert.Equal(t, "unix:///abc/123", have.String())
}

func TestNewAddressFromNetworkValues(t *testing.T) {
	have := NewAddressFromNetworkValues(AddressSchemeUDP, "av", 1)

	require.NotNil(t, have)
	assert.Equal(t, "udp://av:1", have.String())
}

func TestNewSMTPAddress(t *testing.T) {
	testCases := []struct {
		name                            string
		haveScheme                      string
		haveHost                        string
		havePort                        int
		expected                        string
		expectedNetwork, expectedScheme string
		expectedHostname                string
		expectedPort                    int
		expectedExplicitTLS             bool
	}{
		{
			"ShouldParseUnknownSchemePort25",
			"",
			"hosta",
			25,
			"smtp://hosta:25",
			"tcp",
			"smtp",
			"hosta",
			25,
			false,
		},
		{
			"ShouldParseUnknownSchemePort465",
			"",
			"hostb",
			465,
			"submissions://hostb:465",
			"tcp",
			"submissions",
			"hostb",
			465,
			true,
		},
		{
			"ShouldParseUnknownSchemePort587",
			"",
			"hostc",
			587,
			"submission://hostc:587",
			"tcp",
			"submission",
			"hostc",
			587,
			false,
		},
		{
			"ShouldParseUnknownPortSchemeSMTP",
			"smtp",
			"hostd",
			0,
			"smtp://hostd:25",
			"tcp",
			"smtp",
			"hostd",
			25,
			false,
		},
		{
			"ShouldParseUnknownPortSchemeSUBMISSION",
			"submission",
			"hoste",
			0,
			"submission://hoste:587",
			"tcp",
			"submission",
			"hoste",
			587,
			false,
		},
		{
			"ShouldParseUnknownPortSchemeSUBMISSIONS",
			"submissions",
			"hostf",
			0,
			"submissions://hostf:465",
			"tcp",
			"submissions",
			"hostf",
			465,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			have := NewSMTPAddress(tc.haveScheme, tc.haveHost, tc.havePort)

			assert.Equal(t, tc.expected, have.String())
			assert.Equal(t, tc.expectedScheme, have.Scheme())
			assert.Equal(t, tc.expectedNetwork, have.Network())
			assert.Equal(t, tc.expectedHostname, have.Hostname())
			assert.Equal(t, tc.expectedHostname, have.SocketHostname())
			assert.Equal(t, tc.expectedPort, have.Port())
			assert.Equal(t, tc.expectedExplicitTLS, have.IsExplicitlySecure())
		})
	}
}

func TestAddress_Dial(t *testing.T) {
	testCases := []struct {
		name    string
		have    Address
		success bool
		err     string
	}{
		{
			"ShouldNotDialNil",
			Address{true, false, -1, 0, nil},
			false,
			"address url is nil",
		},
		{
			"ShouldNotDialInvalid",
			Address{false, false, -1, 0, &url.URL{}},
			false,
			"address url is nil",
		},
		{
			"ShouldNotDialInvalidAddress",
			Address{true, false, -1, 0, &url.URL{Scheme: "abc", Host: "127.0.0.1:0"}},
			false,
			"dial tcp 127.0.0.1:0: connect: connection refused",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := tc.have.Dial()

			defer func(c net.Conn) {
				if c == nil {
					return
				}

				conn.Close()
			}(conn)

			if tc.success {

			} else {
				assert.Nil(t, conn)

				if tc.err != "" {
					assert.EqualError(t, err, tc.err)
				} else {
					assert.NotNil(t, err)
				}
			}
		})
	}
}

func TestAddress_UnixDomainSocket(t *testing.T) {
	dir := t.TempDir()

	testCases := []struct {
		name     string
		have     string
		socket   bool
		path     string
		rpath    string
		strUmask string
		umask    int
		err      string
	}{
		{
			"ShouldNotBeSocket",
			"tcp://:9091",
			false,
			"",
			"",
			"",
			-1,
			"",
		},
		{
			"ShouldParseSocket",
			fmt.Sprintf("unix://%s", filepath.Join(dir, "example.sock")),
			true,
			filepath.Join(dir, "example.sock"),
			"/",
			"",
			-1,
			"",
		},
		{
			"ShouldParseSocketWithUmask",
			fmt.Sprintf("unix://%s?umask=0022", filepath.Join(dir, "example.sock")),
			true,
			filepath.Join(dir, "example.sock"),
			"/",
			"0022",
			18,
			"",
		},
		{
			"ShouldParseSocketWithUmaskAndPath",
			fmt.Sprintf("unix://%s?umask=0022&path=abc", filepath.Join(dir, "example.sock")),
			true,
			filepath.Join(dir, "example.sock"),
			"/abc",
			"0022",
			18,
			"",
		},
		{
			"ShouldParseSocketWithBadUmask",
			fmt.Sprintf("unix://%s?umask=abc", filepath.Join(dir, "example.sock")),
			true,
			"",
			"",
			"",
			-1,
			fmt.Sprintf("error validating the unix socket address: could not parse address 'unix://%s?umask=abc': the address has a umask value of 'abc' which does not appear to be a valid octal string", filepath.Join(dir, "example.sock")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewAddress(tc.have)

			if tc.err == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.socket, actual.IsUnixDomainSocket())
				assert.Equal(t, tc.path, actual.Path())
				assert.Equal(t, tc.rpath, actual.RouterPath())
				assert.Equal(t, tc.strUmask, actual.Umask())
				assert.Equal(t, tc.umask, actual.umask)

				ln, err := actual.Listener()

				assert.NoError(t, err)
				assert.NotNil(t, ln)

				assert.NoError(t, ln.Close())
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestAddress_SocketHostname(t *testing.T) {
	testCases := []struct {
		name     string
		have     Address
		expected string
	}{
		{
			"ShouldReturnHostname",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeTCP, Host: "examplea:80"}},
			"examplea",
		},
		{
			"ShouldReturnPath",
			Address{true, true, -1, 80, &url.URL{Scheme: AddressSchemeUnix, Path: "/abc/123"}},
			"/abc/123",
		},
		{
			"ShouldReturnNothing",
			Address{false, true, -1, 80, &url.URL{Scheme: AddressSchemeUnix, Path: "/abc/123"}},
			"",
		},
		{
			"ShouldReturnNothingNil",
			Address{true, true, -1, 80, nil},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.SocketHostname())
		})
	}
}

func TestAddress_Path(t *testing.T) {
	testCases := []struct {
		name     string
		have     Address
		expected string
	}{
		{
			"ShouldReturnEmptyPath",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeTCP, Host: "tcphosta"}},
			"",
		},
		{
			"ShouldReturnPath",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeTCP, Host: "tcphosta", Path: "/apath"}},
			"/apath",
		},
		{
			"ShouldNotReturnPathInvalid",
			Address{false, false, -1, 80, &url.URL{Scheme: AddressSchemeTCP, Host: "tcphosta", Path: "/apath"}},
			"",
		},
		{
			"ShouldNotReturnPathNil",
			Address{true, false, -1, 80, nil},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.Path())
		})
	}
}

func TestAddress_RouterPath(t *testing.T) {
	testCases := []struct {
		name     string
		have     Address
		expected string
	}{
		{
			"ShouldReturnEmptyPath",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeTCP, Host: "tcphosta"}},
			"",
		},
		{
			"ShouldReturnPath",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeTCP, Host: "tcphosta", Path: "/apath"}},
			"/apath",
		},
		{
			"ShouldNotReturnPathInvalid",
			Address{false, false, -1, 80, &url.URL{Scheme: AddressSchemeTCP, Host: "tcphosta", Path: "/apath"}},
			"",
		},
		{
			"ShouldNotReturnPathNil",
			Address{true, false, -1, 80, nil},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.RouterPath())
		})
	}
}

func TestAddress_IsTCP_IsUDP(t *testing.T) {
	testCases := []struct {
		name  string
		have  Address
		isTCP bool
		isUDP bool
	}{
		{
			"ShouldReturnTrueTCP",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeTCP, Host: "tcphosta"}},
			true,
			false,
		},
		{
			"ShouldReturnTrueTCP4",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeTCP4, Host: "tcphostb"}},
			true,
			false,
		},
		{
			"ShouldReturnTrueTCP6",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeTCP6, Host: "tcphostc"}},
			true,
			false,
		},
		{
			"ShouldReturnFalseUDP",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeUDP, Host: "tcphostd"}},
			false,
			true,
		},
		{
			"ShouldReturnFalseUDP4",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeUDP4, Host: "tcphoste"}},
			false,
			true,
		},
		{
			"ShouldReturnFalseUDP6",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeUDP6, Host: "tcphostf"}},
			false,
			true,
		},
		{
			"ShouldReturnFalseSMTP",
			Address{true, false, -1, 80, &url.URL{Scheme: AddressSchemeSMTP, Host: "tcphostg"}},
			false,
			false,
		},
		{
			"ShouldReturnFalseUnix",
			Address{true, true, -1, 80, &url.URL{Scheme: AddressSchemeUnix, Host: "tcphosth"}},
			false,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.isTCP, tc.have.IsTCP())
			assert.Equal(t, tc.isUDP, tc.have.IsUDP())
		})
	}
}

func TestNewAddressFromNetworkValuesDefault(t *testing.T) {
	testCases := []struct {
		name                  string
		haveHost              string
		havePort              int
		haveSchemeDefault     string
		haveSchemeDefaultPath string
		expected              string
		expectedErr           string
	}{
		{
			"ShouldParseTCPWithTCPUnix",
			"cba",
			80,
			AddressSchemeTCP,
			AddressSchemeUnix,
			"tcp://cba:80",
			"",
		},
		{
			"ShouldParseTCPWithTCPUnixNoPort",
			"cba",
			0,
			AddressSchemeTCP,
			AddressSchemeUnix,
			"tcp://cba:0",
			"",
		},
		{
			"ShouldParseUnixWithTCPUnix",
			"/abc/123",
			80,
			AddressSchemeTCP,
			AddressSchemeUnix,
			"unix:///abc/123",
			"",
		},
		{
			"ShouldParseUnixWithScheme",
			"unix:///abc/123",
			0,
			AddressSchemeTCP,
			AddressSchemeUnix,
			"unix:///abc/123",
			"",
		},
		{
			"ShouldErrBadURL",
			"tcp://127.0.0.1:abc",
			0,
			AddressSchemeTCP,
			AddressSchemeUnix,
			"",
			"could not parse string 'tcp://127.0.0.1:abc' as address: expected format is [<scheme>://]<hostname>[:<port>]: parse \"tcp://127.0.0.1:abc\": invalid port \":abc\" after host",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, theError := NewAddressFromNetworkValuesDefault(tc.haveHost, tc.havePort, tc.haveSchemeDefault, tc.haveSchemeDefaultPath)

			if tc.expectedErr == "" {
				require.NoError(t, theError)
				assert.Equal(t, tc.expected, actual.String())
			} else {
				assert.EqualError(t, theError, tc.expectedErr)
				assert.Nil(t, actual)
			}
		})
	}
}
