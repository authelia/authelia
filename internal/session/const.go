package session

import (
	"time"
)

const (
	testDomain     = "example.com"
	testName       = "my_session"
	testUsername   = "john"
	testExpiration = time.Second * 40
)

const (
	userSessionStorerKey = "UserSession"
	randomSessionChars   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_!#$%^*"
)
