package storage

import (
	"context"
	"time"

	"github.com/authelia/authelia/v4/internal/models"
)

// Provider is an interface providing storage capabilities for persisting any kind of data related to Authelia.
type Provider interface {
	models.StartupCheck

	RegulatorProvider

	SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error)
	LoadPreferred2FAMethod(ctx context.Context, username string) (method string, err error)

	SaveIdentityVerification(ctx context.Context, verification models.IdentityVerification) (err error)
	RemoveIdentityVerification(ctx context.Context, jti string) (err error)
	FindIdentityVerification(ctx context.Context, jti string) (found bool, err error)

	SaveTOTPConfiguration(ctx context.Context, config models.TOTPConfiguration) (err error)
	DeleteTOTPConfiguration(ctx context.Context, username string) (err error)
	LoadTOTPConfiguration(ctx context.Context, username string) (config *models.TOTPConfiguration, err error)

	SaveU2FDevice(ctx context.Context, device models.U2FDevice) (err error)
	LoadU2FDevice(ctx context.Context, username string) (device *models.U2FDevice, err error)
}

// RegulatorProvider is an interface providing storage capabilities for persisting any kind of data related to the regulator.
type RegulatorProvider interface {
	AppendAuthenticationLog(ctx context.Context, attempt models.AuthenticationAttempt) (err error)
	LoadAuthenticationLogs(ctx context.Context, username string, fromDate time.Time, limit, page int) (attempts []models.AuthenticationAttempt, err error)
}
