package notification

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"
)

type loginAuth struct {
	username string
	password string
	host     string
}

func newLoginAuth(username, password, host string) smtp.Auth {
	return &loginAuth{username, password, host}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && !(server.Name == "localhost" || server.Name == "127.0.0.1" || server.Name == "::1") {
		return "", nil, errors.New("connection over plain-text")
	}

	if server.Name != a.host {
		return "", nil, errors.New("unexpected hostname from server")
	}

	return smtpAUTHMechanismLogin, []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	switch {
	case bytes.Equal(fromServer, []byte("Username:")):
		return []byte(a.username), nil
	case bytes.Equal(fromServer, []byte("Password:")):
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected server challenge: %s", fromServer)
	}
}
