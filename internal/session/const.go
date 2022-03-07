package session

import (
	"time"

	"github.com/authelia/authelia/v4/internal/oidc"
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

var (
	amrFactorSomethingKnown = []string{oidc.AMRPasswordBasedAuthentication}
	amrFactorSomethingHave  = []string{oidc.AMROneTimePassword, oidc.AMRHardwareSecuredKey, oidc.AMRShortMessageService}
	amrChannelBrowser       = []string{oidc.AMRPasswordBasedAuthentication, oidc.AMRHardwareSecuredKey, oidc.AMROneTimePassword}
	amrChannelService       = []string{oidc.AMRShortMessageService}
)
