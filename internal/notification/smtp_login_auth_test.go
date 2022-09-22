package notification

import (
	"fmt"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullLoginAuth(t *testing.T) {
	username := "john"
	password := "strongpw123"
	serverInfo := &smtp.ServerInfo{
		Name: "mail.authelia.com",
		TLS:  true,
		Auth: nil,
	}
	auth := newLoginAuth(username, password, "mail.authelia.com")

	proto, _, err := auth.Start(serverInfo)
	assert.Equal(t, smtpAUTHMechanismLogin, proto)
	require.NoError(t, err)

	toServer, err := auth.Next([]byte("Username:"), true)
	assert.Equal(t, []byte(username), toServer)
	require.NoError(t, err)

	toServer, err = auth.Next([]byte("Password:"), true)
	assert.Equal(t, []byte(password), toServer)
	require.NoError(t, err)

	toServer, err = auth.Next([]byte(nil), false)
	assert.Equal(t, []byte(nil), toServer)
	require.NoError(t, err)

	toServer, err = auth.Next([]byte("test"), true)
	assert.Equal(t, []byte(nil), toServer)
	assert.EqualError(t, err, fmt.Sprintf("unexpected server challenge: %s", []byte("test")))
}

func TestShouldHaveUnexpectedHostname(t *testing.T) {
	serverInfo := &smtp.ServerInfo{
		Name: "localhost",
		TLS:  true,
		Auth: nil,
	}
	auth := newLoginAuth("john", "strongpw123", "mail.authelia.com")
	_, _, err := auth.Start(serverInfo)
	assert.EqualError(t, err, "unexpected hostname from server")
}

func TestTLSNotNeededForLocalhost(t *testing.T) {
	serverInfo := &smtp.ServerInfo{
		Name: "localhost",
		TLS:  false,
		Auth: nil,
	}
	auth := newLoginAuth("john", "strongpw123", "localhost")

	proto, _, err := auth.Start(serverInfo)
	assert.Equal(t, "LOGIN", proto)
	require.NoError(t, err)
}

func TestTLSNeededForNonLocalhost(t *testing.T) {
	serverInfo := &smtp.ServerInfo{
		Name: "mail.authelia.com",
		TLS:  false,
		Auth: nil,
	}
	auth := newLoginAuth("john", "strongpw123", "mail.authelia.com")
	_, _, err := auth.Start(serverInfo)
	assert.EqualError(t, err, "connection over plain-text")
}
