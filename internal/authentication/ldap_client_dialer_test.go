package authentication

import (
	"net"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLDAPClientDialerStandard(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:3389")

	require.NoError(t, err)

	testCases := []struct {
		name    string
		have    string
		timeout time.Duration
		err     string
	}{
		{
			"ShouldFailToDialTimeout",
			"ldap://127.0.0.1:389",
			1 * time.Microsecond,
			"error occurred attempting to dial LDAP server at 'ldap://127.0.0.1:389': LDAP Result Code 200 \"Network Error\": dial tcp 127.0.0.1:389: i/o timeout",
		},
		{
			"ShouldFailToDialConnectionRefused",
			"ldap://127.0.0.1:389",
			20 * time.Millisecond,
			"error occurred attempting to dial LDAP server at 'ldap://127.0.0.1:389': LDAP Result Code 200 \"Network Error\": dial tcp 127.0.0.1:389: connect: connection refused",
		},
		{
			"ShouldSuccessfullyDial",
			"ldap://127.0.0.1:3389",
			20 * time.Millisecond,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewLDAPClientDialerStandard()

			conn, err := client.DialURL(tc.have, ldap.DialWithTLSConfig(nil), ldap.DialWithDialer(&net.Dialer{Timeout: tc.timeout}))
			if tc.err == "" {
				assert.NoError(t, err)
				require.NotNil(t, conn)
				assert.NoError(t, conn.Close())
			} else {
				assert.Nil(t, conn)
				assert.EqualError(t, err, tc.err)
			}
		})
	}

	require.NoError(t, ln.Close())
}
