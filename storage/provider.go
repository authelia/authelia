package storage

import (
	"time"

	"github.com/clems4ever/authelia/models"
)

// Provider is an interface providing storage capabilities for
// persisting any kind of data related to Authelia.
type Provider interface {
	LoadPrefered2FAMethod(username string) (string, error)
	SavePrefered2FAMethod(username string, method string) error

	FindIdentityVerificationToken(token string) (bool, error)
	SaveIdentityVerificationToken(token string) error
	RemoveIdentityVerificationToken(token string) error

	SaveTOTPSecret(username string, secret string) error
	LoadTOTPSecret(username string) (string, error)

	SaveU2FDeviceHandle(username string, device []byte) error
	LoadU2FDeviceHandle(username string) ([]byte, error)

	AppendAuthenticationLog(attempt models.AuthenticationAttempt) error
	LoadLatestAuthenticationLogs(username string, fromDate time.Time) ([]models.AuthenticationAttempt, error)
}
