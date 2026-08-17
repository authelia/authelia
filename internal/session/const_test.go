package session

import "time"

const (
	testDomain     = "example.com"
	testExpiration = time.Second * 40
	testRememberMe = time.Hour * 24
	testName       = "my_session"
	testUsername   = "john"
	testSecret     = "a-secret-value"
	testHMACKey    = "an-hmac-key"
)
