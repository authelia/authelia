package storage

import (
	"context"
	"time"

	"github.com/authelia/authelia/v4/internal/models"
)

// Provider is an interface providing storage capabilities for persisting any kind of data related to Authelia.
type Provider interface {
	RegulatorProvider

	LoadPreferred2FAMethod(ctx context.Context, username string) (method string, err error)
	SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error)

	FindIdentityVerificationToken(ctx context.Context, token string) (found bool, err error)
	SaveIdentityVerificationToken(ctx context.Context, token string) (err error)
	RemoveIdentityVerificationToken(ctx context.Context, token string) (err error)

	SaveTOTPSecret(ctx context.Context, username string, secret string) (err error)
	LoadTOTPSecret(ctx context.Context, username string) (secret string, err error)
	DeleteTOTPSecret(ctx context.Context, username string) (err error)

	SaveU2FDeviceHandle(ctx context.Context, username string, keyHandle []byte, publicKey []byte) (err error)
	LoadU2FDeviceHandle(ctx context.Context, username string) (keyHandle []byte, publicKey []byte, err error)
}

// RegulatorProvider is an interface providing storage capabilities for persisting any kind of data related to the regulator.
type RegulatorProvider interface {
	AppendAuthenticationLog(ctx context.Context, attempt models.AuthenticationAttempt) (err error)
	LoadAuthenticationLogs(ctx context.Context, username string, fromDate time.Time, limit, page int) (attempts []models.AuthenticationAttempt, err error)
}
