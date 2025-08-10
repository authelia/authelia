package authentication

import (
	"net"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
)

func TestNewLDAPClientDialerStandard(t *testing.T) {
	client := NewLDAPClientDialerStandard()

	conn, err := client.DialURL("ldap://localhost:389", ldap.DialWithTLSConfig(nil), ldap.DialWithDialer(&net.Dialer{Timeout: 1 * time.Microsecond}))

	assert.Nil(t, conn)
	assert.EqualError(t, err, "failed to dial LDAP server at ldap://localhost:389: LDAP Result Code 200 \"Network Error\": dial tcp: lookup localhost: i/o timeout")
}
