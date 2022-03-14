package session

import (
	"time"
)

const (
	testDomain     = "example.com"
	testExpiration = time.Second * 40
	testName       = "my_session"
	testUsername   = "john"
)

const (
	userSessionStorerKey = "UserSession"
	randomSessionChars   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_!#$%^*"
)
